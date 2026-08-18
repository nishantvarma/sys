// a stand-in for quark -l: the installed quark will not start without root,
// as it chroots. written from ~/src/quark at 5ad0df9: data.c's listing,
// http.c's redirect and ranges, config.h's types. it serves the page at
// /path.html and every other path from the library, as if the page had been
// copied into the library's root.
// what it gets wrong: no chroot, vhosts, maps or timeouts; error bodies and
// caching headers differ.
import { createServer } from 'node:http';
import { createReadStream } from 'node:fs';
import { readdir, stat } from 'node:fs/promises';

const utf8 = type => `${type}; charset=utf-8`;
// config.h as built. its pdf type, application/x-pdf, makes chrome download
// a pdf instead of showing it
export const types = {
  xml: utf8('application/xml'), xhtml: utf8('application/xhtml+xml'),
  html: utf8('text/html'), htm: utf8('text/html'), css: utf8('text/css'),
  txt: utf8('text/plain'), md: utf8('text/plain'), c: utf8('text/plain'),
  h: utf8('text/plain'), gz: 'application/x-gtar', tar: 'application/tar',
  pdf: 'application/x-pdf', png: 'image/png', gif: 'image/gif',
  jpeg: 'image/jpg', jpg: 'image/jpg', iso: 'application/x-iso9660-image',
  webp: 'image/webp', svg: utf8('image/svg+xml'), flac: 'audio/flac',
  mp3: 'audio/mpeg', ogg: 'audio/ogg', mp4: 'video/mp4', ogv: 'video/ogg',
  webm: 'video/webm',
};
const entity = { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;',
  "'": '&#x27;' };
const esc = s => s.replace(/[&<>"']/g, c => entity[c]);
const suffix = d => d.isDirectory() ? '/' : d.isSymbolicLink() ? '@'
  : d.isFIFO() ? '|' : d.isSocket() ? '=' : '';
// data.c: folders first, then strcmp
const order = (a, b) => (b.isDirectory() - a.isDirectory())
  || (a.name < b.name ? -1 : a.name > b.name ? 1 : 0);

function listing(path, entries) {
  const links = entries.map(d => `<br />\n\t\t<a href="${esc(d.name)}`
    + `${d.isDirectory() ? '/' : ''}">${esc(d.name)}${suffix(d)}</a>`);
  return `<!DOCTYPE html>\n<html>\n\t<head><title>Index of ${esc(path)}`
    + '</title></head>\n\t<body>\n\t\t<a href="..">..</a>'
    + links.join('') + '\n\t</body>\n</html>\n';
}

// http.c: one range, as first-last, first- or -last; a list is refused
function range(header, size) {
  const m = /^bytes=(\d*)-(\d*)$/.exec(header);
  if (!m || (!m[1] && !m[2])) return null;
  const [first, last] = m[1]
    ? [+m[1], m[2] ? Math.min(+m[2], size - 1) : size - 1]
    : [Math.max(size - +m[2], 0), size - 1];
  return first > last ? null : [first, last];
}

export function quark(library, page, mimes = types) {
  return createServer(async (req, res) => {
    const reply = (status, headers = {}, body = '') => {
      res.writeHead(status, headers);
      res.end(req.method === 'HEAD' ? undefined : body);
    };
    if (!['GET', 'HEAD'].includes(req.method)) return reply(405);
    const { pathname } = new URL(req.url, 'http://quark');
    const path = decodeURIComponent(pathname);
    const file = path === '/path.html' ? page : library + path;
    const st = await stat(file).catch(() => null);
    const html = utf8('text/html');
    if (!st) return reply(404, { 'Content-Type': html }, '404 Not Found');
    if (st.isDirectory()) {
      // http.c: a folder asked for without its slash is moved to it
      const moved = `http://${req.headers.host}${pathname}/`;
      if (!path.endsWith('/')) return reply(301, { Location: moved });
      const entries = (await readdir(file, { withFileTypes: true }))
        .filter(d => !d.name.startsWith('.')).sort(order);
      const body = listing(path, entries);
      return reply(200, { 'Content-Type': html,
        'Content-Length': Buffer.byteLength(body) }, body);
    }
    const ext = path.match(/\.([^./]+)$/)?.[1];
    const headers = { 'Content-Type': mimes[ext] ?? 'application/octet-stream',
      'Accept-Ranges': 'bytes' };
    let [first, last] = [0, st.size - 1];
    if (req.headers.range) {
      const r = range(req.headers.range, st.size);
      if (!r) return reply(416, { 'Content-Range': `bytes */${st.size}` });
      [first, last] = r;
      headers['Content-Range'] = `bytes ${first}-${last}/${st.size}`;
    }
    headers['Content-Length'] = st.size ? last - first + 1 : 0;
    res.writeHead(req.headers.range ? 206 : 200, headers);
    if (req.method === 'HEAD' || !st.size) return res.end();
    createReadStream(file, { start: first, end: last }).pipe(res);
  });
}
