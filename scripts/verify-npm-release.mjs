// Verify published npm bytes against downloaded, checksummed GitHub release assets.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import { execFileSync } from 'node:child_process';

const [version, directory = 'dist'] = process.argv.slice(2);
if (!/^\d+\.\d+\.\d+$/.test(version ?? '')) throw new Error('Usage: node scripts/verify-npm-release.mjs <version> [release-directory]');
const root = path.resolve(directory);
const server = JSON.parse(fs.readFileSync('server.json', 'utf8'));
const expectedName = 'io.github.KingPsychopath/raindrop-cli';
if (server.name !== expectedName || server.version !== version) throw new Error('Registry identity/version differs from this release');
const registryPackage = server.packages?.[0];
if (server.packages.length !== 1 || registryPackage.identifier !== '@machineyearn/raindrop-cli' || registryPackage.version !== version || registryPackage.transport.type !== 'stdio') throw new Error('Unexpected registry package');
const checksums = new Map(fs.readFileSync(path.join(root, 'checksums.txt'), 'utf8').trim().split('\n').map(line => {
  const match = /^([a-f0-9]{64})\s+\*?([^/\\\s]+)$/.exec(line);
  if (!match) throw new Error('Invalid release checksum entry');
  return [match[2], match[1]];
}));
const platforms = ['rdrop-darwin-amd64', 'rdrop-darwin-arm64', 'rdrop-linux-amd64', 'rdrop-linux-arm64', 'rdrop-windows-amd64', 'rdrop-windows-arm64'];
const artifacts = [...platforms, 'raindrop-cli'].map(name => {
  const filename = `machineyearn-${name}-${version}.tgz`;
  const archive = path.join(root, filename);
  const bytes = fs.readFileSync(archive);
  if (crypto.createHash('sha256').update(bytes).digest('hex') !== checksums.get(filename)) throw new Error(`Release checksum mismatch: ${filename}`);
  const manifest = JSON.parse(execFileSync('tar', ['-xOf', archive, 'package/package.json'], { encoding: 'utf8' }));
  const packageName = `@machineyearn/${name}`;
  if (manifest.name !== packageName || manifest.version !== version || manifest.private) throw new Error(`Unexpected manifest: ${filename}`);
  if (name === 'raindrop-cli') {
    if (manifest.mcpName !== expectedName) throw new Error('Launcher mcpName does not match server.json');
    const expectedDependencies = Object.fromEntries(platforms.map(platform => [`@machineyearn/${platform}`, version]));
    if (JSON.stringify(Object.entries(manifest.optionalDependencies ?? {}).sort()) !== JSON.stringify(Object.entries(expectedDependencies).sort())) throw new Error('Launcher platform dependencies differ from this release');
  }
  return { name, packageName, integrity: `sha512-${crypto.createHash('sha512').update(bytes).digest('base64')}` };
});
for (const { name, packageName, integrity } of artifacts) {
  const response = await fetch(`https://registry.npmjs.org/${encodeURIComponent(packageName)}/${version}`, { signal: AbortSignal.timeout(30000) });
  if (!response.ok) throw new Error(`npm lookup failed for ${packageName}: HTTP ${response.status}`);
  const published = await response.json();
  if (published.name !== packageName || published.version !== version || published.dist?.integrity !== integrity) throw new Error(`Published bytes differ: ${packageName}@${version}`);
  if (name === 'raindrop-cli' && published.mcpName !== expectedName) throw new Error('Published launcher is missing registry ownership metadata');
  process.stdout.write(`Verified ${packageName}@${version}\n`);
}
