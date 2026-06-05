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
          <span class="label">当前阶段</span>
          <strong>{{ task.stage }}</strong>
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
import { computed, onBeforeUnmount, onMounted } from 'vue';
import { useRoute } from 'vue-router';

import { useConversionTaskStore } from '../stores/conversionTask';

const route = useRoute();
const taskStore = useConversionTaskStore();
let timer: number | undefined;

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

const fetchTask = () => {
  if (taskId.value) {
    void taskStore.fetch(taskId.value);
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
