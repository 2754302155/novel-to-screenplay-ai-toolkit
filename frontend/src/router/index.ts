import { createRouter, createWebHistory } from 'vue-router';

import HomeView from '../views/HomeView.vue';
import ChapterConfirmView from '../views/ChapterConfirmView.vue';
import TaskDetailView from '../views/TaskDetailView.vue';
import TasksView from '../views/TasksView.vue';

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeView
    },
    {
      path: '/tasks',
      name: 'tasks',
      component: TasksView
    },
    {
      path: '/tasks/:id',
      name: 'task-detail',
      component: TaskDetailView
    },
    {
      path: '/chapters/confirm',
      name: 'chapter-confirm',
      component: ChapterConfirmView
    }
  ]
});

export default router;
