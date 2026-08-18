// checks shelf.html against the real library in headless chrome, both ways
// it is read: opened as a file, with the folder chosen in the picker, and
// served over http by a stand-in for quark; then the steps it was built in,
// on the library and on ~/scratch/work. the oracles are the disk itself:
// folder listings, each path.md's stages and spans, fs.stat of every path.
// usage: node run.mjs [library] [page]. the library defaults to ~/Data, the
// page to the shelf.html beside this folder. screenshots land in
// $TMPDIR/shelf.
import { launch } from './cdp.mjs';
import { quark, types } from './quark.mjs';
import { execFileSync } from 'node:child_process';
import { existsSync, readFileSync, readdirSync, statSync } from 'node:fs';
import { once } from 'node:events';
import { mkdir, readFile, readdir, stat, writeFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { posix } from 'node:path';
import { setTimeout as sleep } from 'node:timers/promises';

const home = process.env.HOME;
const library = process.argv[2] ?? `${home}/Data`;
const page =
  process.argv[3] ?? new URL('../shelf.html', import.meta.url).pathname;
const shots = `${tmpdir()}/shelf`;
const collate = new Intl.Collator(undefined, { numeric: true }).compare;
const ls = async dir => (await readdir(`${library}/${dir}`)).sort(collate);
const present = path =>
  stat(`${library}/${path}`).then(
    s => s.isDirectory() || s.size > 0,
    () => false,
  );

// the root as the disk has it, the page aside, and the subjects in it:
// every folder holding a path.md
const top = (await ls('')).filter(n => n !== posix.basename(page));
const subjects = [];
for (const dir of top) {
  const md = await readFile(`${library}/${dir}/path.md`, 'utf8').catch(
    () => null,
  );
  if (md !== null) subjects.push({ dir, md });
}
const heading = ({ dir, md }) =>
  (md.match(/^# (.*)/m)?.[1] ?? dir).replace(/\b'\b/g, '’');
const stages = ({ md }) => md.match(/^## /gm)?.length ?? 0;
// cs's first stage that holds an entry with no title, a bare path: it
// ticks under its path from the root
const bare = subjects
  .find(s => s.dir === 'cs')
  .md.split(/^## /m)
  .slice(1)
  .findIndex(stage => /^(?:\d+\.|-) `[^`*]+`\s*$/m.test(stage));
// an entry is a list item under a stage that names a path or has a title
const entries = ({ md }) =>
  md.slice(md.search(/^## /m)).match(/^(?:\d+\.|-) (?:.*`|\*\*).*$/gm)
    ?.length ?? 0;

// the installed quark sends a pdf as application/x-pdf, which chrome
// downloads; this assumes the one-line fix in its config.h. each server
// is kept with its root, to look a path up on disk
const served = [];
const serve = async root => {
  const s = quark(root, page, { ...types, pdf: 'application/pdf' });
  await once(s.listen(0, '127.0.0.1'), 'listening');
  served.push([s, root]);
  return s;
};
const server = await serve(library);
const c = await launch(`${tmpdir()}/shelfprofile`);
const problems = [];
let failed = 0;
const check = (ok, what, detail = '') => {
  if (!ok) failed++;
  const why = ok || !detail ? '' : ` — ${detail}`;
  console.log(`${ok ? 'ok  ' : 'FAIL'} ${what}${why}`);
};
c.on(({ method, params }) => {
  const e = params.exceptionDetails;
  if (method === 'Runtime.exceptionThrown')
    problems.push(e.exception?.description ?? e.text);
  if (method === 'Runtime.consoleAPICalled' && params.type === 'error')
    problems.push(params.args.map(a => a.value ?? a.description).join(' '));
  // over http, a path the disk lacks is a 404 that chrome logs. it is a
  // problem only if the disk has the path
  if (method === 'Log.entryAdded' && params.entry.level === 'error') {
    const { text, url } = params.entry;
    if (!/\b404\b/.test(text)) return problems.push(url ?? text);
    const { port, pathname } = new URL(url);
    const [, root] = served.find(([s]) => s.address().port === +port);
    const path = root + decodeURIComponent(pathname);
    if (existsSync(path)) problems.push(`404 for ${path}, which is on disk`);
  }
  if (method === 'Page.downloadWillBegin')
    problems.push(`downloaded ${params.suggestedFilename}, not shown`);
});
await c.send('Runtime.enable');
await c.send('Page.enable');
await c.send('Log.enable');
await c.send('Browser.setDownloadBehavior', {
  behavior: 'deny',
  eventsEnabled: true,
});
await mkdir(shots, { recursive: true });

const size = (width, height) =>
  c.send('Emulation.setDeviceMetricsOverride', {
    width,
    height,
    deviceScaleFactor: 1,
    mobile: false,
  });
const theme = value =>
  c.send('Emulation.setEmulatedMedia', {
    features: [{ name: 'prefers-color-scheme', value }],
  });
// chrome's own pdf viewer. off, the page is chrome on android: pdf.js
// draws a pdf
const builtin = on =>
  c.evaluate(
    on
      ? `delete navigator.pdfViewerEnabled; 1`
      : `Object.defineProperty(navigator, 'pdfViewerEnabled',
    { value: false, configurable: true }); 1`,
  );
// the page, or another reached by its session
const shot = async (name, session) =>
  writeFile(
    `${shots}/${name}.png`,
    Buffer.from(
      (await c.send('Page.captureScreenshot', {}, session)).data,
      'base64',
    ),
  );
// until an expression is true in the page, or a test is, or the time is up
const until = async (test, ms = 15000, where = c.evaluate) => {
  const ok = typeof test === 'string' ? () => where(test) : test;
  for (const end = Date.now() + ms; Date.now() < end; await sleep(50))
    if (await ok().catch(() => false)) return true;
  return false;
};
// the sandboxed frame open, which chrome runs in a process of its own:
// devtools reaches it by attaching to it
const framed = async expression => {
  const { targetInfos } = await c.send('Target.getTargets');
  const { targetId } = targetInfos.find(
    t => t.type === 'iframe' && t.url === 'about:srcdoc',
  );
  const { sessionId } = await c.send('Target.attachToTarget', {
    targetId,
    flatten: true,
  });
  try {
    return await c.evaluate(expression, false, sessionId);
  } finally {
    await c.send('Target.detachFromTarget', { sessionId });
  }
};
const json = JSON.stringify;
// a key as the page's table spells it, a name or a character, shift+ in
// front, down and up; then a moment for the page to answer
const codes = {
  Tab: 9,
  Enter: 13,
  Escape: 27,
  ArrowLeft: 37,
  ArrowUp: 38,
  ArrowRight: 39,
  ArrowDown: 40,
};
const press = async (name, session) => {
  const [, shift, key] = name.match(/^(Shift\+)?(.+)/);
  const text = key.length === 1 ? key : undefined;
  const event = {
    key,
    windowsVirtualKeyCode: codes[key] ?? key.toUpperCase().charCodeAt(0),
    modifiers: shift ? 8 : 0,
  };
  const down = { type: text ? 'keyDown' : 'rawKeyDown', text, ...event };
  await c.send('Input.dispatchKeyEvent', down, session);
  await c.send('Input.dispatchKeyEvent', { type: 'keyUp', ...event }, session);
  await sleep(60);
};
// a tap on what an expression finds, or at a point: a finger's, else a
// mouse's click. in a page in a frame too. what is not shown has none, so
// the checks after it fail, as they should
const poke = async (at, finger) => {
  const point = Array.isArray(at)
    ? at
    : await c.evaluate(`(e => {
      if (!e?.checkVisibility()) return null;
      e.scrollIntoViewIfNeeded();
      const r = e.getBoundingClientRect();
      let x = r.x + r.width / 2, y = r.y + r.height / 2;
      for (let f = e.ownerDocument.defaultView.frameElement; f;
        f = f.ownerDocument.defaultView.frameElement) {
        const o = f.getBoundingClientRect();
        x += o.x + f.clientLeft;
        y += o.y + f.clientTop;
      }
      return [x, y]; })(${at})`);
  if (!point) return;
  const [x, y] = point;
  if (finger)
    for (const [type, touchPoints] of [
      ['touchStart', [{ x, y }]],
      ['touchEnd', []],
    ])
      await c.send('Input.dispatchTouchEvent', { type, touchPoints });
  else
    for (const type of ['mouseMoved', 'mousePressed', 'mouseReleased'])
      await c.send('Input.dispatchMouseEvent', {
        type,
        x,
        y,
        button: 'left',
        clickCount: 1,
      });
};
// a url loaded, else the page again. a hash moved within the page loads
// nothing
let loaded;
c.on(({ method }) => method === 'Page.loadEventFired' && loaded?.());
const navigate = async url => {
  const load = new Promise(ok => (loaded = ok));
  if (!url) await c.send('Page.reload');
  else if (!(await c.send('Page.navigate', { url })).loaderId) return;
  await load;
};
// a page in a window of its own beside this one, as wide: an oracle
const beside = async (url, width) => {
  const { targetId } = await c.send('Target.createTarget', {
    url,
    newWindow: true,
  });
  const { sessionId: id } = await c.send('Target.attachToTarget', {
    targetId,
    flatten: true,
  });
  await c.send(
    'Emulation.setDeviceMetricsOverride',
    { width, height: 900, deviceScaleFactor: 1, mobile: false },
    id,
  );
  await until(`document.readyState === 'complete'`, 15000, e =>
    c.evaluate(e, false, id),
  );
  await sleep(300);
  return { id, close: () => c.send('Target.closeTarget', { targetId }) };
};
const stored = async name =>
  (await c.evaluate(`localStorage.getItem(${json(name)})`)) ?? '';
// a row of the pane, by the path of what it names
const node = path => `[...document.querySelectorAll('main li')]
  .find(li => li.node?.path === ${json(path)})`;
const focused = `document.querySelector('.focus')?.parentElement`;
const viewing = `document.body.classList.contains('viewing')`;

// the pane drawn, and the aside showing the row under the cursor
const rendered = () =>
  until(`!!document.querySelector('main ul')
  && document.querySelector('aside').childElementCount > 0`);
// a folder is the place shown once the header names it last, nothing open
const place = async hash => {
  await c.evaluate(`location.hash = ${json(hash)}; 1`);
  const name = posix.basename(hash.slice(1));
  return until(`(h => !${viewing} && (${json(name)}
    ? h.querySelector('b:last-child')?.textContent === ${json(name)}
    : !h.querySelector('a')))(document.querySelector('main header h1'))`);
};
// a stage unfolds in place as on the tablet: a tap, and a second if the
// first folded it. it is shown once its links are looked up
const stage = async i => {
  await c.evaluate(`(li => { li.firstChild.click();
    if (!li.querySelector(':scope > ul')) li.firstChild.click(); })(
    document.querySelectorAll('main .stage')[${i}]); 1`);
  return until(`(li => !!li.querySelector(':scope > ul')
    && [...li.querySelectorAll('ul a')].every(a => a.title))(
    document.querySelectorAll('main .stage')[${i}])`);
};
// the view is this hash's once the bar names it, and it is no longer opening
const open = async (hash, ready) => {
  const name = posix.basename(hash.slice(1));
  await c.evaluate(`location.hash = ${json(hash)}; 1`);
  return until(`(() => {
    const view = document.querySelector('aside > :last-child');
    const named = document.querySelector('aside nav b')?.textContent;
    return decodeURIComponent(location.hash) === ${json(hash)} && view
      && named === ${json(name)} && view.textContent !== 'Opening…'
      && (${ready}); })()`);
};

// the picker cannot be driven, but a dropped folder yields the same kind of
// handle. headless cannot answer a permission prompt, so it is refused, and
// the click always reaches the picker
const drop = async dirs => {
  await c.evaluate(`window.dropped = null;
    addEventListener('dragover', e => e.preventDefault());
    addEventListener('drop', async e => {
      e.preventDefault();
      window.dropped = await Promise.all([...e.dataTransfer.items]
        .map(item => item.getAsFileSystemHandle()));
    }); 1`);
  const data = { items: [], files: dirs, dragOperationsMask: 1 };
  for (const type of ['dragEnter', 'dragOver', 'drop'])
    await c.send('Input.dispatchDragEvent', { type, x: 9, y: 9, data });
  await until('!!window.dropped');
};
const choose = async () => {
  await until(`!!document.querySelector('main ul')
    || !document.querySelector('main button').hidden`);
  if (await c.evaluate(`!document.querySelector('main ul')`)) {
    await drop([library]);
    await c.evaluate(
      `showDirectoryPicker = async () => dropped[0];
      FileSystemHandle.prototype.requestPermission = async () => 'denied';
      document.querySelector('main button').click(); 1`,
      true,
    );
  }
  return rendered();
};

// what path.html kept, under pathreader: and with ~/Data/math as its root
const watched =
  'calc/mit-highlights-of-calculus/' + '01-big-picture-of-calculus.mp4';
const old = {
  'pathreader:open': json([
    'stage:2 — After Algebra and Trig Are Solid',
    'calc/mit-highlights-of-calculus',
  ]),
  'pathreader:done': json([
    'SL Loney Plane Trigonometry',
    'elem/gateways-to-mathematics-herbert-gross',
  ]),
  [`pathreader:at:${watched}`]: '125',
};

const routes = {
  async file(reload) {
    if (!reload) {
      await c.send('Page.navigate', { url: `file://${page}` });
      await until(`!document.querySelector('main button').hidden`);
      await drop([`${library}/math`]);
      await c.evaluate(`new Promise(ok => {
        const r = indexedDB.open('pathreader');
        r.onupgradeneeded = () => r.result.createObjectStore('folders');
        r.onsuccess = () => {
          const t = r.result.transaction('folders', 'readwrite');
          t.objectStore('folders').put(dropped[0], 'root');
          t.oncomplete = ok;
        };
      })`);
      await c.evaluate(`Object.entries(${json(old)})
        .forEach(([k, v]) => localStorage.setItem(k, v)); 1`);
    }
    await c.send('Page.reload');
    return choose();
  },
  async http(reload) {
    if (reload) await c.send('Page.reload');
    else
      await c.send('Page.navigate', {
        url: `http://127.0.0.1:${server.address().port}/shelf.html`,
      });
    return rendered();
  },
};

await size(1440, 900);
for (const [how, start] of Object.entries(routes)) {
  console.log(`-- ${how}`);
  check(await start(), 'renders the root');
  await sleep(500);
  await shot(`${how}-root`);

  const rows = await c.evaluate(`[...document.querySelectorAll('main li')]
    .filter(li => li.node).map(li => li.node.path)`);
  check(
    json(rows) === json(top),
    "the root is the library's listing, this page aside",
    json(rows),
  );
  const counts = await c.evaluate(`Object.fromEntries(
    [...document.querySelectorAll('main li')]
      .filter(li => li.firstChild.dataset.count)
      .map(li => [li.node.path, li.firstChild.dataset.count]))`);
  check(
    json(Object.keys(counts)) === json(subjects.map(s => s.dir)) &&
      subjects.every(s => counts[s.dir].endsWith(`/${entries(s)}`)),
    'a folder holding a path.md shows its ticks beside it, of every entry',
    json(counts),
  );
  const [first] = subjects;
  await c.evaluate(`${node(first.dir)}.firstChild.click(); 1`);
  check(
    await until(`(v => v.localName === 'section'
      && v.path === ${json(first.dir)}
      && ${node(first.dir)}.node.list?.parentNode === ${node(first.dir)})(
      document.querySelector('aside > :last-child'))`),
    'a tap on a subject unfolds it in place, and the right shows its summary',
  );
  await c.evaluate(`${node(first.dir)}.firstChild.click(); 1`);

  if (how === 'file') {
    const moved = {
      at: await stored(`shelf:at:math/${watched}`),
      done: await stored('shelf:done'),
      left: await c.evaluate(`Object.keys(localStorage)
        .filter(k => k.startsWith('shelf:at:calc/')).length`),
      old: await c.evaluate(`indexedDB.databases().then(dbs => [
        ...dbs.map(db => db.name), ...Object.keys(localStorage)]
        .filter(k => k.startsWith('pathreader')))`),
    };
    check(
      !moved.old.length,
      'what path.html kept under pathreader moves under shelf',
      json(moved.old),
    );
    check(
      moved.at === '125' && moved.left === 0,
      'choosing ~/Data over ~/Data/math moves lecture places under math/',
      json(moved),
    );
    check(
      moved.done.includes('SL Loney Plane Trigonometry') &&
        moved.done.includes(
          'math/elem/gateways-to-mathematics-herbert-' + 'gross',
        ),
      'and the ticks, a title as it was',
      moved.done,
    );
    check(
      counts.math?.startsWith('2/'),
      'and the page counts them',
      counts.math,
    );
  }

  for (const s of subjects) {
    check(await place(`#${s.dir}`), `goes into ${s.dir}`);
    const shown = await c.evaluate(`[...document.querySelectorAll('main li')]
      .map(li => li.node?.path ?? li.className)`);
    const files = (await ls(s.dir)).filter(n => n !== 'path.md');
    check(
      json(shown) ===
        json([
          ...Array(stages(s)).fill('stage'),
          ...(files.length ? ['group'] : []),
          ...files.map(n => `${s.dir}/${n}`),
        ]),
      `${s.dir} lists its stages, then a rule, then its files`,
      json(shown),
    );
    if (how === 'file') await shot(s.dir);
  }

  // every stage of every subject gone into, and each link it makes looked
  // up
  const links = {};
  for (const s of subjects) {
    await place(`#${s.dir}`);
    links[s.dir] = [];
    for (let i = 0; i < stages(s); i++) {
      await stage(i);
      links[s.dir].push(
        ...(await c.evaluate(`[...document.querySelectorAll(
          'main .stage')[${i}].querySelectorAll(':scope > ul a')]
          .map(a => ({ text: a.textContent, absent: 'absent' in a.dataset,
            to: decodeURIComponent(a.hash.slice(1)) }))`)),
      );
      if (how === 'file' && s.dir === 'cs' && i === stages(s) - 1)
        await shot('map');
      await press('h');
      await until(`!!document.querySelector('main .stage')`);
    }
    const spans = s.md.slice(s.md.search(/^## /m)).match(/`[^`]+`/g) ?? [];
    check(
      links[s.dir].length === spans.length,
      `every path in ${s.dir}/path.md is a link`,
      `${links[s.dir].length} links, ${spans.length} paths`,
    );
  }
  const all = Object.values(links).flat(),
    wrong = [];
  for (const { text, to, absent } of all)
    if (absent === (await present(to))) wrong.push(`${text} -> ${to}`);
  check(
    !wrong.length,
    'a link is struck through when its path is missing ' + 'or empty',
    wrong.join(', '),
  );
  check(
    links.cs?.some(l => l.text === '../math/dm' && l.to === 'math/dm'),
    'a path from cs/ reaches ../math/dm',
  );
  const refined = Object.entries(links).flatMap(([dir, list]) =>
    list.filter(l => l.to !== posix.join(dir, l.text).replace(/\/$/, '')),
  );
  check(
    refined.length === 3 &&
      refined.every(l => l.text === 'bio/srinivasa-ramanujan/'),
    'three titles in bio/srinivasa-ramanujan/ open their one file; the ' +
      'ambiguous fourth stays the folder',
    refined.map(l => l.to).join(', '),
  );
  check(
    refined.some(l =>
      l.to.endsWith(
        'the-man-who-knew-infinity-a-life-' +
          'of-the-genius-ramanujan-robert-kanigel.pdf',
      ),
    ),
    'kanigel resolves to his book',
  );

  for (const [name, hash, ready] of [
    [
      'pdf',
      '#math/elem/plane-trigonometry-sydney-loney.pdf',
      `view.localName === 'iframe' && !!view.src`,
    ],
    [
      'djvu',
      '#math/anal/elliptic-functions-and-elliptic-integrals-' +
        'prasolov-and-solovyev.djvu',
      `view.localName === 'iframe' && view.sandbox.value === 'allow-scripts'`,
    ],
    [
      'epub',
      '#math/elem/how-to-solve-it-george-polya.epub',
      `view.shadowRoot?.querySelectorAll('section').length > 5
        && !view.shadowRoot.querySelector('parsererror')`,
    ],
    [
      'video',
      '#math/elem/project-mathematics-tom-apostol/01-similarity.mkv',
      `view.localName === 'video' && view.videoWidth > 0`,
    ],
    [
      'audio',
      '#math/bio/john-milnor-abel-prize.mp3',
      `view.localName === 'audio' && view.duration > 0`,
    ],
    [
      'image',
      '#math/bio/srinivasa-ramanujan/timeline-illustration.png',
      `view.localName === 'img' && view.naturalWidth > 0`,
    ],
    [
      'text',
      '#math/notes/exp.txt',
      `view.localName === 'pre' && view.textContent.length > 100`,
    ],
    [
      'empty',
      '#math/calc/infinite-powers-steven-strogatz',
      `view.textContent === 'This file is empty'`,
    ],
    [
      'missing',
      '#math/anal/a-radical-approach-to-real-analysis-david-' + 'bressoud.pdf',
      `view.textContent === 'Nothing is at this path'`,
    ],
    [
      'nowhere',
      '#no/such/folder',
      `view.textContent === 'Nothing is at this path'`,
    ],
  ]) {
    check(await open(hash, ready), `opens ${name}: ${hash.slice(1)}`);
    await sleep({ djvu: 6000, pdf: 3000 }[name] ?? 800);
    if (how === 'file' || name === 'pdf')
      await shot(how === 'file' ? name : `${how}-${name}`);
  }

  // chrome on android has no pdf viewer, so pdf.js draws the pdf. the
  // oracle is pdfinfo: a canvas a page, and the first one inked
  const lament = 'math/elem/a-mathematicians-lament-paul-lockhart.pdf';
  const pages = +execFileSync('pdfinfo', [`${library}/${lament}`], {
    encoding: 'utf8',
  }).match(/^Pages: +(\d+)/m)[1];
  await builtin(false);
  check(
    await open(
      `#${lament}`,
      `view.localName === 'iframe'
    && view.sandbox.value === 'allow-scripts'`,
    ),
    `opens pdf.js: ${lament}`,
  );
  check(
    await until(
      `(() => {
    const all = document.querySelectorAll('canvas'), [one] = all;
    if (all.length !== ${pages} || !one.width) return false;
    const { data } = one.getContext('2d')
      .getImageData(0, 0, one.width, one.height);
    return data.some((v, i) => i % 4 !== 3 && v < 128); })()`,
      15000,
      framed,
    ),
    `pdf.js draws ${pages} pages, the first inked`,
  );
  await shot(`${how}-pdfjs`);
  await builtin(true);

  const course = 'math/calc/the-essence-of-calculus-grant-sanderson';
  const files = await ls(course);
  check(
    (await place(`#${course}`)) &&
      (await c.evaluate(`[...document.querySelectorAll('main li')]
        .map(li => li.node.path).join()`)) ===
        files.map(n => `${course}/${n}`).join(),
    'a folder lists what is on disk',
  );
  check(
    await until(`document.querySelector('aside > :last-child').path
      === ${focused}.node.path`),
    'nothing open, the aside previews the row under the cursor',
  );
  await shot(how === 'file' ? 'folder' : `${how}-folder`);
  check(
    await open(`#${course}/${files[0]}`, `view.localName === 'video'`),
    `opens a lecture: ${files[0]}`,
  );
  await c.evaluate(`document.querySelector('[rel=next]').click(); 1`);
  check(
    await until(`decodeURIComponent(location.hash)
      === ${json(`#${course}/${files[1]}`)}`),
    'next opens the next file in the same folder',
    files[1],
  );

  const lecture = `${course}/05-derivatives-of-exponentials.mp4`;
  await c.evaluate(`localStorage.setItem('shelf:at:${lecture}', '100');
    1`);
  check(
    await open(`#${lecture}`, `view.currentTime >= 100`),
    'a lecture resumes where it was left',
  );

  const begun = Date.now();
  check(
    await open(
      '#math/calc/calculus-an-intuitive-and-physical-approach-' +
        'morris-kline/book.epub',
      `view.shadowRoot?.querySelector('img[src]')`,
    ),
    `a 39 MB epub of 4800 images opens, in ${Date.now() - begun} ms`,
  );
  const [inflated, images] = await c.evaluate(`(r => [
    r.querySelectorAll('img[src]').length, r.querySelectorAll('img').length
  ])(document.querySelector('aside article').shadowRoot)`);
  check(
    inflated < images / 20,
    'and inflates only the images of the chapters in view',
    `${inflated} of ${images}`,
  );
  check(
    await c.evaluate(`(() => {
    const book = document.querySelector('aside article');
    const before = book.scrollTop;
    book.shadowRoot.querySelectorAll('a[data-to]')[12].click();
    return book.scrollTop > before && location.hash.includes('book.epub');
  })()`),
    'a link inside a book scrolls the book, not the page',
  );

  // ticked by hand: loney's entry under its title, petzold's under its path
  await place('#math');
  await stage(0);
  const loney = `document.querySelector(
    '[href="#math/elem/plane-trigonometry-sydney-loney.pdf"]')
    .closest('.entry').querySelector('input')`;
  // cleared, then ticked: the file route planted loney's tick
  await c.evaluate(`(t => { if (t.checked) t.click(); t.click(); })(${loney});
    1`);
  check(
    (await stored('shelf:done')).includes('SL Loney Plane Trigonometry'),
    'ticking an entry keeps it under its title',
  );
  await place('#cs');
  await stage(bare);
  const petzold = `document.querySelector('input[name^="cs/"]')`;
  const name = await c.evaluate(`(t => { t.click(); return t.name; })(
    ${petzold})`);
  check(
    name.startsWith('cs/') && (await stored('shelf:done')).includes(name),
    'and an entry with no title under its path from the root',
    name,
  );
  await start(true);
  check(
    await until(`${focused}?.matches('.stage') && [...document
      .querySelectorAll('main .stage')].indexOf(${focused}) === ${bare}`),
    'a reload lands on the row the cursor was left on',
  );
  await press('l');
  check(await until(`${petzold}?.checked`), 'ticks survive a reload');
}

// the http route left the keys in cs's stage with a bare path
console.log('-- keys');
await place('');
for (let i = 0; i < 60; i++) {
  if (await c.evaluate(`${focused} === ${node('cs')}`)) break;
  await press('j');
}
check(
  await until(`${focused} === ${node('cs')}
  && !document.querySelector('main li > ul')`),
  'j walks to cs, and unfolds nothing it passes',
);
await press('l');
check(
  await until(`(li => li.node.list?.parentNode === li
  && !!li.querySelector('.stage') && ${focused} === li)(${node('cs')})`),
  'l unfolds a folder in place: a subject, its stages first',
);
const staged = `[...${node('cs')}.querySelectorAll(':scope > ul > .stage')]`;
await press('l');
check(
  await until(`${staged}.indexOf(${focused}) === ${bare}`),
  'l steps in, onto the row it was left on: that stage',
);
await press('h');
await press('j');
await press('l');
await press('l');
check(
  await until(
    `${staged}[${bare + 1}].contains(${focused})
    && ${focused}?.matches('.entry')`,
  ),
  'l unfolds the next stage, and steps in onto an entry',
);
await press('x');
check(
  await until(`${focused}.querySelector('input').checked
  && localStorage.getItem('shelf:done')
    .includes(${focused}.querySelector('input').name)`),
  'x ticks the entry under the cursor',
);
const before = await c.evaluate(`${focused}.node.path`);
await press('j');
check(
  await until(`${focused}.previousElementSibling?.node?.path
  === ${json(before)}`),
  'j moves down one row',
);
await press('h');
check(
  await until(`${focused} === ${staged}[${bare + 1}]
  && !${staged}[${bare + 1}].querySelector('ul')`),
  'h on an entry goes up onto its stage, and folds it',
);
await press('h');
check(
  await until(`${focused} === ${node('cs')}
  && !${node('cs')}.querySelector('ul') && location.hash === ''`),
  'h on a folded stage goes up onto its folder, and folds that',
);
await press('l');
await press('l');
check(
  await until(`${staged}.indexOf(${focused}) === ${bare + 1}`),
  'the cursor is kept per folder: back in cs, it is where it was left',
);

await press('/');
await c.send('Input.insertText', { text: 'essence calculus' });
await press('Enter');
const course = 'math/calc/the-essence-of-calculus-grant-sanderson';
check(
  await until(`decodeURIComponent(location.hash) === ${json(`#${course}`)}
  && document.querySelector('main header h1 b')?.textContent
    === ${json(posix.basename(course))}`),
  'find, then enter, goes where the hit names: into its folder',
);
const [one, two] = (await ls(course)).map(n => `${course}/${n}`);
for (let i = 0; i < 40; i++) {
  if (await c.evaluate(`${focused} === ${node(one)}`)) break;
  await press('k');
}
await press('l');
check(
  await until(`decodeURIComponent(location.hash) === ${json(`#${one}`)}
  && !!document.querySelector('aside video')`),
  'l on a file opens it',
);
await press('n');
check(
  await until(`decodeURIComponent(location.hash) === ${json(`#${two}`)}
  && ${focused} === ${node(two)}`),
  'n opens the next file in the folder, the cursor with it',
);
await press('h');
check(
  await until(`!${viewing} && ${focused} === ${node(two)}
  && decodeURIComponent(location.hash) === ${json(`#${course}`)}`),
  'h closes it, the cursor on it',
);

// a book under the cursor, opened: the preview is the viewer, kept
await place('#math');
await stage(0);
const polya = `document.querySelector(
  '[href="#math/elem/how-to-solve-it-george-polya.epub"]').closest('li')`;
for (let i = 0; i < 60 && !(await c.evaluate(`${focused} === ${polya}`)); i++)
  await press('j');
await until(`!!document.querySelector('aside article')
  ?.shadowRoot.querySelector('section')`);
await c.evaluate(`document.querySelector('aside article').kept = true; 1`);
await press('l');
check(
  await until(`${viewing} && !!document.querySelector('aside nav')
  && document.querySelector('aside article').kept`),
  'a file opened under the cursor is its preview, become the viewer',
);
await press('Escape');
check(
  await until(`!${viewing} && !document.querySelector('aside nav')
  && document.querySelector('aside article').kept`),
  'escape closes it, back to the preview',
);

await theme('dark');
check(
  await open(
    '#math/elem/how-to-solve-it-george-polya.epub',
    `view.shadowRoot?.querySelector('section')`,
  ),
  'dark: an epub',
);
await c.evaluate(`document.querySelector('aside article').scrollTop = 2400;
  1`);
await sleep(500);
await shot('dark-epub');
const inverted = `getComputedStyle(document.querySelector('aside .pdf'))
  .filter.includes('invert')`;
check(
  await open('#math/elem/plane-trigonometry-sydney-loney.pdf', inverted),
  'dark: a pdf is inverted',
);
await sleep(3000);
await shot('dark-pdf');
await press('i');
check(await until(`!${inverted}`), 'i turns the pdf back to paper');
await builtin(false);
check(
  await open('#math/elem/a-mathematicians-lament-paul-lockhart.pdf', inverted),
  'dark: a pdf drawn by pdf.js is inverted',
);
await sleep(3000);
await shot('dark-pdfjs');
await builtin(true);
await place('#math');
check(
  await c.evaluate(`!document.querySelector('aside nav')`),
  'closed, the bar goes with what was open',
);
await sleep(500);
await shot('dark-pane');
await theme('light');
await size(390, 844);
await shot('narrow-pane');
const hidden = q => `!document.querySelector('${q}').checkVisibility()`;
check(
  (await place(`#${course}`)) && (await c.evaluate(hidden('aside'))),
  'narrow: a folder is the pane alone',
);
check(
  await c.evaluate(`(r => r.top >= 0 && r.bottom <= innerHeight)(
  document.querySelector('.focus').getBoundingClientRect())`),
  'narrow: the row under the cursor is in view',
);
await shot('narrow-folder');
check(
  (await open(
    '#math/elem/plane-trigonometry-sydney-loney.pdf',
    `view.localName === 'iframe'`,
  )) && (await c.evaluate(hidden('main'))),
  'narrow: a file is the viewer alone',
);
check(
  await c.evaluate(`document.querySelector('nav').scrollWidth
  <= innerWidth`),
  'narrow: the viewer bar fits',
);
await sleep(3000);
await shot('narrow-file');

// the viewer gets most of a tablet's screen, either way up
const box = q =>
  c.evaluate(`(b => [b.x, b.y, b.width, b.height])(
  document.querySelector('${q}').getBoundingClientRect())`);
await size(1333, 800);
await sleep(300);
const [pane, viewer] = [await box('main'), await box('aside')];
check(
  viewer[2] > 1.5 * pane[2],
  'landscape: the viewer is beside the pane, and wider',
  json({ pane, viewer }),
);
await size(800, 1333);
await sleep(300);
const [under, over] = [await box('main'), await box('aside')];
check(
  over[1] === 0 && under[1] >= over[3] && over[3] > under[3],
  'portrait: the viewer is on top, and taller than the pane below',
  json({ under, over }),
);
await shot('portrait');

check(
  !problems.length,
  'no errors, exceptions or downloads; each 404 is a path not on disk',
  problems.join('\n     '),
);
// -- steps --
// the steps of ~/scratch/brief, 1 to 11, as each was checked when built:
// each runs in a page of its own at a width, with a mouse or with touch,
// with no page errors
//   tree  the tree unfolds in place, by keys and by taps (step 1); a tap
//         on a file opens it, at every width: a laptop's, a phone's, a
//         window 700px wide (step 13)
//   tick  a subject's entries keep a tick column; any other file shows a ✓
//         at its edge once ticked, and a path's tick is one wherever it is
//         listed (step 2)
//   frame the view fills the aside under the bar, the bar there or not: a
//         page's frame previewed is as tall as the aside, and so is the
//         tablet's, upright; a card longer than the aside scrolls in it
//         (step 3)
//   outline  a text with headings opens as an outline, folded at first
//         to its headings: md by #, a leaf by its headings, # or
//         underlined or bare lines; a script, #! first, stays text (step 4)
//         python by its classes and defs, nested as python nests them, each
//         headed as its code says it, its id its qualified name, the code
//         before the first def text at the top; ast is the oracle (step 4b)
//   node  md's headings are page.html's nodes: the id the path of its
//         headings, a duplicate's -2; the keys reach each, and pass a
//         component's own details by; the heading the summary's first
//         child, the box's, a figure after it at the edge; a heading Book
//         leaves the epub's #book alone (todo/studio/refactor, step 1)
//   shape every kind's node is one shape, its id, heading, edge and body:
//         a diff's file is headed by its path, its status and +n −m at the
//         edge, a gap apart, its id its path's; python's nodes an id each,
//         none twice (todo/studio/refactor, step 2)
//   sign  python's heading as its code says it, @property def parts(), in
//         the cases model.py has not: an async def, a def in a def, a
//         signature and a decorator over lines, a def named twice, tabs;
//         hovered, the signature as written and the docstring; ast is the
//         oracle (todo/studio/refactor, step 3)
//   keys  tab hands the view the keys, shift+tab, esc and h at the first
//         heading hand them back; an outline takes j k l h o z e c /
//         as page.html's engine does, a page in a frame its own keys; on
//         touch a tap goes to the pane it lands in (step 5)
//   kept  what is unfolded, folders and stages, unfolds again after a
//         reload, and the cursor is where it was, down through it; h at
//         the top leaves what it folded folded (step 6)
//   help  ? as page.html's: one table binds the keys and lists them, a dl
//         of kbd and what each does, its box as page.html's; esc or ?
//         closes it; in a page's frame ? is the page's; on touch, ? keys
//         atop the tree (step 7)
//   answers  the user's answers (step 8): no hover tint left by a tap; a
//         reload keeps the place; the find box as page.html's; a ✓ in an
//         open file's bar; a subject's title atop the tree; cs/path.md one
//         path an entry; on a phone an outline's blocks scroll across; an
//         outline's bar; python's preamble one node, its marks headings;
//         a folder's name medium, a file's regular, both in the ink
//   kind  a file's extension as a tag right after its name, in its hue or
//         muted, none for a leaf or a folder; the name whole in the dom,
//         a long one cut before the tag; the edge keeps t/n and a
//         lecture's clock (step 9)
//   source  js and css open as an outline, a page as the page, u flipping
//         it to its source and back; each section mark heads what follows,
//         css by its rules under it, js by its functions; with no marks,
//         functions alone. the oracle is the file's own lines (step 14)
//   folder  a folder under the cursor shows its card on the right: its
//         count and the file to resume, over its tree as the tree draws
//         it, two levels at first, a folder in a folder, of over 30 rows
//         or named with a dot folded, its count beside it. its keys are
//         an outline's, and a file in it opens in the tree; a subject
//         keeps its summary; a folder's hash roots the tree there (step 11)
// #work is the example the steps walk, ~/scratch/work, served from
// ~/scratch as the library is; the medium is served for page.html's own ?
const scratch = await serve(`${home}/scratch`);
const medium = await serve(new URL('..', import.meta.url).pathname);
const work = `http://127.0.0.1:${scratch.address().port}/shelf.html#work`;
const shelf = `http://127.0.0.1:${server.address().port}/shelf.html`;
// a frame handed the keys, and the keys box, ask for focus: the page has
// it, as a page in front does, though a window opens beside it
await c.send('Emulation.setFocusEmulationEnabled', { enabled: true });

// a page of its own: nothing kept from the last, the screen this wide,
// with touch or a mouse. a file's or a folder's row is found by its path;
// an entry naming the same path is not
async function session(width, touch, tall = width < 600 ? 844 : 900) {
  await navigate('about:blank');
  for (const [s] of served)
    await c.send('Storage.clearDataForOrigin', {
      origin: `http://127.0.0.1:${s.address().port}`,
      storageTypes: 'all',
    });
  await c.send('Emulation.setDeviceMetricsOverride', {
    width,
    height: tall,
    deviceScaleFactor: width < 600 ? 2 : 1,
    mobile: touch,
  });
  await c.send('Emulation.setTouchEmulationEnabled', { enabled: touch });
  // its errors are the problems since it began
  problems.length = 0;
  const js = (f, ...a) => c.evaluate(`(${f})(...${json(a)})`);
  const li = `p => [...document.querySelectorAll('main li.file')]
    .find(li => li.node.path === p)`;
  const cursor = () =>
    js(`() => document.querySelector('.focus')?.parentElement.node?.path`);
  const unfolded = p =>
    js(`p => (li => !!li && li.node.list?.parentNode === li)((${li})(p))`, p);
  const opened = () =>
    js(`() => document.querySelectorAll('main li > ul').length`);
  // j until the cursor is on p, once p is a row, as a tap may still be
  // unfolding its folder; it never unfolds what it passes
  const walk = async p => {
    await until(() => js(`p => !!(${li})(p)`, p));
    for (let i = 0; i < 60 && (await cursor()) !== p; i++) await press('j');
    return (await cursor()) === p;
  };
  // a tap on what an expression finds, with touch, else a click
  const hit = async at => {
    await poke(at, touch);
    await sleep(120);
  };
  // a tap on a row, or on a part of it
  const tap = (p, part) =>
    hit(
      `(${li})(${json(p)})?.firstChild` +
        (part ? `?.querySelector(${json(part)})` : ''),
    );
  const shown = () =>
    js(`() => document.querySelector('aside > :last-child')?.path`);
  const viewing = () =>
    js(`() => document.body.classList.contains('viewing')`);
  const height = p =>
    js(`p => (${li})(p).firstChild.getBoundingClientRect().height`, p);
  const hash = () => js(`() => decodeURIComponent(location.hash)`);
  // a url, else the page again: done once the header names the hash's
  // place last, the root's no folder above it
  const load = async url => {
    const at = url ?? (await c.evaluate('location.href'));
    await navigate(url);
    const name = decodeURIComponent(at.split('#')[1] ?? '')
      .split('/')
      .at(-1);
    await until(
      async () =>
        (await js(`() => (h => h && (${json(name)}
          ? h.querySelector('b')?.textContent === ${json(name)}
          : !h.querySelector('a')))(document.querySelector('main header h1'))
          && document.querySelectorAll('main .row').length > 0`)) === true,
    );
  };
  const find = async text => {
    await press('/');
    await c.send('Input.insertText', { text });
    await sleep(150);
    await press('Enter');
  };
  // a row's tick as it is drawn: shown or not, its glyph, how far from the
  // row's end, and where the row's name starts
  const mark = p =>
    js(
      `p => {
      const row = (${li})(p).firstChild, r = row.getBoundingClientRect();
      const tick = row.querySelector(':scope > .tick');
      const text = row.querySelector(':scope > .text');
      return {
        shown: getComputedStyle(tick).display !== 'none',
        glyph: getComputedStyle(tick, '::before').content,
        checked: tick.querySelector('input').checked,
        end: Math.round(r.right - tick.getBoundingClientRect().right),
        start: Math.round(text.getBoundingClientRect().left - r.left),
      }; }`,
      p,
    );
  // the aside and its view as drawn: heights, whether the view starts
  // right under the bar, whether it scrolls, and whether the page does
  const fits = () =>
    js(`() => {
      const aside = document.querySelector('aside').getBoundingClientRect();
      const nav = document.querySelector('aside > nav');
      const bar = nav ? Math.round(nav.getBoundingClientRect().height) : 0;
      const view = document.querySelector('aside > :last-child');
      const r = view.getBoundingClientRect();
      return {
        path: view.path,
        kind: view.localName,
        aside: Math.round(aside.height),
        bar,
        view: Math.round(r.height),
        under: Math.round(r.top - aside.top) === bar,
        scrolls: view.scrollHeight > view.clientHeight,
        page: document.scrollingElement.scrollHeight <= innerHeight,
      }; }`);
  // opened by its hash: the viewer shows it, done opening
  const open = async url => {
    await navigate(url);
    const name = decodeURIComponent(url.split('#')[1]).split('/').at(-1);
    await until(
      async () =>
        (await viewing()) &&
        (await js(`() => document.querySelector('aside nav b')
          ?.textContent`)) === name &&
        (await js(`() => document.querySelector('aside > :last-child')
          .textContent`)) !== 'Opening…',
    );
  };
  // the view's outline: the headings in sight, each open or not, its bullet
  // ringed or not, how many bodies show, and a heading's height
  const outlined = () =>
    js(`() => {
      const view = document.querySelector('aside > :last-child');
      if (!view.matches('.outline')) return { kind: view.localName };
      const seen = [...view.querySelectorAll('summary')]
        .filter(s => s.checkVisibility());
      return {
        kind: 'outline',
        heads: seen.map(s => s.textContent),
        open: seen.map(s => s.parentElement.open),
        ringed: seen.map(s =>
          getComputedStyle(s, '::before').boxShadow !== 'none'),
        bodies: [...view.querySelectorAll('pre')]
          .filter(p => p.checkVisibility()).length,
        height: Math.round(seen[0].getBoundingClientRect().height),
      }; }`);
  // a tap on a heading of the view; no such heading has none, so the
  // checks after it fail, as they should
  const head = text =>
    hit(`[...document.querySelectorAll('aside summary')]
      .find(s => s.textContent === ${json(text)})`);
  const ticked = name =>
    js(
      `name => JSON.parse(localStorage.getItem('shelf:done') ?? '[]')
        .includes(name)`,
      name,
    );
  return {
    touch,
    width,
    errors: problems,
    js,
    cursor,
    unfolded,
    opened,
    press,
    walk,
    hit,
    tap,
    shown,
    viewing,
    height,
    hash,
    load,
    find,
    mark,
    fits,
    open,
    outlined,
    head,
    ticked,
    li,
  };
}

const flow = 'work/ssv/poc/flow';
const html = 'work/pipeline.html',
  leaf = 'work/ssv/issues/summary.txt';
// a page's frame, previewed where the aside shows a preview, then opened
async function frames(s, tag, previews) {
  let f;
  if (previews) {
    await s.walk(html);
    await until(async () => (await s.fits()).path === html);
    f = await s.fits();
    check(
      f.kind === 'iframe' && f.view === f.aside,
      `${tag} frame: previewed, a page's frame is as tall as the aside`,
      JSON.stringify(f),
    );
    await shot(`${tag}-frame-preview`);
    await s.tap('work/ssv');
    await s.tap('work/ssv/issues');
    await s.walk(leaf);
    await until(async () => (await s.fits()).path === leaf);
    f = await s.fits();
    check(
      f.view === f.aside && f.scrolls && f.page,
      `${tag} frame: previewed, a long text scrolls in the aside, ` +
        'not the page',
      JSON.stringify(f),
    );
  }
  await s.tap(html);
  await until(
    async () => (await s.viewing()) && (await s.fits()).path === html,
  );
  f = await s.fits();
  check(
    f.kind === 'iframe' && f.under && f.view === f.aside - f.bar,
    `${tag} frame: opened, it fills the aside under the bar`,
    JSON.stringify(f),
  );
  await shot(`${tag}-frame-open`);
  check(
    !s.errors.length,
    `${tag} frame: no page errors`,
    s.errors.join(' | '),
  );
}

for (const [width, touch] of [
  [1400, false],
  [390, true],
]) {
  const tag = `${width}${touch ? ' touch' : ''}`;
  const wide = width >= 800;
  console.log(`-- tree, ${tag}`);

  // by keys
  let s = await session(width, touch);
  await s.load(work);
  check(
    (await s.js(`() => matchMedia('(pointer: coarse)').matches`)) === touch,
    `${tag}: pointer is ${touch ? 'coarse' : 'fine'}`,
  );
  check(await s.walk('work/ssv'), `${tag} keys: j walks to ssv`);
  check((await s.opened()) === 0, `${tag} keys: j k unfold nothing passed`);
  await s.press('l');
  check(
    await until(
      async () =>
        (await s.unfolded('work/ssv')) && (await s.cursor()) === 'work/ssv',
    ),
    `${tag} keys: l unfolds ssv in place, the cursor stays on it`,
  );
  if (wide)
    check(
      await until(async () => (await s.shown()) === 'work/ssv'),
      `${tag} keys: the right shows the folder under the cursor, its card`,
    );
  await s.press('l');
  check(
    (await s.cursor())?.startsWith('work/ssv/'),
    `${tag} keys: l on an unfolded folder steps in`,
    await s.cursor(),
  );
  check(await s.walk('work/ssv/poc'), `${tag} keys: j walks to poc`);
  await s.press('l');
  check(
    await until(() => s.unfolded('work/ssv/poc')),
    `${tag} keys: l unfolds poc inside ssv`,
  );
  await s.press('l');
  check(await s.walk(flow), `${tag} keys: j walks to flow`);
  if (wide)
    check(
      await until(async () => (await s.shown()) === flow),
      `${tag} keys: the right previews the file under the cursor`,
    );
  check(
    (await s.height(flow)) === (touch ? 40 : 30),
    `${tag}: a row is ${touch ? 40 : 30}px`,
    String(await s.height(flow)),
  );
  await shot(`${width}-keys-unfolded`);
  await s.press('h');
  check(
    (await s.cursor()) === 'work/ssv/poc' &&
      !(await s.unfolded('work/ssv/poc')),
    `${tag} keys: h on a file goes up onto poc and folds it`,
  );
  await s.press('h');
  check(
    (await s.cursor()) === 'work/ssv' && !(await s.unfolded('work/ssv')),
    `${tag} keys: h on a folded folder goes up onto ssv and folds it`,
  );
  await s.press('Enter');
  check(
    await until(() => s.unfolded('work/ssv')),
    `${tag} keys: enter unfolds`,
  );
  await s.press('Enter');
  check(!(await s.unfolded('work/ssv')), `${tag} keys: enter folds`);
  // the cursor is kept per folder: back in, l l lands where it was left
  await s.press('l');
  await until(() => s.unfolded('work/ssv'));
  await s.press('l');
  check(
    (await s.cursor()) === 'work/ssv/poc',
    `${tag} keys: stepping back in lands on the row it was left on`,
    await s.cursor(),
  );
  await s.press('l');
  await until(() => s.unfolded('work/ssv/poc'));
  await s.press('l');
  check(
    (await s.cursor()) === flow,
    `${tag} keys: and so on down`,
    await s.cursor(),
  );
  await s.press('l');
  check(
    await until(
      async () =>
        (await s.viewing()) &&
        (await s.hash()) === `#${flow}` &&
        (await s.js(`() => document.querySelector('aside nav b')
          ?.textContent`)) === 'flow',
    ),
    `${tag} keys: l on a file opens it`,
  );
  await shot(`${width}-keys-open`);
  await s.press('h');
  check(
    await until(
      async () =>
        !(await s.viewing()) &&
        (await s.hash()) === '#work' &&
        (await s.cursor()) === flow &&
        (await s.unfolded('work/ssv/poc')),
    ),
    `${tag} keys: h closes it, the tree as it was, the cursor on it`,
  );
  // find reveals a hit in the tree, the place kept: poc and ssv folded,
  // a folder hit unfolds them, and a file hit opens
  await s.press('h');
  await s.press('h');
  await s.find('poc');
  check(
    await until(
      async () =>
        (await s.cursor()) === 'work/ssv/poc' &&
        (await s.unfolded('work/ssv')) &&
        (await s.unfolded('work/ssv/poc')) &&
        (await s.hash()) === '#work',
    ),
    `${tag} find: a folder hit is unfolded where it is, the place kept`,
    `${await s.cursor()} ${await s.hash()}`,
  );
  await s.press('h');
  await s.press('h');
  await s.find('model.py');
  check(
    await until(
      async () =>
        (await s.viewing()) &&
        (await s.hash()) === '#work/ssv/poc/model.py' &&
        (await s.cursor()) === 'work/ssv/poc/model.py',
    ),
    `${tag} find: a file hit under folded folders unfolds them, and opens`,
    `${await s.cursor()} ${await s.hash()}`,
  );
  await s.press('h');
  check(
    await until(
      async () =>
        !(await s.viewing()) &&
        (await s.hash()) === '#work' &&
        (await s.cursor()) === 'work/ssv/poc/model.py',
    ),
    `${tag} find: closed, the place is still work`,
  );
  for (let i = 0; i < 2; i++) await s.press('h');
  // at the top, h goes to the folder above, onto the place's row, folded
  await s.press('k');
  for (let i = 0; i < 3; i++) await s.press('h');
  check(
    await until(
      async () => (await s.hash()) === '' && (await s.cursor()) === 'work',
    ),
    `${tag} keys: h at the top goes to the folder above, onto its row`,
    `${await s.hash()} ${await s.cursor()}`,
  );
  check(!s.errors.length, `${tag} keys: no page errors`, s.errors.join(' | '));

  // by taps
  s = await session(width, touch);
  await taps(s, tag);

  // a subject's stages unfold in place, and its entries keep the tick
  // column
  s = await session(width, touch);
  await s.load(`${shelf}#math`);
  check(
    (await s.js(`() => document.querySelectorAll('main li.stage').length`)) >
      0,
    `${tag} subject: math lists its stages`,
  );
  await c.evaluate(`document.querySelector('main li.stage > .row').click()`);
  check(
    await until(() =>
      s.js(`() => !!document.querySelector('main li.stage > ul li.entry')`),
    ),
    `${tag} subject: a tap unfolds a stage in place`,
  );
  await s.press('j');
  await s.press('h');
  check(
    await s.js(`() => document.querySelector('.focus')?.parentElement
      .matches('li.stage') && !document.querySelector('main li.stage > ul')`),
    `${tag} subject: h on an entry goes up onto its stage and folds it`,
  );
  check(
    !s.errors.length,
    `${tag} subject: no page errors`,
    s.errors.join(' | '),
  );

  console.log(`-- tick, ${tag}`);
  // a file outside a subject: no column; x ticks it and a ✓ shows at its
  // edge, only while it is ticked; a tap on the ✓ unticks it. the cursor
  // goes there by keys, as a tap opens it
  s = await session(width, touch);
  await s.load(work);
  await s.tap('work/ssv');
  await s.tap('work/ssv/poc');
  await s.walk(flow);
  let m = await s.mark(flow);
  check(
    !m.shown && m.start === 10,
    `${tag} tick: a file has no tick column; its name starts the row`,
    JSON.stringify(m),
  );
  await s.press('x');
  m = await s.mark(flow);
  check(
    m.shown &&
      m.checked &&
      m.glyph === '"✓"' &&
      m.end === 10 &&
      (await s.ticked(flow)),
    `${tag} tick: x ticks it, and a ✓ shows at the row's edge`,
    JSON.stringify(m),
  );
  await shot(`${width}-tick-file`);
  await s.press('x');
  m = await s.mark(flow);
  check(
    !m.shown && !(await s.ticked(flow)),
    `${tag} tick: x again unticks it, and the ✓ goes`,
    JSON.stringify(m),
  );
  await s.press('x');
  await s.tap(flow, ':scope > .tick');
  check(
    !(await s.mark(flow)).shown &&
      !(await s.ticked(flow)) &&
      !(await s.viewing()),
    `${tag} tick: a tap on the ✓ unticks it, and opens nothing`,
  );
  await s.press('x');
  await s.load();
  // unfolded as it was left, so flow shows again (step 6)
  await until(async () => (await s.cursor()) === flow);
  check((await s.mark(flow)).shown, `${tag} tick: the ✓ survives a reload`);
  check(!s.errors.length, `${tag} tick: no page errors`, s.errors.join(' | '));

  // in a subject: an entry keeps its tick in a column, its ticks counted as
  // t/n. an entry with no title ticks under its path, as the file it names
  // does: tick one, and the other shows it
  s = await session(width, touch);
  await s.load(`${shelf}#cs`);
  const pairs =
    await s.js(`() => [...document.querySelectorAll('main li.stage')]
    .flatMap((li, i) => li.node.kids.filter(k => k.kind === 'entry'
      && !k.title && k.name?.startsWith('cs/')).map(k => [i, k.name]))`);
  const pair = pairs.find(([, p]) => {
    const path = `${home}/Data/${p}`;
    return (
      p.split('/').length === 3 && existsSync(path) && statSync(path).isFile()
    );
  });
  check(!!pair, `${tag} tick: cs has an entry with no title, naming a file`);
  const [i, name] = pair;
  const folder = name.split('/').slice(0, 2).join('/');
  const stage = `document.querySelectorAll('main li.stage')[${i}]`;
  const entry = `${stage}.querySelector('li.entry input[name="${name}"]')`;
  const count = () => s.js(`() => ${stage}.firstChild.dataset.count`);
  // the file's row drawn first, so the entry's tick must reach a row
  // already shown
  await s.tap(folder);
  await c.evaluate(`${stage}.firstChild.click()`);
  await until(() => s.js(`() => !!${entry}`));
  const before = await count();
  const column = await s.js(`() => {
    const r = ${entry}.closest('.row').getBoundingClientRect();
    const t = ${entry}.closest('.tick').getBoundingClientRect();
    return Math.round(t.left - r.left) + ' ' + Math.round(t.width); }`);
  check(
    column === '10 24',
    `${tag} tick: an entry keeps its tick in a column, at the row's start`,
    column,
  );
  await s.hit(`${entry}.closest('.tick')`);
  const [t, n] = before.split('/').map(Number);
  check(
    await until(async () => (await count()) === `${t + 1}/${n}`),
    `${tag} tick: a tap on it ticks it, and its stage counts ${t + 1}/${n}`,
    await count(),
  );
  m = await s.mark(name);
  check(
    m.shown && m.checked,
    `${tag} tick: the file it names, in its folder, shows the ✓`,
    JSON.stringify(m),
  );
  await shot(`${width}-tick-subject`);
  // a ✓ not shown has no tap: the check below then fails, as it should
  if (m.shown) await s.tap(name, ':scope > .tick');
  check(
    await until(
      async () =>
        (await count()) === before &&
        !(await s.js(`() => ${entry}.checked`)) &&
        !(await s.ticked(name)),
    ),
    `${tag} tick: a tap on that ✓ unticks both`,
  );
  check(!s.errors.length, `${tag} tick: no page errors`, s.errors.join(' | '));

  console.log(`-- frame, ${tag}`);
  s = await session(width, touch);
  await s.load(work);
  await frames(s, tag.replace(' ', '-'), wide);
  if (wide) await card(s, tag);

  console.log(`-- outline, ${tag}`);
  s = await session(width, touch);
  await outlines(s, tag.replace(' ', '-'));
  if (wide) await others(s, tag);
  if (wide) await nodes(s, tag);
  if (wide) await shapes(s, tag);
  await python(s, tag.replace(' ', '-'));
  if (wide) await signatures(s, tag);

  console.log(`-- keys, ${tag}`);
  s = await session(width, touch);
  await (wide ? keyed : tapped)(s, tag.replace(' ', '-'));

  console.log(`-- kept, ${tag}`);
  s = await session(width, touch);
  await reload(s, tag.replace(' ', '-'));

  console.log(`-- answers, ${tag}`);
  s = await session(width, touch);
  await answers(s, tag.replace(' ', '-'));
  if (wide) await pathed(s, tag);

  console.log(`-- help, ${tag}`);
  s = await session(width, touch);
  await helped(
    s,
    tag.replace(' ', '-'),
    `http://127.0.0.1:${medium.address().port}`,
  );

  console.log(`-- kind, ${tag}`);
  s = await session(width, touch);
  await kinds(s, tag.replace(' ', '-'));

  console.log(`-- folder, ${tag}`);
  s = await session(width, touch);
  await folders(s, tag.replace(' ', '-'));

  console.log(`-- source, ${tag}`);
  s = await session(width, touch);
  await sources(s, tag.replace(' ', '-'));

  console.log(`-- bar, ${tag}`);
  s = await session(width, touch);
  await scrolled(s, tag.replace(' ', '-'));

  console.log(`-- edge, ${tag}`);
  s = await session(width, touch);
  await edges(s, tag.replace(' ', '-'));

  console.log(`-- top, ${tag}`);
  s = await session(width, touch);
  await tops(s, tag.replace(' ', '-'));

  console.log(`-- picture, ${tag}`);
  s = await session(width, touch);
  await pictured(s, tag.replace(' ', '-'));
  await widened(s, tag.replace(' ', '-'));

  if (!touch) {
    console.log(`-- fold, ${tag}`);
    s = await session(width, touch);
    await still(s, tag);
  }
}

// the oracles, from the disk: md's headings, # outside a fence, or a line
// underlined in = or - after a blank
function headings(path) {
  const lines = readFileSync(`${home}/${path}`, 'utf8').split('\n');
  let fenced = false;
  return lines.flatMap((line, i) => {
    if (/^(```|~~~)/.test(line)) fenced = !fenced;
    if (fenced) return [];
    const atx = line.match(/^#{1,6}\s+(.*?)[\s#]*$/);
    if (atx) return [atx[1]];
    const under = /^(=+|-+)\s*$/.test(lines[i + 1] ?? '');
    return under && /^\S/.test(line) && !lines[i - 1]?.trim()
      ? [line.trim()]
      : [];
  });
}
// flow, a leaf with # headings, opened: folded at first, a page's o, so
// the top heading, holding the rest, is open and they are shut; a tap
// unfolds one and folds it again
async function outlines(s, tag) {
  await s.open(work.replace('#work', `#${flow}`));
  let o = await s.outlined();
  const heads = headings(flow);
  check(
    o.kind === 'outline',
    `${tag} outline: flow opens as an outline`,
    json(o),
  );
  check(
    json(o.heads) === json(heads) &&
      json(o.open) === json(heads.map((_, i) => i === 0)),
    `${tag} outline: folded at first, its headings in sight: the top open`,
    json(o),
  );
  check(
    json(o.ringed) === json(heads.map((_, i) => i > 0)) && o.bodies === 1,
    `${tag} outline: a folded heading's bullet is ringed; one body shows`,
    json(o),
  );
  check(
    o.height === (s.touch ? 40 : 30),
    `${tag} outline: a heading is a row, ${s.touch ? 40 : 30}px`,
    String(o.height),
  );
  await s.head(heads[1]);
  o = await s.outlined();
  check(
    o.open?.[1] && !o.ringed?.[1] && o.bodies === 2,
    `${tag} outline: a tap unfolds a heading: its lines show, its ring goes`,
    json(o),
  );
  const block = await s.js(`() => (p => p && [getComputedStyle(p).whiteSpace,
    p.scrollWidth > p.clientWidth])([...document.querySelectorAll(
    'aside .outline summary')].find(s => s.textContent === ${json(heads[1])})
    ?.parentElement.querySelector(':scope > div > pre'))`);
  const phone = s.width < 800;
  check(
    phone ? block?.[0] === 'pre' && block[1] : block?.[0] === 'pre-wrap',
    phone
      ? `${tag} outline: on a phone a block keeps its columns, ` +
          'and scrolls across'
      : `${tag} outline: wide, a block wraps as before`,
    json(block),
  );
  await shot(`${tag}-outline`);
  await s.head(heads[1]);
  o = await s.outlined();
  check(
    !o.open?.[1] && o.ringed?.[1] && o.bodies === 1,
    `${tag} outline: a tap again folds it`,
    json(o),
  );
  // step 17: the file's bar folds the whole, e and c its icons there; the
  // outline holds no bar of its own
  const own = await s.js(`() => !!document.querySelector(
    'aside .outline :is(button, .folds)')`);
  const press = key =>
    s.hit(`document.querySelector('aside nav [data-key=${key}]')`);
  await press('e');
  const all = (await s.outlined()).open;
  await press('c');
  const none = (await s.outlined()).open;
  check(
    !own && all?.every(Boolean) && none?.length === 1 && !none[0],
    `${tag} outline: the file's bar folds it: expand all, collapse all`,
    json({ own, all, none }),
  );
  check(
    !s.errors.length,
    `${tag} outline: no page errors`,
    s.errors.join(' | '),
  );
}
// the other headings, wide: underlined, bare lines, md's # with a fence;
// a script and a leaf with no headings stay text; and the preview is the
// viewer, its folds kept when it opens
async function others(s, tag) {
  await s.open(work.replace('#work', '#work/pipeline'));
  let o = await s.outlined();
  check(
    json(o.heads) === json(headings('work/pipeline')),
    `${tag} outline: a leaf with underlined headings, by them`,
    json(o),
  );
  await s.open(work.replace('#work', '#brief'));
  o = await s.outlined();
  check(
    ['the idea', 'decided', 'next, in order', 'check'].every(h =>
      o.heads?.includes(h),
    ) &&
      o.heads?.every(h => !/[.:;,!?]$/.test(h)) &&
      !o.open?.some(Boolean),
    `${tag} outline: a leaf with bare lines for headings, by them`,
    json(o),
  );
  const md = 'cs/aiml/notes/0-concepts/2-residual-stream.md';
  await s.open(`${shelf}#${md}`);
  o = await s.outlined();
  const mdheads = headings(`Data/${md}`);
  check(
    json(o.heads) === json(mdheads),
    `${tag} outline: md by its # levels`,
    json(o),
  );
  const fenced = await s.js(`() => [...document.querySelectorAll(
    'aside .outline pre')].some(p => p.textContent.includes(
    '# The embedding table'))
    && ![...document.querySelectorAll('aside summary')].some(s =>
      s.textContent.includes('embedding table'))`);
  check(fenced, `${tag} outline: a # in a fenced block is a line, no heading`);
  for (const [url, what] of [
    [work.replace('#work', '#work/tmux-layout'), 'a script, #! first'],
    [`${shelf}#guidelines/general/check`, 'a leaf with no headings'],
  ]) {
    await s.open(url);
    o = await s.outlined();
    check(o.kind === 'pre', `${tag} outline: ${what} stays text`, json(o));
  }
  await s.load(work);
  await s.tap('work/ssv');
  await s.tap('work/ssv/poc');
  await s.walk(flow);
  await until(async () => (await s.outlined()).kind === 'outline');
  await s.head(headings(flow)[1]);
  await s.js(`() => document.querySelector('aside > :last-child').mine = 1`);
  await s.tap(flow);
  await until(s.viewing);
  o = await s.outlined();
  check(
    o.open?.[1] &&
      (await s.js(`() => document.querySelector('aside > :last-child')
        .mine`)) === 1,
    `${tag} outline: previewed, opened: the same view, its folds kept`,
    json(o),
  );
  check(
    !s.errors.length,
    `${tag} others: no page errors`,
    s.errors.join(' | '),
  );
}

// md's nodes, todo/studio/refactor's step 1: each heading a page's node, its id
// the path of the headings down to it. the oracle is a file written here,
// beside the ids and headings it must have: a duplicate heading, a # in a
// fenced block, a line underlined. the keys reach each node and nothing
// else: a component's own details has no id, and is passed by. the
// heading is the summary's first child, what the box lists, and a figure
// after it sits at the edge. an id leaves the shelf's own alone
async function nodes(s, tag) {
  const md = [
    '# shelf',
    '',
    '## design',
    '',
    '### notes',
    '',
    '### notes',
    '',
    '```sh',
    '# a comment, no heading',
    '```',
    '',
    'setext',
    '------',
  ];
  const ids = [
    'shelf',
    'shelf/design',
    'shelf/design/notes',
    'shelf/design/notes-2',
    'shelf/setext',
  ];
  const heads = ['shelf', 'design', 'notes', 'notes', 'setext'];
  const dir = `${tmpdir()}/shelfnode`;
  await mkdir(dir, { recursive: true });
  await writeFile(`${dir}/node.md`, md.join('\n'));
  const host = await serve(dir);
  await s.open(`http://127.0.0.1:${host.address().port}/shelf.html#node.md`);
  const got = await s.js(`() => {
    const view = document.querySelector('aside > :last-child');
    const nodes = [...view.querySelectorAll('details[id]')];
    return {
      ids: nodes.map(d => d.id),
      heads: nodes.map(d =>
        d.querySelector(':scope > summary > :first-child')?.textContent),
      text: [...view.querySelectorAll('pre')].map(p => p.textContent),
    }; }`);
  check(
    json(got.ids) === json(ids) &&
      new Set(got.ids).size === ids.length &&
      json(got.heads) === json(heads),
    `${tag} node: md's ids unique, each the path of its headings`,
    json(got),
  );
  check(
    got.ids.includes('shelf/design/notes-2'),
    `${tag} node: a duplicate heading gets -2`,
  );
  check(
    got.text.some(t => t.includes('# a comment')) &&
      !got.heads.some(h => h?.includes('a comment')),
    `${tag} node: a # in a fenced block is a line, no heading`,
  );
  check(
    got.ids.includes('shelf/setext'),
    `${tag} node: a line underlined is a heading`,
  );
  // tab, all open, then j over every node. a details with no id put in
  // the first node's lines is no node: the keys pass it by
  await s.js(`() => document.querySelector('aside .outline details[id] > div')
    ?.prepend(Object.assign(document.createElement('details'),
      { innerHTML: '<summary>own</summary>' }))`);
  await s.press('Tab');
  await s.press('e');
  const walked = [];
  for (let i = 0; i < ids.length; i++) {
    walked.push(
      await s.js(`() => document.querySelector(
        'aside .outline details.focus')?.id`),
    );
    await s.press('j');
  }
  check(
    json(walked) === json(ids),
    `${tag} node: the keys reach every heading, and pass a component's ` +
      'own details by',
    json(walked),
  );
  // a figure after the heading sits at the summary's right, and the box
  // lists the headings without it
  const edge = await s.js(`() => {
    const summary = document.querySelector('aside .outline summary');
    const figure = summary?.appendChild(document.createElement('small'));
    if (!figure) return null;
    figure.textContent = '9';
    return Math.round(summary.getBoundingClientRect().right
      - figure.getBoundingClientRect().right); }`);
  check(
    edge === 0,
    `${tag} node: a figure after the heading sits at the edge`,
    String(edge),
  );
  await s.press('/');
  const listed = await s.js(`() => [...document.querySelectorAll(
    'dialog[open] li > span')].map(e => e.textContent)`);
  await s.press('Escape');
  check(
    json(listed) === json(heads),
    `${tag} node: the box lists the headings, not what is at the edge`,
    json(listed),
  );
  // an id leaves the shelf's own alone: a heading Book is #book, as the
  // template an epub is drawn in is. an epub previewed after it opens
  const books = `${tmpdir()}/shelfbook`;
  await mkdir(books, { recursive: true });
  await writeFile(`${books}/a.md`, '# Book\n\ntext.\n');
  await writeFile(
    `${books}/b.epub`,
    readFileSync(`${library}/math/elem/how-to-solve-it-george-polya.epub`),
  );
  const shelved = await serve(books);
  await s.load(`http://127.0.0.1:${shelved.address().port}/shelf.html`);
  await s.walk('a.md');
  await until(async () => (await s.outlined()).kind === 'outline');
  await s.walk('b.epub');
  check(
    await until(
      async () =>
        (await s.js(`() => document.querySelector('aside > :last-child')
          .localName`)) === 'article',
    ),
    `${tag} node: an epub previewed after a heading Book opens: an id ` +
      "leaves the shelf's own alone",
  );
  check(!s.errors.length, `${tag} node: no page errors`, s.errors.join(' | '));
}

// every kind's node one shape, todo/studio/refactor's step 2: its id, its
// heading, the figures at its edge and its body, as part() draws them. the
// oracle is a diff and a python file written here, beside what they must
// show: a file headed by its path, its status and +n −m at its edge, a gap
// apart, its id its path's, no count in it; a node of python's an id each
async function shapes(s, tag) {
  const diff = [
    'diff --git a/medium/x.md b/medium/x.md',
    'index 1111111..2222222 100644',
    '--- a/medium/x.md',
    '+++ b/medium/x.md',
    '@@ -1,2 +1,3 @@',
    ' # x',
    '-old',
    '+new',
    '+more',
    'diff --git a/y.txt b/y.txt',
    'new file mode 100644',
    'index 0000000..3333333',
    '--- /dev/null',
    '+++ b/y.txt',
    '@@ -0,0 +1 @@',
    '+y',
  ];
  const nodes = [
    ['medium-x-md', 'medium/x.md', 'M', '+2', '−1'],
    ['medium-x-md/1-2-1-3', '@@ -1,2 +1,3 @@'],
    ['y-txt', 'y.txt', 'A', '+1', '−0'],
    ['y-txt/0-0-1', '@@ -0,0 +1 @@'],
  ];
  const py = [
    '"""a module."""',
    'import os',
    '',
    '',
    'class A:',
    '    def f(self):',
    '        return 1',
    '',
    '    def f(self):',
    '        return 2',
    '',
    '',
    'def g():',
    '    pass',
  ];
  const dir = `${tmpdir()}/shelfshape`;
  await mkdir(dir, { recursive: true });
  await writeFile(`${dir}/a.diff`, diff.join('\n'));
  await writeFile(`${dir}/a.py`, py.join('\n'));
  const host = await serve(dir);
  const url = `http://127.0.0.1:${host.address().port}/shelf.html`;
  await s.open(`${url}#a.diff`);
  const got = await s.js(`() => [...document.querySelectorAll(
    'aside > :last-child details[id]')].map(d => [d.id,
    ...[...d.querySelector(':scope > summary').children].map(e =>
      e.textContent)])`);
  check(
    json(got.map(([, ...shown]) => shown)) ===
      json(nodes.map(([, ...shown]) => shown)),
    `${tag} shape: a diff's file is headed by its path, its status and ` +
      '+n −m the figures at its edge',
    json(got),
  );
  check(
    json(got.map(([id]) => id)) === json(nodes.map(([id]) => id)),
    `${tag} shape: a diff's file's id is its path's, no count in it`,
    json(got.map(([id]) => id)),
  );
  // the last figure at the summary's right, each a gap from the one
  // before it
  const gaps = await s.js(`() => {
    const summary = document.querySelector('aside .outline summary');
    if (!summary) return null;
    const boxes = [summary, ...summary.children].map(e =>
      e.getBoundingClientRect());
    return [boxes[0].right - boxes.at(-1).right,
      ...boxes.slice(2).map((b, i) => b.left - boxes[i + 1].right)]
      .map(Math.round); }`);
  check(
    gaps?.length === 4 && gaps[0] === 0 && gaps.slice(1).every(gap => gap > 0),
    `${tag} shape: a diff's figures sit at the edge, a gap apart`,
    json(gaps),
  );
  // a class, a def in it twice, a def, and the code above them one node
  await s.open(`${url}#a.py`);
  const ids = await s.js(`() => [...document.querySelectorAll(
    'aside > :last-child details')].map(d => d.id)`);
  check(
    ids.length === 5 && ids.every(Boolean) && new Set(ids).size === 5,
    `${tag} shape: python's nodes have an id each, none twice`,
    json(ids),
  );
  check(
    !s.errors.length,
    `${tag} shape: no page errors`,
    s.errors.join(' | '),
  );
}

// python's own view, the oracle: the code before the first def one
// heading, its first line; each # -- name -- mark at the margin a heading,
// the defs after it under it; each class or def, its depth by ast, its
// heading as its code says it, its decorators inline, its params as
// todo/studio/refactor's oracle.py has them, and its id its qualified name, a
// duplicate -2
function ast(path) {
  return JSON.parse(
    execFileSync(
      'python3',
      [
        '-c',
        `import ast, json, re, sys
src = open(sys.argv[1]).read()
lines = src.split('\\n')
marks = [i + 1 for i, l in enumerate(lines)
         if re.match(r'# -- [a-z]+ --\\s*$', l)]
tree = ast.parse(src)
defs = (ast.FunctionDef, ast.AsyncFunctionDef, ast.ClassDef)
first = lambda n: min([d.lineno for d in n.decorator_list] + [n.lineno])
def params(k):
    a = k.args; pos = a.posonlyargs + a.args
    dfl = [None] * (len(pos) - len(a.defaults)) + list(a.defaults)
    dv = lambda d: '' if d is None else '=' + (
        ast.unparse(d) if len(ast.unparse(d)) <= 10 else '…')
    out = [x.arg + dv(d) for x, d in zip(pos, dfl)
           if x.arg not in ('self', 'cls')]
    if a.vararg: out.append('*' + a.vararg.arg)
    elif a.kwonlyargs: out.append('*')
    out += [x.arg + dv(d) for x, d in zip(a.kwonlyargs, a.kw_defaults)]
    if a.kwarg: out.append('**' + a.kwarg.arg)
    return ', '.join(out)
def heading(n):
    deco = ''.join('@' + ast.unparse(d) + ' ' for d in n.decorator_list)
    if isinstance(n, ast.ClassDef):
        bases = [ast.unparse(b) for b in n.bases] + [
            '%s=%s' % (k.arg, ast.unparse(k.value)) for k in n.keywords]
        return deco + 'class ' + n.name + (
            '(' + ', '.join(bases) + ')' if bases else '')
    a = 'async ' if isinstance(n, ast.AsyncFunctionDef) else ''
    return deco + a + 'def ' + n.name + '(' + params(n) + ')'
out, seen = [], set()
def walk(body, depth, top, owner):
    for n in body:
        if isinstance(n, defs):
            d = depth + (1 if top and any(m < first(n) for m in marks) else 0)
            id, k = owner + n.name, 2
            while id in seen: id, k = '%s-%d' % (owner + n.name, k), k + 1
            seen.add(id)
            out.append([d, heading(n), first(n), id])
            walk(n.body, d + 1, False, owner + n.name + '.')
        else:
            for f in ('body', 'handlers', 'orelse', 'finalbody'):
                walk(getattr(n, f, []) or [], depth, top, owner)
walk(tree.body, 0, True, '')
out += [[0, re.search(r'-- ([a-z]+) --', lines[m - 1]).group(1), m, '']
        for m in marks]
lead = min([o[2] for o in out], default=len(lines) + 1)
before = [l for l in lines[:lead - 1] if l.strip()]
if before and out:
    out.append([0, before[0].strip(), 0, ''])
out.sort(key=lambda o: o[2])
print(json.dumps([[o[0], o[1], o[3]] for o in out]))`,
        path,
      ],
      { encoding: 'utf8' },
    ),
  );
}

// python, opened: its preamble, marks, classes and defs are its headings,
// as the oracle has them, and a class's or def's id its qualified name;
// all in sight at first, for each heading holding defs is open; the code
// before the first def one node, shut; a decorated def and a signature on
// two lines keep to their heading, one line, hovered as written and with
// the docstring; a tap opens a def's body, dedented
async function python(s, tag) {
  const py = 'work/ssv/poc/model.py';
  await s.open(work.replace('#work', `#${py}`));
  const want = ast(`${home}/${py}`);
  const got = await s.js(`() => {
    const view = document.querySelector('aside > :last-child');
    if (!view.matches('.outline')) return { kind: view.localName };
    const all = [...view.querySelectorAll('summary')];
    const depth = s => {
      let d = 0;
      for (let p = s.parentElement.parentElement.closest('details');
        p && view.contains(p); p = p.parentElement.closest('details')) d++;
      return d;
    };
    return {
      kind: 'outline',
      heads: all.map(s => [depth(s), s.textContent.split('\\n')[0].trim()]),
      seen: all.filter(s => s.checkVisibility()).length,
      top: (d => d && !d.open && [d.firstElementChild.textContent,
        d.querySelector(':scope > div').textContent])(
        view.querySelector(':scope > details')),
      ids: all.map(s => s.parentElement.id),
      hovers: Object.fromEntries(all.map(s =>
        [s.textContent, s.firstElementChild.title])),
      height: Math.round(all.filter(s => !s.textContent.includes('\\n'))
        .sort((a, b) => a.textContent.length - b.textContent.length)[0]
        .getBoundingClientRect().height),
    }; }`);
  check(
    got.kind === 'outline',
    `${tag} python: it opens as an outline`,
    json(got.kind),
  );
  check(
    json(got.heads) === json(want.map(([depth, head]) => [depth, head])),
    `${tag} python: its preamble, marks, classes and defs, ` +
      'as the oracle nests them',
    `${got.heads?.length} heads, ast ${want.length}`,
  );
  check(
    want.every(([, , id], i) => !id || got.ids?.[i] === id),
    `${tag} python: a class's or def's id is its qualified name, ` +
      'Field.broken_constraint, a duplicate -2',
    json(got.ids?.filter((id, i) => want[i]?.[2] && id !== want[i][2])),
  );
  check(
    got.seen === want.length,
    `${tag} python: folded at first, every heading in sight`,
    `${got.seen} of ${want.length}`,
  );
  const first = readFileSync(`${home}/${py}`, 'utf8').split('\n');
  check(
    got.top?.[0] === first[0] &&
      got.top[1].includes('log = logging.getLogger(__name__)'),
    `${tag} python: the code before the first def is one node, ` +
      'shut, its first line the heading',
    json(got.top),
  );
  const hovers = Object.entries(got.hovers ?? {});
  check(
    got.hovers?.['@property def warns()'] === '@property\ndef warns(self):' &&
      hovers.some(
        ([head, hover]) =>
          head.startsWith('def __init__(payload') &&
          hover.startsWith('def __init__(self, payload') &&
          hover.split('\n').length === 2,
      ) &&
      got.hovers['def broken_constraint(value)']?.startsWith(
        'def broken_constraint(self, value):\n\nthe constraint this value',
      ),
    `${tag} python: decorators and a two-line signature keep to the ` +
      'heading, one line; hovered, as written, then the docstring',
  );
  check(
    got.height === (s.touch ? 40 : 30),
    `${tag} python: a one-line heading is a row, ${s.touch ? 40 : 30}px`,
    String(got.height),
  );
  await s.head('@property def warns()');
  const body =
    await s.js(`() => [...document.querySelectorAll('aside summary')]
    .find(s => s.textContent === '@property def warns()')
    ?.parentElement.querySelector(':scope > div > pre')?.textContent`);
  check(
    body === 'return self.severity == "warning"',
    `${tag} python: a tap opens a def's body, dedented`,
    json(body),
  );
  await shot(`${tag}-python`);
  check(
    !s.errors.length,
    `${tag} python: no page errors`,
    s.errors.join(' | '),
  );
}

// python's heading, todo/studio/refactor's step 3, in the cases model.py has not:
// an async def, a def in a def, a signature over lines, a decorator over
// lines, a def named twice, and a file indented by tabs. the oracle is
// ast, as for model.py: a case holds when the node ast names is headed as
// ast heads it, and its id is ast's
async function signatures(s, tag) {
  const cases = [
    '"""the cases a heading meets."""',
    '',
    '',
    'class Client:',
    '    async def fetch(self, url, *, timeout=30):',
    '        """the page at url, or an error."""',
    '        return url',
    '',
    '    @property',
    '    def x(self):',
    '        return 1',
    '',
    '    @x.setter',
    '    def x(self, value):',
    '        self._x = value',
    '',
    '',
    'def outer(a, b=1):',
    '    def inner(c):',
    '        return c',
    '',
    '    return inner',
    '',
    '',
    'def wide(',
    '    first: int,',
    "    second: str = 'abcdefghijk',",
    '    *rest,',
    '    key=lambda k: k,',
    '    **options,',
    ') -> None:',
    '    pass',
    '',
    '',
    '@decorated(',
    "    'name',",
    '    level=2,',
    ')',
    'def marked():',
    '    pass',
  ];
  const tabs = [
    'class Tabbed:',
    '\tdef method(self, a, b=2):',
    '\t\t"""a method."""',
    '\t\treturn a',
  ];
  const dir = `${tmpdir()}/shelfsign`;
  await mkdir(dir, { recursive: true });
  await writeFile(`${dir}/cases.py`, cases.join('\n'));
  await writeFile(`${dir}/tabs.py`, tabs.join('\n'));
  const host = await serve(dir);
  const url = `http://127.0.0.1:${host.address().port}/shelf.html`;
  // each node ast has, its heading and id, beside the one drawn there:
  // its heading, id and hover
  const read = async file => {
    await s.open(`${url}#${file}`);
    const got = await s.js(`() => [...document.querySelectorAll(
      'aside > :last-child summary')].map(s => [s.textContent,
      s.parentElement.id, s.firstElementChild.title])`);
    return ast(`${dir}/${file}`).map(([, head, id], i) => [
      head,
      id,
      ...(got[i] ?? []),
    ]);
  };
  const holds = (nodes, id) =>
    nodes.some(
      ([head, want, got, drawn]) =>
        want === id && drawn === id && got === head,
    );
  const hover = (nodes, id) => nodes.find(([, want]) => want === id)?.[4];
  const cased = await read('cases.py');
  check(
    holds(cased, 'Client.fetch') &&
      hover(cased, 'Client.fetch') ===
        'async def fetch(self, url, *, timeout=30):\n\n' +
          'the page at url, or an error.',
    `${tag} sign: an async def is headed async def; hovered, its ` +
      'signature and docstring',
    json(cased.find(([, id]) => id === 'Client.fetch')),
  );
  check(
    holds(cased, 'outer.inner'),
    `${tag} sign: a def in a def is under it, its id outer.inner`,
    json(cased.find(([, id]) => id === 'outer.inner')),
  );
  check(
    holds(cased, 'wide'),
    `${tag} sign: a signature over lines is one line, its annotations ` +
      'let go, a default over 10 characters =…',
    json(cased.find(([, id]) => id === 'wide')),
  );
  check(
    holds(cased, 'marked'),
    `${tag} sign: a decorator over lines is inline, on one line`,
    json(cased.find(([, id]) => id === 'marked')),
  );
  check(
    holds(cased, 'Client.x') && holds(cased, 'Client.x-2'),
    `${tag} sign: a def named twice, a property and its setter, the ` +
      'second -2',
    json(cased.filter(([, id]) => id.startsWith('Client.x'))),
  );
  const tabbed = await read('tabs.py');
  check(
    holds(tabbed, 'Tabbed.method') &&
      hover(tabbed, 'Tabbed.method') ===
        'def method(self, a, b=2):\n\na method.',
    `${tag} sign: a file indented by tabs nests and reads as one by spaces`,
    json(tabbed),
  );
  check(!s.errors.length, `${tag} sign: no page errors`, s.errors.join(' | '));
}

// the keys in the view, wide: an outline's, then a page's in a frame
async function keyed(s, tag) {
  const keys = () =>
    s.js(`() => document.querySelector('aside').classList.contains('keys')`);
  const at = () =>
    s.js(`() => document.querySelector(
      'aside .outline details.focus > summary')?.textContent`);
  const open = () =>
    s.js(`() => [...document.querySelectorAll('aside .outline details')]
      .map(d => d.open)`);
  const heads = headings(flow);
  await s.open(work.replace('#work', `#${flow}`));
  await s.press('Tab');
  check(
    (await keys()) && (await at()) === heads[0],
    `${tag} keys: tab hands the outline the keys, on its first heading`,
    `${await keys()} ${await at()}`,
  );
  await s.press('j');
  const j = await at();
  await s.press('k');
  check(
    j === heads[1] && (await at()) === heads[0],
    `${tag} keys: j the next heading in sight, k the one before`,
    `${j} ${await at()}`,
  );
  await s.press('j');
  await s.press('l');
  const l = (await open())[1];
  await s.press('l');
  check(
    l && (await at()) === heads[2],
    `${tag} keys: l unfolds, and on an open one steps on`,
    `${l} ${await at()}`,
  );
  await s.press('h');
  check(
    (await at()) === heads[0] && !(await open())[0],
    `${tag} keys: h on a shut one goes up, and folds that`,
    `${await at()} ${json(await open())}`,
  );
  await s.press('o');
  check(
    json(await open()) === json(heads.map((_, i) => i === 0)),
    `${tag} keys: o folds to the outline, as at first`,
    json(await open()),
  );
  await s.press('e');
  const e = (await open()).every(Boolean);
  for (let i = 0; i < 3; i++) await s.press('j');
  await s.press('z');
  check(
    e &&
      json(await open()) === json(heads.map((_, i) => i === 0 || i === 3)) &&
      (await at()) === heads[3],
    `${tag} keys: e expands all; z folds the rest`,
    `${e} ${json(await open())} ${await at()}`,
  );
  await s.press('c');
  check(
    !(await open()).some(Boolean) && (await at()) === heads[0],
    `${tag} keys: c collapses all, the cursor up to what shows`,
    `${json(await open())} ${await at()}`,
  );
  await s.press('/');
  await c.send('Input.insertText', { text: 'probe' });
  await sleep(150);
  const listed = await s.js(`() => [...document.querySelectorAll(
    'dialog li')].map(li => li.textContent)`);
  await s.press('Enter');
  check(
    json(listed) === json(heads.filter(h => h.includes('probe'))) &&
      (await at()) === listed[0] &&
      (await open())[0],
    `${tag} keys: / lists the headings found, enter goes to one, shown`,
    `${json(listed)} ${await at()}`,
  );
  await s.press('Escape');
  const tree = await s.cursor();
  await s.press('j');
  check(
    !(await keys()) && (await s.cursor()) !== tree,
    `${tag} keys: esc hands them back, and j moves the tree again`,
  );
  await s.press('k');
  await s.press('Tab');
  await s.press('c');
  await s.press('h');
  check(
    !(await keys()),
    `${tag} keys: h at the first heading hands them back, not stuck`,
  );
  await s.press('Tab');
  await s.press('Shift+Tab');
  check(!(await keys()), `${tag} keys: shift+tab hands them back`);
  // the box opens on the heading under the cursor, in sight: python's
  // outline, all open, the cursor 40 headings down
  await s.open(work.replace('#work', '#work/ssv/poc/model.py'));
  await s.hit(`document.querySelector('aside nav [data-key=e]')`);
  await s.press('Tab');
  for (let i = 0; i < 40; i++) await s.press('j');
  await s.press('/');
  const sight = await s.js(`() => {
    const ol = document.querySelector('dialog[open] ol');
    const li = ol.querySelector('[aria-selected=true]');
    const [a, b] = [ol, li].map(e => e.getBoundingClientRect());
    return [li.textContent.trim().slice(0, 40), Math.round(b.top - a.top),
      Math.round(a.bottom - b.bottom)]; }`);
  await s.press('Escape');
  check(
    sight[1] >= 0 && sight[2] >= 0,
    `${tag} keys: / opens on the heading under the cursor, in sight`,
    json(sight),
  );

  // a page in a frame: the focus handed in, its own keys run
  await s.open(work.replace('#work', '#work/pipeline.html'));
  const frame = f =>
    s.js(
      `() => (${f})(document.querySelector('aside iframe').contentDocument)`,
    );
  await until(() => frame(`d => !!d?.querySelector('details.focus')`));
  const mark = () =>
    frame(`d => d.querySelector('details.focus > summary').textContent`);
  const inside = () =>
    s.js(`() => document.activeElement?.localName === 'iframe'`);
  await s.press('Tab');
  const first = await mark();
  await s.press('j');
  check(
    (await inside()) && (await keys()) && (await mark()) !== first,
    `${tag} keys: tab hands a page in a frame the focus; j is its own`,
    `${await inside()} ${await keys()} ${first} ${await mark()}`,
  );
  await s.press('c');
  check(
    await frame(`d => ![...d.querySelectorAll('details[id]')]
      .some(d => d.open)`),
    `${tag} keys: c is the page's own too: all shut`,
  );
  await s.press('Escape');
  check(
    !(await inside()) && !(await keys()),
    `${tag} keys: esc hands them back out of the frame`,
  );
  await s.press('Tab');
  await s.press('Shift+Tab');
  check(
    !(await inside()) && !(await keys()),
    `${tag} keys: shift+tab out of the frame`,
  );
  await s.press('Tab');
  for (let i = 0; i < 20; i++) await s.press('k');
  await s.press('h');
  check(
    !(await inside()) && !(await keys()),
    `${tag} keys: h at a page's first node hands them back, not stuck`,
  );
  check(!s.errors.length, `${tag} keys: no page errors`, s.errors.join(' | '));
}
// on touch, no tab: a tap goes to the pane it lands in. a tap on a heading
// hands the outline the keys, closing gives them back, and a tap in a
// page's frame is the page's own
async function tapped(s, tag) {
  const heads = headings(flow);
  await s.open(work.replace('#work', `#${flow}`));
  await s.head(heads[1]);
  check(
    (await s.js(`() => document.querySelector('aside')
      .classList.contains('keys')`)) &&
      (await s.js(`() => document.querySelector(
        'aside .outline details.focus > summary')?.textContent`)) === heads[1],
    `${tag} keys: a tap on a heading hands the outline the keys, there`,
  );
  await s.hit(`document.querySelector('aside nav > a:not([rel])')`);
  await until(async () => !(await s.viewing()));
  check(
    !(await s.js(`() => document.querySelector('aside')
      .classList.contains('keys')`)),
    `${tag} keys: closed, the keys are the tree's`,
  );
  await s.open(work.replace('#work', '#work/pipeline.html'));
  const first = `document.querySelector('aside iframe').contentDocument
    ?.querySelector('details[id]')`;
  await until(() => s.js(`() => !!${first}`));
  const was = await s.js(`() => ${first}.open`);
  await s.hit(`${first}.querySelector(':scope > summary')`);
  check(
    (await s.js(`() => ${first}.open`)) !== was,
    `${tag} keys: a tap in a page's frame is the page's own: it folds`,
  );
  check(!s.errors.length, `${tag} taps: no page errors`, s.errors.join(' | '));
}

// what is unfolded is kept across a reload, beside the cursor
async function reload(s, tag) {
  const poc = 'work/ssv/poc';
  await s.load(work);
  await s.tap('work/ssv');
  await s.tap(poc);
  // by keys, as a tap opens a file
  await s.walk(flow);
  await s.load();
  await until(async () => (await s.cursor()) === flow);
  check(
    (await s.unfolded('work/ssv')) &&
      (await s.unfolded(poc)) &&
      (await s.cursor()) === flow,
    `${tag} kept: after a reload, ssv and poc unfolded, the cursor on flow`,
    json([
      await s.unfolded('work/ssv'),
      await s.unfolded(poc),
      await s.cursor(),
    ]),
  );
  await s.press('h');
  await s.load();
  await until(async () => (await s.cursor()) === poc);
  check(
    (await s.unfolded('work/ssv')) &&
      !(await s.unfolded(poc)) &&
      (await s.cursor()) === poc,
    `${tag} kept: h folded poc; after a reload it is folded, the cursor on it`,
    `${await s.unfolded(poc)} ${await s.cursor()}`,
  );
  await s.tap(poc);
  await s.tap('work/ssv');
  await s.load();
  await until(async () => (await s.cursor()) === 'work/ssv');
  const shut = !(await s.unfolded('work/ssv'));
  await s.tap('work/ssv');
  await until(() => s.unfolded(poc));
  check(
    shut && (await s.unfolded(poc)),
    `${tag} kept: a folder folded keeps what was unfolded under it`,
    `${shut} ${await s.unfolded(poc)}`,
  );
  // a stage, and h at the top: the place left folded under the one above
  await s.load(`${shelf}#math`);
  await c.evaluate(`document.querySelector('main li.stage > .row').click()`);
  await s.load();
  check(
    await until(() =>
      s.js(`() => !!document.querySelector('main li.stage > ul')`),
    ),
    `${tag} kept: a stage unfolded is unfolded after a reload`,
  );
  await s.load(`${shelf}#`);
  await s.tap('math');
  await c.evaluate(`location.hash = '#math'`);
  await until(
    async () =>
      (await s.js(`() => document.querySelector('main header h1 b')
      ?.textContent`)) === 'math',
  );
  for (let i = 0; i < 40; i++) {
    if ((await s.js(`() => location.hash`)) === '') break;
    await s.press('h');
  }
  await s.load();
  check(
    (await until(async () => (await s.cursor()) === 'math')) &&
      !(await s.unfolded('math')),
    `${tag} kept: h at the top folds the place it leaves, after a reload too`,
    `${await s.cursor()} ${await s.unfolded('math')}`,
  );
  check(!s.errors.length, `${tag} kept: no page errors`, s.errors.join(' | '));
}

// by taps: a folder unfolds and folds, and a file opens at the first tap,
// at every width (step 13)
async function taps(s, tag) {
  await s.load(work);
  await s.tap('work/ssv');
  check(
    await until(
      async () =>
        (await s.unfolded('work/ssv')) && (await s.cursor()) === 'work/ssv',
    ),
    `${tag} taps: a tap unfolds ssv, the cursor on it`,
  );
  await s.tap('work/ssv/poc');
  check(
    await until(
      async () =>
        (await s.unfolded('work/ssv/poc')) &&
        (await s.cursor()) === 'work/ssv/poc',
    ),
    `${tag} taps: a tap unfolds poc`,
  );
  await shot(`${s.width}-taps-unfolded`);
  await s.tap(flow);
  check(
    await until(
      async () =>
        (await s.viewing()) &&
        (await s.hash()) === `#${flow}` &&
        (await s.cursor()) === flow,
    ),
    `${tag} taps: the first click on a file opens it, the cursor there`,
  );
  if (s.width < 800)
    check(
      await s.js(`() => !document.querySelector('main').checkVisibility()
        && document.querySelector('aside').checkVisibility()`),
      `${tag} taps: open, the viewer is the screen`,
    );
  await shot(`${s.width}-taps-open`);
  // none if it did not open: the checks after it then fail, as they should
  await s.hit(`document.querySelector('aside nav > a:not([rel])')`);
  check(
    await until(
      async () => !(await s.viewing()) && (await s.cursor()) === flow,
    ),
    `${tag} taps: close leaves the cursor on it`,
  );
  await s.press('h');
  await s.press('h');
  check(
    (await s.cursor()) === 'work/ssv' && (await s.opened()) === 0,
    `${tag} taps: h h folds poc, then ssv`,
  );
  await s.tap('work/ssv');
  await s.tap('work/ssv');
  check(
    !(await s.unfolded('work/ssv')) && (await s.cursor()) === 'work/ssv',
    `${tag} taps: a tap on an unfolded folder folds it`,
  );
  check(!s.errors.length, `${tag} taps: no page errors`, s.errors.join(' | '));
}

// ? as page.html has it: one table, each key bound from it and listed
// as a dl of kbd and what it does. the oracle for its look is page.html's
// own ? box, served here read-only from the medium
async function helped(s, tag, medium) {
  const moves = ['j ↓', 'k ↑', 'l →', 'h ←'];
  const every = [...moves, 'enter', 'o', 'z', 'e', 'c', '/', 'esc'];
  const more = ['tap', '.', 'x', 'n', 'p', 'f', 'i', 'u', 'v', 'tab', '?'];
  // a tap, as the head's keys have it
  const tap = 'Unfold or fold; a file opens';
  const looks = `() => {
    const box = [...document.querySelectorAll('dialog')]
      .find(d => d.open && d.querySelector('dl'));
    if (!box) return null;
    const css = (e, ...names) => names.map(n => getComputedStyle(e)[n]);
    const dl = box.querySelector('dl');
    return {
      box: css(box, 'width', 'borderRadius', 'padding', 'marginTop'),
      dl: css(dl, 'display', 'rowGap', 'columnGap', 'padding', 'margin'),
      dt: css(box.querySelector('dt'), 'margin'),
      dd: css(box.querySelector('dd'), 'margin', 'fontSize'),
      kbd: css(box.querySelector('kbd'), 'fontSize'),
      rows: [...dl.querySelectorAll('dt')].map(dt =>
        [dt.textContent, dt.nextElementSibling.textContent]),
    }; }`;
  const open = () =>
    s.js(`() => [...document.querySelectorAll('dialog')]
      .some(d => d.open && d.querySelector('dl'))`);
  await s.load(work);
  if (s.touch) {
    const button = `document.querySelector('main header button')`;
    check(
      (await s.js(`() => ${button}?.textContent`)) === '? keys',
      `${tag} help: atop the tree, a button labelled ? keys`,
    );
    await s.hit(button);
    await sleep(30);
    const fits = await s.js(`() => (r => r && r.left >= 0
      && r.right <= innerWidth && document.scrollingElement.scrollWidth
      <= innerWidth)([...document.querySelectorAll('dialog')]
      .find(d => d.open)?.getBoundingClientRect())`);
    check(
      (await open()) && fits,
      `${tag} help: a tap on it opens the keys, and they fit`,
    );
    const said = (await s.js(looks))?.rows.find(([n]) => n === 'tap')?.[1];
    check(
      said === tap,
      `${tag} help: they say what a tap does: a file opens`,
      json(said),
    );
    await shot(`${tag}-help`);
    await poke([8, 836], true);
    await sleep(150);
    check(!(await open()), `${tag} help: a tap outside closes it`);
  } else {
    await s.press('?');
    const got = await s.js(looks);
    const names = got?.rows.map(([n]) => n) ?? [];
    check(
      [...every, ...more].every(n => names.includes(n)) &&
        got.rows.some(
          ([n, what]) =>
            n === 'l →' && what === 'Unfold, or step in; a file opens',
        ) &&
        got.rows.some(([n, what]) => n === 'tap' && what === tap) &&
        !(await s.js(`() => !!document.querySelector('dialog pre')`)),
      `${tag} help: ? lists one table as a dl: every key, a file's way too`,
      json(names),
    );
    await shot(`${tag}-help`);
    // page.html's own ? box, the same width of screen
    const theirs = await beside(`${medium}/page.html`, s.width);
    await press('?', theirs.id);
    const want = await c.evaluate(`(${looks})()`, false, theirs.id);
    await shot(`${tag}-help-page`, theirs.id);
    await theirs.close();
    const shape = o => o && json([o.box, o.dl, o.dt, o.dd, o.kbd]);
    // and the find box as page.html's box: its width and corners
    const box = `() => (d => d && [getComputedStyle(d).width,
      getComputedStyle(d).borderRadius])(
      document.querySelector('dialog[open]'))`;
    await s.press('?');
    await s.press('/');
    const finding = await s.js(box);
    await s.press('Escape');
    const again = await beside(`${medium}/page.html`, s.width);
    await press('/', again.id);
    const found = await c.evaluate(`(${box})()`, false, again.id);
    await again.close();
    check(
      json(finding) === json(found),
      `${tag} find: its box as page.html's: rounded, as wide`,
      `${json(finding)} vs ${json(found)}`,
    );
    await s.press('?');
    check(
      shape(got) === shape(want),
      `${tag} help: its box, dl and kbd as page.html's own ? box has them`,
      `${shape(got)} vs ${shape(want)}`,
    );
    const at = await s.cursor();
    await s.press('j');
    check(
      (await s.cursor()) === at && (await open()),
      `${tag} help: the keys do nothing behind it`,
    );
    await s.press('?');
    check(!(await open()), `${tag} help: ? closes it`);
    await s.press('?');
    await s.press('Escape');
    check(
      !(await open()) && (await s.hash()) === '#work',
      `${tag} help: esc closes it, and nothing else`,
    );
    await s.open(work.replace('#work', '#work/pipeline.html'));
    const frame = f =>
      s.js(`() => (${f})(document.querySelector('aside iframe')
        .contentDocument)`);
    await until(() => frame(`d => !!d?.querySelector('details.focus')`));
    await s.press('Tab');
    await s.press('?');
    check(
      (await frame(`d => [...d.querySelectorAll('dialog')]
        .some(d => d.open)`)) && !(await open()),
      `${tag} help: in a page's frame, ? is the page's own`,
    );
    await s.press('Escape');
    await s.press('Escape');
    await s.open(work.replace('#work', `#${flow}`));
    await s.press('Tab');
    await s.press('?');
    check(
      await open(),
      `${tag} help: in an outline, ? is this page's: its keys are here`,
    );
    await s.press('Escape');
  }
  check(!s.errors.length, `${tag} help: no page errors`, s.errors.join(' | '));
}

// the user's answers, step 8
async function answers(s, tag) {
  // hover: a pointer's tints a row; a finger's tap leaves no tint
  await s.load(`${shelf}#math`);
  await c.evaluate(`document.querySelector('main li.stage > .row').click()`);
  await until(() =>
    s.js(
      `() => !!document.querySelector('main li.stage > ul li.entry .tick')`,
    ),
  );
  const row = `document.querySelector('main li.stage > ul li.entry > .row')`;
  if (s.touch) {
    await s.hit(`${row}.querySelector('.tick')`);
    await s.hit(`${row}.querySelector('.tick')`);
  } else {
    const [x, y] = await s.js(`() => (r => [r.x + r.width / 2,
      r.y + r.height / 2])(${row}.getBoundingClientRect())`);
    await c.send('Input.dispatchMouseEvent', { type: 'mouseMoved', x, y });
  }
  await sleep(150);
  const tinted = await s.js(`() => (r => !r.classList.contains('focus')
    && getComputedStyle(r).backgroundColor !== 'rgba(0, 0, 0, 0)')(${row})`);
  check(
    s.touch ? !tinted : tinted,
    s.touch
      ? `${tag} hover: a tap leaves no tint behind`
      : `${tag} hover: a pointer over a row tints it`,
  );

  // a subject's # title atop the tree, while it is the place
  const title = readFileSync(`${home}/Data/cs/path.md`, 'utf8').match(
    /^# (.*)/m,
  )[1];
  await s.load(`${shelf}#cs`);
  const shown = await s.js(
    `() => document.querySelector('main header p')?.textContent`,
  );
  await s.load(work);
  check(
    shown === title &&
      !(await s.js(`() => !!document.querySelector('main header p')`)),
    `${tag} title: a subject's # title atop the tree; none for a folder`,
    json(shown),
  );

  // weights: a folder's name medium, a file's regular, both in the ink
  const weights = await s.js(`() => {
    const i = document.body.appendChild(document.createElement('i'));
    i.style.color = 'var(--ink)';
    const ink = getComputedStyle(i).color;
    i.remove();
    return ['li.folder', 'li.file:not(.folder)'].map(li => (t => [t.fontWeight,
      t.color === ink])(getComputedStyle(document.querySelector('main ' + li +
      ' > .row > .text'))));
  }`);
  check(
    json(weights) ===
      json([
        ['500', true],
        ['400', true],
      ]),
    `${tag} weights: a folder's name medium, a file's regular, ` +
      'both in the ink',
    json(weights),
  );

  // a ✓ in an open file's bar ticks it
  await s.open(work.replace('#work', `#${flow}`));
  const pressed = () =>
    s.js(`() => document.querySelector('aside nav [data-key=x]')
      ?.getAttribute('aria-pressed')`);
  const tap = async () => {
    await s.hit(`document.querySelector('aside nav [data-key=x]')`);
    await sleep(30);
  };
  const up = await pressed();
  await tap();
  const ticked = (await s.ticked(flow)) && (await pressed()) === 'true';
  await tap();
  check(
    up === 'false' &&
      ticked &&
      !(await s.ticked(flow)) &&
      (await pressed()) === 'false',
    `${tag} tick: a ✓ in an open file's bar ticks it, and again unticks it`,
    json({ up, ticked }),
  );

  // a reload with a file open keeps the place, and the tree. the way to
  // flow may be open already, kept from before: it is opened where shut
  await s.load(work);
  for (const folder of ['work/ssv', 'work/ssv/poc'])
    if (!(await s.unfolded(folder))) await s.tap(folder);
  await s.tap(flow);
  await until(s.viewing);
  await navigate();
  await until(async () => (await s.viewing()) && (await s.cursor()) === flow);
  const place = await s.js(
    `() => document.querySelector('main header h1 b')?.textContent`,
  );
  check(
    place === 'work' &&
      (await s.unfolded('work/ssv')) &&
      (await s.unfolded('work/ssv/poc')) &&
      (await s.hash()) === `#${flow}`,
    `${tag} place: a reload with a file open keeps the place, and the tree`,
    json(place),
  );
  check(
    !s.errors.length,
    `${tag} answers: no page errors`,
    s.errors.join(' | '),
  );
}

// cs/path.md, one path an entry: each has its title, the second path a
// link in its note
async function pathed(s, tag) {
  const md = readFileSync(`${home}/Data/cs/path.md`, 'utf8');
  const plus = md.match(/^\s*(\d+\.|-) .*`[^`]+` \+ `.*$/gm);
  await s.load(`${shelf}#cs`);
  await c.evaluate(
    `document.querySelectorAll('main li.stage > .row')[1].click()`,
  );
  await until(() =>
    s.js(`() => !!document.querySelector('main li.stage > ul li.entry')`),
  );
  const sicp = await s.js(`() => (i => i && [i.name, !!i.closest('.row')
    .querySelector('.note a')])(document.querySelector('main li.stage > ul ' +
    'input[name^="Structure and Interpretation of Computer Programs"]'))`);
  check(
    !plus && sicp?.[1],
    `${tag} path: cs/path.md one path an entry: its title read, ` +
      'the second a link',
    json({ plus, sicp }),
  );
}

// a file's kind, step 9: its extension as a tag right after its name, html
// py md js pdf each in its hue on a 12% tint of it, any other muted, a leaf
// and a folder none. the name stays whole in the dom; a long one is cut
// before the tag; the edge keeps a subject's t/n and a lecture's clock
async function kinds(s, tag) {
  const lecture =
    'cs/aiml/introduction-to-neural-networks-and-backpropagation-' +
    'building-micrograd-andrej-karpathy.mp4';
  // the rows in sight: the kind each should show, the tag it does and in
  // what paint, and where its name, tag, side and edge are
  const rows = () =>
    s.js(`() => {
      const hued = ['html', 'py', 'md', 'js', 'pdf'];
      const paint = hue => {
        const i = document.body.appendChild(document.createElement('i'));
        i.style.color = hue;
        i.style.background =
          'color-mix(in oklab, ' + hue + ' 12%, transparent)';
        const { color, backgroundColor } = getComputedStyle(i);
        i.remove();
        return color + ' on ' + backgroundColor;
      };
      const box = e => e && (b => ({ left: b.left, right: b.right,
        middle: (b.top + b.bottom) / 2 }))(e.getBoundingClientRect());
      return [...document.querySelectorAll('main li.file > .row')]
        .filter(row => row.checkVisibility())
        .map(row => {
          const li = row.parentElement, name = li.node.path.split('/').at(-1);
          const folder = li.matches('.folder');
          const ext = !folder && name.includes('.')
            ? name.split('.').at(-1).toLowerCase() : null;
          const text = row.querySelector(':scope > .text');
          const tag = row.querySelector(':scope > .kind');
          const side = row.querySelector(':scope > small');
          const k = tag && getComputedStyle(tag), r = getComputedStyle(row);
          return {
            path: li.node.path, folder, ext,
            kind: tag?.dataset.kind ?? null,
            painted: !tag || k.color + ' on ' + k.backgroundColor ===
              paint(hued.includes(ext) ? 'var(--' + ext + ')'
                : 'var(--muted)'),
            text: box(text), tag: box(tag), side: box(side),
            clock: side?.textContent,
            cut: text.scrollWidth > text.clientWidth,
            whole: !tag || tag.scrollWidth <= tag.clientWidth,
            edge: row.getBoundingClientRect().right -
              parseFloat(r.paddingRight) - parseFloat(r.columnGap),
          };
        });
    }`);

  // the library's root: subjects with their t/n, md txt html files, a leaf
  await s.load(shelf);
  const root = await rows();
  // a t/n is a row's ::after, which has no box in the dom: cdp has it, and
  // a span set in its font has its width
  const { root: doc } = await c.send('DOM.getDocument', { depth: -1 });
  const counted = [];
  const walk = n => {
    const after = n.pseudoElements?.find(p => p.pseudoType === 'after');
    if (n.attributes?.includes('data-count') && after) counted.push(after);
    n.children?.forEach(walk);
  };
  walk(doc);
  const counts = [];
  for (const after of counted) {
    const { model } = await c.send('DOM.getBoxModel', {
      nodeId: after.nodeId,
    });
    counts.push({ left: model.content[0], right: model.content[2] });
  }
  const sized = await s.js(`() => [...document.querySelectorAll(
    'main li.file > .row[data-count]')].map(row => {
      const after = getComputedStyle(row, '::after'),
        r = getComputedStyle(row);
      const i = document.body.appendChild(document.createElement('span'));
      for (const p of
        ['fontFamily', 'fontSize', 'fontWeight', 'letterSpacing'])
        i.style[p] = after[p];
      // out of the body's grid, so as wide as its text
      i.style.position = 'fixed';
      i.textContent = row.dataset.count;
      const width = i.getBoundingClientRect().width;
      i.remove();
      return { width, edge: row.getBoundingClientRect().right -
        parseFloat(r.paddingRight) - parseFloat(r.columnGap) };
    })`);

  // cs/aiml: long names, and a lecture left at 10:00
  await s.js(`p => localStorage.setItem('shelf:at:' + p, '600')`, lecture);
  await s.load(`${shelf}#cs/aiml`);
  const aiml = await rows();

  // work/ssv/poc: py js html pdf, toml txt, leaves and folders
  const model = 'work/ssv/poc/model.py';
  await s.load(work.replace('#work', '#work/ssv/poc'));
  const poc = await rows();

  const all = [...root, ...poc];
  const kinds = new Set(all.map(r => r.kind));
  const wrong = all.filter(r => r.kind !== r.ext || !r.painted);
  check(
    !wrong.length &&
      ['html', 'py', 'md', 'js', 'pdf', 'toml'].every(k => kinds.has(k)) &&
      all.some(r => r.folder) &&
      all.some(r => !r.folder && !r.ext),
    `${tag} kind: a file's tag is its extension, in its hue or muted; ` +
      'a leaf and a folder show none',
    json(wrong.map(r => [r.path, r.ext, r.kind, r.painted])),
  );
  const tagged = poc.filter(r => r.tag && !r.cut);
  const off = tagged.filter(
    r =>
      r.tag.left < r.text.right ||
      r.tag.left > r.text.right + 12 ||
      Math.abs(r.tag.middle - r.text.middle) > 2,
  );
  check(
    tagged.length && !off.length,
    `${tag} kind: the tag sits right after the name, on its line`,
    json(off.map(r => [r.path, r.text, r.tag])),
  );
  const look = await s.js(
    `p => (k => k && [k.fontSize, k.fontWeight, k.fontFamily.includes('Mono'),
      k.textTransform, k.borderRadius, k.fontStyle])((t => t &&
      getComputedStyle(t))([...document.querySelectorAll('main li.file')]
      .find(li => li.node.path === p)?.querySelector('.kind')))`,
    model,
  );
  check(
    json(look) === json(['9.5px', '600', true, 'uppercase', '4px', 'normal']),
    `${tag} kind: the tag is exts.html's: 9.5px mono, 600, ` +
      'upper case, radius 4px',
    json(look),
  );

  // the name whole in the dom, its extension not shown, the tag no text;
  // x ticks the file under its path, the ✓ at the edge, the tag in place
  const before = await s.js(
    `p => (row => row && { dom: row.querySelector('.text').textContent,
      shown: row.querySelector('.text').innerText,
      nodes: row.querySelector('.kind')?.childNodes.length,
      drawn: (t => t && getComputedStyle(t, '::before').content)(row
        .querySelector('.kind')) })([...document.querySelectorAll(
      'main li.file')].find(li => li.node.path === p)?.firstChild)`,
    model,
  );
  await s.walk(model);
  await s.press('x');
  const m = await s.mark(model);
  const after = (await rows()).find(r => r.path === model);
  check(
    before?.dom === 'model.py' &&
      before.shown === 'model' &&
      before.nodes === 0 &&
      ![undefined, null, 'none'].includes(before.drawn) &&
      (await s.ticked(model)) &&
      m.shown &&
      m.end === 10 &&
      after?.tag?.left >= after.text.right,
    `${tag} kind: the name stays whole for find and the tick; ` +
      'the tag is drawn',
    json({ before, m }),
  );

  // a long name is cut before its tag, which shows whole
  const long = aiml.filter(r => r.tag && r.cut);
  const hid = long.filter(
    r => !r.whole || r.tag.left < r.text.right || r.tag.right > r.edge,
  );
  check(
    long.length && !hid.length,
    `${tag} kind: a long name is cut before its tag, which shows whole`,
    json(hid.map(r => [r.path, r.text, r.tag])),
  );

  // the edge keeps a subject's t/n and a lecture's clock, clear of the tag
  const near = (a, b) => Math.abs(a - b) < 1;
  const stray = counts.filter(
    (c, i) =>
      !near(c.right, sized[i].edge) || !near(c.right - c.left, sized[i].width),
  );
  const lectured = aiml.find(r => r.path === lecture);
  check(
    counts.length === sized.length &&
      counts.length &&
      !stray.length &&
      lectured?.clock === '10:00' &&
      near(lectured.side.right, lectured.edge) &&
      (!lectured.tag || lectured.side.left >= lectured.tag.right),
    `${tag} kind: the edge keeps a subject's t/n and a lecture's clock, ` +
      'clear of the tag',
    json({ stray, lecture: lectured && [lectured.tag, lectured.side] }),
  );
  check(!s.errors.length, `${tag} kind: no page errors`, s.errors.join(' | '));
}

// a folder's card, under the cursor: math/elem's, longer than the aside,
// scrolls in it, its top in view
async function card(s, tag) {
  await s.load(`${shelf}#math`);
  const rows = `[...document.querySelectorAll('li.stage li.entry > .row')]`;
  await c.evaluate(`document.querySelector('main li.stage > .row')
    .click()`);
  // each row is typed as its lookup comes back, in no set order (a titled
  // folder is listed too, so math/elem comes after bare ones below it):
  // the first folder is known once the stage's lookups are done
  await until(() =>
    s.js(`() => document.querySelector('main li.stage')
      .node.marked?.then(() => true)`),
  );
  // under the cursor by keys, as a tap opens it; done once its card shows
  const folder = await s.js(`() => ${rows}
    .find(r => r.dataset.type === 'folder').parentElement.node.path`);
  await s.walk(folder);
  await until(async () => (await s.fits()).path === folder);
  const f = await s.fits();
  check(
    f.view <= f.aside && f.under && f.scrolls,
    `${tag} frame: previewed, a long card scrolls in the aside, ` +
      'its top in view',
    JSON.stringify(f),
  );
  await shot(`${tag}-card`);
  check(!s.errors.length, `${tag} card: no page errors`, s.errors.join(' | '));
}

// a folder on the right, step 11: its card, its count over its tree drawn
// as the tree draws it. two levels at first: a folder in a folder folded,
// as is one of over 30 rows or named with a dot, a folded one's count
// beside it. its keys are an outline's; a file in it opens in the tree, so
// the right opens no right of its own. a subject keeps its summary.
// step 16: only a folder a path.md names offers the file to resume, as a
// line; its title is a heading. where no preview fits, a folder unfolds in
// place and the right stays shut. the oracle is the disk
async function folders(s, tag) {
  const ssv = 'work/ssv';
  const collate = new Intl.Collator(undefined, { numeric: true }).compare;
  const disk = dir =>
    readdirSync(`${home}/scratch/${dir}`, { withFileTypes: true }).sort(
      (a, b) => collate(a.name, b.name),
    );
  const shut = () =>
    s.js(`() => !document.querySelector('aside').checkVisibility()`);
  await s.load(work);
  if (s.width < 800) {
    await s.tap(ssv);
    check(
      (await until(() => s.unfolded(ssv))) && (await shut()),
      `${tag} folder: no preview fits, so a tap unfolds a folder in place, ` +
        'the right shut',
    );
    await shot(`${tag}-folder`);
    check(
      !s.errors.length,
      `${tag} folder: no page errors`,
      s.errors.join(' | '),
    );
    return;
  }
  // the card's rows in sight: each one's path, depth, and open, folded
  // with its count as drawn, or a file
  const right = `p => [...document.querySelectorAll('aside details')]
    .find(d => d.node.path === p)`;
  const rows = () =>
    s.js(`() => [...document.querySelectorAll('aside details')]
      .filter(d => d.firstChild.checkVisibility())
      .map(d => {
        let depth = 0;
        for (let up = d; (up = up.parentElement.closest('details')); depth++);
        const count = getComputedStyle(d.firstChild, '::after').content;
        return [d.node.path, depth,
          d.open ? 'open' : count === 'none' ? 'file' : +JSON.parse(count)];
      })`);
  const card = () =>
    s.js(`() => (v => v.matches('.card') && v.path)(
      document.querySelector('aside > :last-child'))`);
  await (s.touch ? s.tap(ssv) : s.walk(ssv));
  check(
    await until(async () => (await card()) === ssv),
    `${tag} folder: the right shows the folder under the cursor, its card`,
  );
  // no path.md names ssv, so its card offers no file
  const files = disk(ssv).filter(e => !e.isDirectory());
  const head = await s.js(`() => [...document.querySelectorAll(
    'aside .kicker, aside .go')].map(e => e.textContent)`);
  check(
    json(head) === json([`${files.length} files`]),
    `${tag} folder: its head counts its files; a plain folder's has no go`,
    json(head),
  );
  // the disk's shape: ssv's rows, and under each folder of 30 rows or
  // fewer, its name not a dot one, its rows, each folder there folded
  // with its count
  const shape = [];
  for (const e of disk(ssv)) {
    const p = `${ssv}/${e.name}`;
    if (!e.isDirectory()) {
      shape.push([p, 0, 'file']);
      continue;
    }
    const kids = disk(p),
      open = kids.length <= 30 && !e.name.startsWith('.');
    shape.push([p, 0, open ? 'open' : kids.length]);
    if (open)
      for (const k of kids) {
        const q = `${p}/${k.name}`;
        shape.push([q, 1, k.isDirectory() ? disk(q).length : 'file']);
      }
  }
  const seen = await rows();
  check(
    json(seen) === json(shape),
    `${tag} folder: two levels show at first, what is on disk, in order`,
    json(seen.filter((r, i) => json(r) !== json(shape[i])).slice(0, 3)),
  );
  const issues = `${ssv}/issues`,
    demo = `${ssv}/poc/demo`;
  check(
    json(seen.find(r => r[0] === issues)) ===
      json([issues, 0, disk(issues).length]) && disk(issues).length > 30,
    `${tag} folder: a folder of over 30 rows starts folded, ` +
      'its count beside it',
  );
  check(
    json(seen.find(r => r[0] === demo)) === json([demo, 1, disk(demo).length]),
    `${tag} folder: a folder in a folder starts folded, its count beside it`,
  );
  // its rows as the tree's: the left unfolds ssv, and a folder's, a file's
  // and its tag's paint, and a row's height, are the same either side; so
  // is the guide an open folder's rows hang on, and how far in they sit
  if (!(await s.unfolded(ssv))) await s.press('l');
  await until(() => s.unfolded(ssv));
  const drawn = p =>
    s.js(
      `p => [(${s.li})(p), (${right})(p)].map(at => {
        if (!at) return null;
        const row = at.firstChild, look = e => e && (t => [t.fontSize,
          t.fontWeight, t.color, t.backgroundColor, t.fontFamily])(
          getComputedStyle(e));
        return [look(row.querySelector('.text')), look(row.querySelector(
          '.kind')), Math.round(row.getBoundingClientRect().height)];
      })`,
      p,
    );
  for (const p of [`${ssv}/codebase`, `${ssv}/clicbank.html`]) {
    const [left, here] = await drawn(p);
    check(
      json(left) === json(here),
      `${tag} folder: ${p.split('/').at(-1)} drawn as the tree draws it`,
      json([left, here]),
    );
  }
  const hung = await s.js(
    `([ssv, codebase]) => [[(${s.li})(ssv), 'ul'], [(${right})(codebase),
      'div']].map(([at, under]) => {
      if (!at) return null;
      const guide = getComputedStyle(at.querySelector(':scope > ' + under));
      const x = row => row.querySelector('.text').getBoundingClientRect().x;
      return [guide.borderInlineStart, Math.round(x(at.querySelector(
        ':scope > ' + under + ' .row')) - x(at.firstChild))];
    })`,
    [ssv, `${ssv}/codebase`],
  );
  check(
    json(hung[0]) === json(hung[1]),
    `${tag} folder: an open folder's rows hang on the tree's guide, as far in`,
    json(hung),
  );
  await shot(`${tag}-folder`);
  const keys = () =>
    s.js(`() => document.querySelector('aside').classList.contains('keys')`);
  const model = `${ssv}/poc/model.py`;
  if (!s.touch) {
    // the keys: tab, j to issues, l and h; then a file, and enter
    const at = () =>
      s.js(`() => document.querySelector('aside details.focus')?.node.path`);
    const to = async p => {
      for (let i = 0; i < 80 && (await at()) !== p; i++) await s.press('j');
      return (await at()) === p;
    };
    await s.press('Tab');
    check(
      (await keys()) && (await at()) === shape[0][0],
      `${tag} folder: tab hands the card the keys, on its first row`,
    );
    check(await to(issues), `${tag} folder: j walks the card's rows`);
    await s.press('l');
    check(
      await until(async () => {
        const under = (await rows()).filter(r =>
          r[0].startsWith(`${issues}/`),
        );
        return (
          json((await rows()).find(r => r[0] === issues)) ===
            json([issues, 0, 'open']) && under.length === disk(issues).length
        );
      }),
      `${tag} folder: l unfolds a folded folder, its rows read then, its ` +
        'count gone',
    );
    await s.press('h');
    check(
      json((await rows()).find(r => r[0] === issues)) ===
        json([issues, 0, disk(issues).length]),
      `${tag} folder: h folds it, its count back`,
    );
    await to(model);
    await sleep(400);
    check(
      (await at()) === model && (await card()) === ssv,
      `${tag} folder: a file under the card's cursor leaves the card: the ` +
        'right opens no right of its own',
    );
    await s.press('Enter');
    check(
      await until(
        async () =>
          (await s.viewing()) &&
          (await s.hash()) === `#${model}` &&
          (await s.cursor()) === model &&
          (await s.unfolded(`${ssv}/poc`)) &&
          (await s.shown()) === model,
      ),
      `${tag} folder: enter on a file in it opens it in the tree, the ` +
        'cursor there, its folders unfolded',
    );
    check(!(await keys()), `${tag} folder: the keys are the tree's again`);
    await s.press('Escape');
    await until(async () => !(await s.viewing()));
  }
  // a tap on a file in the card opens it in the tree
  await s.tap(ssv);
  await until(async () => (await card()) === ssv);
  const todo = `${ssv}/todo.html`;
  await s.hit(`(${right})(${json(todo)})?.firstChild`);
  check(
    await until(
      async () =>
        (await s.viewing()) &&
        (await s.cursor()) === todo &&
        (await s.shown()) === todo,
    ),
    `${tag} folder: a tap on a file in it opens it in the tree`,
  );
  // a subject keeps its summary: its stages, as path.md has them
  await s.load(shelf);
  await (s.touch ? s.tap('cs') : s.walk('cs'));
  const stages = readFileSync(`${home}/Data/cs/path.md`, 'utf8').match(
    /^## /gm,
  ).length;
  check(
    await until(
      async () =>
        (await s.js(`() => (v => v.localName === 'section' && v.path === 'cs'
          && v.querySelectorAll('.lines > li').length)(
          document.querySelector('aside > :last-child'))`)) === stages,
    ),
    `${tag} folder: a subject keeps its summary, its stages`,
  );
  // step 16: a folder a path.md names is reading. its card offers the file
  // left part way, else its first, as a line of text in the accent: no
  // fill, one line high. a folder none names offers none, though it sits
  // in a subject's folder. the oracle: each entry's paths, a list item
  // under a stage, from its subject's folder; the lecture left at 1:40
  const named = new Set(
    subjects.flatMap(({ dir, md }) =>
      (md.slice(md.search(/^## /m)).match(/^(?:\d+\.|-) .*$/gm) ?? [])
        .flatMap(item => [...item.matchAll(/`([^`]+)`/g)])
        .map(m => posix.join(dir, m[1]).replace(/\/$/, '')),
    ),
  );
  const calc = 'math/calc',
    course = `${calc}/the-essence-of-calculus-grant-sanderson`,
    lecture = '05-derivatives-of-exponentials';
  const dirs = (await ls(calc))
    .map(n => `${calc}/${n}`)
    .filter(p => statSync(`${library}/${p}`).isDirectory());
  const start = dirs.find(p => named.has(p) && p !== course),
    plain = dirs.find(p => !named.has(p));
  const first = (await ls(start)).find(
    n => !n.startsWith('.') && statSync(`${library}/${start}/${n}`).isFile(),
  );
  await s.load(`${shelf}#${calc}`);
  await c.evaluate(
    `localStorage.setItem('shelf:at:${course}/${lecture}.mp4', '100'); 1`,
  );
  // a folder's go: its text, fill, lines and whether it is the accent
  const go = async p => {
    await s.tap(p);
    await until(async () => (await card()) === p);
    return s.js(`() => {
      const a = document.querySelector('aside .go');
      const i = document.body.appendChild(document.createElement('i'));
      i.style.color = 'var(--accent)';
      const accent = getComputedStyle(i).color;
      i.remove();
      if (!a) return null;
      const css = getComputedStyle(a);
      return [a.textContent, css.backgroundColor,
        Math.round(a.offsetHeight / parseFloat(css.lineHeight)),
        css.color === accent]; }`);
  };
  const gone = await go(plain);
  check(
    gone === null,
    `${tag} folder: a folder no path.md names has no go, in a subject too`,
    `${plain}: ${json(gone)}`,
  );
  const line = text => [text, 'rgba(0, 0, 0, 0)', 1, true];
  const offered = [await go(start), await go(course)];
  check(
    json(offered) ===
      json([
        line(`Start → ${first.replace(/\.\w+$/, '')}`),
        line(`Resume → ${lecture} · 1:40`),
      ]),
    `${tag} folder: a folder a path.md names has its go, a line of text ` +
      'in the accent, no block: its first file, or the one left part way',
    json(offered),
  );
  // its title a heading, a page's h1 in size, as page.html sets it
  const h1 = readFileSync(
    new URL('../page.html', import.meta.url),
    'utf8',
  ).match(/^ {4}h1 \{[^}]*?font-size: ([\d.]+)rem/m)[1];
  check(
    (await s.js(`() => getComputedStyle(
      document.querySelector('aside .card h1')).fontSize`)) === `${h1 * 16}px`,
    `${tag} folder: a card's title is a heading, a page's h1 in size`,
  );
  // the open question: a folder's hash roots the tree there
  await s.load(`${work}/ssv`);
  check(
    json(
      await s.js(`() => [...document.querySelectorAll('main > ul > li')]
        .map(li => li.node.path)`),
    ) === json(shape.filter(r => r[1] === 0).map(r => r[0])),
    `${tag} folder: a folder's hash roots the tree there`,
  );
  check(
    !s.errors.length,
    `${tag} folder: no page errors`,
    s.errors.join(' | '),
  );
}

// source view, step 14: run.mjs and cdp.mjs as js, the page's own style
// as css, and the page itself, flipped by u. the oracle is each file's
// lines: a full-line mark in its comment, and under it, at its indent, a
// css rule's selectors to its {, or a js function's signature
async function sources(s, tag) {
  const med = `http://127.0.0.1:${medium.address().port}/shelf.html`;
  const marks = /^\s*(# |\/\/ |\/\* |<!-- )-- [a-z]+ --( \*\/| -->)?$/;
  const lines = file => readFileSync(file, 'utf8').split('\n');
  const named = line => line.match(/-- ([a-z]+) --/)[1];
  // the outline's nodes under a heading, or its top: each heading's first
  // line, as written. no such heading has none: the check then fails
  const under = title =>
    s.js(
      `t => (v => (at => at && [...at.querySelectorAll(':scope > details')]
        .map(d => d.firstChild.textContent.split('\\n')[0].trim()))(
        t === null ? v : [...v.querySelectorAll('details')]
          .find(d => d.firstChild.textContent.trim() === t)?.lastChild))(
        document.querySelector('aside > :last-child'))`,
      title,
    );
  // a section's headings by the oracle: from its mark to the next, the
  // lines at the indent of the code after it that open a node, less the
  // brace that opens it
  const heads = (all, name, opens) => {
    const from = all.findIndex(l => marks.test(l) && named(l) === name) + 1;
    const to = all.findIndex((l, i) => i >= from && marks.test(l));
    const body = all.slice(from, to < 0 ? undefined : to);
    const at = body.find(l => l.trim()).search(/\S/);
    return body
      .filter(l => l.search(/\S/) === at && opens.test(l))
      .map(l => l.trim().replace(/\s*\{$/, ''));
  };
  if (!s.touch) {
    const run = new URL('run.mjs', import.meta.url).pathname;
    await s.open(`${med}#shelf/run.mjs`);
    const js = lines(run);
    check(
      json(await under(null)) === json([js[0], 'steps']) &&
        json(await under('steps')) ===
          json(heads(js, 'steps', /^(async )?function\b/)),
      `${tag} source: js opens as an outline, a mark heading its functions`,
    );
    await s.open(`${med}#shelf/cdp.mjs`);
    const cdp = lines(new URL('cdp.mjs', import.meta.url).pathname);
    check(
      json(await under(null)) ===
        json([
          cdp[0],
          ...cdp
            .filter(l => /^export async function/.test(l))
            .map(l => l.replace(/\s*\{$/, '')),
        ]),
      `${tag} source: a js with no marks, by its functions alone`,
    );
    // css: the page's own style, as a file of its own
    const css = lines(page).slice(
      lines(page).findIndex(l => l.trim() === '<style>') + 1,
      lines(page).findIndex(l => l.trim() === '</style>'),
    );
    const dir = `${tmpdir()}/shelfsource`;
    await mkdir(dir, { recursive: true });
    await writeFile(`${dir}/style.css`, css.join('\n'));
    const styled = await serve(dir);
    await s.open(
      `http://127.0.0.1:${styled.address().port}/shelf.html#style.css`,
    );
    const csstop = css.filter(l => marks.test(l)).map(named);
    const rules = await under('style');
    check(
      json(await under(null)) === json(csstop) &&
        json(rules) ===
          json(
            heads(css, 'style', /[{,]$/).filter(
              (l, i, all) => !/,$/.test(all[i - 1] ?? ''),
            ),
          ),
      `${tag} source: css opens as an outline, its rules under its marks`,
      json(rules?.slice(0, 4)),
    );
  }
  // the page: itself first, then its source, and back, its frame kept
  await s.open(`${med}#shelf.html`);
  const view = () =>
    s.js(`() => (v => v.matches('iframe') ? 'page' : v.className)(
      document.querySelector('aside > :last-child'))`);
  const flip = s.touch
    ? to =>
        s.hit(`[...document.querySelectorAll('aside nav [data-key=u]')]
        .find(b => b.value === ${json(to)})`)
    : () => s.press('u');
  check((await view()) === 'page', `${tag} source: a page opens as the page`);
  await s.js(`() => document.querySelector('aside iframe').contentWindow
    .kept = 1`);
  await flip('source');
  check(
    await until(async () => (await view()) === 'outline html'),
    s.touch
      ? `${tag} source: source in its bar flips it for a finger`
      : `${tag} source: u flips a page to its source`,
  );
  const pressed = () =>
    s.js(`() => [...document.querySelectorAll('aside nav [data-key=u]')]
      .map(b => b.value + ' ' + b.ariaPressed).join()`);
  check(
    (await pressed()) === 'page false,source true',
    `${tag} source: page | source in the bar says which shows`,
    await pressed(),
  );
  if (!s.touch) {
    const all = lines(page);
    const top = all.filter(l => marks.test(l)).map(named);
    check(
      json(await under(null)) === json(top),
      `${tag} source: a page's source, its marks at the top, as the map ` +
        'has them',
    );
    check(
      json(await under('tokens')) === json(heads(all, 'tokens', /\{$/)) &&
        json(await under('engine')) ===
          json(heads(all, 'engine', /^\s*(async )?function\b/)),
      `${tag} source: under a css mark its rules, under a js mark its ` +
        'functions',
    );
  }
  await flip('page');
  check(
    await until(
      async () =>
        (await view()) === 'page' &&
        (await s.js(`() => document.querySelector('aside > :last-child')
          .contentWindow.kept`)) === 1,
    ),
    `${tag} source: and back to the page, its frame as it was`,
  );
  // deck.html's source as step 14's sketch has it: each mark by its bare
  // name, no doctype, no lone closer, no brace on a heading
  const deck = lines(new URL('../deck.html', import.meta.url).pathname);
  await s.open(`${med}#deck.html`);
  await flip('source');
  await until(async () => (await view()) === 'outline html');
  check(
    json(await under(null)) ===
      json(deck.filter(l => marks.test(l)).map(named)),
    `${tag} source: a mark's node reads its bare name`,
    json(await under(null)),
  );
  check(
    await s.js(`() => {
      const shown = q => [...document.querySelectorAll('aside .outline ' + q)]
        .map(e => e.textContent.trim());
      return !shown('summary, pre').some(t =>
        /^(}|<\\/(style|script)>|<!doctype html>)$/i.test(t))
        && !shown('summary').some(t => /\\{$/.test(t)); }`),
    `${tag} source: no doctype, no lone closer, no brace on a heading`,
  );
  check(
    !s.errors.length,
    `${tag} source: no page errors`,
    s.errors.join(' | '),
  );
}

// the file's bar, step 17, in place of 14c's strip: one bar, nothing
// between it and the text. a leaf's, md's, python's and a page's source's
// first line sits under it at the view's top padding, and is there again
// after a scroll and a collapse; a scroll moves nothing but the text. each
// tool an icon, named with its key, its icon beside that key in ?; on a
// phone too the tools sit on one line, none cut
async function scrolled(s, tag) {
  const med = `http://127.0.0.1:${medium.address().port}/shelf.html`;
  const views = [
    `${work}/ssv/poc/flow`,
    `${work}/ssv/integration/diff.md`,
    `${work}/ssv/poc/model.py`,
    `${med}#deck.html`,
  ];
  // the first line from the bar, less the view's top padding; whether the
  // view's first element holds it; the bar's top; the line's top
  const where = () =>
    s.js(`() => {
      const nav = document.querySelector('aside > nav')
        .getBoundingClientRect();
      const v = document.querySelector('aside > :last-child');
      const first = v.querySelector('summary');
      const top = first.getBoundingClientRect().top;
      return [Math.round(top - nav.bottom
        - parseFloat(getComputedStyle(v).paddingTop)),
        v.firstElementChild.contains(first), Math.round(nav.top),
        Math.round(top)]; }`);
  // the tools at fault: unnamed, or not beside their key in ?; and the bar
  // on two lines, or cut
  const named = () =>
    s.js(`() => {
      const nav = document.querySelector('aside > nav');
      const tools = [...nav.querySelectorAll('.tool')]
        .filter(t => t.checkVisibility());
      const keys = [...document.querySelectorAll('dialog dt')];
      const d = e => e?.querySelector('path')?.getAttribute('d');
      const odd = tools.filter(t => {
        const dt = keys.find(dt =>
          dt.querySelector('kbd').textContent === t.dataset.key);
        return !(d(t) && t.title === t.ariaLabel
          && t.ariaLabel.endsWith('(' + t.dataset.key + ')')
          && d(dt) === d(t));
      }).map(t => t.dataset.key);
      const lines = new Set(tools.map(t =>
        Math.round(t.getBoundingClientRect().top))).size;
      if (lines !== 1) odd.push(lines + ' lines');
      if (nav.scrollWidth > nav.clientWidth
        || document.scrollingElement.scrollWidth > innerWidth) odd.push('cut');
      return odd; }`);
  const bad = [],
    unnamed = [];
  for (const url of views) {
    await s.open(url);
    if (url.endsWith('.html')) {
      await (s.touch
        ? s.hit(`document.querySelector('aside nav [value=source]')`)
        : s.press('u'));
      await until(() =>
        s.js(`() => document.querySelector('aside > :last-child')
          .matches('.outline')`),
      );
    }
    const name = url.split('#')[1];
    const was = await where();
    await s.hit(`document.querySelector('aside nav [data-key=e]')`);
    await s.js(`() => {
      document.querySelector('aside > :last-child').scrollTop = 400; }`);
    const scrolled = await where();
    await s.hit(`document.querySelector('aside nav [data-key=c]')`);
    const after = await where();
    const held =
      was[0] === 0 && was[1] && scrolled[2] === was[2] && after[3] === was[3];
    if (!held) bad.push(`${name}: ${json([was, scrolled, after])}`);
    const odd = await named();
    if (odd.length) unnamed.push(`${name}: ${odd.join()}`);
  }
  check(
    !bad.length,
    `${tag} bar: nothing between it and the first line, which a scroll ` +
      'and a collapse leave where it was',
    bad.join(' | '),
  );
  check(
    !unnamed.length,
    `${tag} bar: each tool an icon named with its key, beside it in ?, ` +
      'the bar one line',
    unnamed.join(' | '),
  );
}

// where the right's rows start, step 14d: side by side, a folder's card,
// python, md and a page's source start at the aside's left, a row in the
// gutter the tree's rows have; on a phone each keeps the margin it had
async function edges(s, tag) {
  const med = `http://127.0.0.1:${medium.address().port}/shelf.html`;
  const wide = s.width >= 800;
  // the first row's left edge from the aside's, and the tree's from main's
  const edge = () =>
    s.js(`() => {
      const left = q => document.querySelector(q)?.getBoundingClientRect()
        .left;
      return [left('aside > :last-child summary') - left('aside'),
        left('main .row') - left('main')].map(Math.round); }`);
  const got = [];
  for (const url of [
    `${work}/ssv/poc/model.py`,
    `${work}/ssv/integration/diff.md`,
    `${med}#deck.html`,
  ]) {
    await s.open(url);
    if (url.endsWith('.html')) {
      await (s.touch
        ? s.hit(`document.querySelector('aside nav [value=source]')`)
        : s.press('u'));
      await until(() =>
        s.js(`() => document.querySelector('aside > :last-child')
          .matches('.outline')`),
      );
    }
    got.push([url.split('#')[1], ...(await edge())]);
  }
  if (wide) {
    await s.load(work);
    await (s.touch ? s.tap('work/ssv') : s.walk('work/ssv'));
    await until(async () => (await s.shown()) === 'work/ssv');
    got.push(['work/ssv', ...(await edge())]);
  }
  // a phone's outline kept its 1.5rem, 24px
  check(
    got.every(([, row, gutter]) =>
      wide ? row >= 0 && row <= gutter : row === 24,
    ),
    wide
      ? `${tag} edge: the right's rows start at its left, in the tree's gutter`
      : `${tag} edge: on a phone the right keeps its margins`,
    json(got),
  );
}

// a view starts at the top, step 14e: a long folder's card, a short
// one's, a subject's summary, a message and an outline all start at one
// place, under the bar or atop the aside, and a view grown keeps its first
// line where it was. on a phone only what is open shows
async function tops(s, tag) {
  // the view's first line from where its row starts, then with the view
  // grown by a tall block at its end
  const top = () =>
    s.js(`() => {
      const v = document.querySelector('aside > :last-child');
      const nav = document.querySelector('aside > nav');
      const from = nav ? nav.getBoundingClientRect().bottom
        : document.querySelector('aside').getBoundingClientRect().top;
      const head = v.firstElementChild;
      const first = () => head ? head.getBoundingClientRect().top
        : v.getBoundingClientRect().top + parseFloat(
          getComputedStyle(v).paddingTop);
      const was = first();
      const grown = v.appendChild(document.createElement('div'));
      grown.style.blockSize = '3000px';
      const now = first();
      grown.remove();
      return [was - from, now - from].map(Math.round); }`);
  const got = [];
  if (s.width >= 800) {
    await s.load(work);
    // and each part of the card, its bar, head and tree: a short card's
    // sit where a long one's do, not spread down the aside
    const parts = [];
    for (const folder of ['work/ssv', 'work/tests']) {
      await (s.touch ? s.tap(folder) : s.walk(folder));
      await until(async () => (await s.shown()) === folder);
      got.push([folder, ...(await top())]);
      parts.push(
        await s.js(`() => [...document.querySelector('aside > :last-child')
          .children].map(e => Math.round(e.getBoundingClientRect().top))`),
      );
    }
    check(
      json(parts[0]) === json(parts[1]),
      `${tag} top: a short card's parts sit where a long one's do`,
      json(parts),
    );
    await s.load(shelf);
    await (s.touch ? s.tap('cs') : s.walk('cs'));
    await until(async () => (await s.shown()) === 'cs');
    got.push(['cs', ...(await top())]);
  }
  for (const p of ['work/no/such/path', 'work/ssv/poc/model.py']) {
    await s.open(work.replace('#work', `#${p}`));
    got.push([p, ...(await top())]);
  }
  check(
    got.every(([, was, now]) => was === got[0][1] && now === was),
    `${tag} top: a view starts at the top, at one place, long or short`,
    json(got),
  );
}

// a picture, step 18: at the view's top left, not across its middle; a
// player, step 19, where text starts, as an outline's first heading does.
// a video's box is the whole view under the bar, a black stage, its
// picture contained in the middle, side by side and at full width. a
// picture taller than the view, a player, a 4:3 video, and side by side
// that video at full width
async function pictured(s, tag) {
  // from the aside's left and the bar's bottom: where each starts, its box,
  // or an outline's first heading, then the box's top
  const placed = () =>
    s.js(`() => {
      const left = document.querySelector('aside').getBoundingClientRect()
        .left;
      const top = document.querySelector('aside > nav')
        .getBoundingClientRect().bottom;
      const v = document.querySelector('aside > :last-child');
      const r = v.getBoundingClientRect();
      const head = v.querySelector('summary')?.getBoundingClientRect() ?? r;
      return [v.localName, ...[head.left - left, head.top - top, r.top - top]
        .map(Math.round)]; }`);
  // a video: how far its box is in from the view, the aside under the bar,
  // at each edge; its fit, where that puts the picture, and its colour
  const fitted = () =>
    s.js(`() => {
      const a = document.querySelector('aside').getBoundingClientRect(),
        bar = document.querySelector('aside > nav').getBoundingClientRect(),
        v = document.querySelector('aside > video'),
        r = v.getBoundingClientRect(),
        c = getComputedStyle(v);
      return [[r.left - a.left, r.top - bar.bottom, a.right - r.right,
        a.bottom - r.bottom].map(Math.round), c.objectFit,
        c.objectPosition, c.backgroundColor]; }`);
  const at = {};
  for (const [name, p, ready] of [
    [
      'text',
      'math/path.md',
      "v.matches('.outline') && !!v.querySelector('summary')",
    ],
    ['picture', 'math/gen/mathematics-trench.jpg', 'v.naturalWidth > 0'],
    ['player', 'math/bio/john-milnor-abel-prize.mp3', 'v.duration > 0'],
    [
      'video',
      'math/elem/project-mathematics-tom-apostol/01-similarity.mkv',
      'v.videoWidth > 0',
    ],
  ]) {
    await s.open(`${shelf}#${p}`);
    await until(() =>
      s.js(`() => (v => ${ready})(
        document.querySelector('aside > :last-child'))`),
    );
    at[name] = await (name === 'video' ? fitted : placed)();
  }
  const videos = [at.video];
  if (s.width >= 800 && !s.touch) {
    const f = `document.querySelector('aside nav [data-key=f]')`;
    await s.hit(f);
    videos.push(await fitted());
    await s.hit(f);
  }
  check(
    at.picture[1] === 0 && at.picture[2] === 0,
    `${tag} picture: a picture starts at the view's left, at its top`,
    json(at.picture),
  );
  check(
    json(at.player.slice(1, 3)) === json(at.text.slice(1, 3)) &&
      at.text[1] > 0,
    `${tag} picture: a player starts where text does, in from the view's ` +
      'left and under the bar',
    json([at.text, at.player]),
  );
  check(
    videos.every(
      ([edges, fit, middle, black]) =>
        json(edges) === json([0, 0, 0, 0]) &&
        fit === 'contain' &&
        middle === '50% 50%' &&
        black === 'rgb(0, 0, 0)',
    ),
    `${tag} picture: a video's box is the whole view, black, its picture ` +
      'contained in the middle, side by side and at full width',
    json(videos),
  );
}

// a fold moves no text across, step 19: an outline grown past the view's
// height and a tree unfolded past its own keep each row's left and right
// where they were, as each keeps its scrollbar's room. only where a
// scrollbar takes room, a mouse's: a finger's overlays the page
async function still(s, tag) {
  // whether a scroller runs past its height, and each row in sight's left
  // and right, by its text
  const across = (scroller, rows) =>
    s.js(
      `(scroller, rows) => {
      const e = document.querySelector(scroller);
      return [e.scrollHeight > e.clientHeight,
        Object.fromEntries([...e.querySelectorAll(rows)]
          .filter(r => r.checkVisibility())
          .map(r => (b => [r.textContent, [b.left, b.right].map(Math.round)])(
            r.getBoundingClientRect())))]; }`,
      scroller,
      rows,
    );
  // past its height after, not before, and the rows in sight both times
  // that moved
  const moved = ([was, a], [now, b]) => [
    was,
    now,
    Object.keys(a).filter(t => t in b && json(a[t]) !== json(b[t])),
  ];
  const tool = key =>
    s.hit(`document.querySelector('aside nav [data-key=${key}]')`);
  await s.open(work.replace('#work', '#work/ssv/integration/diff.md'));
  await tool('c');
  const shut = await across('aside > :last-child', 'summary');
  await tool('e');
  const view = moved(shut, await across('aside > :last-child', 'summary'));
  await s.load(work);
  const tree = await across('main', '.row');
  await s.tap('work/ai');
  await until(() => s.unfolded('work/ai'));
  const rows = moved(tree, await across('main', '.row'));
  const kept = json([false, true, []]);
  check(
    json(view) === kept && json(rows) === kept,
    `${tag} fold: the text's x does not move on fold, as an outline grows ` +
      'past the view or a folder runs the tree past its height',
    json([view, rows]),
  );
  check(!s.errors.length, `${tag} fold: no page errors`, s.errors.join(' | '));
}

// full width, step 18: side by side or upright, f shows the file alone, the
// tree hidden, and esc, or the tool again on touch, brings the tree back,
// the file still open. on a phone the file is the whole width already: no
// tool for it, and the path has its room
async function widened(s, tag) {
  const f = `document.querySelector('aside nav [data-key=f]')`;
  // the tool shown, the tree shown, and the aside's distance from each
  // edge of the screen
  const seen = () =>
    s.js(`() => {
      const a = document.querySelector('aside').getBoundingClientRect();
      return [${f}.checkVisibility(),
        document.querySelector('main').checkVisibility(),
        ...[a.left, a.top, innerWidth - a.right, innerHeight - a.bottom]
          .map(Math.round)]; }`);
  await s.open(work.replace('#work', '#work/ssv/integration/diff.md'));
  const was = await seen();
  if (s.width < 800)
    return check(
      !was[0],
      `${tag} bar: no full width on a phone: the file is the whole width`,
      json(was),
    );
  await s.hit(f);
  const wide = await seen();
  await (s.touch ? s.hit(f) : s.press('Escape'));
  const back = await seen();
  check(
    json(wide) === json([true, false, 0, 0, 0, 0]) &&
      json(back) === json(was) &&
      was[1] &&
      (await s.viewing()),
    `${tag} bar: full width shows the file alone; ` +
      `${s.touch ? 'the tool' : 'esc'} again brings the tree back`,
    json([was, wide, back]),
  );
}

// the tablet upright: the viewer on top, its preview as tall as its own
console.log('-- frame, 800 upright touch');
const upright = await session(800, true, 1333);
await upright.load(work);
await frames(upright, '800-upright-touch', true);
await card(upright, '800-upright-touch');
console.log('-- folder, 800 upright touch');
const above = await session(800, true, 1333);
await folders(above, '800-upright-touch');
await edges(above, '800-upright-touch');
await tops(above, '800-upright-touch');
await pictured(above, '800-upright-touch');
await widened(above, '800-upright-touch');
// the answers where the tablet upright has them: the outline's bar and
// blocks, no hover tint, the title, the bar's ✓, the place kept
console.log('-- answers, 800 upright touch');
const tablet = await session(800, true, 1333);
await outlines(tablet, '800-upright-touch');
await answers(tablet, '800-upright-touch');
// a narrow window, with a mouse: no room for a preview, as on a phone
console.log('-- taps, 700');
const narrow = await session(700, false);
await taps(narrow, '700');

console.log(`shots in ${shots}`);
await c.close();
for (const [s] of served) s.close();
process.exit(failed ? 1 : 0);
