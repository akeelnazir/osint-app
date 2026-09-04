// Shared API types matching the Go backend DTOs.

export type Role = 'admin' | 'analyst' | 'viewer'
export type CaseStatus = 'open' | 'closed' | 'archived'
export type CaseVisibility = 'private' | 'public'
export type EvidenceType = 'text' | 'url' | 'image' | 'video' | 'document' | 'raw'
export type EntityType = 'person' | 'organization' | 'location' | 'other'

export interface User {
  id: number
  email: string
  username: string
  role: Role
}

export interface AuthResponse {
  access_token: string
  refresh_token: string
  expires_in: number
  user: User
}

export interface Case {
  id: number
  title: string
  description: string
  status: CaseStatus
  visibility: CaseVisibility
  owner_id: number
  owner_name: string
  tags: string[]
  created_at: string
  updated_at: string
}

export interface Evidence {
  id: number
  case_id: number
  type: EvidenceType
  title: string
  content: string
  file_path?: string
  mime_type?: string
  source?: string
  evidence_date?: string | null
  latitude?: number | null
  longitude?: number | null
  description?: string
  metadata?: Record<string, unknown>
  created_by: number
  created_at: string
  updated_at: string
  tags: string[]
}

export interface Entity {
  id: number
  evidence_id: number
  type: EntityType
  value: string
  confidence: number
}

export interface Comment {
  id: number
  author_id: number
  author: string
  body: string
  mentions: string[]
  created_at: string
}

export interface Member {
  user_id: number
  username: string
  email: string
  role: string
}

export interface TimelineEvent {
  evidence_id: number
  title: string
  type: EvidenceType
  date: string | null
  description: string
}

export interface Activity {
  id: number
  username: string
  action: string
  target_type?: string
  target_id?: string
  created_at: string
}

export interface DashboardStats {
  stats: { cases: number; evidence: number; entities: number; users: number }
  recent_cases: { id: number; title: string; status: CaseStatus; updated_at: string }[]
  activity: Activity[]
}

export interface SearchResult {
  type: 'case' | 'evidence'
  id: number
  title: string
  description: string
  case_id?: number
  evidence_type?: EvidenceType
  date?: string | null
}

export interface Paginated<T> {
  data: T[]
  total: number
  page: number
  limit: number
}

export interface GeoJSONFeatureCollection {
  type: 'FeatureCollection'
  features: {
    type: 'Feature'
    geometry: { type: string; coordinates: [number, number] }
    properties: { id: number; title: string; type: EvidenceType; date: string | null }
  }[]
}
