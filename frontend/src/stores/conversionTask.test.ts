import { createPinia, setActivePinia } from 'pinia';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { useConversionTaskStore } from './conversionTask';

vi.mock('../services/api', () => ({
  createConversionTask: vi.fn(async () => ({
    id: 'task-001',
    status: 'pending',
    progress: 5,
    stage: '任务已创建，等待进入转换流程。',
    chapters: [],
    created_at: '2026-06-05T10:00:00Z',
    updated_at: '2026-06-05T10:00:00Z'
  })),
  getConversionTask: vi.fn(async () => ({
    id: 'task-001',
    status: 'processing',
    progress: 45,
    stage: '正在整理章节内容并准备剧本转换。',
    chapters: [],
    created_at: '2026-06-05T10:00:00Z',
    updated_at: '2026-06-05T10:00:02Z'
  }))
}));

describe('conversion task store', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it('creates and fetches a conversion task', async () => {
    const store = useConversionTaskStore();

    await store.create('正文', []);
    expect(store.currentTask?.id).toBe('task-001');

    await store.fetch('task-001');
    expect(store.currentTask?.status).toBe('processing');
  });
});
