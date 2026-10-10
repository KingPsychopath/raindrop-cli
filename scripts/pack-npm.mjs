import fs from 'node:fs';
import { npmCommand } from './npm-command.mjs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { execFileSync } from 'node:child_process';
const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const dist = path.join(root, 'dist');
const directories = fs.readdirSync(path.join(dist, 'npm')).sort();
for (const name of directories) {
  const [command, args] = npmCommand(['pack', '--ignore-scripts', '--pack-destination', dist]);
  execFileSync(command, args, { cwd: path.join(dist, 'npm', name), stdio: 'inherit', shell: false });
}
