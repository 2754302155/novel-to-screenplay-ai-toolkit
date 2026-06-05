import { defineStore } from 'pinia';

import { parseChapters, type ParseChaptersResponse } from '../services/api';

export const useImportSessionStore = defineStore('importSession', {
  state: () => ({
    rawText: '',
    parseResult: null as ParseChaptersResponse | null,
    isParsing: false,
    errorMessage: ''
  }),
  getters: {
    charCount: (state) => Array.from(state.rawText.trim()).length,
    canParse: (state) => state.rawText.trim().length > 0 && !state.isParsing,
    canConfirm: (state) => {
      const result = state.parseResult;
      return Boolean(result && result.chapters.length >= 3 && !result.blocking_errors?.length);
    }
  },
  actions: {
    setText(text: string) {
      this.rawText = text;
      this.errorMessage = '';
      this.parseResult = null;
    },
    clear() {
      this.rawText = '';
      this.parseResult = null;
      this.errorMessage = '';
    },
    async parse() {
      if (!this.rawText.trim()) {
        this.errorMessage = '请先粘贴或上传小说正文。';
        return null;
      }

      this.isParsing = true;
      this.errorMessage = '';

      try {
        const result = await parseChapters(this.rawText);
        this.parseResult = result;
        if (result.blocking_errors?.length) {
          this.errorMessage = result.blocking_errors.map((issue) => issue.message).join(' ');
        }
        return result;
      } catch {
        this.errorMessage = '章节解析暂时不可用，请稍后重试。';
        return null;
      } finally {
        this.isParsing = false;
      }
    }
  }
});
