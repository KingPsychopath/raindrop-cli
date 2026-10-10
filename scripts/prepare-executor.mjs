// Use the exact version returned by the TARGET HOST's executor apps framework.
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
const version = process.argv[2];
if (!/^\d+\.\d+\.\d+(?:-[a-zA-Z0-9.-]+)?$/.test(version ?? '')) throw new Error('Pass the exact apps version from executor apps framework');
const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const source = path.join(root, 'integrations', 'executor');
const tested = JSON.parse(fs.readFileSync(path.join(source, 'package.json'))).dependencies.apps;
if (version !== tested) throw new Error(`Adapter tested with apps@${tested}; update and verify the adapter before targeting ${version}`);
const output = path.join(root, 'dist', 'executor-app');
fs.mkdirSync(output, { recursive: true });
fs.copyFileSync(path.join(source, 'index.ts'), path.join(output, 'index.ts'));
fs.writeFileSync(path.join(output, 'package.json'), JSON.stringify({ private: true, type: 'module', dependencies: { apps: version } }, null, 2) + '\n');
process.stdout.write(`Prepared ${output} for apps@${version}\n`);
