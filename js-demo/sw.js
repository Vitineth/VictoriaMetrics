importScripts('/wasm_exec.js');

let interceptConfig;

function updateVmStatus(status, client) {
    client.postMessage(JSON.stringify({
        type: 'status',
        status,
    }));
}

function wrapFetchResponseWithProgress(response, progressHandler) {
    console.log(response);

    let loaded = 0;
    const contentLength = response.headers.get('Content-Length');
    const total = parseInt(contentLength, 10);

    const res = new Response(new ReadableStream({
        async start(controller) {
            const reader = response.body.getReader();
            for (; ;) {
                let {done, value} = await reader.read();

                if (done) {
                    progressHandler(total, total)
                    break
                }

                loaded += value.byteLength;
                progressHandler(loaded, total)
                controller.enqueue(value);
            }
            controller.close();
        },
    }, {
        "status": response.status,
        "statusText": response.statusText
    }));

// Make sure to copy the headers!
// Wasm is very picky with it's headers and it will fail to compile if they are not
// specified correctly.
    for (let pair of response.headers.entries()) {
        res.headers.set(pair[0], pair[1]);
    }

    return res;
}

async function registerWasmHTTPListener(wasm, {base, cacheName, args = []} = {}, client) {
    try {
        updateVmStatus('Launching', client);
        await self.clients.claim();

        let path = new URL(registration.scope).pathname
        if (base && base !== '') path = `${trimEnd(path, '/')}/${trimStart(base, '/')}`

        console.log('Using path', path);

        const handlerPromise = new Promise(setHandler => {
            self.wasmhttp = {
                path,
                setHandler,
            }
        })

        const go = new Go();
        go.argv = [wasm, ...args]

        console.log('Instantiated Go instance with argv', wasm, args);
        // const source = cacheName
        //     ? caches.open(cacheName).then((cache) => cache.match(wasm)).then((response) => response ?? fetch(wasm))
        //     : caches.match(wasm).then(response => (response) ?? fetch(wasm))
        const source = wrapFetchResponseWithProgress(await fetch(wasm), function (loaded, total) {
            console.log(`WASM Loading progress: ${loaded}/${total} = ${Math.round((loaded / total) * 100) / 100}`)
        });
        console.log('Launched fetch');
        WebAssembly.instantiateStreaming(source, go.importObject).then(({instance}) => {
            try {
                updateVmStatus('Running', client);
                go.run(instance);
            } catch (e) {
                updateVmStatus('Errored: ' + (e.message ?? 'Unknown Error'), client);
            }
        });

        handlerPromise.then((e) => {
            console.log('Intercept has been set');
            interceptConfig = {
                path,
                handler: e,
            }
        })

        console.log('WASM Instantiated 3');
    } catch (e) {
        updateVmStatus('Errored: ' + (e.message ?? 'Unknown Error'), client);
    }
}

self.addEventListener('fetch', e => {
    console.log(e, interceptConfig);
    const {pathname} = new URL(e.request.url)
    console.log('Received request for', e.request.url, 'with pathname', pathname, 'matched = ', pathname.startsWith(interceptConfig.path));
    if (!pathname.startsWith(interceptConfig.path)) return

    e.respondWith(interceptConfig.handler(e.request));// handlerPromise.then(handler => handler(e.request)))
})

self.addEventListener('activate', function (event) {
    console.log('Claiming control');
    return self.clients.claim();
});

function trimStart(s, c) {
    let r = s
    while (r.startsWith(c)) r = r.slice(c.length)
    return r
}

function trimEnd(s, c) {
    let r = s
    while (r.endsWith(c)) r = r.slice(0, -c.length)
    return r
}

console.log('Service Worker Registered, preparing to launch');

addEventListener('message', async (m) => {
    // Exit early if we don't have access to the client.
    // Eg, if it's cross-origin.
    if (!m.source.id) {
        console.error('No client ID on request')
        return;
    }

    // Get the client.
    const client = await self.clients.get(m.source.id);
    // Exit early if we don't get the client.
    // Eg, if it closed.
    if (!client) {
        console.error('No client');
        return;
    }

    if (m.data === 'terminate') {
        updateVmStatus('Exited', client);
        terminateVictoriaMetricsInstance()
    }
    if (m.data === 'launch') {
        await registerWasmHTTPListener('/victoria-metrics-js-wasm-prod.wasm', {base: '/victoria-metrics'}, client);
    }
});