import { getTokenCookie, removeTokenCookie, removeUserStorage } from './utils';
import { LoginResponse, User, Folder, Document, ApiError } from '@/type';

const API_BASE = 'http://localhost:8080/api';

type RequestMethod = 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH';

// ---- JSON fetch (with auth) ----
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

// ---- FormData fetch (for file uploads) ----
async function fetchWithFormData<T>(
  endpoint: string,
  method: 'POST' | 'PUT' = 'POST',
  body: FormData
): Promise<T> {
  const token = getTokenCookie();
  const headers: HeadersInit = {};
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  // ⚠️ Do NOT set Content-Type – browser will set multipart boundary

  const res = await fetch(`${API_BASE}${endpoint}`, {
    method,
    headers,
    body,
  });

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

// ---- Password ----
export const updatePassword = (id: number, newPassword: string) =>
  fetchWithAuth<{ message: string }>(`/users/${id}/password`, 'PUT', {
    new_password: newPassword,
  });

// ---- Folders ----
export const getFolders = () => fetchWithAuth<Folder[]>('/folders');
export const createFolder = (name: string) =>
  fetchWithAuth<Folder>('/folders', 'POST', { name });
export const deleteFolder = (id: number) =>
  fetchWithAuth<{ message: string }>(`/folders/${id}`, 'DELETE');
export const updateFolder = (id: number, name: string) =>
  fetchWithAuth<{ message: string }>(`/folders/${id}`, 'PUT', { name });

// ---- Documents ----
export const uploadDocument = (folderId: number, file: File, description?: string) => {
  const formData = new FormData();
  formData.append('file', file);
  if (description) {
    formData.append('description', description);
  }
  return fetchWithFormData<{ message: string; document: Document }>(
    `/folders/${folderId}/documents`,
    'POST',
    formData
  );
};

export const getDocuments = (folderId: number) =>
  fetchWithAuth<Document[]>(`/folders/${folderId}/documents`);

export const deleteDocument = (docId: number) =>
  fetchWithAuth<{ message: string }>(`/documents/${docId}`, 'DELETE');

export const downloadDocument = async (docId: number, fileName?: string) => {
  const token = getTokenCookie();
  const res = await fetch(`${API_BASE}/documents/${docId}`, {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
  if (!res.ok) {
    const text = await res.text();
    throw new Error(text || 'Download failed');
  }
  const blob = await res.blob();
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = fileName || 'document';
  document.body.appendChild(a);
  a.click();
  a.remove();
  URL.revokeObjectURL(url);
};

export const renameDocument = (docId: number, newName: string) =>
  fetchWithAuth<{ message: string }>(`/documents/${docId}`, 'PATCH', { name: newName });

export const moveDocument = (docId: number, newFolderId: number) =>
  fetchWithAuth<{ message: string }>(`/documents/${docId}/move`, 'PATCH', { folder_id: newFolderId });

export const searchDocuments = (search: string, folderId?: number) => {
  let url = `/documents?search=${encodeURIComponent(search)}`;
  if (folderId !== undefined) {
    url += `&folder_id=${folderId}`;
  }
  return fetchWithAuth<Document[]>(url, 'GET');
};

export const updateDocumentContent = (docId: number, file: File) => {
  const formData = new FormData();
  formData.append('document', file);
  return fetchWithFormData<{ message: string }>(
    `/documents/${docId}/content`,
    'PUT',
    formData
  );
};

// ---- New: Document Status Update ----
export const updateDocumentStatus = (docId: number, status: string, comment?: string) =>
  fetchWithAuth<{ message: string }>(`/documents/${docId}/status`, 'PATCH', {
    status,
    ...(comment && { comment }),
  });

// ---- Document Version types and endpoints ----
export interface DocumentVersion {
  id: number;
  document_id: number;
  version: number;
  user_id: number;
  file_path: string;
  hashed_string: string;
  created_at: string;
}

export const getDocumentVersions = (docId: number) =>
  fetchWithAuth<DocumentVersion[]>(`/documentversion?document_id=${docId}`);

export const downloadDocumentVersion = async (versionId: number, baseName: string) => {
  const token = getTokenCookie();
  const res = await fetch(`${API_BASE}/documentversion/${versionId}/download`, {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
  if (!res.ok) {
    const text = await res.text();
    throw new Error(text || 'Download failed');
  }
  const blob = await res.blob();
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `version_${baseName}`;
  document.body.appendChild(a);
  a.click();
  a.remove();
  URL.revokeObjectURL(url);
};