// checks path.html against the real library in headless chrome, both ways
// it is read: opened as a file, with the folder chosen in the picker, and
// served over http by a stand-in for quark. the oracles are the disk itself:
// each path.md's spans, fs.stat of every path, folder listings.
// usage: node run.mjs [library] [page]. the library defaults to ~/Data, the
// page to the path.html beside this folder. screenshots land in $TMPDIR/path.
import { launch } from './cdp.mjs';
import { quark, types } from './quark.mjs';
import { once } from 'node:events';
import { mkdir, readFile, readdir, stat, writeFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { posix } from 'node:path';
import { setTimeout as sleep } from 'node:timers/promises';

const library = process.argv[2] ?? `${process.env.HOME}/Data`;
const page = process.argv[3]
  ?? new URL('../path.html', import.meta.url).pathname;
const shots = `${tmpdir()}/path`;
const collate = new Intl.Collator(undefined, { numeric: true }).compare;
const visible = async dir => (await readdir(`${library}/${dir}`))
  .filter(n => !n.startsWith('.')).sort(collate);
const present = path => stat(`${library}/${path}`)
  .then(s => s.isDirectory() || s.size > 0, () => false);

// the subjects, as the disk has them: every folder holding a path.md
const subjects = [];
for (const dir of await visible('')) {
  const md = await readFile(`${library}/${dir}/path.md`, 'utf8')
    .catch(() => null);
  if (md !== null) subjects.push({ dir, md });
}
const heading = ({ dir, md }) =>
  (md.match(/^# (.*)/m)?.[1] ?? dir).replace(/\b'\b/g, '’');

// the installed quark sends a pdf as application/x-pdf, which chrome
// downloads; this assumes the one-line fix in its config.h
const server = quark(library, page, { ...types, pdf: 'application/pdf' })
  .listen(0, '127.0.0.1');
await once(server, 'listening');
const c = await launch(`${tmpdir()}/pathprofile`);
const problems = [], misses = [];
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
    (/\b404\b/.test(params.entry.text) ? misses : problems)
      .push(params.entry.url ?? params.entry.text);
  if (method === 'Page.downloadWillBegin')
    problems.push(`downloaded ${params.suggestedFilename}, not shown`);
});
await c.send('Runtime.enable');
await c.send('Page.enable');
await c.send('Log.enable');
await c.send('Browser.setDownloadBehavior',
  { behavior: 'deny', eventsEnabled: true });
await mkdir(shots, { recursive: true });

const size = (width, height) => c.send('Emulation.setDeviceMetricsOverride',
  { width, height, deviceScaleFactor: 1, mobile: false });
const theme = value => c.send('Emulation.setEmulatedMedia',
  { features: [{ name: 'prefers-color-scheme', value }] });
const shot = async name => writeFile(`${shots}/${name}.png`, Buffer.from(
  (await c.send('Page.captureScreenshot')).data, 'base64'));
const until = async (expression, ms = 15000) => {
  for (const end = Date.now() + ms; Date.now() < end; await sleep(50))
    if (await c.evaluate(expression).catch(() => false)) return true;
  return false;
};
const json = JSON.stringify;
const press = key =>
  c.send('Input.dispatchKeyEvent', { type: 'keyDown', key });
const stored = async name =>
  await c.evaluate(`localStorage.getItem(${json(name)})`) ?? '';
// a node of the page, by its path from the root
const node = path => `[...document.querySelectorAll('[role=treeitem]')]
  .find(li => li.node.path === ${json(path)})`;
const focused = `document.querySelector('.focus').parentElement`;

// drawn, every link path.md makes looked up, and the hash routed: mark
// gives each link a title, and route fills the aside
const rendered = () => until(`!!document.querySelector('[role=tree]')
  && [...document.querySelectorAll('main a')]
    .every(a => a.closest('.file') || a.title)
  && document.querySelector('aside').childElementCount > 0`);
// the view is this hash's once the bar names it, and it is no longer opening
const open = async (hash, ready) => {
  const name = posix.basename(hash.slice(1));
  await c.evaluate(`location.hash = ${json(hash)}; 1`);
  return until(`(() => {
    const view = document.querySelector('aside > :last-child');
    const named = document.querySelector('aside nav b')?.textContent;
    return decodeURIComponent(location.hash) === ${json(hash)} && view
      && (${json(!name)} || named === ${json(name)})
      && view.textContent !== 'opening…' && (${ready}); })()`);
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
  await until(`!!document.querySelector('[role=tree]')
    || !document.querySelector('main button').hidden`);
  if (await c.evaluate(`!document.querySelector('[role=tree]')`)) {
    await drop([library]);
    await c.evaluate(`showDirectoryPicker = async () => dropped[0];
      FileSystemHandle.prototype.requestPermission = async () => 'denied';
      document.querySelector('main button').click(); 1`, true);
  }
  return rendered();
};

// what the last version kept, with ~/Data/math as its root
const watched = 'calc/mit-highlights-of-calculus/'
  + '01-big-picture-of-calculus.mp4';
const old = {
  'pathreader:open': json(['stage:2 — After Algebra and Trig Are Solid',
    'calc/mit-highlights-of-calculus']),
  'pathreader:done': json(['SL Loney Plane Trigonometry',
    'elem/gateways-to-mathematics-herbert-gross']),
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
    else await c.send('Page.navigate',
      { url: `http://127.0.0.1:${server.address().port}/path.html` });
    return rendered();
  },
};

await size(1440, 900);
for (const [how, start] of Object.entries(routes)) {
  console.log(`-- ${how}`);
  check(await start(), 'renders every subject');
  await sleep(500);
  await shot(`${how}-path`);

  const drawn = await c.evaluate(`[...document.querySelectorAll('.subject')]
    .map(li => [li.node.path, li.firstChild.textContent])`);
  const want = subjects.map(s => [s.dir, heading(s)]);
  check(json(drawn) === json(want),
    'the subjects are the folders that hold a path.md, titled by it',
    json(drawn));

  if (how === 'file') {
    const moved = {
      at: await stored(`pathreader:at:math/${watched}`),
      open: await stored('pathreader:open'),
      done: await stored('pathreader:done'),
      left: await c.evaluate(`Object.keys(localStorage)
        .filter(k => k.startsWith('pathreader:at:calc/')).length`),
    };
    check(moved.at === '125' && moved.left === 0,
      'choosing ~/Data over ~/Data/math moves lecture places under math/',
      json(moved));
    check(moved.open.includes('math/stage:2 — After Algebra')
      && moved.open.includes('math/calc/mit-highlights-of-calculus'),
    'and the folds', moved.open);
    check(moved.done.includes('SL Loney Plane Trigonometry')
      && moved.done.includes('math/elem/gateways-to-mathematics-herbert-'
        + 'gross'), 'and the ticks, a title as it was', moved.done);
    check(await c.evaluate(`${node('math/calc/mit-highlights-of-calculus')}
      .ariaExpanded === 'true' && document.querySelector('input[name='
      + '"math/elem/gateways-to-mathematics-herbert-gross"]').checked`),
    'and the page shows them');
  }

  const links = Object.fromEntries(await c.evaluate(`[...document
    .querySelectorAll('.subject')].map(li => [li.node.path,
      [...li.querySelectorAll('a')].filter(a => !a.closest('.file'))
        .map(a => ({ text: a.textContent, absent: 'absent' in a.dataset,
          to: decodeURIComponent(a.hash.slice(1)) }))])`));
  for (const { dir, md } of subjects) {
    const spans = md.slice(md.search(/^## /m)).match(/`[^`]+`/g) ?? [];
    check(links[dir].length === spans.length,
      `every path in ${dir}/path.md is a link`,
      `${links[dir].length} links, ${spans.length} paths`);
  }
  const all = Object.values(links).flat(), wrong = [];
  for (const { text, to, absent } of all)
    if (absent === await present(to)) wrong.push(`${text} -> ${to}`);
  check(!wrong.length, 'a link is struck through when its path is missing '
    + 'or empty', wrong.join(', '));
  check(links.cs?.some(l => l.text === '../math/dm' && l.to === 'math/dm'),
    'a path from cs/ reaches ../math/dm');
  const refined = Object.entries(links).flatMap(([dir, list]) => list
    .filter(l => l.to !== posix.join(dir, l.text).replace(/\/$/, '')));
  check(refined.length === 4
    && refined.every(l => l.text === 'bio/srinivasa-ramanujan/'),
  'four titles in bio/srinivasa-ramanujan/ open their one file; the '
    + 'ambiguous fifth stays the folder', refined.map(l => l.to).join(', '));
  check(refined.some(l => l.to.endsWith('the-man-who-knew-infinity-a-life-'
    + 'of-the-genius-ramanujan-robert-kanigel.pdf')),
  'kanigel resolves to his book');

  for (const [name, hash, ready] of [
    ['pdf', '#math/elem/plane-trigonometry-sydney-loney.pdf',
      `view.localName === 'iframe' && !!view.src`],
    ['djvu', '#math/anal/elliptic-functions-and-elliptic-integrals-'
      + 'prasolov-and-solovyev.djvu',
    `view.localName === 'iframe' && view.sandbox.value === 'allow-scripts'`],
    ['epub', '#math/elem/how-to-solve-it-george-polya.epub',
      `view.shadowRoot?.querySelectorAll('section').length > 5
        && !view.shadowRoot.querySelector('parsererror')`],
    ['video', '#math/elem/project-mathematics-tom-apostol/01-similarity.mkv',
      `view.localName === 'video' && view.videoWidth > 0`],
    ['audio', '#math/bio/john-milnor-abel-prize.mp3',
      `view.localName === 'audio' && view.duration > 0`],
    ['image', '#math/bio/srinivasa-ramanujan/timeline-illustration.png',
      `view.localName === 'img' && view.naturalWidth > 0`],
    ['text', '#math/notes/exp.txt',
      `view.localName === 'pre' && view.textContent.length > 100`],
    ['empty', '#math/calc/infinite-powers-steven-strogatz',
      `view.textContent === 'this file is empty'`],
    ['missing', '#math/anal/a-radical-approach-to-real-analysis-david-'
      + 'bressoud.pdf', `view.textContent === 'nothing is at this path'`],
    ['nowhere', '#no/such/folder',
      `view.textContent === 'nothing is at this path'`],
    ['subject', '#cs', `view.querySelector('h1').textContent === 'cs/path'
      && / of \\d+ ticked/.test(view.textContent)`],
  ]) {
    check(await open(hash, ready), `opens ${name}: ${hash.slice(1)}`);
    await sleep({ djvu: 6000, pdf: 3000 }[name] ?? 800);
    if (how === 'file' || name === 'pdf')
      await shot(how === 'file' ? name : `${how}-${name}`);
  }
  if (how === 'file') {
    // cs's table and map, kept as written
    const cs = node('cs');
    await c.evaluate(`${cs}.querySelectorAll('.stage > .row')
      .forEach(row => row.click()); 1`);
    await sleep(300);
    await shot('cs');
    await c.evaluate(`${cs}.querySelector('.stage:last-child')
      .scrollIntoView(); 1`);
    await shot('map');
  }

  check(await open('', `view.querySelectorAll('.card').length
    === ${subjects.length}`), 'nothing open shows every subject');

  const course = 'math/calc/the-essence-of-calculus-grant-sanderson';
  const files = await visible(course);
  check(await open(`#${course}`, `${node(course)}.ariaCurrent === 'page'
    && ${node(course)}.querySelectorAll(':scope > ul > li').length
      === ${files.length}`), 'a folder unfolds to what is on disk');
  await shot(how === 'file' ? 'folder' : `${how}-folder`);
  await c.evaluate(`document.querySelector('[rel=next]').click(); 1`);
  const calc = await visible('math/calc');
  const after = `math/calc/${calc[calc.indexOf(posix.basename(course)) + 1]}`;
  check(await until(`decodeURIComponent(location.hash) === ${json(
    `#${after}`)}`), 'next opens the next entry in the same folder', after);

  const lecture = `${course}/05-derivatives-of-exponentials.mp4`;
  await c.evaluate(`localStorage.setItem('pathreader:at:${lecture}', '100');
    1`);
  check(await open(`#${lecture}`, `view.currentTime >= 100`),
    'a lecture resumes where it was left');

  const begun = Date.now();
  check(await open('#math/calc/calculus-an-intuitive-and-physical-approach-'
    + 'morris-kline/book.epub', `view.shadowRoot?.querySelector('img[src]')`),
  `a 39 MB epub of 4800 images opens, in ${Date.now() - begun} ms`);
  const [inflated, images] = await c.evaluate(`(r => [
    r.querySelectorAll('img[src]').length, r.querySelectorAll('img').length
  ])(document.querySelector('aside article').shadowRoot)`);
  check(inflated < images / 20,
    'and inflates only the images of the chapters in view',
    `${inflated} of ${images}`);
  check(await c.evaluate(`(() => {
    const book = document.querySelector('aside article');
    const before = book.scrollTop;
    book.shadowRoot.querySelectorAll('a[data-to]')[12].click();
    return book.scrollTop > before && location.hash.includes('book.epub');
  })()`), 'a link inside a book scrolls the book, not the page');

  await c.evaluate(`location.hash = ''; 1`);
  const loney = `${node('math')}.querySelector('.stage')
    .querySelectorAll('input')[1]`;
  const petzold = `document.querySelector('input[name^="cs/"]')`;
  // each is cleared, then ticked: the file route planted loney's tick
  const name = await c.evaluate(`[${loney}, ${petzold}].map(t => {
    if (t.checked) t.click();
    t.click();
    return t.name; })[1]`, true);
  const done = await stored('pathreader:done');
  check(done.includes('SL Loney Plane Trigonometry'),
    'ticking an entry keeps it under its title', done);
  check(name.startsWith('cs/elem/') && done.includes(name),
    'and an entry with no title under its path from the root', name);
  await start(true);
  check(await c.evaluate(`${loney}.checked && ${petzold}.checked`),
    'ticks survive a reload');
}

// the reload left the keys on the first row, the first subject's
console.log('-- keys');
await press('2');
check(await until(`${focused} === ${node('math')}
  && ${node('math')}.ariaExpanded === 'true'`),
'on a subject\'s row, a number unfolds that subject and takes the keys to it');
await press('l');
check(await until(`${focused} === ${node('math')}.querySelector('.stage')`),
  'right, on an unfolded subject, steps into its first stage');
await press('2');
check(await until(`${focused} === ${node('math')}
  .querySelector('.stage:nth-child(2)[aria-expanded=true]')`),
'in a subject, a number unfolds that stage and takes the keys to it');
await press('l');
check(await until(`${focused}.matches('.stage:nth-child(2) .entry')`),
  'right, on an unfolded stage, steps into its first entry');
await press('x');
check(await until(`${focused}.querySelector('input').checked
  && localStorage.getItem('pathreader:done')
    .includes(${focused}.querySelector('input').name)`),
'x ticks the entry the keys are on');
await press('j');
check(await until(`${focused}.matches(
  '.stage:nth-child(2) .entry:nth-of-type(2)')`), 'j moves down one row');
await press('h');
check(await until(`${focused}.matches('.stage:nth-child(2)')`),
  'left, on an entry, goes to its stage');
await press('h');
check(await until(`${focused}.ariaExpanded === 'false'`),
  'left again folds the stage');
await press('1');
await press('z');
check(await until(`[...document.querySelectorAll('[aria-expanded=true]')]
  .filter(li => !li.matches('.entry, .file'))
  .every(li => li.contains(${focused}))`),
'z folds every subject and stage but the ones the keys are in');
await press('/');
await c.send('Input.insertText', { text: 'essence calculus' });
await press('Enter');
check(await until(`decodeURIComponent(location.hash)
  === '#math/calc/the-essence-of-calculus-grant-sanderson'
  && ${focused}.ariaCurrent === 'page'`),
'find, then enter, opens the entry and puts the keys on it');
await press('l');
check(await until(`${focused}.matches('.file')`),
  'right, on an unfolded folder, steps into its first file');
await press('Enter');
check(await until(`location.hash.includes('00-what-they-dont-teach-you')
  && document.querySelector('aside video')`), 'enter opens a file');
await press('n');
check(await until(`location.hash.includes('01-introduction')`),
  'n opens the next file in the folder');
await press('Escape');
check(await until(`location.hash === ''`), 'escape closes what is open');

await theme('dark');
check(await open('#math/elem/how-to-solve-it-george-polya.epub',
  `view.shadowRoot?.querySelector('section')`), 'dark: an epub');
await c.evaluate(`document.querySelector('aside article').scrollTop = 2400;
  1`);
await sleep(500);
await shot('dark-epub');
await c.evaluate(`location.hash = ''; 1`);
await sleep(300);
await c.evaluate(`document.querySelector('main').scrollTop = 380; 1`);
await shot('dark-path');
await theme('light');
await size(390, 844);
await shot('narrow-path');
check(await open(`#math/calc/the-essence-of-calculus-grant-sanderson`,
  `!!${node('math/calc/the-essence-of-calculus-grant-sanderson')}
    .querySelector('li')`), 'narrow: a folder');
await sleep(300);
check(await c.evaluate(`(r => r.top >= 0 && r.bottom <= innerHeight)(
  document.querySelector('[aria-current] > .row').getBoundingClientRect())`),
'narrow: the folder unfolds in view');
await shot('narrow-folder');
check(await open('#math/elem/plane-trigonometry-sydney-loney.pdf',
  `view.localName === 'iframe'`), 'narrow: a file');
check(await c.evaluate(`document.querySelector('nav').scrollWidth
  <= innerWidth`), 'narrow: the viewer bar fits');
await sleep(3000);
await shot('narrow-file');

for (const url of misses) {
  const path = decodeURIComponent(new URL(url).pathname);
  if (await stat(library + path).then(() => true, () => false))
    problems.push(`404 for ${path}, which is on disk`);
}
check(!problems.length, 'no errors, exceptions or downloads; each 404 is a '
  + 'path not on disk', problems.join('\n     '));
console.log(`shots in ${shots}`);
await c.close();
server.close();
process.exit(failed ? 1 : 0);
