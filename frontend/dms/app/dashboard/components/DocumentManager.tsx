'use client';

import { useState, useEffect, useCallback } from 'react';
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
  Upload,
  Trash2,
  Download,
  FileText,
  Pencil,
  FolderInput,
  X,
  Check,
  RefreshCw,
  History,
  CheckCircle,
  XCircle,
  AlertCircle,
  Clock,
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

  // Which document's "Update Status" editor (comment + Approve/Reject) is
  // currently open. Only one at a time, and only relevant for managers/admins.
  const [statusEditorOpenFor, setStatusEditorOpenFor] = useState<number | null>(null);

  const [selectedDocForHistory, setSelectedDocForHistory] = useState<Document | null>(null);
  const [versions, setVersions] = useState<DocumentVersion[]>([]);
  const [loadingVersions, setLoadingVersions] = useState(false);

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

  const handleDownloadVersionFile = async (versionId: number, docName: string) => {
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

  const renderStatusBadge = (status: string) => {
    if (!status) return null;
    const base =
      'inline-flex items-center gap-1 text-xs font-medium px-2 py-0.5 rounded-full mt-1';
    let colorClass = '';
    let Icon = AlertCircle;
    if (status === 'approved') {
      colorClass = 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200';
      Icon = CheckCircle;
    } else if (status === 'rejected') {
      colorClass = 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200';
      Icon = XCircle;
    } else {
      colorClass = 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200';
      Icon = AlertCircle;
    }
    return (
      <span className={`${base} ${colorClass}`}>
        <Icon className="w-3 h-3" />
        {status}
      </span>
    );
  };

  const isManagerOrAdmin = (() => {
    const role = effectiveRole.toLowerCase();
    return role === 'manager' || role === 'admin';
  })();

  return (
    <div className="mt-4 space-y-4">
      {/* Search and Upload bar */}
      <div className="flex flex-wrap items-center gap-2">
        <div className="flex-1 min-w-[200px]">
          <input
            type="text"
            placeholder="Search documents..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-gray-50 dark:bg-gray-700 p-2 text-sm text-gray-900 dark:text-white"
          />
        </div>
        {canUpload && (
          <label className="cursor-pointer bg-blue-500 hover:bg-blue-600 text-white px-3 py-1.5 rounded-md text-sm flex items-center gap-2 transition whitespace-nowrap">
            <Upload className="w-4 h-4" />
            Upload
            <input
              type="file"
              onChange={handleUpload}
              className="hidden"
              disabled={uploading}
            />
          </label>
        )}
      </div>

      {loading && (
        <div className="flex justify-center py-4">
          <div className="animate-spin rounded-full h-6 w-6 border-t-2 border-b-2 border-blue-500"></div>
        </div>
      )}

      {!loading && documents.length === 0 && (
        <p className="text-sm text-gray-500 dark:text-gray-400 italic">
          {searchTerm ? 'No matching documents' : 'No documents yet'}
        </p>
      )}

      <ul className="divide-y divide-gray-200 dark:divide-gray-700">
        {documents.map((doc) => {
          const normalizedStatus = getRawStatus(doc);
          const docComment = getRawComment(doc);
          const isFinalized =
            normalizedStatus === 'approved' || normalizedStatus === 'rejected';
          const isEditingStatus = statusEditorOpenFor === doc.id;

          return (
            <li key={doc.id} className="py-2 flex flex-col gap-2">
              {/* Main row: icon, name, actions */}
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3 min-w-0 flex-1">
                  <FileText className="w-5 h-5 text-gray-400 flex-shrink-0" />
                  <div className="min-w-0 flex-1">
                    {editingName === doc.id ? (
                      <div className="flex items-center gap-2">
                        <input
                          type="text"
                          value={newName}
                          onChange={(e) => setNewName(e.target.value)}
                          className="border border-gray-300 dark:border-gray-600 rounded px-2 py-1 text-sm w-full max-w-xs bg-white dark:bg-gray-800 text-gray-900 dark:text-white"
                          autoFocus
                        />
                        <button
                          onClick={() => handleRename(doc.id)}
                          className="text-green-600 hover:text-green-800"
                        >
                          <Check className="w-4 h-4" />
                        </button>
                        <button
                          onClick={() => setEditingName(null)}
                          className="text-gray-500 hover:text-gray-700"
                        >
                          <X className="w-4 h-4" />
                        </button>
                      </div>
                    ) : (
                      <>
                        <div className="flex items-center gap-2 flex-wrap">
                          <p className="text-sm font-medium text-gray-900 dark:text-white truncate">
                            {doc.name}{' '}
                            {doc.version ? (
                              <span className="text-xs text-blue-500 font-semibold">
                                (v{doc.version})
                              </span>
                            ) : null}
                          </p>
                          {renderStatusBadge(normalizedStatus)}
                        </div>
                        <p className="text-xs text-gray-500 dark:text-gray-400">
                          {formatFileSize(doc.size)} •{' '}
                          {doc.uploaded_at
                            ? new Date(doc.uploaded_at).toLocaleDateString()
                            : ''}
                        </p>
                      </>
                    )}
                  </div>
                </div>

                {/* Action buttons */}
                <div className="flex items-center gap-2 flex-shrink-0">
                  <button
                    onClick={() => handleOpenHistory(doc)}
                    className="text-gray-600 hover:text-gray-800 dark:text-gray-400 dark:hover:text-gray-200 transition"
                    title="View Version History"
                  >
                    <History className="w-4 h-4" />
                  </button>

                  <label
                    className={`cursor-pointer text-indigo-600 hover:text-indigo-800 dark:text-indigo-400 dark:hover:text-indigo-300 transition ${
                      updatingDocId === doc.id ? 'opacity-50 cursor-not-allowed' : ''
                    }`}
                    title="Upload new version"
                  >
                    <RefreshCw
                      className={`w-4 h-4 ${updatingDocId === doc.id ? 'animate-spin' : ''}`}
                    />
                    <input
                      type="file"
                      onChange={(e) => handleUpdateContent(doc.id, e)}
                      className="hidden"
                      disabled={updatingDocId === doc.id}
                    />
                  </label>

                  <button
                    onClick={() => handleDownload(doc)}
                    className="text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300 transition"
                    title="Download Latest"
                  >
                    <Download className="w-4 h-4" />
                  </button>

                  <button
                    onClick={() => {
                      setEditingName(doc.id);
                      setNewName(doc.name);
                    }}
                    className="text-yellow-600 hover:text-yellow-800 dark:text-yellow-400 dark:hover:text-yellow-300 transition"
                    title="Rename"
                  >
                    <Pencil className="w-4 h-4" />
                  </button>

                  {allFolders.length > 0 && (
                    <div className="relative">
                      <button
                        onClick={() =>
                          setMovingDoc(movingDoc === doc.id ? null : doc.id)
                        }
                        className="text-purple-600 hover:text-purple-800 dark:text-purple-400 dark:hover:text-purple-300 transition"
                        title="Move to folder"
                      >
                        <FolderInput className="w-4 h-4" />
                      </button>
                      {movingDoc === doc.id && (
                        <div className="absolute right-0 mt-2 w-48 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-lg z-10 max-h-60 overflow-y-auto">
                          {allFolders.map((f) => (
                            <button
                              key={f.id}
                              onClick={() => handleMove(doc.id, f.id)}
                              className={`block w-full text-left px-4 py-2 text-sm hover:bg-gray-100 dark:hover:bg-gray-700 ${
                                f.id === folderId
                                  ? 'text-gray-400 cursor-not-allowed'
                                  : 'text-gray-700 dark:text-gray-200'
                              }`}
                              disabled={f.id === folderId}
                            >
                              {f.name} {f.id === folderId && '(current)'}
                            </button>
                          ))}
                        </div>
                      )}
                    </div>
                  )}

                  {canDelete && (
                    <button
                      onClick={() => handleDelete(doc.id)}
                      className="text-red-600 hover:text-red-800 dark:text-red-400 dark:hover:text-red-300 transition"
                      title="Delete"
                    >
                      <Trash2 className="w-4 h-4" />
                    </button>
                  )}
                </div>
              </div>

              {/* ---- Manager/Admin: status + comment (if decided) plus an
                   "Update Status" button that opens Approve/Reject controls,
                   usable anytime — even to change a prior decision. ---- */}
              {isManagerOrAdmin && (
                <div className="ml-10 mt-1 space-y-2">
                  {isFinalized ? (
                    <ReviewSummary status={normalizedStatus} comment={docComment} />
                  ) : (
                    <span className="inline-flex items-center gap-1.5 text-xs text-gray-400 dark:text-gray-500">
                      <Clock className="w-3.5 h-3.5" />
                      Waiting for approval
                    </span>
                  )}

                  {isEditingStatus ? (
                    <div className="flex items-center gap-3">
                      <textarea
                        placeholder="Add a comment (optional)"
                        value={comment[doc.id] || ''}
                        onChange={(e) =>
                          setComment((prev) => ({ ...prev, [doc.id]: e.target.value }))
                        }
                        className="flex-1 min-w-[120px] text-sm border border-gray-300 dark:border-gray-600 rounded px-2 py-1 bg-white dark:bg-gray-800 text-gray-900 dark:text-white resize-none"
                        rows={1}
                        autoFocus
                      />
                      <button
                        onClick={() => handleStatusUpdate(doc.id, 'approved')}
                        disabled={processingStatus === doc.id}
                        className="bg-green-500 hover:bg-green-600 disabled:bg-green-300 text-white px-3 py-1.5 rounded-md text-xs flex items-center gap-1 transition"
                      >
                        <CheckCircle className="w-4 h-4" /> Approve
                      </button>
                      <button
                        onClick={() => handleStatusUpdate(doc.id, 'rejected')}
                        disabled={processingStatus === doc.id}
                        className="bg-red-500 hover:bg-red-600 disabled:bg-red-300 text-white px-3 py-1.5 rounded-md text-xs flex items-center gap-1 transition"
                      >
                        <XCircle className="w-4 h-4" /> Reject
                      </button>
                      <button
                        onClick={() => setStatusEditorOpenFor(null)}
                        disabled={processingStatus === doc.id}
                        className="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 transition"
                        title="Cancel"
                      >
                        <X className="w-4 h-4" />
                      </button>
                    </div>
                  ) : (
                    <button
                      onClick={() => setStatusEditorOpenFor(doc.id)}
                      className="inline-flex items-center gap-1.5 text-xs font-medium text-indigo-600 hover:text-indigo-800 dark:text-indigo-400 dark:hover:text-indigo-300 border border-indigo-200 dark:border-indigo-800 rounded-md px-2.5 py-1 transition"
                    >
                      <ShieldCheck className="w-3.5 h-3.5" />
                      Update Status
                    </button>
                  )}
                </div>
              )}

              {/* ---- Regular user: fetch-and-display only, read-only,
                   never any buttons regardless of status. ---- */}
              {!isManagerOrAdmin && (
                <div className="ml-10 mt-1">
                  {isFinalized ? (
                    <ReviewSummary status={normalizedStatus} comment={docComment} />
                  ) : (
                    <span className="inline-flex items-center gap-1.5 text-xs text-gray-400 dark:text-gray-500">
                      <Clock className="w-3.5 h-3.5" />
                      Waiting for approval
                    </span>
                  )}
                </div>
              )}
            </li>
          );
        })}
      </ul>

      {/* Version History Modal (unchanged) */}
      {selectedDocForHistory && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-xl w-full p-6 space-y-4 max-h-[80vh] flex flex-col">
            <div className="flex items-center justify-between border-b pb-3 border-gray-200 dark:border-gray-700">
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
                Version History: {selectedDocForHistory.name}
              </h3>
              <button
                onClick={() => setSelectedDocForHistory(null)}
                className="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            <div className="flex-1 overflow-y-auto space-y-3">
              {loadingVersions ? (
                <div className="flex justify-center py-6">
                  <div className="animate-spin rounded-full h-6 w-6 border-t-2 border-b-2 border-blue-500"></div>
                </div>
              ) : versions.length === 0 ? (
                <p className="text-sm text-gray-500 dark:text-gray-400 italic text-center py-4">
                  No version history found.
                </p>
              ) : (
                <div className="space-y-2">
                  {versions.map((v) => (
                    <div
                      key={v.id}
                      className="p-3 rounded-lg border border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-700/50 flex items-center justify-between gap-4"
                    >
                      <div className="min-w-0 flex-1">
                        <div className="flex items-center gap-2">
                          <p className="text-sm font-semibold text-gray-900 dark:text-white">
                            Version {v.version}
                          </p>
                        </div>
                        <p className="text-xs text-gray-500 dark:text-gray-400">
                          Uploaded:{' '}
                          {v.created_at
                            ? new Date(v.created_at).toLocaleString()
                            : 'N/A'}
                        </p>
                        <p className="text-xs text-gray-400 dark:text-gray-500 font-mono mt-1 truncate">
                          Hash: {v.hashed_string}
                        </p>
                      </div>

                      <button
                        onClick={() =>
                          handleDownloadVersionFile(
                            v.id,
                            selectedDocForHistory.name
                          )
                        }
                        className="bg-blue-500 hover:bg-blue-600 text-white px-3 py-1.5 rounded-md text-xs flex items-center gap-1.5 transition whitespace-nowrap flex-shrink-0"
                        title="Download this version"
                      >
                        <Download className="w-3.5 h-3.5" />
                        Download v{v.version}
                      </button>
                    </div>
                  ))}
                </div>
              )}
            </div>

            <div className="border-t pt-3 flex justify-end border-gray-200 dark:border-gray-700">
              <button
                onClick={() => setSelectedDocForHistory(null)}
                className="bg-gray-200 hover:bg-gray-300 dark:bg-gray-700 dark:hover:bg-gray-600 text-gray-800 dark:text-gray-200 px-4 py-2 rounded-md text-sm transition"
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
  const wrapperClass = isApproved
    ? 'border-green-500 bg-green-50 dark:bg-green-900/20'
    : 'border-red-500 bg-red-50 dark:bg-red-900/20';
  const labelClass = isApproved
    ? 'text-green-700 dark:text-green-300'
    : 'text-red-700 dark:text-red-300';

  return (
    <div
      className={`rounded-md border-l-4 px-3 py-2 max-w-md ${wrapperClass}`}
      style={{ borderRadius: '0 6px 6px 0' }}
    >
      <div className={`flex items-center gap-1.5 text-sm font-medium ${labelClass}`}>
        <Icon className="w-4 h-4" />
        {label}
      </div>
      {comment ? (
        <p className="mt-1 flex items-start gap-1.5 text-xs text-gray-600 dark:text-gray-300">
          <MessageSquareText className="w-3.5 h-3.5 mt-0.5 flex-shrink-0 opacity-60" />
          <span className="italic">{comment}</span>
        </p>
      ) : (
        <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">No comment added</p>
      )}
    </div>
  );
}