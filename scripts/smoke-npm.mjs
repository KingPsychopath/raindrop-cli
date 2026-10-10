// Install the current host's packed launcher and binary without registry access.
import fs from 'node:fs';
import { npmCommand } from './npm-command.mjs';
import os from 'node:os';
import path from 'node:path';
import { execFileSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';
const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const npmOS = { darwin: 'darwin', linux: 'linux', win32: 'windows' }[process.platform];
const arch = { x64: 'amd64', arm64: 'arm64' }[process.arch];
if (!npmOS || !arch) throw new Error('Unsupported smoke-test host');
const dist = path.join(root, 'dist');
const find = suffix => {
  const candidates = fs.readdirSync(dist).filter(name => name.includes(suffix) && name.endsWith('.tgz'));
  if (candidates.length !== 1) throw new Error(`Expected exactly one tarball for ${suffix}; clean dist before testing`);
  return path.join(dist, candidates[0]);
};
const directory = fs.mkdtempSync(path.join(os.tmpdir(), 'rdrop-npm-smoke-'));
try {
  fs.writeFileSync(path.join(directory, 'package.json'), '{"private":true}\n');
  const [command, args] = npmCommand(['install', '--offline', '--ignore-scripts', '--no-audit', '--no-fund', find(`rdrop-${npmOS}-${arch}-`), find('raindrop-cli-')]);
  execFileSync(command, args, { cwd: directory, stdio: 'inherit' });
  const scope = process.env.RDROP_NPM_SCOPE ?? '@kingpsychopath';
  const launcher = path.join(directory, 'node_modules', scope, 'raindrop-cli', 'launcher.cjs');
  const result = execFileSync(process.execPath, [launcher, 'version'], { encoding: 'utf8', cwd: directory });
  if (!result.includes('commit:')) throw new Error(`Bad installed binary response: ${result}`);
  const catalog = JSON.parse(execFileSync(process.execPath, [launcher, 'mcp', 'tools'], { encoding: 'utf8', cwd: directory }));
  if (!catalog.some(tool => tool.name === 'raindrop_context')) throw new Error('Installed package lacks context tools');
  process.stdout.write(`Installed npm launcher smoke passed: ${result}`);
} finally { fs.rmSync(directory, { recursive: true, force: true }); }
