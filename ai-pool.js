// Spreads the computer's search over one Web Worker per CPU core. lib/ai/search_js.go drives it:
// each worker grows its own tree (cmd/aiworker) and the page merges their statistics.
(() => {
  if (!window.Worker || !WebAssembly.compileStreaming) {
    return;
  }
  const size = navigator.hardwareConcurrency || 4;
  const requests = new Map(); // search id -> {stats, remaining, resolve, reject}
  let current = null; // request of the latest search, for progress
  let nextId = 0;

  // A worker that cannot run would leave searches waiting forever: give up on the pool,
  // and the game falls back to searching on the page itself.
  const fail = (error) => {
    if (window.aiPool !== pool) {
      return;
    }
    delete window.aiPool;
    workers.forEach((worker) => worker.terminate());
    requests.forEach(({ reject }) => reject(error));
    requests.clear();
  };

  const workers = Array.from({ length: size }, (_, index) => {
    const worker = new Worker("ai-worker.js");
    worker.onmessage = ({ data }) => {
      if (data.error) {
        fail(new Error(data.error));
        return;
      }
      const request = requests.get(data.id);
      if (!request) {
        return;
      }
      request.stats[index] = data.stats;
      if (data.done && --request.remaining === 0) {
        requests.delete(data.id);
        request.resolve(request.stats);
      }
    };
    worker.onerror = (event) => fail(new Error(event.message));
    return worker;
  });

  // Each worker is told to work on its share of the threads, the others stop.
  const dispatch = (threads, message) => {
    const used = Math.min(Math.max(threads || size, 1), size);
    workers.forEach((worker, i) => worker.postMessage(i < used ? message : { type: "stop" }));
    return used;
  };

  const pool = {
    size,

    // Resolves with the JSON root statistics of every worker used.
    search(state, ms, simulations, threads) {
      const id = nextId++;
      const used = Math.min(Math.max(threads || size, 1), size);
      const request = { stats: Array(used).fill("[]"), remaining: used };
      const promise = new Promise((resolve, reject) => Object.assign(request, { resolve, reject }));
      requests.set(id, request);
      current = request;
      const perWorker = simulations > 0 ? Math.ceil(simulations / used) : 0;
      dispatch(used, { type: "search", id, state, ms, simulations: perWorker });
      return promise;
    },

    progress: () => (current ? current.stats.slice() : []),

    ponder(state, ms, threads) {
      dispatch(threads, { type: "ponder", state, ms });
    },

    stop() {
      workers.forEach((worker) => worker.postMessage({ type: "stop" }));
    },
  };

  WebAssembly.compileStreaming(fetch("aiworker.wasm"))
    .then((module) => workers.forEach((worker) => worker.postMessage({ type: "init", module, size })))
    .catch(fail);
  window.aiPool = pool;
})();
