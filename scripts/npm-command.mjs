import fs from 'node:fs';
import path from 'node:path';
// Windows .cmd launchers cannot be invoked by execFile without a shell. Run
// npm's JS entry directly; paths and package arguments remain literal.
export function npmCommand(args) {
  if (process.platform !== 'win32') return ['npm', args];
  const cli = process.env.RDROP_NPM_CLI ?? path.join(path.dirname(process.execPath), 'node_modules', 'npm', 'bin', 'npm-cli.js');
  if (!fs.existsSync(cli)) throw new Error('Set RDROP_NPM_CLI to the installed npm-cli.js path on this Windows host');
  return [process.execPath, [cli, ...args]];
}
