import { getTokenCookie, removeTokenCookie, removeUserStorage } from './utils';
import { LoginResponse, User, Folder, ApiError } from '@/type';

const API_BASE = 'http://localhost:8080/api';

type RequestMethod = 'GET' | 'POST' | 'PUT' | 'DELETE';

async function fetchWithAuth<T>(
  endpoint: string,
  method: RequestMethod = 'GET',
  body?: any
): Promise<T> {
  const token = getTokenCookie();
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
  };
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const options: RequestInit = {
    method,
    headers,
  };
  if (body) {
    options.body = JSON.stringify(body);
    // Debug log – check the payload in the console
    console.log(`📤 ${method} ${endpoint} payload:`, body);
  }

  const res = await fetch(`${API_BASE}${endpoint}`, options);

  let data;
  const contentType = res.headers.get('content-type');
  if (contentType && contentType.includes('application/json')) {
    data = await res.json();
  } else {
    const text = await res.text();
    data = { message: text || 'Request failed' };
  }

  if (!res.ok) {
    if (res.status === 401) {
      removeTokenCookie();
      removeUserStorage();
      if (typeof window !== 'undefined') {
        window.location.href = '/login';
      }
      throw new Error(data.message || 'Session expired. Please log in again.');
    }
    console.error(`❌ ${method} ${endpoint} failed:`, data);
    throw new Error(data.message || `Request failed with status ${res.status}`);
  }

  return data as T;
}

// ---- Auth ----
export const login = (email: string, password: string) =>
  fetchWithAuth<LoginResponse>('/auth/login', 'POST', { email, password });

export const register = (username: string, email: string, password: string, role: string) =>
  fetchWithAuth<{ message: string }>('/users/register', 'POST', { username, email, password, role });

// ---- Users ----
export const getUsers = () => fetchWithAuth<User[]>('/users');
export const getUser = (id: number) => fetchWithAuth<User>(`/users/${id}`);
export const deleteUser = (id: number) =>
  fetchWithAuth<{ message: string }>(`/users/${id}`, 'DELETE');

export const updateProfile = (id: number, username: string, email: string) =>
  fetchWithAuth<{ message: string }>(`/users/${id}`, 'PUT', { username, email });

export const adminUpdateUser = (id: number, username: string, email: string, role: string) =>
  fetchWithAuth<{ message: string }>(`/users/${id}`, 'PUT', { username, email, role });

export const updateUser = (
  id: number,
  data: { username: string; email: string; role?: string }
) => {
  if (data.role === undefined) {
    return updateProfile(id, data.username, data.email);
  }
  return adminUpdateUser(id, data.username, data.email, data.role);
};

/**
 * Password update – try different field names if this fails.
 * 
 * Common backend expectations:
 * - { currentPassword, newPassword }   ← most common
 * - { oldPassword, newPassword }
 * - { old_password, new_password }
 * - { password } (if old password is verified elsewhere)
 * 
 * Uncomment the line that matches your backend.
 */
export const updatePassword = (id: number, newPassword: string) =>
  fetchWithAuth<{ message: string }>(`/users/${id}/password`, 'PUT', {
    new_password: newPassword,
  });
  // Option B (if your backend expects oldPassword)
  // fetchWithAuth<{ message: string }>(`/users/${id}/password`, 'PUT', {
  //   oldPassword,
  //   newPassword,
  // });
  // Option C (snake_case)
  // fetchWithAuth<{ message: string }>(`/users/${id}/password`, 'PUT', {
  //   old_password: oldPassword,
  //   new_password: newPassword,
  // });

// ---- Folders ----
export const getFolders = () => fetchWithAuth<Folder[]>('/folders');
export const createFolder = (name: string) =>
  fetchWithAuth<Folder>('/folders', 'POST', { name });
export const deleteFolder = (id: number) =>
  fetchWithAuth<{ message: string }>(`/folders/${id}`, 'DELETE');
export const updateFolder = (id: number, name: string) =>
  fetchWithAuth<{ message: string }>(`/folders/${id}`, 'PUT', { name });