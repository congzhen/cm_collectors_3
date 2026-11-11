import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { I_resource } from '@/dataType/resource.dataType';
import { resourceBatchDelete, resourceDelete } from '../resource';

const mocks = vi.hoisted(() => ({
  confirm: vi.fn(), remove: vi.fn(), success: vi.fn(), error: vi.fn(), warning: vi.fn(),
  show: vi.fn(), hide: vi.fn(), refresh: vi.fn(),
}));
vi.mock('@/common/messageBox', () => ({ messageBoxConfirm: mocks.confirm }));
vi.mock('@/server/resource.server', () => ({ resourceServer: { delete: mocks.remove } }));
vi.mock('element-plus', () => ({ ElMessage: { success: mocks.success, error: mocks.error, warning: mocks.warning } }));
vi.mock('@/assets/loading', () => ({ LoadingService: { show: mocks.show, hide: mocks.hide } }));
vi.mock('@/storeData/app.storeData', () => ({ appStoreData: () => ({
  currentFilesBases: { id: 'library' }, refreshCurrentTagData: mocks.refresh,
}) }));

const resources = ['a', 'b', 'c'].map(id => ({ id, title: id, filesBases_id: 'library' }) as I_resource);
const confirm = async () => { await mocks.confirm.mock.calls[0][0].successCallBack(); };
beforeEach(() => { vi.resetAllMocks(); });

describe('删除结果提示', () => {
  it('全部成功只提示一次，列表和标签各刷新一次，关闭遮罩', async () => {
    mocks.remove.mockResolvedValue({ status: true });
    const refreshList = vi.fn();
    resourceBatchDelete(resources, refreshList);
    expect(mocks.remove).not.toHaveBeenCalled();
    await confirm();
    expect(mocks.remove.mock.calls.map(call => call[0])).toEqual(['a', 'b', 'c']);
    expect(mocks.success).toHaveBeenCalledExactlyOnceWith('成功删除 3 项资源');
    expect(mocks.error).not.toHaveBeenCalled();
    expect(mocks.warning).not.toHaveBeenCalled();
    expect(refreshList).toHaveBeenCalledOnce();
    expect(mocks.refresh).toHaveBeenCalledOnce();
    expect(mocks.hide).toHaveBeenCalledOnce();
  });

  it('业务失败和空响应都计入失败，只显示一次部分失败汇总', async () => {
    mocks.remove.mockResolvedValueOnce({ status: false, msg: '失败' })
      .mockResolvedValueOnce(undefined).mockResolvedValueOnce({ status: true });
    resourceBatchDelete(resources);
    await confirm();
    expect(mocks.warning).toHaveBeenCalledExactlyOnceWith('成功删除 1 项资源，失败 2 项');
    expect(mocks.success).not.toHaveBeenCalled();
    expect(mocks.error).not.toHaveBeenCalled();
    expect(mocks.refresh).toHaveBeenCalledOnce();
    expect(mocks.hide).toHaveBeenCalledOnce();
  });

  it('请求抛出异常后继续删除剩余资源，并释放遮罩', async () => {
    const log = vi.spyOn(console, 'error').mockImplementation(() => {});
    try {
      mocks.remove.mockRejectedValueOnce(new Error('网络异常')).mockResolvedValue({ status: true });
      resourceBatchDelete(resources);
      await confirm();
      expect(mocks.remove).toHaveBeenCalledTimes(3);
      expect(mocks.warning).toHaveBeenCalledExactlyOnceWith('成功删除 2 项资源，失败 1 项');
      expect(mocks.hide).toHaveBeenCalledOnce();
    } finally { log.mockRestore(); }
  });

  it('全部失败不刷新标签，使用错误汇总', async () => {
    mocks.remove.mockResolvedValue({ status: false, msg: '失败' });
    resourceBatchDelete(resources);
    await confirm();
    expect(mocks.error).toHaveBeenCalledExactlyOnceWith('删除失败，共 3 项资源');
    expect(mocks.success).not.toHaveBeenCalled();
    expect(mocks.warning).not.toHaveBeenCalled();
    expect(mocks.refresh).not.toHaveBeenCalled();
    expect(mocks.hide).toHaveBeenCalledOnce();
  });

  it('空选择不打开确认框，不请求删除', () => {
    resourceBatchDelete([]);
    expect(mocks.error).toHaveBeenCalledExactlyOnceWith('请选择要删除的资源');
    expect(mocks.confirm).not.toHaveBeenCalled();
    expect(mocks.remove).not.toHaveBeenCalled();
  });

  it('单条删除保留成功提示和刷新', async () => {
    mocks.remove.mockResolvedValue({ status: true });
    const refreshList = vi.fn();
    resourceDelete(resources[0], refreshList);
    await confirm();
    expect(mocks.success).toHaveBeenCalledExactlyOnceWith('删除成功');
    expect(refreshList).toHaveBeenCalledOnce();
    expect(mocks.refresh).toHaveBeenCalledOnce();
  });
});
