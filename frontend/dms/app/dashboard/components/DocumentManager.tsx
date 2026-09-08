'use client';

import { useState, useEffect, useCallback } from 'react';
import { 
  getDocuments, uploadDocument, deleteDocument, downloadDocument, 
  renameDocument, moveDocument, searchDocuments, updateDocumentContent,
  getDocumentVersions, downloadDocumentVersion, DocumentVersion 
} from '@/lib/api';
import { Document, Folder } from '@/type';
import { Upload, Trash2, Download, FileText, Pencil, FolderInput, X, Check, RefreshCw, History } from 'lucide-react';
import toast from 'react-hot-toast';

interface Props {
  folderId: number;
  canUpload: boolean;
  canDelete?: boolean;
  allFolders: Folder[]; // needed for move dropdown
}

export default function DocumentManager({ folderId, canUpload, canDelete = true, allFolders }: Props) {
  const [documents, setDocuments] = useState<Document[]>([]);
  const [loading, setLoading] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [searchTerm, setSearchTerm] = useState('');
  const [editingName, setEditingName] = useState<number | null>(null);
  const [newName, setNewName] = useState('');
  const [movingDoc, setMovingDoc] = useState<number | null>(null);
  const [updatingDocId, setUpdatingDocId] = useState<number | null>(null);

  // Version History Modal State
  const [selectedDocForHistory, setSelectedDocForHistory] = useState<Document | null>(null);
  const [versions, setVersions] = useState<DocumentVersion[]>([]);
  const [loadingVersions, setLoadingVersions] = useState(false);

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

  const formatFileSize = (bytes: number) => {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1048576) return (bytes / 1024).toFixed(1) + ' KB';
    return (bytes / 1048576).toFixed(1) + ' MB';
  };

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
        {documents.map((doc) => (
          <li key={doc.id} className="py-2 flex items-center justify-between">
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
                    <p className="text-sm font-medium text-gray-900 dark:text-white truncate">
                      {doc.name} {doc.version ? <span className="text-xs text-blue-500 font-semibold">(v{doc.version})</span> : null}
                    </p>
                    <p className="text-xs text-gray-500 dark:text-gray-400">
                      {formatFileSize(doc.size)} • {doc.uploaded_at ? new Date(doc.uploaded_at).toLocaleDateString() : ''}
                    </p>
                  </>
                )}
              </div>
            </div>

            <div className="flex items-center gap-2 flex-shrink-0">
              {/* Version History Button */}
              <button
                onClick={() => handleOpenHistory(doc)}
                className="text-gray-600 hover:text-gray-800 dark:text-gray-400 dark:hover:text-gray-200 transition"
                title="View Version History"
              >
                <History className="w-4 h-4" />
              </button>

              {/* Update Content File Input (New Version) */}
              <label 
                className={`cursor-pointer text-indigo-600 hover:text-indigo-800 dark:text-indigo-400 dark:hover:text-indigo-300 transition ${updatingDocId === doc.id ? 'opacity-50 cursor-not-allowed' : ''}`}
                title="Upload new version"
              >
                <RefreshCw className={`w-4 h-4 ${updatingDocId === doc.id ? 'animate-spin' : ''}`} />
                <input
                  type="file"
                  onChange={(e) => handleUpdateContent(doc.id, e)}
                  className="hidden"
                  disabled={updatingDocId === doc.id}
                />
              </label>

              {/* Download */}
              <button
                onClick={() => handleDownload(doc)}
                className="text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300 transition"
                title="Download Latest"
              >
                <Download className="w-4 h-4" />
              </button>

              {/* Rename */}
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

              {/* Move */}
              {allFolders.length > 0 && (
                <div className="relative">
                  <button
                    onClick={() => setMovingDoc(movingDoc === doc.id ? null : doc.id)}
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
                          className={`block w-full text-left px-4 py-2 text-sm hover:bg-gray-100 dark:hover:bg-gray-700 ${f.id === folderId ? 'text-gray-400 cursor-not-allowed' : 'text-gray-700 dark:text-gray-200'}`}
                          disabled={f.id === folderId}
                        >
                          {f.name} {f.id === folderId && '(current)'}
                        </button>
                      ))}
                    </div>
                  )}
                </div>
              )}

              {/* Delete */}
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
          </li>
        ))}
      </ul>

      {/* Version History Modal */}
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
                          Uploaded: {v.created_at ? new Date(v.created_at).toLocaleString() : 'N/A'}
                        </p>
                        <p className="text-xs text-gray-400 dark:text-gray-500 font-mono mt-1 truncate">
                          Hash: {v.hashed_string}
                        </p>
                      </div>

                      <button
                        onClick={() => handleDownloadVersionFile(v.id, selectedDocForHistory.name)}
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