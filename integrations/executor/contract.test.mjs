import { test } from 'node:test';
import assert from 'node:assert/strict';
import app from './index.ts';
import { createAppHandler, hostContext } from 'apps/host';

const handle = createAppHandler(app);
async function invoke(body, context) {
  const response = await handle(new Request('http://fixture/', { method: 'POST', body: JSON.stringify(body) }), context);
  return { status: response.status, ...await response.json() };
}
test('portable host enforces exact approval and preserves existing tags', async () => {
  const requirements = await invoke({ operation: 'requirements' }, hostContext({}));
  assert.equal(requirements.ok, true);
  const accounts = { raindrop: { id: 'acc_fixture', provider: requirements.value.accounts.raindrop.definition, method: 'apiKey', fields: { token: 'fixture-token' } } };
  const context = hostContext(accounts);
  const catalog = await invoke({ operation: 'inspect' }, context);
  assert.equal(catalog.ok, true);
  const index = new Map(catalog.value.tools.map(tool => [tool.name, tool]));
  assert.equal(index.get('cleanup.report').readOnly, true);
  assert.equal(index.get('bookmarks.appendTags').readOnly, false);
  const input = { id: 42, tags: ['reference'], confirm: true };
  let reads = 0, writes = 0;
  const originalFetch = globalThis.fetch;
  globalThis.fetch = async (url, options) => {
    assert.equal(String(url), 'https://api.raindrop.io/rest/v1/raindrop/42');
    assert.equal(options.redirect, 'manual');
    assert.equal(new Headers(options.headers).get('Authorization'), 'Bearer fixture-token');
    const item = { _id: 42, title: 'Docs', link: 'https://fixture.test', tags: ['My own tag'], collection: { $id: -1 } };
    if (options.method === 'PUT') {
      writes++;
      assert.deepEqual(JSON.parse(await new Response(options.body).text()), { tags: ['My own tag', 'reference'] });
    } else { reads++; }
    return Response.json({ result: true, item });
  };
  try {
    const denied = await invoke({ operation: 'call', tool: 'bookmarks.appendTags', input }, context);
    assert.equal(denied.error._tag, 'HostToolApprovalRequired');
    assert.equal(reads + writes, 0);
    const wrongApproval = hostContext(accounts, { tool: 'bookmarks.appendTags', input: { ...input, id: 99 } });
    const wrong = await invoke({ operation: 'call', tool: 'bookmarks.appendTags', input }, wrongApproval);
    assert.equal(wrong.ok, false);
    assert.equal(reads + writes, 0);
    const approved = hostContext(accounts, { tool: 'bookmarks.appendTags', input });
    const result = await invoke({ operation: 'call', tool: 'bookmarks.appendTags', input }, approved);
    assert.equal(result.ok, true, JSON.stringify(result));
    assert.equal(reads, 1);
    assert.equal(writes, 1);
    const invalid = await invoke({ operation: 'call', tool: 'bookmarks.list', input: { perPage: 51 } }, context);
    assert.equal(invalid.ok, false);
    assert.equal(reads + writes, 2);
  } finally { globalThis.fetch = originalFetch; }
});
