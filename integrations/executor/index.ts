// A portable Executor app: REST calls run on its host; no Go executable needed.
import { array, boolean, defineApp, defineProvider, mutation, number, object, query, router, secrets, string, type QueryContext } from 'apps';
import { always, never } from 'apps/operations/approval';

const raindrop = defineProvider({
  name: 'Raindrop.io',
  auth: { apiKey: secrets({ label: 'Raindrop API token', fields: object({ token: string({ minLength: 1 }) }) }) },
});
const requirements = { accounts: { raindrop } };
type Context = QueryContext<typeof requirements>;

type Bookmark = { _id: number; title: string; link: string; tags?: string[]; collection: { $id: number } };
type Collection = { _id: number; title: string; count: number; parent?: { $id: number } | null };
type Envelope = { result?: boolean; items?: unknown[]; item?: unknown };

function positiveID(value: number): number {
  if (!Number.isSafeInteger(value) || value <= 0) throw new Error('ID must be a positive safe integer');
  return value;
}
function collectionID(value: number): number {
  if (!Number.isSafeInteger(value) || (value <= 0 && ![0, -1, -99].includes(value))) throw new Error('Collection must be positive, 0, -1, or -99');
  return value;
}
async function request<T>(ctx: Context, method: string, path: string, body?: unknown): Promise<T> {
  const response = await ctx.fetch(`https://api.raindrop.io/rest/v1/${path}`, {
    method, redirect: 'manual',
    headers: { Authorization: `Bearer ${ctx.accounts.raindrop.fields.token}`, Accept: 'application/json', ...(body === undefined ? {} : { 'Content-Type': 'application/json' }) },
    ...(body === undefined ? {} : { body: JSON.stringify(body) }),
  });
  if (!response.ok) throw new Error(`Raindrop API HTTP ${response.status}`);
  const value = await response.json() as T & Envelope;
  if (value.result === false) throw new Error('Raindrop API reported result:false');
  return value;
}
function bookmark(value: unknown): Bookmark {
  const item = value as Bookmark | undefined;
  if (!item || !Number.isSafeInteger(item._id) || typeof item.title !== 'string' || typeof item.link !== 'string' || !item.collection || !Number.isSafeInteger(item.collection.$id)) throw new Error('Invalid Raindrop bookmark response');
  return item;
}
const scope = object({ collection: number().default(0), search: string().default('') });
const idInput = object({ id: number() });

const report = query({ description: 'Start organizing here. Read library cleanup counts; never mutates.', input: scope, approval: never() }, async (ctx: Context, input) => {
  const path = `filters/${collectionID(input.collection)}?${new URLSearchParams({ search: input.search })}`;
  return { collectionId: input.collection, search: input.search, filters: await request<Envelope>(ctx, 'GET', path), next: 'Inspect exact bookmark IDs, propose a taxonomy, and obtain review before applying writes. Use local rdrop snapshot plans for bulk changes.' };
});
const list = query({ description: 'Read one bounded bookmark page using Raindrop search syntax.', input: object({ collection: number().default(0), search: string().default(''), page: number().default(0), perPage: number().default(25) }), approval: never() }, async (ctx: Context, input) => {
  if (!Number.isSafeInteger(input.page) || input.page < 0 || !Number.isSafeInteger(input.perPage) || input.perPage < 1 || input.perPage > 50) throw new Error('page must be nonnegative; perPage must be 1..50');
  const params = new URLSearchParams({ search: input.search, page: String(input.page), perpage: String(input.perPage) });
  return request<Envelope>(ctx, 'GET', `raindrops/${collectionID(input.collection)}?${params}`);
});
const get = query({ description: 'Read one bookmark by ID.', input: idInput, approval: never() }, async (ctx: Context, { id }) => bookmark((await request<Envelope>(ctx, 'GET', `raindrop/${positiveID(id)}`)).item));
const collections = query({ description: 'Read root and nested collections.', input: object({}), approval: never() }, async (ctx: Context) => {
  const roots = await request<{ items: Collection[] }>(ctx, 'GET', 'collections');
  const children = await request<{ items: Collection[] }>(ctx, 'GET', 'collections/childrens');
  if (!Array.isArray(roots.items) || !Array.isArray(children.items)) throw new Error('Invalid collection response');
  return [...new Map([...roots.items, ...children.items].map(item => [item._id, item])).values()];
});
const tags = query({ description: 'Read tag names and counts.', input: object({ collection: number().default(0) }), approval: never() }, async (ctx: Context, { collection }) => request<Envelope>(ctx, 'GET', `tags/${collectionID(collection)}`));
const highlights = query({ description: 'Read highlights for all bookmarks or a collection.', input: object({ collection: number().default(0) }), approval: never() }, async (ctx: Context, { collection }) => request<Envelope>(ctx, 'GET', collectionID(collection) === 0 ? 'highlights' : `highlights/${collection}`));

const add = mutation({ description: 'Create one reviewed bookmark. Executor requires human approval.', input: object({ url: string({ minLength: 1 }), title: string().optional(), tags: array(string({ minLength: 1 })).optional(), collection: number().default(-1), confirm: boolean() }), approval: always() }, async (ctx: Context, input) => {
  if (!input.confirm) throw new Error('confirm must be true');
  const url = new URL(input.url);
  if (!['http:', 'https:'].includes(url.protocol)) throw new Error('URL must use HTTP(S)');
  if (input.collection !== -1) positiveID(input.collection);
  const { item } = await request<Envelope>(ctx, 'POST', 'raindrop', { link: input.url, title: input.title, tags: input.tags, collection: { $id: input.collection } });
  return bookmark(item);
});
const appendTags = mutation({ description: 'Append tags to one exact bookmark; preserve existing tags. Executor requires human approval.', input: object({ id: number(), tags: array(string({ minLength: 1 })), confirm: boolean() }), approval: always() }, async (ctx: Context, input) => {
  if (!input.confirm || input.tags.length === 0) throw new Error('Nonempty tags and confirm:true required');
  const id = positiveID(input.id);
  const current = bookmark((await request<Envelope>(ctx, 'GET', `raindrop/${id}`)).item);
  const tags = [...new Set([...(current.tags ?? []), ...input.tags])];
  return bookmark((await request<Envelope>(ctx, 'PUT', `raindrop/${id}`, { tags })).item);
});

export default defineApp(requirements, { tools: router({ cleanup: router({ report }), bookmarks: router({ list, get, add, appendTags }), collections, tags, highlights }) });
