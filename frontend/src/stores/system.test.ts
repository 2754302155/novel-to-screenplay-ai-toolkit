import { createPinia, setActivePinia } from 'pinia';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { useSystemStore } from './system';

vi.mock('../services/api', () => ({
  getHealth: vi.fn(async () => ({
    status: 'ok',
    service: 'novel-to-screenplay-api',
    version: '0.1.0'
  }))
}));

describe('system store', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it('marks the API as healthy when the health check succeeds', async () => {
    const store = useSystemStore();

    await store.fetchHealth();

    expect(store.healthStatus).toBe('healthy');
    expect(store.apiVersion).toBe('0.1.0');
  });
});
