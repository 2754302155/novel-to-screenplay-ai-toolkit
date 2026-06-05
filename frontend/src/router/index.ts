import { createRouter, createWebHistory } from 'vue-router';

import HomeView from '../views/HomeView.vue';
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
    }
  ]
});

export default router;
