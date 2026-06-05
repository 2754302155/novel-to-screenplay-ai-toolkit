import axios from 'axios';

export interface HealthResponse {
  status: string;
  service: string;
  version: string;
}

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL ?? '/api',
  timeout: 8000
});

export async function getHealth(): Promise<HealthResponse> {
  const response = await api.get<HealthResponse>('/healthz');
  return response.data;
}
