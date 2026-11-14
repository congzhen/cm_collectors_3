import { mount, flushPromises } from '@vue/test-utils';
import { beforeEach, afterEach, describe, expect, it, vi } from 'vitest';
import ImportDialog from '../modeScanDiskImportDataDialog.vue';
import ScraperDialog from '../scraperDataProcessDialog.vue';
import { defualtConfigScanDisk, defualtConfigScraperData } from '@/dataType/config.dataType';
import { importDataServer } from '@/server/importData.server';
import { scraperDataServer } from '@/server/scraper.server';

vi.mock('@/assets/debounce', () => ({ debounceNow: (fn: unknown) => fn }));
vi.mock('@/components/com/dialog/dialog-common.vue', () => ({ default: { name: 'dialogCommon', template: '<div />' } }));
vi.mock('@/server/importData.server', () => ({ importDataServer: { scanDiskImportData: vi.fn(), batch: vi.fn() } }));
vi.mock('@/server/scraper.server', () => ({ scraperDataServer: { scraperDataProcess: vi.fn() } }));
vi.mock('element-plus', () => ({ ElMessageBox: { alert: vi.fn() }, ElTable: { name: 'ElTable', template: '<div />' } }));

const options = { global: { stubs: {
  ElButton: true, ElTableColumn: true, ElIcon: true, Loading: true, Select: true, CloseBold: true, Paperclip: true,
  dialogCommon: { name: 'TaskDialogStub', emits: ['submit', 'closed'], methods: { open() {}, disabledSubmit() {} }, template: '<div><button @click="$emit(\'submit\')">开始</button><slot /></div>' },
  ElTable: { methods: { getSelectionRows: () => [], toggleAllSelection() {} }, template: '<div />' },
} } };

describe('批量任务配置快照', () => {
  beforeEach(() => {
    vi.resetAllMocks();
    vi.mocked(importDataServer.batch).mockResolvedValue({ status: true, statusCode: 0, msg: '', data: 'batch-1' });
    vi.mocked(importDataServer.scanDiskImportData).mockResolvedValue({ status: true, statusCode: 200, msg: '', data: true });
    vi.mocked(scraperDataServer.scraperDataProcess).mockResolvedValue({ status: true, statusCode: 200, msg: '', data: true });
  });
  afterEach(() => vi.useRealTimers());

  it('导入固定建任务时的库、目录和嵌套 NFO 配置，不引用页面草稿', async () => {
    const w = mount(ImportDialog, options);
    const draft = JSON.parse(JSON.stringify(defualtConfigScanDisk));
    draft.scanDiskPaths = ['A-path']; draft.nfo.titles = ['original-title'];
    w.vm.open(['A-file-1', 'A-file-2'], draft, 'A');
    draft.scanDiskPaths[0] = 'B-path'; draft.nfo.titles[0] = 'changed-title';
    await w.get('button').trigger('click'); await flushPromises();
    expect(importDataServer.scanDiskImportData).toHaveBeenCalledTimes(2);
    for (const [id, , config] of vi.mocked(importDataServer.scanDiskImportData).mock.calls) {
      expect(id).toBe('A'); expect(config.scanDiskPaths).toEqual(['A-path']); expect(config.nfo.titles).toEqual(['original-title']);
    }
    w.unmount();
  });

  it('刮削固定目标库和参数，并发执行期间修改草稿不改变任务', async () => {
    vi.useFakeTimers();
    const w = mount(ScraperDialog, options);
    const draft = { ...defualtConfigScraperData, scanDiskPaths: ['A-path'], scraperConfigs: ['original'], concurrency: 1 };
    w.vm.open(['A-file-1', 'A-file-2'], draft, 'A');
    draft.scraperConfigs[0] = 'changed'; draft.timeout = 120;
    await w.get('button').trigger('click');
    await vi.advanceTimersByTimeAsync(5000);
    expect(scraperDataServer.scraperDataProcess).toHaveBeenCalledTimes(2);
    for (const [id, , config] of vi.mocked(scraperDataServer.scraperDataProcess).mock.calls) {
      expect(id).toBe('A'); expect(config.scraperConfigs).toEqual(['original']); expect(config.timeout).toBe(30);
    }
    w.unmount();
  });
  it('关闭后重新打开刮削列表，旧批次等待中的请求不会混入新批次', async () => {
    vi.useFakeTimers();
    const w = mount(ScraperDialog, options);
    w.vm.open(['old-file'], { ...defualtConfigScraperData, scraperConfigs: ['old'] }, 'A');
    await w.get('button').trigger('click');
    w.getComponent({ name: 'TaskDialogStub' }).vm.$emit('closed');
    w.vm.open(['new-file'], { ...defualtConfigScraperData, scraperConfigs: ['new'] }, 'B');
    await w.get('button').trigger('click');
    await vi.advanceTimersByTimeAsync(5000);
    expect(scraperDataServer.scraperDataProcess).toHaveBeenCalledTimes(1);
    expect(scraperDataServer.scraperDataProcess).toHaveBeenCalledWith('B', 'new-file', expect.objectContaining({ scraperConfigs: ['new'] }));
    w.unmount();
  });
});
