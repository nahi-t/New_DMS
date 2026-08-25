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