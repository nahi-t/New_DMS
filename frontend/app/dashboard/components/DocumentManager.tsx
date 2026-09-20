'use client';

import { useState, useEffect, useCallback, useRef, type ComponentType } from 'react';
import { useAuth } from '@/context/AuthContext';
import {
  getDocuments,
  uploadDocument,
  deleteDocument,
  downloadDocument,
  renameDocument,
  moveDocument,
  searchDocuments,
  updateDocumentContent,
  getDocumentVersions,
  downloadDocumentVersion,
  updateDocumentStatus,
  DocumentVersion,
} from '@/lib/api';
import { Document, Folder } from '@/type';
import {
  Trash2,
  Download,
  FileText,
  FileSpreadsheet,
  FileImage,
  FileArchive,
  File as FileIcon,
  FileUp,
  UploadCloud,
  Pencil,
  FolderInput,
  FolderOpen,
  X,
  Check,
  Search,
  History,
  CheckCircle,
  XCircle,
  Clock,
  Loader2,
  Inbox,
  MessageSquareText,
  ShieldCheck,
} from 'lucide-react';
import toast from 'react-hot-toast';

interface Props {
  folderId: number;
  canUpload: boolean;
  canDelete?: boolean;
  allFolders: Folder[];
  userRole?: string;
}

// ---- Robust status resolution ----
function getRawStatus(doc: any): string {
  const candidate =
    doc?.status ??
    doc?.review_status ??
    doc?.approval_status ??
    doc?.state ??
    doc?.doc_status ??
    '';

  if (candidate && typeof candidate === 'object') {
    return String(candidate.value ?? candidate.name ?? '').trim().toLowerCase();
  }

  return String(candidate ?? '').trim().toLowerCase();
}

function getRawComment(doc: any): string {
  return doc?.comment ?? doc?.review_comment ?? doc?.status_comment ?? '';
}

// ---- File-type icon + color, purely visual, based on file extension ----
type FileMeta = { Icon: ComponentType<{ className?: string }>; iconClass: string; bgClass: string };

function getFileMeta(name: string): FileMeta {
  const ext = (name.split('.').pop() || '').toLowerCase();

  if (ext === 'pdf') return { Icon: FileText, iconClass: 'text-rose-600', bgClass: 'bg-rose-50' };
  if (['doc', 'docx', 'txt', 'rtf', 'odt'].includes(ext)) {
    return { Icon: FileText, iconClass: 'text-blue-600', bgClass: 'bg-blue-50' };
  }
  if (['xls', 'xlsx', 'csv'].includes(ext)) {
    return { Icon: FileSpreadsheet, iconClass: 'text-emerald-600', bgClass: 'bg-emerald-50' };
  }
  if (['png', 'jpg', 'jpeg', 'gif', 'webp', 'svg'].includes(ext)) {
    return { Icon: FileImage, iconClass: 'text-violet-600', bgClass: 'bg-violet-50' };
  }
  if (['zip', 'rar', '7z', 'tar', 'gz'].includes(ext)) {
    return { Icon: FileArchive, iconClass: 'text-amber-600', bgClass: 'bg-amber-50' };
  }
  return { Icon: FileIcon, iconClass: 'text-slate-500', bgClass: 'bg-slate-50' };
}

export default function DocumentManager({
  folderId,
  canUpload = true,
  canDelete = true,
  allFolders,
  userRole: propUserRole,
}: Props) {
  const { user } = useAuth();
  const effectiveRole = propUserRole || user?.role || 'user';

  const [documents, setDocuments] = useState<Document[]>([]);
  const [loading, setLoading] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [searchTerm, setSearchTerm] = useState('');
  const [editingName, setEditingName] = useState<number | null>(null);
  const [newName, setNewName] = useState('');
  const [movingDoc, setMovingDoc] = useState<number | null>(null);
  const [updatingDocId, setUpdatingDocId] = useState<number | null>(null);

  const [comment, setComment] = useState<{ [docId: number]: string }>({});
  const [processingStatus, setProcessingStatus] = useState<number | null>(null);

  // Which document's "Review" editor (comment + Approve/Reject) is currently
  // open. Only one at a time, and only relevant for managers/admins.
  const [statusEditorOpenFor, setStatusEditorOpenFor] = useState<number | null>(null);

  const [selectedDocForHistory, setSelectedDocForHistory] = useState<Document | null>(null);
  const [versions, setVersions] = useState<DocumentVersion[]>([]);
  const [loadingVersions, setLoadingVersions] = useState(false);

  const moveMenuRef = useRef<HTMLDivElement | null>(null);

  // Close the "move to folder" menu on an outside click or Escape.
  useEffect(() => {
    if (movingDoc === null) return;

    function handlePointerDown(e: MouseEvent) {
      if (moveMenuRef.current && !moveMenuRef.current.contains(e.target as Node)) {
        setMovingDoc(null);
      }
    }
    function handleKeyDown(e: KeyboardEvent) {
      if (e.key === 'Escape') setMovingDoc(null);
    }

    document.addEventListener('mousedown', handlePointerDown);
    document.addEventListener('keydown', handleKeyDown);
    return () => {
      document.removeEventListener('mousedown', handlePointerDown);
      document.removeEventListener('keydown', handleKeyDown);
    };
  }, [movingDoc]);

  // Fetch all documents, then each row extracts its own status via
  // getRawStatus/getRawComment at render time. Always trusts the DB —
  // no local caching or overriding of status.
  const fetchDocuments = useCallback(async () => {
    setLoading(true);
    try {
      let data;
      if (searchTerm.trim() === '') {
        data = await getDocuments(folderId);
      } else {
        data = await searchDocuments(searchTerm, folderId);
      }
      setDocuments(Array.isArray(data) ? data : []);
    } catch (error: any) {
      toast.error(error.message || 'Failed to load documents');
      setDocuments([]);
    } finally {
      setLoading(false);
    }
  }, [folderId, searchTerm]);

  useEffect(() => {
    fetchDocuments();
  }, [fetchDocuments]);

  const handleUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    setUploading(true);
    try {
      await uploadDocument(folderId, file);
      toast.success('Document uploaded');
      fetchDocuments();
    } catch (error: any) {
      toast.error(error.message || 'Upload failed');
    } finally {
      setUploading(false);
      e.target.value = '';
    }
  };

  const handleUpdateContent = async (docId: number, e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    setUpdatingDocId(docId);
    try {
      await updateDocumentContent(docId, file);
      toast.success('Document content updated successfully (New version created)');
      fetchDocuments();
    } catch (error: any) {
      toast.error(error.message || 'Failed to update document content');
    } finally {
      setUpdatingDocId(null);
      e.target.value = '';
    }
  };

  const handleDelete = async (docId: number) => {
    if (!confirm('Delete this document?')) return;
    try {
      await deleteDocument(docId);
      toast.success('Document deleted');
      fetchDocuments();
    } catch (error: any) {
      toast.error(error.message || 'Delete failed');
    }
  };

  const handleDownload = async (doc: Document) => {
    try {
      await downloadDocument(doc.id, doc.name);
    } catch (error: any) {
      toast.error(error.message || 'Download failed');
    }
  };

  const handleRename = async (docId: number) => {
    if (!newName.trim()) {
      toast.error('Name cannot be empty');
      return;
    }
    try {
      await renameDocument(docId, newName);
      toast.success('Document renamed');
      setEditingName(null);
      fetchDocuments();
    } catch (error: any) {
      toast.error(error.message || 'Rename failed');
    }
  };

  const handleMove = async (docId: number, targetFolderId: number) => {
    if (targetFolderId === folderId) {
      toast.error('Document is already in this folder');
      return;
    }
    try {
      await moveDocument(docId, targetFolderId);
      toast.success('Document moved');
      setMovingDoc(null);
      fetchDocuments();
    } catch (error: any) {
      toast.error(error.message || 'Move failed');
    }
  };

  // ---- Approve / Reject: manager/admin can do this anytime, including
  // changing a previous decision. Closes the inline editor on success. ----
  const handleStatusUpdate = async (docId: number, status: string) => {
    const commentText = comment[docId] || '';
    const previousDocs = documents;

    setProcessingStatus(docId);

    // Optimistic update so the UI reacts instantly.
    setDocuments((prev) =>
      prev.map((d) =>
        d.id === docId ? { ...d, status, comment: commentText } : d
      )
    );

    try {
      await updateDocumentStatus(docId, status, commentText);
      toast.success(`Document ${status}`);
      setComment((prev) => ({ ...prev, [docId]: '' }));
      setStatusEditorOpenFor(null);
      // Re-sync with the server — the DB is the single source of truth.
      fetchDocuments();
    } catch (error: any) {
      toast.error(error.message || 'Status update failed');
      // Revert the optimistic change on failure; keep the editor open
      // so the manager can retry without losing their comment.
      setDocuments(previousDocs);
    } finally {
      setProcessingStatus(null);
    }
  };

  const handleOpenHistory = async (doc: Document) => {
    setSelectedDocForHistory(doc);
    setLoadingVersions(true);
    try {
      const data = await getDocumentVersions(doc.id);
      setVersions(Array.isArray(data) ? data : []);
    } catch (error: any) {
      toast.error(error.message || 'Failed to fetch version history');
      setVersions([]);
    } finally {
      setLoadingVersions(false);
    }
  };

  const handleDownloadVersionFile = async (versionId: string, docName: string) => {
    try {
      await downloadDocumentVersion(versionId, docName);
      toast.success('Version downloaded');
    } catch (error: any) {
      toast.error(error.message || 'Failed to download version');
    }
  };

  const formatFileSize = (bytes: number) => {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1048576) return (bytes / 1024).toFixed(1) + ' KB';
    return (bytes / 1048576).toFixed(1) + ' MB';
  };

  const isManagerOrAdmin = (() => {
    const role = effectiveRole.toLowerCase();
    return role === 'manager' || role === 'admin';
  })();

  return (
    <div className="mt-4">
      <div className="overflow-hidden rounded-2xl border border-gray-200 bg-white">
        {/* Toolbar */}
        <div className="flex flex-wrap items-center gap-3 border-b border-gray-100 p-4">
          <div className="relative min-w-[220px] flex-1">
            <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
            <input
              type="text"
              placeholder="Search documents"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="w-full rounded-lg border border-gray-200 bg-gray-50 py-2 pl-9 pr-3 text-sm text-gray-900 placeholder:text-gray-400 transition-colors focus:border-indigo-400 focus:bg-white focus:outline-none focus:ring-2 focus:ring-indigo-500/20"
            />
          </div>

          {!loading && (
            <span className="whitespace-nowrap text-xs text-gray-400">
              {documents.length} {documents.length === 1 ? 'document' : 'documents'}
            </span>
          )}

          {canUpload && (
            <label
              className={`inline-flex cursor-pointer items-center gap-2 rounded-lg px-3.5 py-2 text-sm font-medium text-white shadow-sm transition-colors ${
                uploading ? 'cursor-wait bg-indigo-400' : 'bg-indigo-600 hover:bg-indigo-700'
              }`}
            >
              {uploading ? <Loader2 className="h-4 w-4 animate-spin" /> : <UploadCloud className="h-4 w-4" />}
              {uploading ? 'Uploading…' : 'Upload'}
              <input type="file" onChange={handleUpload} className="hidden" disabled={uploading} />
            </label>
          )}
        </div>

        {/* Loading skeleton */}
        {loading && (
          <div className="divide-y divide-gray-100">
            {Array.from({ length: 4 }).map((_, i) => (
              <div key={i} className="flex animate-pulse items-center gap-3 px-4 py-3.5">
                <div className="h-10 w-10 rounded-lg bg-gray-100" />
                <div className="flex-1 space-y-2">
                  <div className="h-3.5 w-1/3 rounded bg-gray-100" />
                  <div className="h-2.5 w-1/5 rounded bg-gray-100" />
                </div>
              </div>
            ))}
          </div>
        )}

        {/* Empty state */}
        {!loading && documents.length === 0 && (
          <div className="flex flex-col items-center justify-center gap-3 px-4 py-14 text-center">
            <div className="flex h-12 w-12 items-center justify-center rounded-full bg-gray-50">
              <Inbox className="h-6 w-6 text-gray-300" />
            </div>
            <div>
              <p className="text-sm font-medium text-gray-900">
                {searchTerm ? 'No matching documents' : 'No documents yet'}
              </p>
              <p className="mt-1 text-xs text-gray-400">
                {searchTerm
                  ? 'Try a different search term.'
                  : canUpload
                  ? 'Upload your first document to get started.'
                  : 'Documents added to this folder will show up here.'}
              </p>
            </div>
          </div>
        )}

        {/* Document list */}
        {!loading && documents.length > 0 && (
          <ul className="divide-y divide-gray-100">
            {documents.map((doc) => {
              const normalizedStatus = getRawStatus(doc);
              const docComment = getRawComment(doc);
              const isFinalized = normalizedStatus === 'approved' || normalizedStatus === 'rejected';
              const isEditingStatus = statusEditorOpenFor === doc.id;
              const { Icon: FileTypeIcon, iconClass, bgClass } = getFileMeta(doc.name);

              return (
                <li key={doc.id} className="group px-4 py-3.5">
                  {/* Main row: icon, name, actions */}
                  <div className="flex items-center gap-3">
                    <div
                      className={`flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-lg ${bgClass}`}
                    >
                      <FileTypeIcon className={`h-5 w-5 ${iconClass}`} />
                    </div>

                    <div className="min-w-0 flex-1">
                      {editingName === doc.id ? (
                        <div className="flex items-center gap-2">
                          <input
                            type="text"
                            value={newName}
                            onChange={(e) => setNewName(e.target.value)}
                            onKeyDown={(e) => {
                              if (e.key === 'Enter') handleRename(doc.id);
                              if (e.key === 'Escape') setEditingName(null);
                            }}
                            className="w-full max-w-xs rounded-lg border border-gray-200 bg-white px-2.5 py-1.5 text-sm text-gray-900 focus:border-indigo-400 focus:outline-none focus:ring-2 focus:ring-indigo-500/20"
                            autoFocus
                          />
                          <button
                            onClick={() => handleRename(doc.id)}
                            className="rounded-md p-1.5 text-emerald-600 transition-colors hover:bg-emerald-50 focus:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500/40"
                            title="Save name"
                            aria-label="Save name"
                          >
                            <Check className="h-4 w-4" />
                          </button>
                          <button
                            onClick={() => setEditingName(null)}
                            className="rounded-md p-1.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600 focus:outline-none focus-visible:ring-2 focus-visible:ring-gray-400/40"
                            title="Cancel"
                            aria-label="Cancel rename"
                          >
                            <X className="h-4 w-4" />
                          </button>
                        </div>
                      ) : (
                        <>
                          <p className="truncate text-sm font-medium text-gray-900">
                            {doc.name}
                            {doc.version ? (
                              <span className="ml-2 rounded bg-indigo-50 px-1.5 py-0.5 text-xs font-semibold text-indigo-600">
                                v{doc.version}
                              </span>
                            ) : null}
                          </p>
                          <p className="mt-0.5 text-xs text-gray-400">
                            {formatFileSize(doc.size)}
                            {doc.uploaded_at
                              ? ` • ${new Date(doc.uploaded_at).toLocaleDateString()}`
                              : ''}
                          </p>
                        </>
                      )}
                    </div>

                    {/* Action buttons — always visible on touch, revealed on hover for pointer users */}
                    <div className="flex flex-shrink-0 items-center gap-0.5 opacity-100 transition-opacity sm:opacity-0 sm:group-hover:opacity-100 sm:group-focus-within:opacity-100">
                      <button
                        onClick={() => handleOpenHistory(doc)}
                        className="rounded-md p-1.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-700 focus:outline-none focus-visible:ring-2 focus-visible:ring-gray-400/40"
                        title="Version history"
                        aria-label="View version history"
                      >
                        <History className="h-4 w-4" />
                      </button>

                      <label
                        className={`rounded-md p-1.5 transition-colors focus-within:ring-2 focus-within:ring-indigo-500/40 ${
                          updatingDocId === doc.id
                            ? 'cursor-wait text-indigo-400'
                            : 'cursor-pointer text-gray-400 hover:bg-indigo-50 hover:text-indigo-600'
                        }`}
                        title="Upload new version"
                        aria-label="Upload new version"
                      >
                        {updatingDocId === doc.id ? (
                          <Loader2 className="h-4 w-4 animate-spin" />
                        ) : (
                          <FileUp className="h-4 w-4" />
                        )}
                        <input
                          type="file"
                          onChange={(e) => handleUpdateContent(doc.id, e)}
                          className="hidden"
                          disabled={updatingDocId === doc.id}
                        />
                      </label>

                      <button
                        onClick={() => handleDownload(doc)}
                        className="rounded-md p-1.5 text-gray-400 transition-colors hover:bg-blue-50 hover:text-blue-600 focus:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/40"
                        title="Download latest"
                        aria-label="Download latest version"
                      >
                        <Download className="h-4 w-4" />
                      </button>

                      <button
                        onClick={() => {
                          setEditingName(doc.id);
                          setNewName(doc.name);
                        }}
                        className="rounded-md p-1.5 text-gray-400 transition-colors hover:bg-amber-50 hover:text-amber-600 focus:outline-none focus-visible:ring-2 focus-visible:ring-amber-500/40"
                        title="Rename"
                        aria-label="Rename document"
                      >
                        <Pencil className="h-4 w-4" />
                      </button>

                      {allFolders.length > 0 && (
                        <div
                          className="relative"
                          ref={movingDoc === doc.id ? moveMenuRef : undefined}
                        >
                          <button
                            onClick={() => setMovingDoc(movingDoc === doc.id ? null : doc.id)}
                            className="rounded-md p-1.5 text-gray-400 transition-colors hover:bg-purple-50 hover:text-purple-600 focus:outline-none focus-visible:ring-2 focus-visible:ring-purple-500/40"
                            title="Move to folder"
                            aria-label="Move to another folder"
                          >
                            <FolderInput className="h-4 w-4" />
                          </button>
                          {movingDoc === doc.id && (
                            <div className="absolute right-0 z-10 mt-2 max-h-60 w-52 overflow-y-auto rounded-xl border border-gray-100 bg-white py-1 shadow-lg">
                              {allFolders.map((f) => {
                                const isCurrent = f.id === folderId;
                                return (
                                  <button
                                    key={f.id}
                                    onClick={() => handleMove(doc.id, f.id)}
                                    disabled={isCurrent}
                                    className={`flex w-full items-center gap-2 px-3 py-2 text-left text-sm transition-colors ${
                                      isCurrent
                                        ? 'cursor-not-allowed text-gray-300'
                                        : 'text-gray-700 hover:bg-gray-50'
                                    }`}
                                  >
                                    <FolderOpen className="h-4 w-4 flex-shrink-0" />
                                    <span className="truncate">{f.name}</span>
                                    {isCurrent && (
                                      <Check className="ml-auto h-4 w-4 flex-shrink-0 text-emerald-500" />
                                    )}
                                  </button>
                                );
                              })}
                            </div>
                          )}
                        </div>
                      )}

                      {canDelete && (
                        <button
                          onClick={() => handleDelete(doc.id)}
                          className="rounded-md p-1.5 text-gray-400 transition-colors hover:bg-rose-50 hover:text-rose-600 focus:outline-none focus-visible:ring-2 focus-visible:ring-rose-500/40"
                          title="Delete"
                          aria-label="Delete document"
                        >
                          <Trash2 className="h-4 w-4" />
                        </button>
                      )}
                    </div>
                  </div>

                  {/* Review status + (manager/admin only) the review controls */}
                  <div className="ml-[52px] mt-2 space-y-2">
                    {isFinalized ? (
                      <ReviewSummary status={normalizedStatus} comment={docComment} />
                    ) : (
                      <span className="inline-flex items-center gap-1.5 rounded-full bg-amber-50 px-2.5 py-1 text-xs font-medium text-amber-700">
                        <Clock className="h-3.5 w-3.5" />
                        Awaiting review
                      </span>
                    )}

                    {isManagerOrAdmin &&
                      (isEditingStatus ? (
                        <div className="max-w-md space-y-2 rounded-xl border border-gray-200 bg-gray-50 p-3">
                          <textarea
                            placeholder="Add a note for this decision (optional)"
                            value={comment[doc.id] || ''}
                            onChange={(e) =>
                              setComment((prev) => ({ ...prev, [doc.id]: e.target.value }))
                            }
                            onKeyDown={(e) => {
                              if (e.key === 'Escape') setStatusEditorOpenFor(null);
                            }}
                            className="w-full resize-none rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm text-gray-900 placeholder:text-gray-400 focus:border-indigo-400 focus:outline-none focus:ring-2 focus:ring-indigo-500/20"
                            rows={2}
                            autoFocus
                          />
                          <div className="flex items-center gap-2">
                            <button
                              onClick={() => handleStatusUpdate(doc.id, 'approved')}
                              disabled={processingStatus === doc.id}
                              className="inline-flex items-center gap-1.5 rounded-lg bg-emerald-600 px-3 py-1.5 text-xs font-medium text-white transition-colors hover:bg-emerald-700 disabled:cursor-not-allowed disabled:opacity-50"
                            >
                              {processingStatus === doc.id ? (
                                <Loader2 className="h-3.5 w-3.5 animate-spin" />
                              ) : (
                                <CheckCircle className="h-3.5 w-3.5" />
                              )}
                              Approve
                            </button>
                            <button
                              onClick={() => handleStatusUpdate(doc.id, 'rejected')}
                              disabled={processingStatus === doc.id}
                              className="inline-flex items-center gap-1.5 rounded-lg bg-rose-600 px-3 py-1.5 text-xs font-medium text-white transition-colors hover:bg-rose-700 disabled:cursor-not-allowed disabled:opacity-50"
                            >
                              {processingStatus === doc.id ? (
                                <Loader2 className="h-3.5 w-3.5 animate-spin" />
                              ) : (
                                <XCircle className="h-3.5 w-3.5" />
                              )}
                              Reject
                            </button>
                            <button
                              onClick={() => setStatusEditorOpenFor(null)}
                              disabled={processingStatus === doc.id}
                              className="rounded-lg p-1.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600 disabled:cursor-not-allowed disabled:opacity-50"
                              title="Cancel"
                              aria-label="Cancel review"
                            >
                              <X className="h-4 w-4" />
                            </button>
                          </div>
                        </div>
                      ) : (
                        <button
                          onClick={() => setStatusEditorOpenFor(doc.id)}
                          className="inline-flex items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-2.5 py-1 text-xs font-medium text-gray-600 transition-colors hover:border-indigo-200 hover:bg-indigo-50 hover:text-indigo-700 focus:outline-none focus-visible:ring-2 focus-visible:ring-indigo-500/40"
                        >
                          <ShieldCheck className="h-3.5 w-3.5" />
                          {isFinalized ? 'Change decision' : 'Review'}
                        </button>
                      ))}
                  </div>
                </li>
              );
            })}
          </ul>
        )}
      </div>

      {/* Version History Modal */}
      {selectedDocForHistory && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-gray-900/40 p-4 backdrop-blur-sm">
          <div className="flex max-h-[80vh] w-full max-w-lg flex-col overflow-hidden rounded-2xl border border-gray-100 bg-white shadow-xl">
            <div className="flex items-center justify-between border-b border-gray-100 px-6 py-4">
              <div className="min-w-0">
                <h3 className="text-sm font-semibold text-gray-900">Version history</h3>
                <p className="truncate text-xs text-gray-400">{selectedDocForHistory.name}</p>
              </div>
              <button
                onClick={() => setSelectedDocForHistory(null)}
                className="rounded-full p-1.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600"
                aria-label="Close version history"
              >
                <X className="h-5 w-5" />
              </button>
            </div>

            <div className="flex-1 space-y-2 overflow-y-auto px-6 py-4">
              {loadingVersions ? (
                <div className="space-y-2">
                  {Array.from({ length: 3 }).map((_, i) => (
                    <div key={i} className="h-14 animate-pulse rounded-xl bg-gray-100" />
                  ))}
                </div>
              ) : versions.length === 0 ? (
                <div className="flex flex-col items-center gap-2 py-8 text-center">
                  <History className="h-6 w-6 text-gray-300" />
                  <p className="text-sm text-gray-400">No version history found.</p>
                </div>
              ) : (
                versions.map((v) => (
                  <div
                    key={v.id}
                    className="flex items-center justify-between gap-4 rounded-xl border border-gray-100 p-3 transition-colors hover:border-gray-200"
                  >
                    <div className="min-w-0 flex-1">
                      <div className="flex items-center gap-2">
                        <span className="rounded-md bg-indigo-50 px-2 py-0.5 text-xs font-semibold text-indigo-700">
                          v{v.version}
                        </span>
                        <span className="text-xs text-gray-400">
                          {v.created_at ? new Date(v.created_at).toLocaleString() : 'N/A'}
                        </span>
                      </div>
                      <p className="mt-1 truncate font-mono text-[11px] text-gray-400">
                        {v.hashed_string}
                      </p>
                    </div>

                    <button
                      onClick={() => handleDownloadVersionFile(v.id, selectedDocForHistory.name)}
                      className="inline-flex flex-shrink-0 items-center gap-1.5 rounded-lg border border-gray-200 px-2.5 py-1.5 text-xs font-medium text-gray-600 transition-colors hover:border-indigo-200 hover:bg-indigo-50 hover:text-indigo-700"
                      title={`Download version ${v.version}`}
                    >
                      <Download className="h-3.5 w-3.5" />
                      v{v.version}
                    </button>
                  </div>
                ))
              )}
            </div>

            <div className="flex justify-end border-t border-gray-100 px-6 py-3">
              <button
                onClick={() => setSelectedDocForHistory(null)}
                className="rounded-lg px-4 py-2 text-sm font-medium text-gray-600 transition-colors hover:bg-gray-100"
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

// ---- Shared review summary panel ----
function ReviewSummary({
  status,
  comment,
}: {
  status: string;
  comment?: string | null;
}) {
  const isApproved = status === 'approved';
  const Icon = isApproved ? CheckCircle : XCircle;
  const label = isApproved ? 'Approved' : 'Rejected';

  return (
    <div
      className={`max-w-md rounded-r-lg border-l-4 px-3 py-2 ${
        isApproved ? 'border-emerald-400 bg-emerald-50' : 'border-rose-400 bg-rose-50'
      }`}
    >
      <div
        className={`flex items-center gap-1.5 text-sm font-medium ${
          isApproved ? 'text-emerald-700' : 'text-rose-700'
        }`}
      >
        <Icon className="h-4 w-4" />
        {label}
      </div>
      {comment ? (
        <p className="mt-1 flex items-start gap-1.5 text-xs text-gray-600">
          <MessageSquareText className="mt-0.5 h-3.5 w-3.5 flex-shrink-0 opacity-60" />
          <span className="italic">{comment}</span>
        </p>
      ) : (
        <p className="mt-1 text-xs text-gray-400">No comment added</p>
      )}
    </div>
  );
}