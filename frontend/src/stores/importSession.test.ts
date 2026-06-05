import { createPinia, setActivePinia } from 'pinia';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { useImportSessionStore } from './importSession';

vi.mock('../services/api', () => ({
  parseChapters: vi.fn(async () => ({
    chapters: [
      { id: 'CH001', title: '第一章', word_count: 10, body: '正文', inferred_title: false },
      { id: 'CH002', title: '第二章', word_count: 10, body: '正文', inferred_title: false },
      { id: 'CH003', title: '第三章', word_count: 10, body: '正文', inferred_title: false }
    ],
    cleaned_text: '正文',
    original_chars: 30,
    cleaned_chars: 30,
    chinese_ratio: 1,
    warnings: [],
    blocking_errors: []
  }))
}));

describe('import session store', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it('parses text and allows chapter confirmation', async () => {
    const store = useImportSessionStore();
    store.setText('第一章\n正文\n第二章\n正文\n第三章\n正文');

    await store.parse();

    expect(store.parseResult?.chapters).toHaveLength(3);
    expect(store.canConfirm).toBe(true);
  });
});
