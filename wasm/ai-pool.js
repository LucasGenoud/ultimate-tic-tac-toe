// Spreads the computer's search over one Web Worker per CPU core.
// lib/ai/search_js.go calls aiPool.search and merges the statistics returned by every worker.
(() => {
  if (!window.Worker) {
    return;
  }
  const size = navigator.hardwareConcurrency || 4;
  const pending = new Map();
  let nextId = 0;

  // A worker that cannot start would leave searches waiting forever: give up on the pool,
  // and the game falls back to searching on the page itself.
  const fail = (error) => {
    delete window.aiPool;
    for (const { reject } of pending.values()) {
      reject(error);
    }
    pending.clear();
  };

  const workers = Array.from({ length: size }, () => {
    const worker = new Worker("ai-worker.js");
    worker.onmessage = ({ data }) => {
      const request = pending.get(data.id);
      pending.delete(data.id);
      if (data.error) {
        request?.reject(new Error(data.error));
      } else {
        request?.resolve(data.stats);
      }
    };
    worker.onerror = (event) => fail(new Error(event.message));
    return worker;
  });

  const ask = (worker, state, ms) =>
    new Promise((resolve, reject) => {
      const id = nextId++;
      pending.set(id, { resolve, reject });
      worker.postMessage({ id, state, ms });
    });

  window.aiPool = {
    size,
    search: (state, ms) => Promise.all(workers.map((worker) => ask(worker, state, ms))),
  };
})();
