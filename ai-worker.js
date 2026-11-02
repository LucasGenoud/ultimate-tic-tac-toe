// Runs on its own thread for ai-pool.js. The tree grows in short slices so that new requests
// (search, ponder, stop) are handled between them.
importScripts("wasm_exec.js");

const SLICE_MS = 40;
let step = null; // set by cmd/aiworker once it runs
let job = null; // {id, state, deadline, simulations, untilFull, played, stats}; id is undefined when pondering
const queued = []; // requests received before step is ready

const report = (done) => {
  if (job.id !== undefined) {
    postMessage({ id: job.id, stats: job.stats, done });
  }
};

const finish = () => {
  if (job) {
    report(true);
    job = null;
  }
};

const tick = new MessageChannel();
let scheduled = false;
const schedule = () => {
  if (!scheduled) {
    scheduled = true;
    tick.port2.postMessage(null);
  }
};

tick.port1.onmessage = () => {
  scheduled = false;
  if (!job) {
    return;
  }
  const remaining = job.simulations > 0 ? job.simulations - job.played : 0;
  const ms = Math.max(1, Math.min(SLICE_MS, job.deadline - Date.now()));
  const result = step(job.state, ms, remaining, job.untilFull);
  if (result instanceof Error) {
    postMessage({ error: String(result) });
    job = null;
    return;
  }
  job.played += result.played;
  job.stats = result.stats;
  const done = result.played === 0 || Date.now() >= job.deadline ||
    (job.simulations > 0 && job.played >= job.simulations);
  if (done) {
    finish();
  } else {
    report(false);
    schedule();
  }
};

const handle = (data) => {
  finish();
  if (data.type === "search" || data.type === "ponder") {
    const searching = data.type === "search";
    job = {
      id: searching ? data.id : undefined,
      state: data.state,
      deadline: Date.now() + data.ms,
      simulations: searching ? data.simulations : 0,
      untilFull: !searching,
      played: 0,
      stats: "[]",
    };
    schedule();
  }
};

self.onmessage = async ({ data }) => {
  if (data.type !== "init") {
    if (step) {
      handle(data);
    } else {
      queued.push(data);
    }
    return;
  }
  try {
    self.poolSize = data.size;
    const ready = new Promise((resolve) => { self.onStepReady = resolve; });
    const go = new Go();
    go.run(await WebAssembly.instantiate(data.module, go.importObject));
    step = await ready;
    queued.splice(0).forEach(handle);
  } catch (error) {
    postMessage({ error: String(error) });
  }
};
