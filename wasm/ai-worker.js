// Runs searches for ai-pool.js on its own thread, using cmd/aiworker compiled to WebAssembly.
importScripts("wasm_exec.js");

const search = new Promise((resolve, reject) => {
  self.onSearchReady = resolve;
  const go = new Go();
  WebAssembly.instantiateStreaming(fetch("aiworker.wasm"), go.importObject)
    .then(({ instance }) => go.run(instance))
    .catch(reject);
});

self.onmessage = async ({ data: { id, state, ms } }) => {
  try {
    const stats = (await search)(state, ms);
    if (stats instanceof Error) {
      throw stats;
    }
    postMessage({ id, stats });
  } catch (error) {
    postMessage({ id, error: String(error) });
  }
};
