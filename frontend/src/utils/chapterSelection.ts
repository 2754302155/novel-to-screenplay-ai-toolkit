export const MIN_SELECTED_CHAPTERS = 3;

export interface ChapterRange {
  start: number;
  end: number;
  count: number;
  isValid: boolean;
}

export interface ChapterTextPart {
  title: string;
  body?: string;
}

export function normalizeChapterRange(
  totalChapters: number,
  startIndex: number,
  endIndex: number
): ChapterRange {
  if (totalChapters <= 0) {
    return { start: 0, end: 0, count: 0, isValid: false };
  }

  const maxIndex = totalChapters - 1;
  let start = clampIndex(startIndex, maxIndex);
  let end = clampIndex(endIndex, maxIndex);
  if (end < start) {
    [start, end] = [end, start];
  }

  if (totalChapters >= MIN_SELECTED_CHAPTERS && end - start + 1 < MIN_SELECTED_CHAPTERS) {
    end = Math.min(maxIndex, start + MIN_SELECTED_CHAPTERS - 1);
    start = Math.max(0, end - MIN_SELECTED_CHAPTERS + 1);
  }

  const count = end - start + 1;
  return {
    start,
    end,
    count,
    isValid: totalChapters >= MIN_SELECTED_CHAPTERS && count >= MIN_SELECTED_CHAPTERS
  };
}

export function isChapterInRange(index: number, range: ChapterRange) {
  return range.isValid && index >= range.start && index <= range.end;
}

export function buildSelectedChapterSourceText(chapters: ChapterTextPart[]) {
  return chapters
    .map((chapter) => [chapter.title.trim(), chapter.body?.trim()].filter(Boolean).join('\n'))
    .filter(Boolean)
    .join('\n\n');
}

function clampIndex(index: number, maxIndex: number) {
  if (!Number.isFinite(index)) {
    return 0;
  }
  return Math.min(Math.max(Math.trunc(index), 0), maxIndex);
}
