import { createPinia, setActivePinia } from 'pinia';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { useAISettingsStore } from './aiSettings';

vi.mock('../services/api', () => ({
  testAIConnection: vi.fn(async () => ({
    ok: true,
    message: 'AI 联通测试成功。'
  }))
}));

describe('ai settings store', () => {
  beforeEach(() => {
    window.localStorage.clear();
    setActivePinia(createPinia());
  });

  it('persists settings and tests connection', async () => {
    const store = useAISettingsStore();

    store.update({
      base_url: 'https://api.example.com/v1',
      model: 'demo-model',
      api_key: 'secret'
    });

    const ok = await store.test();

    expect(ok).toBe(true);
    expect(store.testOK).toBe(true);
    expect(JSON.parse(window.localStorage.getItem('novel-to-screenplay-ai-settings') ?? '{}').model).toBe(
      'demo-model'
    );
  });
});
