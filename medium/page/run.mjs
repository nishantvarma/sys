// checks pages built on page.html in headless chrome: folds, the bar, the
// box, keys, hash reveal, print, narrow, dark. the oracle is each page's own
// nodes, its <details id>, so any page can be checked, not only the skeleton.
// usage: node run.mjs [page]... the page defaults to the page.html beside
// this folder. screenshots land in $TMPDIR/page, named folder-page-what.
import { launch } from '../path/cdp.mjs';
import { mkdir, readFile, writeFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { basename, dirname, resolve } from 'node:path';
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

const size = (width, height) =>
  c.send('Emulation.setDeviceMetricsOverride', {
    width,
    height,
    deviceScaleFactor: 1,
    mobile: false,
  });
const media = (media, value = 'light') =>
  c.send('Emulation.setEmulatedMedia', {
    media,
    features: [{ name: 'prefers-color-scheme', value }],
  });
const json = JSON.stringify;
const until = async (expression, ms = 3000) => {
  for (const end = Date.now() + ms; Date.now() < end; await sleep(30))
    if (await c.evaluate(expression).catch(() => false)) return true;
  return false;
};
// a dialog closes on escape only when it comes with its key code
const press = key =>
  c.send('Input.dispatchKeyEvent', {
    type: 'keyDown',
    key,
    windowsVirtualKeyCode: key === 'Escape' ? 27 : 0,
  });
const focused = `document.querySelector('details.focus')`;
const node = id => `document.getElementById(${json(id)})`;
const nodes = `[...document.querySelectorAll('main details[id]')]`;
const heading = d => `${d}.querySelector(':scope > summary > :first-child')`;
const opened = `${nodes}.map(d => d.open)`;
// the box is the dialog with an input; the keys, the one with a list of them
const box = `document.querySelector('dialog:has(input)')`;
const help = `document.querySelector('dialog:has(dl)')`;
const rows = `[...${box}.querySelectorAll('li')]`;
const selected = `${box}.querySelector('[aria-selected=true]')`;

for (const page of pages) {
  const name = `${basename(dirname(page))}-${basename(page, '.html')}`;
  const shot = async what =>
    writeFile(
      `${shots}/${name}-${what}.png`,
      Buffer.from((await c.send('Page.captureScreenshot')).data, 'base64'),
    );
  console.log(`-- ${page}`);
  await size(1440, 900);
  await media('');
  const before = problems.length;
  await c.send('Page.navigate', { url: `file://${page}` });
  check(
    await until(`document.querySelector('.bar')
    && document.readyState === 'complete'`),
    'renders',
  );
  await shot('wide');

  // what the author wrote, read from the file: the open attributes
  const tags = [
    ...(await readFile(page, 'utf8')).matchAll(/<details\b([^>]*)>/g),
  ]
    .map(m => m[1])
    .filter(a => / id="/.test(a));
  const ids = tags.map(a => a.match(/ id="([^"]*)"/)[1]);
  const authored = tags.map(a => /\bopen\b/.test(a));
  check(
    json(await c.evaluate(`${nodes}.map(d => d.id)`)) === json(ids),
    'the page holds the nodes the file does, in order',
  );
  check(
    json(await c.evaluate(opened)) === json(authored),
    'the page opens folded as the open attributes say',
  );
  check(
    await c.evaluate(`[...document.querySelectorAll('.bar button')]
    .map(b => b.textContent).join('|')
      === 'outline|expand all|collapse all|? keys'
    && document.querySelector('.bar').nextElementSibling === ${nodes}[0]`),
    'the bar sits before the first node: outline, expand all, collapse all',
  );

  const first = ids[0];
  const was = await c.evaluate(`${node(first)}.open`);
  await c.evaluate(`${node(first)}.querySelector('summary').click(); 1`);
  check(
    await until(`${node(first)}.open === ${!was}`),
    'a heading clicked folds',
  );
  await c.evaluate(`${node(first)}.querySelector('summary').click(); 1`);
  check(
    await until(`${node(first)}.open === ${was} && location.hash === ''`),
    'and clicked again unfolds, and goes nowhere',
  );

  // keys. parent is the first top node with a node in it, deep the
  // deepest node, where the folds above it are the most
  const tops = await c.evaluate(`${nodes}
    .filter(d => !d.parentElement.closest('details[id]')).map(d => d.id)`);
  const [parent, kid] = await c.evaluate(`(p => [p?.id,
    p?.querySelector('details[id]').id])(${json(tops)}.map(id =>
      document.getElementById(id))
    .find(d => d.querySelector('details[id]')))`);
  const deep = await c.evaluate(`(all => {
    const depth = d => all.filter(a => a !== d && a.contains(d)).length;
    return all.reduce((a, d) => depth(d) >= depth(a) ? d : a).id;
  })(${nodes})`);
  await c.evaluate(`scrollTo(0, 0); 1`);
  await press('c');
  check(await until(`${opened}.every(o => !o)`), 'c folds every node');
  await press('k');
  check(
    await until(`${focused}?.id === ${json(first)}`),
    'the keys start on the first node',
  );
  await press('j');
  check(
    await until(`${focused}.id === ${json(tops[1] ?? tops[0])}`),
    'j moves to the next heading shown',
    await c.evaluate(`${focused}.id`),
  );
  await press('k');
  check(await until(`${focused}.id === ${json(first)}`), 'k moves back');
  if (parent) {
    await c.evaluate(`${node(parent)}.querySelector('summary').click(); 1`);
    await press('h');
    check(
      await until(`${focused}.id === ${json(parent)}
      && !${node(parent)}.open`),
      'h folds the node',
    );
    await press('l');
    check(await until(`${node(parent)}.open`), 'l unfolds a folded node');
    await press('l');
    check(
      await until(`${focused}.id === ${json(kid)}`),
      'l on an unfolded node steps into it',
    );
    await press('h');
    check(
      await until(`${focused}.id === ${json(parent)}`),
      'h on a folded node goes to its parent',
    );
  }
  await press('o');
  check(
    await until(`${nodes}.every(d =>
      d.open === !!d.querySelector('details[id]'))`),
    'o folds to the outline: a node is open where it holds nodes',
  );
  await press('e');
  check(await until(`${opened}.every(o => o)`), 'e unfolds every node');
  await c.evaluate(`${node(deep)}.querySelector('summary').click();
    ${node(deep)}.querySelector('summary').click(); 1`);
  await press('z');
  check(
    await until(`[...document.querySelectorAll('main details[open]')]
    .every(d => d.contains(${focused}) || ${focused}.contains(d))
    && ${focused}.id === ${json(deep)}`),
    'z folds every node but the current, its parents and its own',
  );
  await press('c');
  check(
    await until(`${opened}.every(o => !o)
    && !${focused}.parentElement.closest('details[id]')`),
    'a fold that hides the keys takes them',
  );
  // the bar's buttons are the keys o, e and c
  const button = label => `[...document.querySelectorAll('.bar button')]
    .find(b => b.textContent === ${json(label)}).click(); 1`;
  await c.evaluate(button('expand all'));
  check(await until(`${opened}.every(o => o)`), 'expand all unfolds all');
  await c.evaluate(button('outline'));
  check(
    await until(`${nodes}.every(d =>
      d.open === !!d.querySelector('details[id]'))`),
    'outline folds to the outline',
  );
  await c.evaluate(button('collapse all'));
  check(await until(`${opened}.every(o => !o)`), 'collapse all folds all');

  await press('?');
  check(
    await until(`${help}.open && ['j', 'k', 'h', 'l', 'enter', '/', 't',
      'o', 'z', 'e', 'c', '?', 'esc'].every(k => [...${help}
      .querySelectorAll('kbd')].some(b => b.textContent.split(' ')
      .includes(k)))`),
    '? lists the keys',
  );
  await press('?');
  check(await until(`!${help}.open`), 'and ? again closes it');

  // the box. empty, it is the outline, the node being read selected; esc
  // leaves the page as it was
  const state = `JSON.stringify([${opened}, ${focused}.id, scrollY])`;
  const left = await c.evaluate(state);
  await press('t');
  check(
    await until(`${box}.open && ${rows}.map(li => li.textContent).join('|')
      === ${nodes}.map(d => ${heading('d')}.textContent).join('|')`),
    't opens the box: every heading, in order',
  );
  check(
    await c.evaluate(`${selected}?.textContent
      === ${heading(focused)}.textContent
      && ${selected}.ariaCurrent === 'true'`),
    'the box opens on the node being read',
  );
  const at = await c.evaluate(`${rows}.indexOf(${selected})`);
  await press('ArrowDown');
  check(
    await until(`${rows}.indexOf(${selected})
      === ${Math.min(at + 1, ids.length - 1)}`),
    'and ↓ selects the next heading',
  );
  await press('Escape');
  check(
    await until(`!${box}.open && ${state} === ${json(left)}
      && !${box}.contains(document.activeElement)`),
    'esc leaves the page as it was',
  );

  // find, and the hash. the box selects the first heading that matches,
  // which need not be deep's own. the hash names it already, folded: a pick
  // then fires no hashchange and the browser opens nothing, so the way to
  // it is the engine's to open
  const says = await c.evaluate(`${heading(node(deep))}.textContent`);
  const words = json(says.toLowerCase().split(/\s+/));
  const matches = h => `${words}.every(w => ${h}.toLowerCase().includes(w))`;
  const hit = await c.evaluate(`${nodes}.find(d =>
    ${matches(`${heading('d')}.textContent`)}).id`);
  await c.evaluate(`location.hash = ${json(`#${hit}`)}; 1`);
  await until(`${focused}.id === ${json(hit)}`);
  await press('c');
  await press('/');
  await c.send('Input.insertText', { text: says });
  check(
    await until(`${selected} === ${rows}[0]
      && ${rows}.length && ${rows}.every(li => ${matches('li.textContent')})`),
    'typed, the box lists the headings that match, the first selected',
    says,
  );
  await press('Enter');
  check(
    await until(`location.hash === ${json(`#${hit}`)}
    && ${focused}.id === ${json(hit)}
    && [...document.querySelectorAll('details')].filter(d =>
      d.contains(${node(hit)})).every(d => d.open)`),
    'enter on a find opens the way to it and puts the keys on it',
  );
  // the hash on deep's top node, folded, then enter on it. the keys are
  // read only once they settle: a fold moves them from its toggle event,
  // which is queued, and the scroll after it can move them again
  const top = await c.evaluate(`${json(tops)}.find(id =>
    document.getElementById(id).contains(${node(deep)}))`);
  await c.evaluate(`location.hash = ${json(`#${top}`)}; 1`);
  check(
    await until(`${focused}.id === ${json(top)} && ${node(top)}.open`),
    'a hash puts the keys on its node',
  );
  await press('c');
  await press('Enter');
  // in view as the engine's inview says: scroll is whole pixels and a
  // node need not be, so a node scrolled to the top can sit at -0.2
  check(
    await until(`location.hash === ${json(`#${top}`)}
    && ${node(top)}.open && (r => r.bottom > 0 && r.top < innerHeight)(
      ${node(top)}.firstElementChild.getBoundingClientRect())`),
    'enter on the hash already shown opens it again, in view',
  );
  await press('c');
  await c.evaluate(`location.hash = ''; 1`);
  await c.evaluate(`location.hash = ${json(`#${deep}`)}; 1`);
  check(
    await until(`[...document.querySelectorAll('details')].filter(d =>
    d.contains(${node(deep)})).every(d => d.open)`),
    'a hash reveals its target through closed folds',
  );
  await c.send('Page.navigate', { url: `file://${page}#${deep}` });
  check(
    await until(`document.querySelector('.bar') && ${node(deep)}.open
    && ${focused}?.id === ${json(deep)}`),
    'and so does a page opened at it',
  );
  await shot('hash');

  // a scroll away takes the keys to what is read. short, so it can
  await size(1440, 240);
  await press('e');
  await c.evaluate(`${node(deep)}.scrollIntoView(); 1`);
  await sleep(100);
  await c.evaluate(`scrollTo(0, 0); 1`);
  check(
    await until(`${focused}.id === ${json(first)}`),
    'the keys follow a scroll that leaves them',
  );
  await size(1440, 900);

  await press('c');
  await c.send('Page.printToPDF');
  check(await until(`${opened}.every(o => o)`), 'print unfolds every node');
  await media('print');
  check(
    await c.evaluate(`getComputedStyle(document.querySelector('.bar'))
    .display === 'none'`),
    'print leaves the bar out',
  );

  await media('', 'dark');
  check(
    await c.evaluate(`getComputedStyle(document.body).backgroundColor
    .match(/\\d+/g).slice(0, 3).every(v => v < 64)`),
    'dark: the page is dark',
  );
  await c.evaluate(`scrollTo(0, 0); 1`);
  await shot('dark');
  await media('');
  await size(390, 844);
  await press('e');
  await until(`${opened}.every(o => o)`);
  check(
    await c.evaluate(`document.documentElement.scrollWidth <= innerWidth`),
    'narrow, every node open: no sideways scroll',
    await c.evaluate(`document.documentElement.scrollWidth`),
  );
  await press('c');
  await press('k');
  await c.evaluate(`scrollTo(0, 0); 1`);
  await shot('narrow');
  await press('j');
  check(
    await until(`${focused}.id === ${json(tops[1] ?? tops[0])}`),
    'narrow: the keys move over the headings',
  );
  check(
    problems.length === before,
    'no errors or exceptions',
    problems.slice(before).join('\n     '),
  );
}
console.log(`shots in ${shots}`);
await c.close();
process.exit(failed ? 1 : 0);
