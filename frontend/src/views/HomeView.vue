<template>
  <section class="page-section">
    <div class="section-heading">
      <p class="eyebrow">导入</p>
      <h2>小说文本导入</h2>
      <p>粘贴或上传 3 个章节以上的中文小说正文，系统会先解析章节边界、标题和字数。</p>
    </div>

    <div class="import-layout">
      <div class="import-main">
        <label class="field-label" for="novelText">小说正文</label>
        <textarea
          id="novelText"
          v-model="text"
          class="text-input"
          placeholder="第一章&#10;这里粘贴小说正文..."
        />
        <div class="toolbar">
          <label class="file-button" for="novelFile">上传 TXT</label>
          <input id="novelFile" class="file-input" type="file" accept=".txt,text/plain" @change="readFile" />
          <button type="button" :disabled="!store.canParse" @click="parseText">
            {{ store.isParsing ? '解析中' : '解析章节' }}
          </button>
          <button type="button" :disabled="store.isParsing || !store.rawText" @click="store.clear()">清空</button>
        </div>
      </div>

      <aside class="import-summary">
        <span class="label">当前字数</span>
        <strong>{{ store.charCount }}</strong>
        <p>建议单次导入 3 至 10 章。少于 3 章时不会进入转换流程。</p>
      </aside>
    </div>

    <div class="ai-settings-panel">
      <div class="section-heading compact">
        <p class="eyebrow">AI</p>
        <h3>模型设置</h3>
      </div>
      <div class="settings-grid">
        <label>
          <span class="field-label">Base URL</span>
          <input
            :value="aiSettings.settings.base_url"
            class="text-field"
            placeholder="https://api.openai.com/v1"
            @input="updateAISetting('base_url', $event)"
          />
        </label>
        <label>
          <span class="field-label">模型</span>
          <input
            :value="aiSettings.settings.model"
            class="text-field"
            placeholder="gpt-4.1-mini"
            @input="updateAISetting('model', $event)"
          />
        </label>
        <label class="settings-wide">
          <span class="field-label">API Key</span>
          <input
            :value="aiSettings.settings.api_key"
            class="text-field"
            type="password"
            placeholder="sk-..."
            autocomplete="off"
            @input="updateAISetting('api_key', $event)"
          />
        </label>
      </div>
      <div class="toolbar">
        <button type="button" :disabled="aiSettings.isTesting" @click="aiSettings.test()">
          {{ aiSettings.isTesting ? '测试中' : '测试联通' }}
        </button>
        <span v-if="aiSettings.testMessage" :class="aiSettings.testOK ? 'inline-success' : 'inline-error'">
          {{ aiSettings.testMessage }}
        </span>
      </div>
    </div>

    <div v-if="store.errorMessage" class="alert alert-error">
      {{ store.errorMessage }}
    </div>

    <div v-if="fileError" class="alert alert-error">
      {{ fileError }}
    </div>

    <div v-if="store.parseResult" class="parse-preview">
      <div>
        <span class="label">识别章节</span>
        <strong>{{ store.parseResult.chapters.length }} 章</strong>
      </div>
      <div>
        <span class="label">清洗后字数</span>
        <strong>{{ store.parseResult.cleaned_chars }}</strong>
      </div>
      <div>
        <span class="label">中文比例</span>
        <strong>{{ chineseRatioLabel }}</strong>
      </div>
      <RouterLink v-if="store.canConfirm" class="primary-link" to="/chapters/confirm">确认章节顺序</RouterLink>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import { useRouter } from 'vue-router';

import { useAISettingsStore } from '../stores/aiSettings';
import { useImportSessionStore } from '../stores/importSession';

const store = useImportSessionStore();
const aiSettings = useAISettingsStore();
const router = useRouter();
const fileError = ref('');
const maxUploadSizeMB = 10;
const maxUploadSizeBytes = maxUploadSizeMB * 1024 * 1024;

const text = computed({
  get: () => store.rawText,
  set: (value: string) => store.setText(value)
});

const chineseRatioLabel = computed(() => {
  const ratio = store.parseResult?.chinese_ratio ?? 0;
  return `${Math.round(ratio * 100)}%`;
});

const parseText = async () => {
  const result = await store.parse();
  if (result && !result.blocking_errors?.length && result.chapters.length >= 3) {
    void router.push('/chapters/confirm');
  }
};

const readFile = async (event: Event) => {
  fileError.value = '';
  const input = event.target as HTMLInputElement;
  const file = input.files?.[0];
  if (!file) {
    return;
  }

  if (!file.name.endsWith('.txt') && file.type !== 'text/plain') {
    fileError.value = '请上传 TXT 文本文件。';
    input.value = '';
    return;
  }

  if (file.size > maxUploadSizeBytes) {
    fileError.value = `文件超过 ${maxUploadSizeMB}MB，请拆分后再上传。`;
    input.value = '';
    return;
  }

  store.setText(await file.text());
  input.value = '';
};

const updateAISetting = (key: 'base_url' | 'model' | 'api_key', event: Event) => {
  aiSettings.update({ [key]: (event.target as HTMLInputElement).value });
};
</script>
