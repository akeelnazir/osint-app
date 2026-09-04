import { api } from './api'
import type {
  AuthResponse,
  Case,
  Evidence,
  Entity,
  Comment,
  Member,
  TimelineEvent,
  Activity,
  DashboardStats,
  SearchResult,
  Paginated,
  GeoJSONFeatureCollection,
} from './types'

// --- Auth ---
export const authApi = {
  register: (data: { email: string; username: string; password: string; role?: string }) =>
    api.post<AuthResponse>('/auth/register', data).then((r) => r.data),
  login: (data: { email: string; password: string }) =>
    api.post<AuthResponse>('/auth/login', data).then((r) => r.data),
  refresh: (refresh_token: string) =>
    api.post<AuthResponse>('/auth/refresh', { refresh_token }).then((r) => r.data),
}

// --- User ---
export const userApi = {
  me: () => api.get<{ id: number; email: string; username: string; role: string }>('/users/me').then((r) => r.data),
}

// --- Cases ---
export const caseApi = {
  list: (params?: Record<string, string | number>) =>
    api.get<Paginated<Case>>('/cases', { params }).then((r) => r.data),
  get: (id: number) => api.get<Case>(`/cases/${id}`).then((r) => r.data),
  create: (data: { title: string; description?: string; visibility?: string; tags?: string[] }) =>
    api.post<Case>('/cases', data).then((r) => r.data),
  update: (id: number, data: Partial<Case>) =>
    api.put<Case>(`/cases/${id}`, data).then((r) => r.data),
  delete: (id: number) => api.delete(`/cases/${id}`),
  listMembers: (id: number) => api.get<Member[]>(`/cases/${id}/members`).then((r) => r.data),
  addMember: (id: number, data: { username: string; role: string }) =>
    api.post<Member>(`/cases/${id}/members`, data).then((r) => r.data),
  removeMember: (id: number, userId: number) => api.delete(`/cases/${id}/members/${userId}`),
  timeline: (id: number, params?: Record<string, string>) =>
    api.get<TimelineEvent[]>(`/cases/${id}/timeline`, { params }).then((r) => r.data),
  listComments: (id: number) => api.get<Comment[]>(`/cases/${id}/comments`).then((r) => r.data),
  createComment: (id: number, body: string) =>
    api.post<Comment>(`/cases/${id}/comments`, { body }).then((r) => r.data),
  listActivity: (id: number) => api.get<Activity[]>(`/cases/${id}/activity`).then((r) => r.data),
}

// --- Evidence ---
export const evidenceApi = {
  list: (caseId: number, params?: Record<string, string | number>) =>
    api.get<Paginated<Evidence>>(`/cases/${caseId}/evidence`, { params }).then((r) => r.data),
  get: (caseId: number, evidenceId: number) =>
    api.get<Evidence>(`/cases/${caseId}/evidence/${evidenceId}`).then((r) => r.data),
  create: (caseId: number, data: Record<string, unknown>) =>
    api.post<Evidence>(`/cases/${caseId}/evidence`, data).then((r) => r.data),
  createFile: (caseId: number, formData: FormData) =>
    api.post<Evidence>(`/cases/${caseId}/evidence`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    }).then((r) => r.data),
  update: (caseId: number, evidenceId: number, data: Partial<Evidence>) =>
    api.put<Evidence>(`/cases/${caseId}/evidence/${evidenceId}`, data).then((r) => r.data),
  delete: (caseId: number, evidenceId: number) =>
    api.delete(`/cases/${caseId}/evidence/${evidenceId}`),
  geojson: (caseId: number) =>
    api.get<GeoJSONFeatureCollection>(`/cases/${caseId}/evidence/geojson`).then((r) => r.data),
  extractEntities: (caseId: number, evidenceId: number) =>
    api.post<Entity[]>(`/cases/${caseId}/evidence/${evidenceId}/entities`).then((r) => r.data),
  listEntities: (caseId: number, evidenceId: number) =>
    api.get<Entity[]>(`/cases/${caseId}/evidence/${evidenceId}/entities`).then((r) => r.data),
  listComments: (caseId: number, evidenceId: number) =>
    api.get<Comment[]>(`/cases/${caseId}/evidence/${evidenceId}/comments`).then((r) => r.data),
  createComment: (caseId: number, evidenceId: number, body: string) =>
    api.post<Comment>(`/cases/${caseId}/evidence/${evidenceId}/comments`, { body }).then((r) => r.data),
}

// --- Search ---
export const searchApi = {
  search: (params: Record<string, string | number>) =>
    api.get<{ data: SearchResult[]; query: string; page: number; limit: number }>('/search', { params }).then((r) => r.data),
}

// --- Dashboard ---
export const dashboardApi = {
  stats: () => api.get<DashboardStats>('/dashboard').then((r) => r.data),
}

// --- Export ---
export const exportApi = {
  casePDF: (id: number) =>
    `${api.defaults.baseURL}/export/case/${id}/pdf`,
  evidenceCSV: (id: number) =>
    `${api.defaults.baseURL}/export/case/${id}/evidence.csv`,
}
