declare module '*.wasm' {
    interface WasmExports {
        handleRequest(r: Request): any;
    }
    const wasmModule: (imports?: WebAssembly.Imports) => Promise<{ instance: WebAssembly.Instance & { exports: WasmExports } }>;
    export default wasmModule;
}
