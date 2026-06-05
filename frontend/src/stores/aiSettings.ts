import { defineStore } from 'pinia';

import { testAIConnection, type AIProviderConfig } from '../services/api';

const storageKey = 'novel-to-screenplay-ai-settings';

interface StoredAISettings {
  provider: string;
  base_url: string;
  model: string;
  api_key: string;
}

function defaultSettings(): StoredAISettings {
  return {
    provider: 'openai_compatible',
    base_url: 'https://api.openai.com/v1',
    model: 'gpt-4.1-mini',
    api_key: ''
  };
}

function loadSettings(): StoredAISettings {
  try {
    const stored = window.localStorage.getItem(storageKey);
    if (!stored) {
      return defaultSettings();
    }
    return { ...defaultSettings(), ...JSON.parse(stored) };
  } catch {
    return defaultSettings();
  }
}

export const useAISettingsStore = defineStore('aiSettings', {
  state: () => ({
    settings: loadSettings(),
    isTesting: false,
    testMessage: '',
    testOK: false
  }),
  getters: {
    hasConfig: (state) =>
      Boolean(state.settings.base_url.trim() && state.settings.model.trim() && state.settings.api_key.trim()),
    sanitizedConfig: (state): AIProviderConfig => ({
      provider: state.settings.provider,
      base_url: state.settings.base_url.trim(),
      model: state.settings.model.trim(),
      api_key: state.settings.api_key.trim()
    })
  },
  actions: {
    update(patch: Partial<StoredAISettings>) {
      this.settings = { ...this.settings, ...patch };
      this.persist();
      this.testMessage = '';
      this.testOK = false;
    },
    persist() {
      window.localStorage.setItem(storageKey, JSON.stringify(this.settings));
    },
    async test() {
      if (!this.hasConfig) {
        this.testOK = false;
        this.testMessage = '请先填写 Base URL、模型名和 API Key。';
        return false;
      }

      this.isTesting = true;
      this.testMessage = '';
      this.testOK = false;

      try {
        const result = await testAIConnection(this.sanitizedConfig);
        this.testOK = result.ok;
        this.testMessage = result.message;
        return result.ok;
      } catch {
        this.testOK = false;
        this.testMessage = 'AI 联通测试失败，请检查网络或配置。';
        return false;
      } finally {
        this.isTesting = false;
      }
    }
  }
});
