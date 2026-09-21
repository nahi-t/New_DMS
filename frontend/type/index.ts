export type Role = 'admin' | 'manager' | 'user';

export interface User {
  id: number;
  username: string;
  email: string;
  role: Role;
  created_at: string;
}

export interface Folder {
  id: number;
  name: string;
  created_by: number;
  created_at: string;
}

export interface LoginResponse {
  token: string;
  user: User;
}

export interface ApiError {
  message: string;
}

// type/index.ts

export interface Document {
  id: number;
  folder_id: number;
  user_id: number;
  name: string;
  file_path: string;
  mime_type: string;
  size: number;
  uploaded_at: string;
  description?: string;
  version?:number;
comment?:string;
  status?: string; 
}

export interface AuditLog {
  id: number;
  user_id: number;
  user_name: string;
  event: string;
  event_happened_time: string; // ISO 8601
  created_at: string;          // ISO 8601
}

export interface AuditLogFilters {
  limit?: number;
  offset?: number;
  user_id?: number | string;
  event?: string;
  /** RFC3339 timestamp, e.g. "2026-09-21T00:00:00Z" */
  from?: string;
  /** RFC3339 timestamp */
  to?: string;
}

export interface AuditLogPage {
  data: AuditLog[];
  total: number;
  limit: number;
  offset: number;
}