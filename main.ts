import app from "./bin/app.wasm";
import "./bin/wasm_exec.js";

globalThis.cf = {}
globalThis.tryCatch = (fn) => {
    try {
        return {
            data: fn(),
        };
    } catch (error) {
        if (!(error instanceof Error)) {
            if (error instanceof Object) {
                error = JSON.stringify(error)
            }

            error = new Error(error || 'no error message')
        }
        return {
            error,
        };
    }
}

function init() {
    const go = new Go()
    go.run(new WebAssembly.Instance(app, go.importObject))
}

async function fetch(req: Request, env: Env, ctx: ExecutionContext) {
    init()
    return await globalThis.cf.fetch(req, env, ctx);
}

export default {
    fetch,
};
