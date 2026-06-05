import { defineStore } from 'pinia';

import { getHealth } from '../services/api';

type HealthStatus = 'idle' | 'checking' | 'healthy' | 'unhealthy';

export const useSystemStore = defineStore('system', {
  state: () => ({
    healthStatus: 'idle' as HealthStatus,
    apiVersion: ''
  }),
  actions: {
    async fetchHealth() {
      this.healthStatus = 'checking';

      try {
        const health = await getHealth();
        this.apiVersion = health.version;
        this.healthStatus = health.status === 'ok' ? 'healthy' : 'unhealthy';
      } catch {
        this.healthStatus = 'unhealthy';
      }
    }
  }
});
