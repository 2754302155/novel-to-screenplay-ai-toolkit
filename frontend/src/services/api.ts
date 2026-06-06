import axios from 'axios';

export interface HealthResponse {
  status: string;
  service: string;
  version: string;
}

export interface ParseIssue {
  code: string;
  message: string;
}

export interface ParsedChapter {
  id: string;
  title: string;
  word_count: number;
  body: string;
  inferred_title: boolean;
}

export interface QualityReport {
  coverage: {
    converted_chapters: number;
    estimated_unconverted_ratio: number;
  };
  warnings: string[];
  human_review_required: string[];
}

export interface ScreenplayDraft {
  quality_report?: QualityReport;
}

export interface ValidationIssue {
  path: string;
  message: string;
}

export interface ValidateYAMLResponse {
  valid: boolean;
  issues: ValidationIssue[];
  quality_report: QualityReport;
}

export interface ConversionTask {
  id: string;
  status: 'pending' | 'processing' | 'validating' | 'completed' | 'failed';
  progress: number;
  stage: string;
  source_text?: string;
  chapters: ParsedChapter[];
  draft?: ScreenplayDraft;
  yaml?: string;
  error_message?: string;
  total_chunks?: number;
  completed_chunks?: number;
  current_chunk?: string;
  created_at: string;
  updated_at: string;
}

export interface ConversionTaskSummary {
  id: string;
  status: ConversionTask['status'];
  progress: number;
  stage: string;
  chapter_count: number;
  total_chunks?: number;
  completed_chunks?: number;
  current_chunk?: string;
  error_message?: string;
  created_at: string;
  updated_at: string;
}

export interface ConversionTaskListResponse {
  tasks: ConversionTaskSummary[];
}

export interface AIProviderConfig {
  provider: string;
  base_url: string;
  model: string;
  api_key: string;
}

export interface TestAIResponse {
  ok: boolean;
  message: string;
}

export interface ParseChaptersResponse {
  chapters: ParsedChapter[];
  cleaned_text?: string;
  original_chars: number;
  cleaned_chars: number;
  chinese_ratio: number;
  warnings: ParseIssue[];
  blocking_errors: ParseIssue[] | null;
}

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL ?? '/api',
  timeout: 8000
});

export async function getHealth(): Promise<HealthResponse> {
  const response = await api.get<HealthResponse>('/healthz');
  return response.data;
}

export async function parseChapters(text: string): Promise<ParseChaptersResponse> {
  try {
    const response = await api.post<ParseChaptersResponse>('/chapters/parse', { text }, {
      timeout: 60000
    });
    return response.data;
  } catch (error) {
    if (axios.isAxiosError<ParseChaptersResponse>(error) && error.response?.data) {
      return error.response.data;
    }

    throw error;
  }
}

export async function createConversionTask(
  sourceText: string,
  chapters: ParsedChapter[],
  aiConfig?: AIProviderConfig
): Promise<ConversionTask> {
  const response = await api.post<ConversionTask>('/conversion-tasks', {
    source_text: sourceText,
    chapters,
    ai_config: aiConfig
  });
  return response.data;
}

export async function getConversionTask(taskId: string): Promise<ConversionTask> {
  const response = await api.get<ConversionTask>(`/conversion-tasks/${taskId}`);
  return response.data;
}

export async function listConversionTasks(): Promise<ConversionTaskSummary[]> {
  const response = await api.get<ConversionTaskListResponse>('/conversion-tasks');
  return response.data.tasks;
}

export async function validateYAML(yamlText: string): Promise<ValidateYAMLResponse> {
  const response = await api.post<ValidateYAMLResponse>('/yaml/validate', {
    yaml: yamlText
  });
  return response.data;
}

export async function testAIConnection(aiConfig: AIProviderConfig): Promise<TestAIResponse> {
  try {
    const response = await api.post<TestAIResponse>('/ai/test', aiConfig, {
      timeout: 35000
    });
    return response.data;
  } catch (error) {
    if (axios.isAxiosError<TestAIResponse>(error) && error.response?.data) {
      return error.response.data;
    }

    throw error;
  }
}
