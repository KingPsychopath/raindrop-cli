// Build versioned platform packages. Nothing is downloaded or published here.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import { execFileSync } from 'node:child_process';
import { releaseReadme } from './release-readme.mjs';
import { fileURLToPath } from 'node:url';

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const version = process.argv[2];
const scope = process.env.RDROP_NPM_SCOPE ?? '@machineyearn';
if (!/^\d+\.\d+\.\d+(?:-[a-zA-Z0-9.-]+)?$/.test(version ?? '')) throw new Error('Usage: node scripts/build-npm.mjs <release-version>');
if (!/^@[a-z0-9][a-z0-9-]*$/.test(scope)) throw new Error('RDROP_NPM_SCOPE must be a valid npm scope');
const output = path.join(root, 'dist', 'npm');
fs.mkdirSync(output, { recursive: true });
const commit = execFileSync('git', ['rev-parse', 'HEAD'], { cwd: root, encoding: 'utf8' }).trim();
const date = new Date().toISOString();
const optionalDependencies = {};
const targets = [['darwin', 'amd64', 'darwin', 'x64'], ['darwin', 'arm64', 'darwin', 'arm64'], ['linux', 'amd64', 'linux', 'x64'], ['linux', 'arm64', 'linux', 'arm64'], ['windows', 'amd64', 'win32', 'x64'], ['windows', 'arm64', 'win32', 'arm64']];
for (const [goos, goarch, npmOS, npmCPU] of targets) {
  const directory = path.join(output, `rdrop-${goos}-${goarch}`);
  fs.mkdirSync(directory, { recursive: true });
  const binary = goos === 'windows' ? 'rdrop.exe' : 'rdrop';
  execFileSync('go', ['build', '-trimpath', '-ldflags', `-s -w -X main.version=${version} -X main.commit=${commit} -X main.date=${date}`, '-o', path.join(directory, binary), './cmd/rdrop'], { cwd: root, env: { ...process.env, CGO_ENABLED: '0', GOOS: goos, GOARCH: goarch }, stdio: 'inherit' });
  const name = `${scope}/rdrop-${goos}-${goarch}`;
  const sha256 = crypto.createHash('sha256').update(fs.readFileSync(path.join(directory, binary))).digest('hex');
  fs.writeFileSync(path.join(directory, 'integrity.json'), JSON.stringify({ sha256, commit, version }) + '\n');
  fs.writeFileSync(path.join(directory, 'package.json'), JSON.stringify({ name, version, description: `Native ${goos}/${goarch} executable for rdrop for Raindrop.io`, license: 'MIT', homepage: 'https://github.com/KingPsychopath/raindrop-cli#readme', repository: { type: 'git', url: 'git+https://github.com/KingPsychopath/raindrop-cli.git' }, os: [npmOS], cpu: [npmCPU], files: [binary, 'integrity.json', 'LICENSE'] }, null, 2) + '\n');
  fs.copyFileSync(path.join(root, 'LICENSE'), path.join(directory, 'LICENSE'));
  optionalDependencies[name] = version;
}
const launcher = path.join(output, 'raindrop-cli');
fs.mkdirSync(launcher, { recursive: true });
const metadata = JSON.parse(fs.readFileSync(path.join(root, 'packaging', 'npm', 'package.json')));
delete metadata.private;
Object.assign(metadata, { name: `${scope}/raindrop-cli`, version, optionalDependencies });
fs.writeFileSync(path.join(launcher, 'package.json'), JSON.stringify(metadata, null, 2) + '\n');
fs.copyFileSync(path.join(root, 'LICENSE'), path.join(launcher, 'LICENSE'));
fs.writeFileSync(path.join(launcher, 'README.md'), releaseReadme(version));
fs.copyFileSync(path.join(root, 'packaging', 'npm', 'launcher.cjs'), path.join(launcher, 'launcher.cjs'));
fs.chmodSync(path.join(launcher, 'launcher.cjs'), 0o755);
process.stdout.write(`Built six platform packages and the launcher under ${output}\n`);
