// Publish the exact release tarballs, platform packages first and launcher last.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import { execFileSync } from 'node:child_process';
import { npmCommand } from './npm-command.mjs';

const version = process.argv[2];
if (!/^\d+\.\d+\.\d+(?:-[a-zA-Z0-9.-]+)?$/.test(version ?? '')) throw new Error('Usage: node scripts/publish-npm.mjs <release-version>');
const scope = process.env.RDROP_NPM_SCOPE ?? '@machineyearn';
if (!/^@[a-z0-9][a-z0-9-]*$/.test(scope)) throw new Error('Invalid npm scope');
const directory = path.resolve(process.env.RDROP_RELEASE_DIR ?? 'dist');
const names = ['rdrop-darwin-amd64', 'rdrop-darwin-arm64', 'rdrop-linux-amd64', 'rdrop-linux-arm64', 'rdrop-windows-amd64', 'rdrop-windows-arm64', 'raindrop-cli'];
const npm = (args, options = {}) => {
  const [command, parameters] = npmCommand([...args, '--registry', 'https://registry.npmjs.org']);
  return execFileSync(command, parameters, { shell: false, ...options });
};
const artifacts = names.map(name => {
  const filename = `${scope.slice(1)}-${name}-${version}.tgz`;
  const archive = path.join(directory, filename);
  const manifest = JSON.parse(execFileSync('tar', ['-xOf', archive, 'package/package.json'], { encoding: 'utf8' }));
  if (manifest.name !== `${scope}/${name}` || manifest.version !== version || manifest.private) throw new Error(`Unexpected manifest in ${filename}`);
  return { name: manifest.name, archive, integrity: `sha512-${crypto.createHash('sha512').update(fs.readFileSync(archive)).digest('base64')}` };
});
npm(['whoami'], { stdio: 'inherit' });
for (const artifact of artifacts) {
  const response = await fetch(`https://registry.npmjs.org/${encodeURIComponent(artifact.name)}/${version}`);
  if (response.ok) {
    const existing = await response.json();
    if (existing.dist?.integrity !== artifact.integrity) throw new Error(`Published ${artifact.name}@${version} differs from this release. Never overwrite or republish it.`);
    process.stdout.write(`Already published and verified: ${artifact.name}@${version}\n`);
    continue;
  }
  if (response.status !== 404) throw new Error(`Registry lookup failed: HTTP ${response.status}`);
  npm(['publish', artifact.archive, '--access', 'public', '--ignore-scripts', '--tag', version.includes('-') ? 'next' : 'latest'], { stdio: 'inherit' });
}
