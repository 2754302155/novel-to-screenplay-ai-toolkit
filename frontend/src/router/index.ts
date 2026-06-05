import { createRouter, createWebHistory } from 'vue-router';

import HomeView from '../views/HomeView.vue';
import ChapterConfirmView from '../views/ChapterConfirmView.vue';
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
      path: '/chapters/confirm',
      name: 'chapter-confirm',
      component: ChapterConfirmView
    }
  ]
});

export default router;
