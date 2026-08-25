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
  }

  const res = await fetch(`${API_BASE}${endpoint}`, options);
  
  // Parse response – handle both JSON and plain text
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

// NEW – update user
export const updateUser = (id: number, data: { username: string; email: string; role: string }) =>
  fetchWithAuth<{ message: string }>(`/users/${id}`, 'PUT', data);

// ---- Folders ----
export const getFolders = () => fetchWithAuth<Folder[]>('/folders');
export const createFolder = (name: string) =>
  fetchWithAuth<Folder>('/folders', 'POST', { name });
export const deleteFolder = (id: number) =>
  fetchWithAuth<{ message: string }>(`/folders/${id}`, 'DELETE');

// NEW – update folder
export const updateFolder = (id: number, name: string) =>
  fetchWithAuth<{ message: string }>(`/folders/${id}`, 'PUT', { name });