#!/usr/bin/env node
'use strict';

const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const { spawnSync } = require('node:child_process');

function resolveBinary(packageName, platform, arch, resolve = require.resolve) {
  const supported = ['darwin-arm64', 'darwin-amd64', 'linux-arm64', 'linux-amd64', 'windows-arm64', 'windows-amd64'];
  const os = platform === 'win32' ? 'windows' : platform;
  const cpu = arch === 'x64' ? 'amd64' : arch;
  const target = `${os}-${cpu}`;
  if (!supported.includes(target)) throw new Error(`Unsupported platform ${platform}/${arch}. Install rdrop from source.`);
  const scope = packageName.slice(0, packageName.indexOf('/'));
  const dependency = `${scope}/rdrop-${target}`;
  let root;
  try { root = path.dirname(resolve(`${dependency}/package.json`)); }
  catch { throw new Error(`Missing ${dependency}. Reinstall with optional dependencies enabled (--include=optional), or use a native release.`); }
  const binary = path.join(root, os === 'windows' ? 'rdrop.exe' : 'rdrop');
  const expected = JSON.parse(fs.readFileSync(path.join(root, 'integrity.json'), 'utf8')).sha256;
  const actual = crypto.createHash('sha256').update(fs.readFileSync(binary)).digest('hex');
  if (!/^[a-f0-9]{64}$/.test(expected) || actual !== expected) throw new Error('rdrop binary checksum mismatch. Reinstall the package.');
  return binary;
}

if (require.main === module) {
  try {
    const metadata = require('./package.json');
    const binary = resolveBinary(metadata.name, process.platform, process.arch);
    const result = spawnSync(binary, process.argv.slice(2), { stdio: 'inherit', windowsHide: true });
    if (result.error) throw result.error;
    if (result.signal) process.kill(process.pid, result.signal);
    process.exit(result.status ?? 1);
  } catch (error) {
    process.stderr.write(`rdrop: ${error.message}\n`);
    process.exit(1);
  }
}

module.exports = { resolveBinary };
