import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { execFileSync } from 'node:child_process';
const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const version = process.argv[2];
if (!/^\d+\.\d+\.\d+(?:-[a-zA-Z0-9.-]+)?$/.test(version ?? '')) throw new Error('Usage: node scripts/pack-plugin.mjs <version>');
const plugin = path.join(root, 'integrations', 'agent-plugin');
for (const name of ['plugin.json', '.codex-plugin/plugin.json', '.claude-plugin/plugin.json']) {
  const manifest = JSON.parse(fs.readFileSync(path.join(plugin, name)));
  if (manifest.name !== 'raindrop' || manifest.version !== version) throw new Error(`Update ${name} to the release version before packaging`);
}
const config = JSON.parse(fs.readFileSync(path.join(plugin, '.mcp.json')));
if (!config.mcpServers.raindrop.args.includes(`@machineyearn/raindrop-cli@${version}`)) throw new Error('Plugin MCP package pin differs from this release');
const dist = path.join(root, 'dist');
fs.mkdirSync(dist, { recursive: true });
const output = path.join(dist, `rdrop-agent-plugin_${version}.zip`);
if (fs.existsSync(output)) fs.unlinkSync(output);
execFileSync('zip', ['-qr', output, 'agent-plugin'], { cwd: path.join(root, 'integrations'), stdio: 'inherit', shell: false });
process.stdout.write(`Packed ${output}\n`);
