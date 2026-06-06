<template>
  <section class="page-section">
    <div class="section-heading">
      <p class="eyebrow">进度</p>
      <h2>转换任务进度</h2>
      <p>系统正在为后续 AI 剧本生成准备任务状态。当前阶段会展示任务创建、处理和校验进度。</p>
    </div>

    <div v-if="taskStore.errorMessage" class="alert alert-error">
      {{ taskStore.errorMessage }}
    </div>

    <div v-if="task" class="task-progress">
      <div class="task-progress-header">
        <div>
          <span class="label">任务 ID</span>
          <strong>{{ task.id }}</strong>
        </div>
        <span class="status-badge">{{ statusLabel }}</span>
      </div>

      <div class="progress-track" aria-label="转换任务进度">
        <div class="progress-bar" :style="{ width: `${task.progress}%` }" />
      </div>

      <div class="task-meta">
        <div>
          <span class="label">当前进度</span>
          <strong>{{ task.progress }}%</strong>
        </div>
        <div>
          <span class="label">章节数量</span>
          <strong>{{ task.chapters.length }}</strong>
        </div>
        <div>
          <span class="label">文本块</span>
          <strong>{{ chunkProgress }}</strong>
        </div>
        <div>
          <span class="label">当前阶段</span>
          <strong>{{ task.stage }}</strong>
        </div>
        <div v-if="task.current_chunk">
          <span class="label">当前文本块</span>
          <strong>{{ task.current_chunk }}</strong>
        </div>
      </div>

      <div v-if="task.status === 'failed' && task.error_message" class="alert alert-error">
        {{ task.error_message }}
      </div>

      <div v-if="qualityReport" class="quality-report">
        <div class="quality-report-header">
          <div>
            <span class="label">质量报告</span>
            <strong>覆盖率与人工确认项</strong>
          </div>
        </div>

        <div class="quality-grid">
          <div>
            <span class="label">已覆盖章节</span>
            <strong>{{ qualityReport.coverage.converted_chapters }}</strong>
          </div>
          <div>
            <span class="label">未处理比例</span>
            <strong>{{ unconvertedRate }}</strong>
          </div>
          <div>
            <span class="label">警告数量</span>
            <strong>{{ qualityReport.warnings.length }}</strong>
          </div>
          <div>
            <span class="label">人工确认项</span>
            <strong>{{ qualityReport.human_review_required.length }}</strong>
          </div>
        </div>

        <div class="quality-columns">
          <div>
            <span class="field-label">警告</span>
            <ul v-if="qualityReport.warnings.length > 0">
              <li v-for="warning in qualityReport.warnings" :key="warning">{{ warning }}</li>
            </ul>
            <p v-else>暂无警告。</p>
          </div>
          <div>
            <span class="field-label">人工确认项</span>
            <ul v-if="qualityReport.human_review_required.length > 0">
              <li v-for="item in qualityReport.human_review_required" :key="item">{{ item }}</li>
            </ul>
            <p v-else>暂无人工确认项。</p>
          </div>
        </div>
      </div>

      <div v-if="task.status === 'completed' && task.yaml" class="yaml-preview">
        <div class="yaml-preview-header">
          <div>
            <span class="label">生成结果</span>
            <strong>YAML 剧本初稿编辑</strong>
          </div>
          <div class="yaml-actions">
            <button type="button" :disabled="isValidating || !yamlText.trim()" @click="runValidation">
              {{ isValidating ? '校验中' : '重新校验' }}
            </button>
            <button type="button" :disabled="!yamlText.trim()" @click="downloadYAML">下载当前 YAML</button>
            <a class="button-link" :href="yamlHelpPDFUrl" download>下载 YAML 帮助 PDF</a>
            <button type="button" @click="resetYAML">恢复生成结果</button>
          </div>
        </div>
        <textarea
          v-model="yamlText"
          class="yaml-editor"
          spellcheck="false"
          aria-label="YAML 剧本初稿编辑器"
        />

        <p v-if="editorMessage" class="editor-message">{{ editorMessage }}</p>

        <div
          v-if="validationResult"
          class="validation-panel"
          :class="{ success: validationResult.valid, error: !validationResult.valid }"
        >
          <div class="validation-summary">
            <div>
              <span class="label">校验结果</span>
              <strong>{{ validationResult.valid ? 'YAML 结构有效' : '发现需要修复的问题' }}</strong>
            </div>
            <span class="status-badge compact">{{ validationResult.issues.length }} 项</span>
          </div>

          <ul v-if="validationResult.issues.length > 0" class="validation-issues">
            <li v-for="issue in validationResult.issues" :key="`${issue.path}-${issue.message}`">
              <span>{{ issue.path || 'root' }}</span>
              <strong>{{ issue.message }}</strong>
            </li>
          </ul>

          <div v-if="validationResult.quality_report" class="quality-grid compact-grid">
            <div>
              <span class="label">已覆盖章节</span>
              <strong>{{ validationResult.quality_report.coverage.converted_chapters }}</strong>
            </div>
            <div>
              <span class="label">未处理比例</span>
              <strong>{{ validationUnconvertedRate }}</strong>
            </div>
            <div>
              <span class="label">警告数量</span>
              <strong>{{ validationResult.quality_report.warnings.length }}</strong>
            </div>
            <div>
              <span class="label">人工确认项</span>
              <strong>{{ validationResult.quality_report.human_review_required.length }}</strong>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-else class="empty-state">
      <strong>{{ taskStore.isLoading ? '正在加载任务' : '暂无任务数据' }}</strong>
      <p>如果任务不存在，请返回导入页重新创建转换任务。</p>
      <RouterLink class="primary-link" to="/">返回导入</RouterLink>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useRoute } from 'vue-router';

import { validateYAML, type ValidateYAMLResponse } from '../services/api';
import { useConversionTaskStore } from '../stores/conversionTask';

const route = useRoute();
const taskStore = useConversionTaskStore();
let timer: number | undefined;

const yamlText = ref('');
const editorMessage = ref('');
const validationResult = ref<ValidateYAMLResponse | null>(null);
const isValidating = ref(false);
const yamlHelpPDFUrl = '/docs/yaml-screenplay-guide.pdf';
const taskId = computed(() => String(route.params.id ?? ''));
const task = computed(() => taskStore.currentTask);
const statusLabel = computed(() => {
  switch (task.value?.status) {
    case 'processing':
      return '处理中';
    case 'validating':
      return '校验中';
    case 'completed':
      return '已完成';
    case 'failed':
      return '失败';
    default:
      return '等待中';
  }
});
const chunkProgress = computed(() => {
  if (!task.value?.total_chunks) {
    return '-';
  }
  return `${task.value.completed_chunks ?? 0}/${task.value.total_chunks}`;
});
const qualityReport = computed(() => task.value?.draft?.quality_report);
const unconvertedRate = computed(() => {
  const rate = qualityReport.value?.coverage.estimated_unconverted_ratio ?? 0;
  return `${Math.round(rate * 100)}%`;
});
const validationUnconvertedRate = computed(() => {
  const rate = validationResult.value?.quality_report.coverage.estimated_unconverted_ratio ?? 0;
  return `${Math.round(rate * 100)}%`;
});

watch(
  () => task.value?.yaml,
  (yaml) => {
    if (yaml && yamlText.value === '') {
      yamlText.value = yaml;
    }
  },
  { immediate: true }
);

const resetYAML = () => {
  yamlText.value = task.value?.yaml ?? '';
  validationResult.value = null;
  editorMessage.value = '已恢复为 AI 生成的原始 YAML。';
};

const runValidation = async () => {
  if (!yamlText.value.trim()) {
    validationResult.value = null;
    editorMessage.value = 'YAML 内容为空，无法校验。';
    return;
  }

  isValidating.value = true;
  editorMessage.value = '';
  try {
    validationResult.value = await validateYAML(yamlText.value);
    editorMessage.value = validationResult.value.valid
      ? 'YAML 校验通过，可下载或继续编辑。'
      : 'YAML 校验发现问题，请按提示修正后再次校验。';
  } catch {
    validationResult.value = null;
    editorMessage.value = 'YAML 校验请求失败，请稍后重试。';
  } finally {
    isValidating.value = false;
  }
};

const downloadYAML = () => {
  if (!yamlText.value.trim()) {
    editorMessage.value = 'YAML 内容为空，无法下载。';
    return;
  }

  const blob = new Blob([yamlText.value], { type: 'application/x-yaml;charset=utf-8' });
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = `screenplay-${task.value?.id ?? 'draft'}.yaml`;
  link.click();
  URL.revokeObjectURL(url);
  editorMessage.value = '已下载当前编辑内容。';
};

const fetchTask = () => {
  if (taskId.value) {
    void taskStore.fetch(taskId.value).then((task) => {
      if ((task?.status === 'completed' || task?.status === 'failed') && timer) {
        window.clearInterval(timer);
        timer = undefined;
      }
    });
  }
};

onMounted(() => {
  fetchTask();
  timer = window.setInterval(fetchTask, 1500);
});

onBeforeUnmount(() => {
  if (timer) {
    window.clearInterval(timer);
  }
});
</script>
