// a minimal chrome devtools protocol client: launch, send, listen, evaluate
import { spawn } from 'node:child_process';
import { readFile, rm } from 'node:fs/promises';
import { setTimeout as sleep } from 'node:timers/promises';

const chrome = `${process.env.HOME}/.cache/ms-playwright/`
  + 'chromium-1228/chrome-linux64/chrome';

// ubuntu's apparmor denies chrome the user namespaces its sandbox needs;
// this browser loads only the page under test and djvu.js
export async function launch(profile) {
  await rm(profile, { recursive: true, force: true });
  const proc = spawn(chrome, ['--headless=new', '--remote-debugging-port=0',
    `--user-data-dir=${profile}`, '--no-first-run',
    '--no-default-browser-check', '--hide-scrollbars', '--no-sandbox',
    'about:blank'], { stdio: 'ignore' });
  // a check that throws never reaches close; chrome must not outlive it
  process.on('exit', () => proc.kill());
  let port;
  for (let i = 0; i < 100 && !port; i++) {
    await sleep(100);
    port = await readFile(`${profile}/DevToolsActivePort`, 'utf8')
      .then(s => s.split('\n')[0], () => null);
  }
  const targets = await fetch(`http://127.0.0.1:${port}/json/list`);
  const [page] = (await targets.json()).filter(t => t.type === 'page');
  const ws = new WebSocket(page.webSocketDebuggerUrl);
  await new Promise(ok => ws.onopen = ok);
  let id = 0;
  const pending = new Map(), listeners = [];
  ws.onmessage = ({ data }) => {
    const msg = JSON.parse(data);
    if (!msg.id) return listeners.forEach(f => f(msg));
    const { ok, fail } = pending.get(msg.id);
    pending.delete(msg.id);
    msg.error ? fail(new Error(JSON.stringify(msg.error))) : ok(msg.result);
  };
  const send = (method, params = {}) => new Promise((ok, fail) => {
    pending.set(++id, { ok, fail });
    ws.send(JSON.stringify({ id, method, params }));
  });
  const on = f => listeners.push(f);
  const close = async () => {
    ws.close();
    proc.kill();
    await sleep(300);
    await rm(profile, { recursive: true, force: true });
  };
  const evaluate = async (expression, gesture = false) => {
    const r = await send('Runtime.evaluate', { expression,
      awaitPromise: true, returnByValue: true, userGesture: gesture });
    const e = r.exceptionDetails;
    if (e) throw new Error(e.exception?.description ?? e.text);
    return r.result.value;
  };
  return { send, on, close, evaluate };
}
