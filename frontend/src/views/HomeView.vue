<template>
  <section class="page-section">
    <div class="section-heading">
      <p class="eyebrow">第 1 阶段</p>
      <h2>工程骨架已就绪</h2>
      <p>当前版本提供前端入口、路由、状态管理和后端健康检查连接，为后续导入和转换流程打底。</p>
    </div>

    <div class="status-panel">
      <div>
        <span class="label">后端服务</span>
        <strong>{{ healthLabel }}</strong>
      </div>
      <button type="button" @click="checkHealth">刷新</button>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue';

import { useSystemStore } from '../stores/system';

const systemStore = useSystemStore();

const healthLabel = computed(() => {
  if (systemStore.healthStatus === 'healthy') {
    return '运行正常';
  }

  if (systemStore.healthStatus === 'unhealthy') {
    return '暂不可用';
  }

  return '检查中';
});

const checkHealth = () => {
  void systemStore.fetchHealth();
};

onMounted(checkHealth);
</script>
