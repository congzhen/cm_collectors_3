import { beforeEach, expect, it, vi } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import { watch } from 'vue';
import { appStoreData } from '../app.storeData';

const api = vi.hoisted(() => ({ info: vi.fn(), actors: vi.fn(), tags: vi.fn() }));
vi.mock('@/server/filesBases.server', () => ({ filesBasesServer: { infoById: api.info } }));
vi.mock('@/server/performer.server', () => ({ performerServer: { listTopPreferredPerformers: api.actors } }));
vi.mock('@/server/tag.server', () => ({ tagServer: { tagDataByFilesBasesId: api.tags } }));
vi.mock('@/server/app.server', () => ({ appDataServer: {} }));

const library = (id: string) => ({ status: true, data: {
  id, name: id + '-name', filesRelatedPerformerBases: [{ main: true, performerBases_id: id + '-actors' }],
  filesBasesSetting: { config_json_data: JSON.stringify({ performerShowNum: 25, showCustomTagResourceCount: false }) },
} });
const tags = (id: string) => ({ status: true, data: { tag: [{ id }], tagClass: [{ id }] } });
const actors = (id: string) => ({ status: true, data: [{ id }] });
function deferred<T>() {
  let resolve!: (value: T) => void;
  let reject!: (reason: unknown) => void;
  const promise = new Promise<T>((yes, no) => { resolve = yes; reject = no; });
  return { promise, resolve, reject };
}
const snapshot = () => {
  const s = appStoreData();
  return JSON.parse(JSON.stringify([s.currentFilesBases, s.currentConfigApp, s.currentPerformerBasesIds,
    s.currentMainPerformerBasesId, s.currentTag, s.currentTagClass, s.currentTopPreferredPerformers]));
};
beforeEach(async () => {
  vi.resetAllMocks();
  setActivePinia(createPinia());
  api.info.mockImplementation(async (id: string) => library(id));
  api.tags.mockImplementation(async (id: string) => tags(id));
  api.actors.mockImplementation(async (...args: unknown[]) => actors(String(args[4])));
  await appStoreData().initCurrentFilesBases('old');
});

it('同步库监听器读取到完整的新库对象，旧引用不被原地改写', async () => {
  const store = appStoreData();
  const previousLibrary = store.currentFilesBases;
  const observed = vi.fn(() => [store.currentFilesBases.name, store.currentFilesBases.filesRelatedPerformerBases[0].performerBases_id]);
  const stop = watch(() => store.currentFilesBases.id, observed, { flush: 'sync' });
  try {
    await store.initCurrentFilesBases('new');
    expect(observed.mock.results[0].value).toEqual(['new-name', 'new-actors']);
    expect(previousLibrary.id).toBe('old');
    expect(previousLibrary.name).toBe('old-name');
  } finally { stop(); }
});

it.each(['info', 'tags', 'actors'] as const)('%s 业务失败保留原库的全部状态', async (endpoint) => {
  const before = snapshot();
  api[endpoint].mockResolvedValueOnce({ status: false, msg: '请求失败' });
  expect(await appStoreData().initCurrentFilesBases('new')).toMatchObject({ status: false, message: '请求失败' });
  expect(snapshot()).toEqual(before);
});

it.each(['info', 'tags', 'actors'] as const)('%s 抛出异常保留原库的全部状态', async (endpoint) => {
  const before = snapshot();
  api[endpoint].mockRejectedValueOnce(new Error('网络异常'));
  expect(await appStoreData().initCurrentFilesBases('new')).toMatchObject({ status: false, message: '网络异常' });
  expect(snapshot()).toEqual(before);
});

it('配置解析失败不提交部分状态', async () => {
  const before = snapshot();
  const response = library('new');
  response.data.filesBasesSetting.config_json_data = '{';
  api.info.mockResolvedValueOnce(response);
  expect((await appStoreData().initCurrentFilesBases('new')).status).toBe(false);
  expect(snapshot()).toEqual(before);
});

it('所有依赖成功后提交，同步库监听器能读取新配置和标签', async () => {
  const pending = deferred<ReturnType<typeof actors>>();
  api.actors.mockReturnValueOnce(pending.promise);
  const before = snapshot();
  const store = appStoreData();
  const observed = vi.fn(() => [store.currentMainPerformerBasesId, store.currentTag[0].id, store.currentConfigApp.performerShowNum]);
  const stop = watch(() => store.currentFilesBases.id, observed, { flush: 'sync' });
  const switching = store.initCurrentFilesBases('new');
  await vi.waitFor(() => expect(api.tags).toHaveBeenCalledWith('new', false));
  expect(snapshot()).toEqual(before);
  pending.resolve(actors('new'));
  expect((await switching).status).toBe(true);
  expect(store.currentFilesBases.id).toBe('new');
  expect(store.currentTopPreferredPerformers[0].id).toBe('new');
  expect(observed.mock.results[0].value).toEqual(['new-actors', 'new', 25]);
  stop();
});

it.each([false, true])('信息请求乱序不会覆盖最新库，旧请求异常=%s', async (reject) => {
  const pending = deferred<ReturnType<typeof library>>();
  api.info.mockReturnValueOnce(pending.promise);
  const oldRequest = appStoreData().initCurrentFilesBases('A');
  await appStoreData().initCurrentFilesBases('B');
  const before = snapshot();
  if (reject) pending.reject(new Error('旧请求失败')); else pending.resolve(library('A'));
  expect(await oldRequest).toMatchObject({ stale: true });
  expect(snapshot()).toEqual(before);
});

it('同一个库重复请求也只接受最新结果', async () => {
  const pending = deferred<ReturnType<typeof tags>>();
  api.tags.mockReturnValueOnce(pending.promise);
  const oldRequest = appStoreData().initCurrentFilesBases('new');
  await vi.waitFor(() => expect(api.tags).toHaveBeenCalledWith('new', false));
  await appStoreData().initCurrentFilesBases('new');
  pending.resolve(tags('stale'));
  expect(await oldRequest).toMatchObject({ stale: true });
  expect(appStoreData().currentTag[0].id).toBe('new');
});

it('最后一次切换失败也不会让较早请求偷偷切库', async () => {
  const before = snapshot();
  const pending = deferred<ReturnType<typeof tags>>();
  api.tags.mockReturnValueOnce(pending.promise);
  const oldRequest = appStoreData().initCurrentFilesBases('A');
  await vi.waitFor(() => expect(api.tags).toHaveBeenCalledWith('A', false));
  api.info.mockResolvedValueOnce({ status: false, msg: '失败' });
  expect((await appStoreData().initCurrentFilesBases('B')).status).toBe(false);
  pending.resolve(tags('A'));
  expect(await oldRequest).toMatchObject({ stale: true });
  expect(snapshot()).toEqual(before);
});

it('切库失败后迟到的标签结果不改变旧库', async () => {
  const before = snapshot();
  const pending = deferred<ReturnType<typeof tags>>();
  api.tags.mockReturnValueOnce(pending.promise);
  api.actors.mockRejectedValueOnce(new Error('失败'));
  expect((await appStoreData().initCurrentFilesBases('new')).status).toBe(false);
  pending.resolve(tags('new'));
  await pending.promise;
  expect(snapshot()).toEqual(before);
});

it('同库重新加载后不被此前的标签刷新覆盖', async () => {
  const pending = deferred<ReturnType<typeof tags>>();
  api.tags.mockReturnValueOnce(pending.promise);
  const refreshing = appStoreData().refreshCurrentTagData();
  await appStoreData().initCurrentFilesBases('old');
  pending.resolve(tags('stale'));
  await refreshing;
  expect(appStoreData().currentTag[0].id).toBe('old');
});

it.each([false, true])('无主演员库时使用本库首项，没有关联时清空：%s', async (empty) => {
  const response = library('new');
  response.data.filesRelatedPerformerBases = empty ? [] : [{ main: false, performerBases_id: 'fallback' }];
  api.info.mockResolvedValueOnce(response);
  await appStoreData().initCurrentFilesBases('new');
  expect(appStoreData().currentMainPerformerBasesId).toBe(empty ? '' : 'fallback');
});
