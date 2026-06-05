<template>
  <section class="page-section">
    <div class="section-heading">
      <p class="eyebrow">任务</p>
      <h2>转换任务</h2>
      <p>查看已创建的小说转剧本任务，进入详情页跟踪 AI 生成进度和 YAML 结果。</p>
    </div>

    <div v-if="taskStore.errorMessage" class="alert alert-error">
      {{ taskStore.errorMessage }}
    </div>

    <div v-if="taskStore.tasks.length > 0" class="table-wrap">
      <table>
        <thead>
          <tr>
            <th>任务 ID</th>
            <th>状态</th>
            <th>进度</th>
            <th>章节</th>
            <th>当前阶段</th>
            <th>创建时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="task in taskStore.tasks" :key="task.id">
            <td>
              <code class="task-id">{{ task.id }}</code>
            </td>
            <td>
              <span class="status-badge compact">{{ statusLabel(task.status) }}</span>
            </td>
            <td>{{ task.progress }}%</td>
            <td>{{ task.chapters.length }}</td>
            <td>{{ task.stage }}</td>
            <td>{{ formatTime(task.created_at) }}</td>
            <td>
              <RouterLink class="secondary-link" :to="`/tasks/${task.id}`">查看</RouterLink>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-else class="empty-state">
      <strong>{{ taskStore.isLoading ? '正在加载任务' : '暂无转换任务' }}</strong>
      <p>完成小说导入和章节确认后，创建的转换任务会显示在这里。</p>
      <RouterLink class="primary-link" to="/">导入小说</RouterLink>
    </div>
  </section>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted } from 'vue';

import { useConversionTaskStore } from '../stores/conversionTask';
import type { ConversionTask } from '../services/api';

const taskStore = useConversionTaskStore();
let timer: number | undefined;

const statusLabel = (status: ConversionTask['status']) => {
  switch (status) {
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
};

const formatTime = (value: string) => {
  return new Intl.DateTimeFormat('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  }).format(new Date(value));
};

onMounted(() => {
  void taskStore.fetchList();
  timer = window.setInterval(() => {
    void taskStore.fetchList();
  }, 3000);
});

onBeforeUnmount(() => {
  if (timer) {
    window.clearInterval(timer);
  }
});
</script>
