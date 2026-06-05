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

export interface ParseChaptersResponse {
  chapters: ParsedChapter[];
  cleaned_text: string;
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
    const response = await api.post<ParseChaptersResponse>('/chapters/parse', { text });
    return response.data;
  } catch (error) {
    if (axios.isAxiosError<ParseChaptersResponse>(error) && error.response?.data) {
      return error.response.data;
    }

    throw error;
  }
}
