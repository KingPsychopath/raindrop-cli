// Keep packaged README links and graphics usable outside the Git checkout.
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
export function releaseReadme(version) {
  const markdown = fs.readFileSync(path.join(root, 'README.md'), 'utf8');
  return markdown.replace(/\]\(([^\s)]+)\)/g, (match, target, offset) => {
    if (/^(?:[a-z][a-z0-9+.-]*:|#|\/)/i.test(target)) return match;
    const image = /!\[[^\]]*$/.test(markdown.slice(0, offset));
    const base = image ? `https://raw.githubusercontent.com/machineyearningengineer/raindrop-cli/v${version}/` : `https://github.com/machineyearningengineer/raindrop-cli/blob/v${version}/`;
    return `](${base}${target})`;
  });
}
if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  const [version, destination] = process.argv.slice(2);
  if (!/^\d+\.\d+\.\d+(?:-[a-zA-Z0-9.-]+)?$/.test(version ?? '') || !destination) throw new Error('Usage: node scripts/release-readme.mjs <version> <output-file>');
  fs.writeFileSync(destination, releaseReadme(version));
}
