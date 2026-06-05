import { defineStore } from 'pinia';

import {
  createConversionTask,
  getConversionTask,
  type ConversionTask,
  type ParsedChapter
} from '../services/api';

export const useConversionTaskStore = defineStore('conversionTask', {
  state: () => ({
    currentTask: null as ConversionTask | null,
    isCreating: false,
    isLoading: false,
    errorMessage: ''
  }),
  actions: {
    async create(sourceText: string, chapters: ParsedChapter[]) {
      this.isCreating = true;
      this.errorMessage = '';

      try {
        const task = await createConversionTask(sourceText, chapters);
        this.currentTask = task;
        return task;
      } catch {
        this.errorMessage = '转换任务创建失败，请稍后重试。';
        return null;
      } finally {
        this.isCreating = false;
      }
    },
    async fetch(taskId: string) {
      this.isLoading = true;
      this.errorMessage = '';

      try {
        const task = await getConversionTask(taskId);
        this.currentTask = task;
        return task;
      } catch {
        this.errorMessage = '转换任务查询失败，请稍后重试。';
        return null;
      } finally {
        this.isLoading = false;
      }
    }
  }
});
