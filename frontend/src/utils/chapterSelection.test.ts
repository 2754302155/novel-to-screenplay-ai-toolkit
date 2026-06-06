import { describe, expect, it } from 'vitest';

import {
  buildSelectedChapterSourceText,
  isChapterInRange,
  MIN_SELECTED_CHAPTERS,
  normalizeChapterRange
} from './chapterSelection';

describe('chapter selection utilities', () => {
  it('keeps a continuous range with at least three chapters', () => {
    const range = normalizeChapterRange(8, 4, 4);

    expect(range).toMatchObject({
      start: 4,
      end: 6,
      count: MIN_SELECTED_CHAPTERS,
      isValid: true
    });
  });

  it('normalizes reversed chapter boundaries', () => {
    const range = normalizeChapterRange(8, 6, 2);

    expect(range.start).toBe(2);
    expect(range.end).toBe(6);
    expect(isChapterInRange(4, range)).toBe(true);
    expect(isChapterInRange(7, range)).toBe(false);
  });

  it('marks ranges below the minimum as invalid when total chapters are insufficient', () => {
    const range = normalizeChapterRange(2, 0, 1);

    expect(range.count).toBe(2);
    expect(range.isValid).toBe(false);
  });

  it('builds source text from selected chapters only', () => {
    const sourceText = buildSelectedChapterSourceText([
      { title: '第一章', body: '正文一' },
      { title: '第二章', body: '正文二' }
    ]);

    expect(sourceText).toBe('第一章\n正文一\n\n第二章\n正文二');
  });
});
