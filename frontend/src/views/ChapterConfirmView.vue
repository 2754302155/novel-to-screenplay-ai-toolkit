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

      <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>编号</th>
              <th>标题</th>
              <th>字数</th>
              <th>状态</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="chapter in store.parseResult.chapters" :key="chapter.id">
              <td>{{ chapter.id }}</td>
              <td>{{ chapter.title }}</td>
              <td>{{ chapter.word_count }}</td>
              <td>{{ chapter.inferred_title ? '标题由系统推断' : '已识别标题' }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="toolbar">
        <RouterLink class="secondary-link" to="/">返回修改</RouterLink>
        <button type="button" :disabled="!store.canConfirm || taskStore.isCreating" @click="createTask">
          {{ taskStore.isCreating ? '创建中' : '确认章节' }}
        </button>
      </div>

      <div v-if="taskStore.errorMessage" class="alert alert-error">
        {{ taskStore.errorMessage }}
      </div>
    </template>
  </section>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router';

import { useConversionTaskStore } from '../stores/conversionTask';
import { useImportSessionStore } from '../stores/importSession';

const store = useImportSessionStore();
const taskStore = useConversionTaskStore();
const router = useRouter();

const createTask = async () => {
  if (!store.parseResult || !store.canConfirm) {
    return;
  }

  const task = await taskStore.create(store.rawText, store.parseResult.chapters);
  if (task) {
    void router.push(`/tasks/${task.id}`);
  }
};
</script>
