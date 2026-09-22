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

// src/type.ts

export interface AuditLog {
  id: number;
  user_id: number | null;      // nullable — failed logins have no user id
  user_name: string;
  event: string;
  event_happened_time: string;
  created_at: string;

  ip_address?: string | null;
  user_agent?: string | null;
  success: boolean;
}

export interface AuditLogFilters {
  limit?: number;
  offset?: number;
  user_id?: number | string;
  event?: string;
  from?: string;
  to?: string;
}

export interface AuditLogPage {
  data: AuditLog[];
  total: number;
  limit: number;
  offset: number;
}

// @/type.ts



export interface DocumentShare {
  id: number;
  document_id: number;
  document_name: string;
  document_description?: string;
  folder_id: number;
  version: number;

  user_id: number;
  user_name: string;
  user_email: string;

  permission: 'viewer' | 'editor';
  shared_by: number;
  shared_by_name: string;
  shared_at: string;
}