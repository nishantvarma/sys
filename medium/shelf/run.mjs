// checks shelf.html against the real library in headless chrome, both ways
// it is read: opened as a file, with the folder chosen in the picker, and
// served over http by a stand-in for quark. the oracles are the disk itself:
// folder listings, each path.md's stages and spans, fs.stat of every path.
// usage: node run.mjs [library] [page]. the library defaults to ~/Data, the
// page to the shelf.html beside this folder. screenshots land in
// $TMPDIR/shelf.
import { launch } from './cdp.mjs';
import { quark, types } from './quark.mjs';
import { execFileSync } from 'node:child_process';
import { once } from 'node:events';
import { mkdir, readFile, readdir, stat, writeFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { posix } from 'node:path';
import { setTimeout as sleep } from 'node:timers/promises';

const library = process.argv[2] ?? `${process.env.HOME}/Data`;
const page =
  process.argv[3] ?? new URL('../shelf.html', import.meta.url).pathname;
const shots = `${tmpdir()}/shelf`;
const collate = new Intl.Collator(undefined, { numeric: true }).compare;
const visible = async dir =>
  (await readdir(`${library}/${dir}`))
    .filter(n => !n.startsWith('.'))
    .sort(collate);
const present = path =>
  stat(`${library}/${path}`).then(
    s => s.isDirectory() || s.size > 0,
    () => false,
  );

// the root as the disk has it, the page aside, and the subjects in it:
// every folder holding a path.md
const top = (await visible('')).filter(n => n !== posix.basename(page));
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
// an entry is a list item under a stage that names a path or has a title
const entries = ({ md }) =>
  md.slice(md.search(/^## /m)).match(/^(?:\d+\.|-) (?:.*`|\*\*).*$/gm)
    ?.length ?? 0;

// the installed quark sends a pdf as application/x-pdf, which chrome
// downloads; this assumes the one-line fix in its config.h
const server = quark(library, page, {
  ...types,
  pdf: 'application/pdf',
}).listen(0, '127.0.0.1');
await once(server, 'listening');
const c = await launch(`${tmpdir()}/shelfprofile`);
const problems = [],
  misses = [];
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
  if (method === 'Log.entryAdded' && params.entry.level === 'error')
    (/\b404\b/.test(params.entry.text) ? misses : problems).push(
      params.entry.url ?? params.entry.text,
    );
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
const shot = async name =>
  writeFile(
    `${shots}/${name}.png`,
    Buffer.from((await c.send('Page.captureScreenshot')).data, 'base64'),
  );
const until = async (expression, ms = 15000, where = c.evaluate) => {
  for (const end = Date.now() + ms; Date.now() < end; await sleep(50))
    if (await where(expression).catch(() => false)) return true;
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
const press = key =>
  c.send('Input.dispatchKeyEvent', { type: 'keyDown', key });
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
// a stage is gone into as on the tablet: a tap puts the cursor on its row,
// and a tap on the row under it goes in. it is shown once its links are
// looked up
const stage = async i => {
  await c.evaluate(`(row => { row.click(); row.click(); })(
    document.querySelectorAll('main .stage > .row')[${i}]); 1`);
  return until(`!document.querySelector('main .stage')
    && [...document.querySelectorAll('main ul a')].every(a => a.title)`);
};
// the view is this hash's once the bar names it, and it is no longer opening
const open = async (hash, ready) => {
  const name = posix.basename(hash.slice(1));
  await c.evaluate(`location.hash = ${json(hash)}; 1`);
  return until(`(() => {
    const view = document.querySelector('aside > :last-child');
    const named = document.querySelector('aside nav b')?.textContent;
    return decodeURIComponent(location.hash) === ${json(hash)} && view
      && named === ${json(name)} && view.textContent !== 'opening…'
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
    await until(`document.querySelector('aside h1')?.textContent
      === ${json(heading(first))}`),
    "under the cursor, a subject's preview is its path, titled by it",
  );

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
    const files = (await visible(s.dir)).filter(n => n !== 'path.md');
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
        ...(await c.evaluate(`[...document.querySelectorAll('main ul a')]
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
      `view.textContent === 'this file is empty'`,
    ],
    [
      'missing',
      '#math/anal/a-radical-approach-to-real-analysis-david-' + 'bressoud.pdf',
      `view.textContent === 'nothing is at this path'`,
    ],
    [
      'nowhere',
      '#no/such/folder',
      `view.textContent === 'nothing is at this path'`,
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
  const files = await visible(course);
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
  await stage(1);
  const petzold = `document.querySelector('input[name^="cs/"]')`;
  const name = await c.evaluate(`(t => { t.click(); return t.name; })(
    ${petzold})`);
  check(
    name.startsWith('cs/elem/') && (await stored('shelf:done')).includes(name),
    'and an entry with no title under its path from the root',
    name,
  );
  await start(true);
  check(
    await until(`${focused}?.matches('.stage')
      && ${focused}.querySelector('.n').textContent === '2'`),
    'a reload lands on the row the cursor was left on',
  );
  await press('l');
  check(await until(`${petzold}?.checked`), 'ticks survive a reload');
}

// the http route left the keys in cs's second stage
console.log('-- keys');
await place('');
const cs = String(top.indexOf('cs') + 1);
await press(cs);
check(
  await until(`${focused} === ${node('cs')}`),
  'a number puts the cursor on the row of that number',
);
await press('l');
check(
  await until(`!!document.querySelector('main .stage')
  && document.querySelector('main header h1 b')?.textContent === 'cs'`),
  'l goes into a folder: a subject, its stages first',
);
await press('3');
await press('l');
check(
  await until(`!document.querySelector('main .stage')
  && ${focused}?.matches('.entry')`),
  'l goes into a stage, onto an entry',
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
  await until(`${focused}?.matches('.stage')
  && ${focused}.querySelector('.n').textContent === '3'`),
  'h leaves a stage, onto its row',
);
await press('h');
check(
  await until(`${focused} === ${node('cs')} && location.hash === ''`),
  'h leaves a folder, onto its row in the one above',
);
await press('l');
check(
  await until(`${focused}?.querySelector('.n')?.textContent === '3'`),
  'the cursor is kept per place: back in cs, it is where it was left',
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
const [one, two] = (await visible(course)).map(n => `${course}/${n}`);
await press('1');
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
  '[href="#math/elem/how-to-solve-it-george-polya.epub"]').closest('.row')`;
await c.evaluate(`${polya}.click(); 1`);
await until(`!!document.querySelector('aside article')
  ?.shadowRoot.querySelector('section')`);
await c.evaluate(`document.querySelector('aside article').kept = true;
  ${polya}.click(); 1`);
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
await until(`!!document.querySelector('aside .card')`);
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

for (const url of misses) {
  const path = decodeURIComponent(new URL(url).pathname);
  if (
    await stat(library + path).then(
      () => true,
      () => false,
    )
  )
    problems.push(`404 for ${path}, which is on disk`);
}
check(
  !problems.length,
  'no errors, exceptions or downloads; each 404 is a ' + 'path not on disk',
  problems.join('\n     '),
);
console.log(`shots in ${shots}`);
await c.close();
server.close();
process.exit(failed ? 1 : 0);
