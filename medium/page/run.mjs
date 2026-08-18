// checks pages built on page.html in headless chrome: folds, keys, hash
// reveal, print, narrow, dark. the oracle is each page's own nodes, its
// <details id>, so any page can be checked, not only the skeleton.
// usage: node run.mjs [page]... the page defaults to the page.html beside
// this folder. screenshots land in $TMPDIR/page.
import { launch } from '../path/cdp.mjs';
import { mkdir, readFile, writeFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { basename, resolve } from 'node:path';
import { setTimeout as sleep } from 'node:timers/promises';

const pages = process.argv.slice(2).map(p => resolve(p));
if (!pages.length)
  pages.push(new URL('../page.html', import.meta.url).pathname);
const shots = `${tmpdir()}/page`;
const c = await launch(`${tmpdir()}/pageprofile`);
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
  if (method === 'Log.entryAdded' && params.entry.level === 'error')
    problems.push(params.entry.url ?? params.entry.text);
});
await c.send('Runtime.enable');
await c.send('Page.enable');
await c.send('Log.enable');
await mkdir(shots, { recursive: true });

const size = (width, height) => c.send('Emulation.setDeviceMetricsOverride',
  { width, height, deviceScaleFactor: 1, mobile: false });
const media = (media, value = 'light') => c.send('Emulation.setEmulatedMedia',
  { media, features: [{ name: 'prefers-color-scheme', value }] });
const json = JSON.stringify;
const until = async (expression, ms = 3000) => {
  for (const end = Date.now() + ms; Date.now() < end; await sleep(30))
    if (await c.evaluate(expression).catch(() => false)) return true;
  return false;
};
const press = key =>
  c.send('Input.dispatchKeyEvent', { type: 'keyDown', key });
const focused = `document.querySelector('details.focus')`;
const node = id => `document.getElementById(${json(id)})`;
// every node open or not, as its index row says and as it is
const folds = `[...document.querySelectorAll('main details[id]')]
  .map(d => [d.open, d.li.ariaExpanded === String(d.open)])`;
const mirrored = `${folds}.every(([, same]) => same)`;
const opened = `${folds}.map(([open]) => open)`;

for (const page of pages) {
  const name = basename(page, '.html');
  const shot = async what => writeFile(`${shots}/${name}-${what}.png`,
    Buffer.from((await c.send('Page.captureScreenshot')).data, 'base64'));
  console.log(`-- ${page}`);
  await size(1440, 900);
  await media('');
  const before = problems.length;
  await c.send('Page.navigate', { url: `file://${page}` });
  check(await until(`document.querySelector('nav li')
    && document.readyState === 'complete'`), 'renders');
  await shot('wide');

  // what the author wrote, read from the file: the open attributes
  const tags = [...(await readFile(page, 'utf8'))
    .matchAll(/<details\b([^>]*)>/g)].map(m => m[1])
    .filter(a => / id="/.test(a));
  const ids = tags.map(a => a.match(/ id="([^"]*)"/)[1]);
  const authored = tags.map(a => /\bopen\b/.test(a));
  const rows = await c.evaluate(`[...document.querySelectorAll('nav a')]
    .map(a => [a.hash.slice(1), a.querySelector('.n').textContent])`);
  check(json(rows.map(r => r[0])) === json(ids),
    'the index has one row per node, in order', json(rows.map(r => r[0])));
  check(await c.evaluate(`[...document.querySelectorAll('main details[id]')]
    .every(d => d.querySelector('summary .n').textContent
      === d.li.querySelector('.n').textContent)`),
  'each heading carries its index number');
  check(await c.evaluate(`!document.querySelector('nav button')`),
    'the nav holds only the index');
  check(json(await c.evaluate(opened)) === json(authored),
    'the page opens folded as the open attributes say');
  check(await c.evaluate(mirrored), 'the index shows the folds');

  // one state, toggled from either side
  const first = ids.find(id => true);
  const was = await c.evaluate(`${node(first)}.open`);
  await c.evaluate(`${node(first)}.querySelector('summary').click(); 1`);
  check(await until(`${node(first)}.open === ${!was} && ${mirrored}`),
    'a heading clicked folds, and the index follows');
  await c.evaluate(`${node(first)}.li.querySelector('.twisty').click(); 1`);
  check(await until(`${node(first)}.open === ${was} && ${mirrored}
    && location.hash === ''`),
  'the index twisty folds the node, and goes nowhere');

  // keys. parent is the first top node with a node in it, deep the
  // deepest node, where the folds above it are the most
  const tops = await c.evaluate(`[...document.querySelectorAll(
    'main details[id]')].filter(d => !d.parentElement.closest('details[id]'))
    .map(d => d.id)`);
  const [parent, kid] = await c.evaluate(`(p => [p?.id,
    p?.querySelector('details[id]').id])(${json(tops)}.map(id =>
      document.getElementById(id))
    .find(d => d.querySelector('details[id]')))`);
  const deep = await c.evaluate(`(all => {
    const depth = d => all.filter(a => a !== d && a.contains(d)).length;
    return all.reduce((a, d) => depth(d) >= depth(a) ? d : a).id;
  })([...document.querySelectorAll('main details[id]')])`);
  await c.evaluate(`scrollTo(0, 0); 1`);
  await press('c');
  check(await until(`${opened}.every(o => !o) && ${mirrored}`),
    'c folds every node');
  await press('k');
  check(await until(`${focused}?.id === ${json(first)}`),
    'the keys start on the first node');
  await press('j');
  check(await until(`${focused}.id === ${json(tops[1] ?? tops[0])}`),
    'j moves to the next heading shown', await c.evaluate(`${focused}.id`));
  await press('k');
  check(await until(`${focused}.id === ${json(first)}`), 'k moves back');
  if (parent) {
    await c.evaluate(`${node(parent)}.querySelector('summary').click(); 1`);
    await press('h');
    check(await until(`${focused}.id === ${json(parent)}
      && !${node(parent)}.open && ${mirrored}`), 'h folds the node');
    await press('l');
    check(await until(`${node(parent)}.open && ${mirrored}`),
      'l unfolds a folded node');
    await press('l');
    check(await until(`${focused}.id === ${json(kid)}`),
      'l on an unfolded node steps into it');
    await press('h');
    check(await until(`${focused}.id === ${json(parent)}`),
      'h on a folded node goes to its parent');
  }
  await press('e');
  check(await until(`${opened}.every(o => o) && ${mirrored}`),
    'e unfolds every node');
  await c.evaluate(`${node(deep)}.querySelector('summary').click();
    ${node(deep)}.querySelector('summary').click(); 1`);
  await press('z');
  check(await until(`[...document.querySelectorAll('main details[open]')]
    .every(d => d.contains(${focused}) || ${focused}.contains(d))
    && ${focused}.id === ${json(deep)}`),
  'z folds every node but the current, its parents and its own');
  await press('c');
  check(await until(`${opened}.every(o => !o)
    && !${focused}.parentElement.closest('details[id]')`),
  'a fold that hides the keys takes them');

  // find, and the hash
  const says = await c.evaluate(`${node(deep)}
    .querySelector('summary > :last-child').textContent`);
  await press('/');
  await c.send('Input.insertText', { text: says });
  check(await until(`document.querySelector('dialog [aria-selected=true]')`),
    'find lists the heading typed', says);
  await press('Enter');
  check(await until(`location.hash === ${json(`#${deep}`)}
    && ${focused}.id === ${json(deep)}
    && [...document.querySelectorAll('details')].filter(d =>
      d.contains(${node(deep)})).every(d => d.open)`),
  'enter on a find opens the way to it and puts the keys on it');
  // c leaves the keys on deep's top node; enter goes there, then again
  await press('c');
  const top = await c.evaluate(`${focused}.id`);
  await press('Enter');
  await press('c');
  await press('Enter');
  check(await until(`location.hash === ${json(`#${top}`)}
    && ${node(top)}.open && (r => r.top >= 0 && r.top < innerHeight)(
      ${node(top)}.getBoundingClientRect())`),
  'enter on the hash already shown opens it again, in view');
  await press('c');
  await c.evaluate(`location.hash = ''; 1`);
  await c.evaluate(`location.hash = ${json(`#${deep}`)}; 1`);
  check(await until(`[...document.querySelectorAll('details')].filter(d =>
    d.contains(${node(deep)})).every(d => d.open)`),
  'a hash reveals its target through closed folds');
  await c.send('Page.navigate', { url: `file://${page}#${deep}` });
  check(await until(`document.querySelector('nav li') && ${node(deep)}.open
    && ${focused}?.id === ${json(deep)}`),
  'and so does a page opened at it');
  await shot('hash');

  // a scroll away takes the keys to what is read. short, so it can
  await size(1440, 240);
  await press('e');
  await c.evaluate(`${node(deep)}.scrollIntoView(); 1`);
  await sleep(100);
  await c.evaluate(`scrollTo(0, 0); 1`);
  check(await until(`${focused}.id === ${json(first)}`),
    'the keys follow a scroll that leaves them');
  await size(1440, 900);

  await press('c');
  await c.send('Page.printToPDF');
  check(await until(`${opened}.every(o => o)`), 'print unfolds every node');
  await media('print');
  check(await c.evaluate(`getComputedStyle(document.querySelector('nav'))
    .display === 'none'`), 'print leaves the index out');

  await media('', 'dark');
  check(await c.evaluate(`getComputedStyle(document.body).backgroundColor
    === 'rgb(15, 15, 15)'`), 'dark: the page is dark');
  await c.evaluate(`scrollTo(0, 0); 1`);
  await shot('dark');
  await media('');
  await size(390, 844);
  await press('c');
  check(await c.evaluate(`getComputedStyle(document.querySelector('nav'))
    .display === 'none' && document.documentElement.scrollWidth
      <= innerWidth`), 'narrow: the text alone, and no sideways scroll',
  await c.evaluate(`document.documentElement.scrollWidth`));
  await press('k');
  await c.evaluate(`scrollTo(0, 0); 1`);
  await shot('narrow');
  await press('j');
  check(await until(`${focused}.id === ${json(tops[1] ?? tops[0])}`),
    'narrow: the keys move over the headings');
  check(problems.length === before, 'no errors or exceptions',
    problems.slice(before).join('\n     '));
}
console.log(`shots in ${shots}`);
await c.close();
process.exit(failed ? 1 : 0);
