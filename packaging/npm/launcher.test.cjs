const { test } = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const crypto = require('node:crypto');
const { resolveBinary } = require('./launcher.cjs');

test('selects native package and rejects a changed executable', () => {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), 'rdrop-npm-test-'));
  try {
    const binary = path.join(root, 'rdrop');
    fs.writeFileSync(binary, 'fixture executable');
    fs.writeFileSync(path.join(root, 'integrity.json'), JSON.stringify({ sha256: crypto.createHash('sha256').update('fixture executable').digest('hex') }));
    const resolve = (name) => { assert.equal(name, '@fixture/rdrop-darwin-arm64/package.json'); return path.join(root, 'package.json'); };
    assert.equal(resolveBinary('@fixture/raindrop-cli', 'darwin', 'arm64', resolve), binary);
    fs.writeFileSync(binary, 'changed');
    assert.throws(() => resolveBinary('@fixture/raindrop-cli', 'darwin', 'arm64', resolve), /checksum mismatch/);
    assert.throws(() => resolveBinary('@fixture/raindrop-cli', 'linux', 'riscv64'), /Unsupported/);
    assert.throws(() => resolveBinary('@fixture/raindrop-cli', 'linux', 'x64', () => { throw new Error(); }), /optional dependencies/);
  } finally { fs.rmSync(root, { recursive: true, force: true }); }
});
