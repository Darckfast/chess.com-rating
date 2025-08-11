import { connect } from "cloudflare:sockets";
import app from "./app.wasm";
import "./wasm_exec.js";

globalThis.tryCatch = (fn) => {
    try {
        return {
            result: fn(),
        };
    } catch (e) {
        return {
            error: e,
        };
    }
};

async function run(ctx) {
    return new Promise(async (resolve) => {
        const go = new Go();

        let imports = go.importObject
        imports.workers = {
            ready: () => {
                resolve(true);
            },
        }
        let instance = new WebAssembly.Instance(app, imports);
        go.run(instance, ctx);
    })
}

async function fetch(req, env, ctx) {
    const binding = {};
    await run({ env, ctx, binding, connect });
    return binding.handleRequest(req);
}

// async function scheduled(event, env, ctx) {
//   const binding = {};
//   await run(createRuntimeContext({ env, ctx, binding }));
//   return binding.runScheduler(event);
// }

// async function queue(batch, env, ctx) {
//   const binding = {};
//   await run(createRuntimeContext({ env, ctx, binding }));
//   return binding.handleQueueMessageBatch(batch);
// }

// onRequest handles request to Cloudflare Pages
// async function onRequest(ctx) {
//     const binding = {};
//     const { request, env } = ctx;
//     await run({ env, ctx, binding, connect });
//     return binding.handleRequest(request);
// }

export default {
    fetch,
    // scheduled,
    // queue,
    // onRequest,
};
