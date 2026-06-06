<template>
  <section class="page-section">
    <div class="section-heading">
      <p class="eyebrow">确认</p>
      <h2>章节解析结果</h2>
      <p>转换前请确认章节顺序、标题和字数，避免复制遗漏导致剧情顺序混乱。</p>
    </div>

    <div v-if="!store.parseResult" class="empty-state">
      <strong>还没有解析结果</strong>
      <p>请先导入小说正文并完成章节解析。</p>
      <RouterLink class="primary-link" to="/">返回导入</RouterLink>
    </div>

    <template v-else>
      <div v-if="store.parseResult.blocking_errors?.length" class="alert alert-error">
        {{ store.parseResult.blocking_errors.map((issue) => issue.message).join(' ') }}
      </div>

      <div v-if="store.parseResult.warnings.length" class="alert alert-warning">
        {{ store.parseResult.warnings.map((issue) => issue.message).join(' ') }}
      </div>

      <div class="confirm-summary">
        <div>
          <span class="label">章节数量</span>
          <strong>{{ store.parseResult.chapters.length }}</strong>
        </div>
        <div>
          <span class="label">原始字数</span>
          <strong>{{ store.parseResult.original_chars }}</strong>
        </div>
        <div>
          <span class="label">清洗后字数</span>
          <strong>{{ store.parseResult.cleaned_chars }}</strong>
        </div>
      </div>

      <div class="chapter-selection-panel">
        <div>
          <span class="label">生成范围</span>
          <strong>{{ selectedRangeLabel }}</strong>
        </div>
        <div>
          <span class="label">所选章节</span>
          <strong>{{ selectedChapterCount }}</strong>
        </div>
        <div>
          <span class="label">所选字数</span>
          <strong :class="{ 'over-limit-text': isSelectedWordLimitExceeded }">
            {{ selectedWordCount }} / {{ maxSelectedWords }}
          </strong>
        </div>
        <div class="selection-controls">
          <label>
            <span class="field-label">起始章节</span>
            <select v-model.number="startIndexModel" class="select-field">
              <option
                v-for="(chapter, index) in store.parseResult.chapters"
                :key="`start-${chapter.id}`"
                :value="index"
                :disabled="index > maxStartIndex"
              >
                {{ chapter.id }} {{ chapter.title }}
              </option>
            </select>
          </label>
          <label>
            <span class="field-label">结束章节</span>
            <select v-model.number="endIndexModel" class="select-field">
              <option
                v-for="(chapter, index) in store.parseResult.chapters"
                :key="`end-${chapter.id}`"
                :value="index"
                :disabled="index < minEndIndex"
              >
                {{ chapter.id }} {{ chapter.title }}
              </option>
            </select>
          </label>
        </div>
      </div>

      <div class="pagination-bar">
        <div>
          <span class="label">当前页</span>
          <strong>{{ pageRangeLabel }}</strong>
        </div>
        <div class="pagination-actions">
          <button type="button" :disabled="currentPage <= 1" @click="goToPreviousPage">上一页</button>
          <span>{{ currentPage }} / {{ totalPages }}</span>
          <button type="button" :disabled="currentPage >= totalPages" @click="goToNextPage">下一页</button>
        </div>
      </div>

      <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>选择</th>
              <th>编号</th>
              <th>标题</th>
              <th>字数</th>
              <th>状态</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="item in pagedChapters"
              :key="item.chapter.id"
              :class="{ 'selected-row': isSelectedChapter(item.index) }"
            >
              <td>
                <span class="selection-mark" :class="{ active: isSelectedChapter(item.index) }">
                  {{ isSelectedChapter(item.index) ? '生成' : '-' }}
                </span>
              </td>
              <td>{{ item.chapter.id }}</td>
              <td>{{ item.chapter.title }}</td>
              <td>{{ item.chapter.word_count }}</td>
              <td>{{ item.chapter.inferred_title ? '标题由系统推断' : '已识别标题' }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="toolbar">
        <RouterLink class="secondary-link" to="/">返回修改</RouterLink>
        <button type="button" :disabled="!canCreateSelectedTask" @click="createTask">
          {{ taskStore.isCreating ? '创建中' : '生成所选章节' }}
        </button>
      </div>

      <div v-if="selectionMessage" class="alert alert-warning">
        {{ selectionMessage }}
      </div>

      <div v-if="taskStore.errorMessage" class="alert alert-error">
        {{ taskStore.errorMessage }}
      </div>
    </template>
  </section>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useRouter } from 'vue-router';

import { useAISettingsStore } from '../stores/aiSettings';
import { useConversionTaskStore } from '../stores/conversionTask';
import { useImportSessionStore } from '../stores/importSession';
import {
  buildSelectedChapterSourceText,
  isChapterInRange,
  MAX_SELECTED_WORDS,
  MIN_SELECTED_CHAPTERS,
  normalizeChapterRange,
  sumChapterWordCount
} from '../utils/chapterSelection';

const store = useImportSessionStore();
const taskStore = useConversionTaskStore();
const aiSettings = useAISettingsStore();
const router = useRouter();
const selectedStartIndex = ref(0);
const selectedEndIndex = ref(0);
const currentPage = ref(1);
const pageSize = 30;
const maxSelectedWords = MAX_SELECTED_WORDS;

const chapters = computed(() => store.parseResult?.chapters ?? []);
const chapterCount = computed(() => chapters.value.length);
const totalPages = computed(() => Math.max(1, Math.ceil(chapterCount.value / pageSize)));
const pageStartIndex = computed(() => (currentPage.value - 1) * pageSize);
const pageEndIndex = computed(() => Math.min(chapterCount.value, pageStartIndex.value + pageSize));
const pagedChapters = computed(() =>
  chapters.value.slice(pageStartIndex.value, pageEndIndex.value).map((chapter, offset) => ({
    chapter,
    index: pageStartIndex.value + offset
  }))
);
const pageRangeLabel = computed(() => {
  if (chapterCount.value === 0) {
    return '-';
  }
  return `${pageStartIndex.value + 1}-${pageEndIndex.value} / ${chapterCount.value}`;
});
const maxStartIndex = computed(() => Math.max(0, chapterCount.value - MIN_SELECTED_CHAPTERS));
const minEndIndex = computed(() => Math.min(chapterCount.value - 1, selectedStartIndex.value + MIN_SELECTED_CHAPTERS - 1));
const selectedRange = computed(() =>
  normalizeChapterRange(chapterCount.value, selectedStartIndex.value, selectedEndIndex.value)
);
const selectedChapters = computed(() =>
  chapters.value.slice(selectedRange.value.start, selectedRange.value.end + 1)
);
const selectedChapterCount = computed(() => selectedChapters.value.length);
const selectedWordCount = computed(() => sumChapterWordCount(selectedChapters.value));
const isSelectedWordLimitExceeded = computed(() => selectedWordCount.value > MAX_SELECTED_WORDS);
const selectedRangeLabel = computed(() => {
  if (!selectedRange.value.isValid || selectedChapters.value.length === 0) {
    return '-';
  }

  const first = selectedChapters.value[0];
  const last = selectedChapters.value[selectedChapters.value.length - 1];
  return `${first.id} - ${last.id}`;
});
const selectionMessage = computed(() => {
  if (!store.canConfirm) {
    return '';
  }
  if (!selectedRange.value.isValid) {
    return `至少选择 ${MIN_SELECTED_CHAPTERS} 个连续章节后才能生成剧本。`;
  }
  if (isSelectedWordLimitExceeded.value) {
    return `所选章节字数为 ${selectedWordCount.value}，超过单次生成上限 ${MAX_SELECTED_WORDS} 字，请缩小连续章节范围。`;
  }
  return '';
});
const canCreateSelectedTask = computed(
  () => store.canConfirm && selectedRange.value.isValid && !isSelectedWordLimitExceeded.value && !taskStore.isCreating
);
const startIndexModel = computed({
  get: () => selectedRange.value.start,
  set: (value: number) => {
    selectedStartIndex.value = value;
    if (selectedEndIndex.value < value + MIN_SELECTED_CHAPTERS - 1) {
      selectedEndIndex.value = Math.min(chapterCount.value - 1, value + MIN_SELECTED_CHAPTERS - 1);
    }
  }
});
const endIndexModel = computed({
  get: () => selectedRange.value.end,
  set: (value: number) => {
    selectedEndIndex.value = value;
    if (selectedStartIndex.value > value - MIN_SELECTED_CHAPTERS + 1) {
      selectedStartIndex.value = Math.max(0, value - MIN_SELECTED_CHAPTERS + 1);
    }
  }
});

watch(
  chapterCount,
  (count) => {
    selectedStartIndex.value = 0;
    selectedEndIndex.value = Math.max(0, count - 1);
    currentPage.value = 1;
  },
  { immediate: true }
);

watch(totalPages, (pages) => {
  if (currentPage.value > pages) {
    currentPage.value = pages;
  }
});

const isSelectedChapter = (index: number) => isChapterInRange(index, selectedRange.value);

const goToPreviousPage = () => {
  currentPage.value = Math.max(1, currentPage.value - 1);
};

const goToNextPage = () => {
  currentPage.value = Math.min(totalPages.value, currentPage.value + 1);
};

const createTask = async () => {
  if (!store.parseResult || !canCreateSelectedTask.value) {
    return;
  }

  const sourceText = buildSelectedChapterSourceText(selectedChapters.value) || store.rawText;
  const task = await taskStore.create(
    sourceText,
    selectedChapters.value,
    aiSettings.hasConfig ? aiSettings.sanitizedConfig : undefined
  );
  if (task) {
    void router.push(`/tasks/${task.id}`);
  }
};
</script>
