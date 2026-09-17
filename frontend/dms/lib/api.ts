// import { getTokenCookie, removeTokenCookie, removeUserStorage } from './utils';
// import { LoginResponse, User, Folder, Document, ApiError } from '@/type';

// const API_BASE = 'http://localhost:8080/api';

// type RequestMethod = 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH';

// // ---- JSON fetch (with auth) ----
// async function fetchWithAuth<T>(
//   endpoint: string,
//   method: RequestMethod = 'GET',
//   body?: any
// ): Promise<T> {
//   const token = getTokenCookie();
//   const headers: HeadersInit = {
//     'Content-Type': 'application/json',
//   };
//   if (token) {
//     headers['Authorization'] = `Bearer ${token}`;
//   }

//   const options: RequestInit = {
//     method,
//     headers,
//   };
//   if (body) {
//     options.body = JSON.stringify(body);
//     console.log(`📤 ${method} ${endpoint} payload:`, body);
//   }

//   const res = await fetch(`${API_BASE}${endpoint}`, options);

//   let data;
//   const contentType = res.headers.get('content-type');
//   if (contentType && contentType.includes('application/json')) {
//     data = await res.json();
//   } else {
//     const text = await res.text();
//     data = { message: text || 'Request failed' };
//   }

//   if (!res.ok) {
//     if (res.status === 401) {
//       removeTokenCookie();
//       removeUserStorage();
//       if (typeof window !== 'undefined') {
//         window.location.href = '/login';
//       }
//       throw new Error(data.message || 'Session expired. Please log in again.');
//     }
//     console.error(`❌ ${method} ${endpoint} failed:`, data);
//     throw new Error(data.message || `Request failed with status ${res.status}`);
//   }

//   return data as T;
// }

// // ---- FormData fetch (for file uploads) ----
// async function fetchWithFormData<T>(
//   endpoint: string,
//   method: 'POST' | 'PUT' = 'POST',
//   body: FormData
// ): Promise<T> {
//   const token = getTokenCookie();
//   const headers: HeadersInit = {};
//   if (token) {
//     headers['Authorization'] = `Bearer ${token}`;
//   }
//   // ⚠️ Do NOT set Content-Type – browser will set multipart boundary

//   const res = await fetch(`${API_BASE}${endpoint}`, {
//     method,
//     headers,
//     body,
//   });

//   let data;
//   const contentType = res.headers.get('content-type');
//   if (contentType && contentType.includes('application/json')) {
//     data = await res.json();
//   } else {
//     const text = await res.text();
//     data = { message: text || 'Request failed' };
//   }

//   if (!res.ok) {
//     if (res.status === 401) {
//       removeTokenCookie();
//       removeUserStorage();
//       if (typeof window !== 'undefined') {
//         window.location.href = '/login';
//       }
//       throw new Error(data.message || 'Session expired. Please log in again.');
//     }
//     console.error(`❌ ${method} ${endpoint} failed:`, data);
//     throw new Error(data.message || `Request failed with status ${res.status}`);
//   }

//   return data as T;
// }

// // ---- Auth ----
// export const login = (email: string, password: string) =>
//   fetchWithAuth<LoginResponse>('/auth/login', 'POST', { email, password });

// export const register = (username: string, email: string, password: string, role: string) =>
//   fetchWithAuth<{ message: string }>('/users/register', 'POST', { username, email, password, role });

// // ---- Users ----
// export const getUsers = () => fetchWithAuth<User[]>('/users');
// export const getUser = (id: number) => fetchWithAuth<User>(`/users/${id}`);
// export const deleteUser = (id: number) =>
//   fetchWithAuth<{ message: string }>(`/users/${id}`, 'DELETE');

// export const updateProfile = (id: number, username: string, email: string) =>
//   fetchWithAuth<{ message: string }>(`/users/${id}`, 'PUT', { username, email });

// export const adminUpdateUser = (id: number, username: string, email: string, role: string) =>
//   fetchWithAuth<{ message: string }>(`/users/${id}`, 'PUT', { username, email, role });

// export const updateUser = (
//   id: number,
//   data: { username: string; email: string; role?: string }
// ) => {
//   if (data.role === undefined) {
//     return updateProfile(id, data.username, data.email);
//   }
//   return adminUpdateUser(id, data.username, data.email, data.role);
// };

// // ---- Password ----
// export const updatePassword = (id: number, newPassword: string) =>
//   fetchWithAuth<{ message: string }>(`/users/${id}/password`, 'PUT', {
//     new_password: newPassword,
//   });

// // ---- Folders ----
// export const getFolders = () => fetchWithAuth<Folder[]>('/folders');
// export const createFolder = (name: string) =>
//   fetchWithAuth<Folder>('/folders', 'POST', { name });
// export const deleteFolder = (id: number) =>
//   fetchWithAuth<{ message: string }>(`/folders/${id}`, 'DELETE');
// export const updateFolder = (id: number, name: string) =>
//   fetchWithAuth<{ message: string }>(`/folders/${id}`, 'PUT', { name });

// // ---- Documents ----
// export const uploadDocument = (folderId: number, file: File, description?: string) => {
//   const formData = new FormData();
//   formData.append('file', file);
//   if (description) {
//     formData.append('description', description);
//   }
//   return fetchWithFormData<{ message: string; document: Document }>(
//     `/folders/${folderId}/documents`,
//     'POST',
//     formData
//   );
// };

// export const getDocuments = (folderId: number) =>
//   fetchWithAuth<Document[]>(`/folders/${folderId}/documents`);

// export const deleteDocument = (docId: number) =>
//   fetchWithAuth<{ message: string }>(`/documents/${docId}`, 'DELETE');

// export const downloadDocument = async (docId: number, fileName?: string) => {
//   const token = getTokenCookie();
//   const res = await fetch(`${API_BASE}/documents/${docId}`, {
//     headers: {
//       Authorization: `Bearer ${token}`,
//     },
//   });
//   if (!res.ok) {
//     const text = await res.text();
//     throw new Error(text || 'Download failed');
//   }
//   const blob = await res.blob();
//   const url = URL.createObjectURL(blob);
//   const a = document.createElement('a');
//   a.href = url;
//   a.download = fileName || 'document';
//   document.body.appendChild(a);
//   a.click();
//   a.remove();
//   URL.revokeObjectURL(url);
// };

// export const renameDocument = (docId: number, newName: string) =>
//   fetchWithAuth<{ message: string }>(`/documents/${docId}`, 'PATCH', { name: newName });

// export const moveDocument = (docId: number, newFolderId: number) =>
//   fetchWithAuth<{ message: string }>(`/documents/${docId}/move`, 'PATCH', { folder_id: newFolderId });

// export const searchDocuments = (search: string, folderId?: number) => {
//   let url = `/documents?search=${encodeURIComponent(search)}`;
//   if (folderId !== undefined) {
//     url += `&folder_id=${folderId}`;
//   }
//   return fetchWithAuth<Document[]>(url, 'GET');
// };

// export const updateDocumentContent = (docId: number, file: File) => {
//   const formData = new FormData();
//   formData.append('document', file);
//   return fetchWithFormData<{ message: string }>(
//     `/documents/${docId}/content`,
//     'PUT',
//     formData
//   );
// };

// // ---- New: Document Status Update ----
// export const updateDocumentStatus = (docId: number, status: string, comment?: string) =>
//   fetchWithAuth<{ message: string }>(`/documents/${docId}/status`, 'PATCH', {
//     status,
//     ...(comment && { comment }),
//   });

// // ---- Document Version types and endpoints ----
// export interface DocumentVersion {
//   id: number;
//   document_id: number;
//   version: number;
//   user_id: number;
//   file_path: string;
//   hashed_string: string;
//   created_at: string;
// }

// export const getDocumentVersions = (docId: number) =>
//   fetchWithAuth<DocumentVersion[]>(`/documentversion?document_id=${docId}`);

// export const downloadDocumentVersion = async (versionId: number, baseName: string) => {
//   const token = getTokenCookie();
//   const res = await fetch(`${API_BASE}/documentversion/${versionId}/download`, {
//     headers: {
//       Authorization: `Bearer ${token}`,
//     },
//   });
//   if (!res.ok) {
//     const text = await res.text();
//     throw new Error(text || 'Download failed');
//   }
//   const blob = await res.blob();
//   const url = URL.createObjectURL(blob);
//   const a = document.createElement('a');
//   a.href = url;
//   a.download = `version_${baseName}`;
//   document.body.appendChild(a);
//   a.click();
//   a.remove();
//   URL.revokeObjectURL(url);
// };

import { getTokenCookie, removeTokenCookie, removeUserStorage } from './utils';
import { LoginResponse, User, Folder, Document, ApiError } from '@/type';

// const API_BASE = 'http://localhost:8080/api';
const API_BASE='https://new-dms.onrender.com/api'

type RequestMethod = 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH';

/* ================================================================== */
/*  Core fetchers                                                      */
/* ================================================================== */

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

/* ================================================================== */
/*  Shared download helper                                             */
/* ================================================================== */

/**
 * Generic download helper.
 *
 * The backend either:
 *   1. Returns JSON: { url: "https://res.cloudinary.com/..." } → we open
 *      the Cloudinary URL directly (fast, no bytes through your server).
 *   2. Returns the file bytes (fallback) → we build a blob and download.
 */
async function downloadFromEndpoint(
  endpoint: string,
  fallbackFileName: string
): Promise<void> {
  const token = getTokenCookie();
  const res = await fetch(`${API_BASE}${endpoint}`, {
    headers: token ? { Authorization: `Bearer ${token}` } : {},
    redirect: 'follow', // follow any 302 redirects automatically
  });

  if (!res.ok) {
    const text = await res.text();
    throw new Error(text || 'Download failed');
  }

  const ct = res.headers.get('content-type') || '';

  // Case 1: backend returned a URL (recommended Cloudinary flow)
  if (ct.includes('application/json')) {
    const data = await res.json();
    const url = data.url ?? data.download_url;
    if (!url) throw new Error('No download URL returned from server');
    // Open directly — browser downloads straight from Cloudinary.
    const a = document.createElement('a');
    a.href = url;
    a.download = fallbackFileName;
    a.target = '_blank';
    a.rel = 'noopener';
    document.body.appendChild(a);
    a.click();
    a.remove();
    return;
  }

  // Case 2: backend streamed the file bytes (blob download)
  const blob = await res.blob();
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = fallbackFileName;
  document.body.appendChild(a);
  a.click();
  a.remove();
  URL.revokeObjectURL(url);
}

/* ================================================================== */
/*  Auth                                                               */
/* ================================================================== */

export const login = (email: string, password: string) =>
  fetchWithAuth<LoginResponse>('/auth/login', 'POST', { email, password });

export const register = (
  username: string,
  email: string,
  password: string,
  role: string
) =>
  fetchWithAuth<{ message: string }>('/users/register', 'POST', {
    username,
    email,
    password,
    role,
  });

/* ================================================================== */
/*  Users                                                              */
/* ================================================================== */

export const getUsers = () => fetchWithAuth<User[]>('/users');

export const getUser = (id: number) => fetchWithAuth<User>(`/users/${id}`);

export const deleteUser = (id: number) =>
  fetchWithAuth<{ message: string }>(`/users/${id}`, 'DELETE');

export const updateProfile = (id: number, username: string, email: string) =>
  fetchWithAuth<{ message: string }>(`/users/${id}`, 'PUT', { username, email });

export const adminUpdateUser = (
  id: number,
  username: string,
  email: string,
  role: string
) =>
  fetchWithAuth<{ message: string }>(`/users/${id}`, 'PUT', {
    username,
    email,
    role,
  });

export const updateUser = (
  id: number,
  data: { username: string; email: string; role?: string }
) => {
  if (data.role === undefined) {
    return updateProfile(id, data.username, data.email);
  }
  return adminUpdateUser(id, data.username, data.email, data.role);
};

/* ================================================================== */
/*  Password                                                           */
/* ================================================================== */

export const updatePassword = (id: number, newPassword: string) =>
  fetchWithAuth<{ message: string }>(`/users/${id}/password`, 'PUT', {
    new_password: newPassword,
  });

/* ================================================================== */
/*  Folders                                                            */
/* ================================================================== */

export const getFolders = () => fetchWithAuth<Folder[]>('/folders');

export const createFolder = (name: string) =>
  fetchWithAuth<Folder>('/folders', 'POST', { name });

export const deleteFolder = (id: number) =>
  fetchWithAuth<{ message: string }>(`/folders/${id}`, 'DELETE');

export const updateFolder = (id: number, name: string) =>
  fetchWithAuth<{ message: string }>(`/folders/${id}`, 'PUT', { name });

/* ================================================================== */
/*  Documents                                                          */
/* ================================================================== */

/**
 * Upload a new document (creates v1).
 * POST /folders/{folderId}/documents  (multipart)
 */
export const uploadDocument = (
  folderId: number,
  file: File,
  description?: string
) => {
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

/** List documents inside a folder. */
export const getDocuments = (folderId: number) =>
  fetchWithAuth<Document[]>(`/folders/${folderId}/documents`);

/** Delete a document and every version of it. */
export const deleteDocument = (docId: number) =>
  fetchWithAuth<{ message: string }>(`/documents/${docId}`, 'DELETE');

/**
 * Download the CURRENT version of a document.
 *
 * Backend redirects to Cloudinary (or returns { url }).
 */
export const downloadDocument = (docId: number, fileName?: string) =>
  downloadFromEndpoint(`/documents/${docId}`, fileName || 'document');

/** Rename a document (metadata only, file unchanged). */
export const renameDocument = (docId: number, newName: string) =>
  fetchWithAuth<{ message: string }>(`/documents/${docId}`, 'PATCH', {
    name: newName,
  });

/** Move a document to a different folder. */
export const moveDocument = (docId: number, newFolderId: number) =>
  fetchWithAuth<{ message: string }>(`/documents/${docId}/move`, 'PATCH', {
    folder_id: newFolderId,
  });

/** Search documents by name/description. */
export const searchDocuments = (search: string, folderId?: number) => {
  let url = `/documents?search=${encodeURIComponent(search)}`;
  if (folderId !== undefined) {
    url += `&folder_id=${folderId}`;
  }
  return fetchWithAuth<Document[]>(url, 'GET');
};

/**
 * Upload a NEW version of an existing document (creates v2, v3, ...).
 * PUT /documents/{id}/content  (multipart)
 */
export const updateDocumentContent = (docId: number, file: File) => {
  const formData = new FormData();
  formData.append('document', file);
  return fetchWithFormData<{ message: string }>(
    `/documents/${docId}/content`,
    'PUT',
    formData
  );
};

/** Update workflow status (and optional comment). */
export const updateDocumentStatus = (
  docId: number,
  status: string,
  comment?: string
) =>
  fetchWithAuth<{ message: string }>(`/documents/${docId}/status`, 'PATCH', {
    status,
    ...(comment && { comment }),
  });

/* ================================================================== */
/*  Document Versions                                                  */
/* ================================================================== */

export interface DocumentVersion {
  /** UUID string */
  id: string;
  document_id: number;
  user_id: number;
  version: number;

  /** Cloudinary public_id (was: file_path). */
  public_id: string;

  /** Original filename of this version. */
  original_filename?: string;

  /** SHA-256 hash used for dedup on next upload. */
  hashed_string: string;

  /** ISO timestamp */
  created_at: string;

  /** Download URL (with fl_attachment) — returned by the handler. */
  url?: string;

  /** Inline preview URL — returned by the handler. */
  preview_url?: string;

  /** True if this version matches documents.version. */
  is_current?: boolean;
}

/**
 * List every version of a document (newest first).
 * GET /documents/{id}/versions
 */
export const getDocumentVersions = (docId: number) =>
  fetchWithAuth<DocumentVersion[]>(`/documents/${docId}/versions`);

/**
 * Get a single version's metadata by its UUID.
 * GET /versions/{id}
 */
export const getDocumentVersion = (versionId: string) =>
  fetchWithAuth<DocumentVersion>(`/versions/${versionId}`);

/**
 * Download a specific version.
 * GET /versions/{id}/download
 *
 * Backend redirects to Cloudinary (or returns { url }).
 */
export const downloadDocumentVersion = (versionId: string, baseName: string) =>
  downloadFromEndpoint(`/versions/${versionId}/download`, `v_${baseName}`);

/**
 * Restore an old version — re-points the master document to it.
 * POST /versions/{id}/restore
 */
export const restoreDocumentVersion = (versionId: string) =>
  fetchWithAuth<{ message: string }>(`/versions/${versionId}/restore`, 'POST');

/**
 * Delete a specific version (admin only; cannot delete the current one).
 * DELETE /versions/{id}
 */
export const deleteDocumentVersion = (versionId: string) =>
  fetchWithAuth<{ message: string }>(`/versions/${versionId}`, 'DELETE');

/**
 * List all versions in the system (admin only).
 * GET /versions
 */
export const getAllDocumentVersions = () =>
  fetchWithAuth<DocumentVersion[]>('/versions');