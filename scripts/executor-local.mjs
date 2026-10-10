#!/usr/bin/env node
// Run the installed Executor CLI with its private local file-backed API key.
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { spawn } from 'node:child_process';

const directory = process.env.EXECUTOR_DATA_DIR ?? path.join(os.homedir(), '.local', 'share', 'rdrop-executor');
const environment = { ...process.env, EXECUTOR_DATA_DIR: directory };
if (!environment.EXECUTOR_API_KEY) {
  const keys = JSON.parse(fs.readFileSync(path.join(directory, 'keys.json'), 'utf8'));
  if (typeof keys.apiKey !== 'string' || !keys.apiKey) throw new Error('Local Executor API key is missing');
  environment.EXECUTOR_API_KEY = keys.apiKey;
}
const child = spawn('executor', process.argv.slice(2), { env: environment, stdio: 'inherit' });
child.on('error', error => { process.stderr.write(`${error.message}\n`); process.exitCode = 1; });
child.on('exit', (code, signal) => {
  if (signal) process.kill(process.pid, signal);
  else process.exitCode = code ?? 1;
});
