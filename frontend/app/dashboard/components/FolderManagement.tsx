'use client';

import { Folder, User } from '@/type';
import { deleteFolder } from '@/lib/api';
import {
  Trash2,
  Plus,
  Pencil,
  ChevronDown,
  ChevronRight,
  Folder as FolderIcon,
  FolderOpen,
  Calendar,
  FileText,
} from 'lucide-react';
import toast from 'react-hot-toast';
import React, { useState } from 'react';
import CreateFolderModal from './CreateFolderModal';
import EditFolderModal from './EditFolderModal';
import DocumentManager from './DocumentManager';

interface Props {
  folders: Folder[];
  onFolderChange: () => void;
  canCreate: boolean;
  canDelete: boolean;
  users: User[];            // ← new
  currentUser: User | null; // ← new
}

export default function FolderManagement({
  folders,
  onFolderChange,
  canCreate,
  canDelete,
  users,
  currentUser,
}: Props) {
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [editingFolder, setEditingFolder] = useState<{ id: number; name: string } | null>(null);
  const [expandedFolders, setExpandedFolders] = useState<Set<number>>(new Set());

  const toggleExpand = (folderId: number) => {
    setExpandedFolders((prev) => {
      const next = new Set(prev);
      next.has(folderId) ? next.delete(folderId) : next.add(folderId);
      return next;
    });
  };

  const handleDelete = async (id: number) => {
    if (!confirm('Delete this folder and all its documents?')) return;
    try {
      await deleteFolder(id);
      toast.success('Folder deleted');
      onFolderChange();
    } catch (error: any) {
      toast.error(error.message || 'Failed to delete folder');
    }
  };

  const formatDate = (value?: string) => {
    if (!value) return '—';
    const d = new Date(value);
    if (isNaN(d.getTime())) return '—';
    return d.toLocaleDateString(undefined, {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  };

  const colSpan = canDelete ? 4 : 3;

  return (
    <div className="space-y-5">
      {/* ---------------------------- Header ---------------------------- */}
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-gradient-to-br from-blue-500 to-indigo-600 text-white shadow-lg shadow-blue-500/25">
            <FolderIcon className="h-5 w-5" />
          </div>
          <div>
            <h3 className="text-base font-bold text-slate-900">Folders</h3>
            <p className="text-xs text-slate-500">
              {folders.length} {folders.length === 1 ? 'folder' : 'folders'} in workspace
            </p>
          </div>
        </div>

        {canCreate && (
          <button
            onClick={() => setShowCreateModal(true)}
            className="group inline-flex items-center gap-2 rounded-xl bg-gradient-to-br from-blue-600 to-indigo-600 px-4 py-2.5 text-sm font-semibold text-white shadow-md shadow-blue-500/20 transition-all hover:-translate-y-0.5 hover:shadow-lg hover:shadow-blue-500/30 active:translate-y-0"
          >
            <Plus className="h-4 w-4 transition-transform group-hover:rotate-90" />
            New Folder
          </button>
        )}
      </div>

      {/* ---------------------------- Table ---------------------------- */}
      {folders.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-2xl border border-dashed border-slate-200 bg-slate-50/60 px-6 py-14 text-center">
          <div className="mb-3 flex h-14 w-14 items-center justify-center rounded-2xl bg-white text-slate-300 shadow-sm ring-1 ring-slate-200">
            <FolderIcon className="h-7 w-7" />
          </div>
          <p className="text-sm font-semibold text-slate-700">No folders yet</p>
          <p className="mt-1 max-w-xs text-xs text-slate-400">
            Create your first folder to start organizing documents.
          </p>
          {canCreate && (
            <button
              onClick={() => setShowCreateModal(true)}
              className="mt-4 inline-flex items-center gap-1.5 rounded-lg bg-slate-900 px-4 py-2 text-xs font-semibold text-white transition-colors hover:bg-slate-800"
            >
              <Plus className="h-3.5 w-3.5" />
              Create folder
            </button>
          )}
        </div>
      ) : (
        <div className="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
          <div className="overflow-x-auto">
            <table className="min-w-full">
              {/* Head */}
              <thead>
                <tr className="border-b border-slate-100 bg-slate-50/70">
                  <th className="px-5 py-3 text-left text-[11px] font-bold uppercase tracking-wider text-slate-400">
                    Folder
                  </th>
                  <th className="hidden px-5 py-3 text-left text-[11px] font-bold uppercase tracking-wider text-slate-400 sm:table-cell">
                    Created By
                  </th>
                  <th className="hidden px-5 py-3 text-left text-[11px] font-bold uppercase tracking-wider text-slate-400 md:table-cell">
                    Created
                  </th>
                  {canDelete && (
                    <th className="px-5 py-3 text-right text-[11px] font-bold uppercase tracking-wider text-slate-400">
                      Actions
                    </th>
                  )}
                </tr>
              </thead>

              {/* Body */}
              <tbody className="divide-y divide-slate-100">
                {folders.map((f) => {
                  const isExpanded = expandedFolders.has(f.id);
                  return (
                    <React.Fragment key={f.id}>
                      <tr
                        className={`group transition-colors ${
                          isExpanded ? 'bg-blue-50/40' : 'hover:bg-slate-50'
                        }`}
                      >
                        {/* Folder name cell */}
                        <td className="px-5 py-4">
                          <button
                            onClick={() => toggleExpand(f.id)}
                            className="flex w-full items-center gap-3 text-left"
                          >
                            <span
                              className={`flex h-6 w-6 shrink-0 items-center justify-center rounded-lg transition-all ${
                                isExpanded
                                  ? 'bg-blue-100 text-blue-600'
                                  : 'bg-slate-100 text-slate-400 group-hover:bg-slate-200 group-hover:text-slate-600'
                              }`}
                            >
                              {isExpanded ? (
                                <ChevronDown className="h-3.5 w-3.5" strokeWidth={2.5} />
                              ) : (
                                <ChevronRight className="h-3.5 w-3.5" strokeWidth={2.5} />
                              )}
                            </span>

                            <span
                              className={`flex h-10 w-10 shrink-0 items-center justify-center rounded-xl transition-all ${
                                isExpanded
                                  ? 'bg-gradient-to-br from-blue-500 to-indigo-600 text-white shadow-md shadow-blue-500/25'
                                  : 'bg-blue-50 text-blue-600 group-hover:bg-blue-100'
                              }`}
                            >
                              {isExpanded ? (
                                <FolderOpen className="h-5 w-5" />
                              ) : (
                                <FolderIcon className="h-5 w-5" />
                              )}
                            </span>

                            <span className="min-w-0 flex-1">
                              <span
                                className={`block truncate text-sm font-semibold transition-colors ${
                                  isExpanded
                                    ? 'text-blue-700'
                                    : 'text-slate-800 group-hover:text-blue-700'
                                }`}
                              >
                                {f.name}
                              </span>
                              <span className="mt-0.5 flex items-center gap-1 text-[11px] text-slate-400">
                                <FileText className="h-3 w-3" />
                                {isExpanded ? 'Click to collapse' : 'Click to open documents'}
                              </span>
                            </span>
                          </button>
                        </td>

                        {/* Created by */}
                        <td className="hidden whitespace-nowrap px-5 py-4 text-sm text-slate-600 sm:table-cell">
                          <span className="inline-flex items-center gap-2">
                            <span className="flex h-6 w-6 items-center justify-center rounded-full bg-slate-100 text-[10px] font-bold uppercase text-slate-500">
                              {(f.created_by as number)?.toString()?.charAt(0) ?? '?'}
                            </span>
                            <span className="truncate">{f.created_by || '—'}</span>
                          </span>
                        </td>

                        {/* Date */}
                        <td className="hidden whitespace-nowrap px-5 py-4 text-sm text-slate-500 md:table-cell">
                          <span className="inline-flex items-center gap-1.5">
                            <Calendar className="h-3.5 w-3.5 text-slate-400" />
                            {formatDate(f.created_at as any)}
                          </span>
                        </td>

                        {/* Actions */}
                        {canDelete && (
                          <td className="whitespace-nowrap px-5 py-4 text-right">
                            <div className="inline-flex items-center gap-1">
                              <button
                                onClick={() =>
                                  setEditingFolder({ id: f.id, name: f.name })
                                }
                                title="Edit folder"
                                className="flex h-8 w-8 items-center justify-center rounded-lg text-slate-400 transition-all hover:bg-blue-50 hover:text-blue-600"
                              >
                                <Pencil className="h-4 w-4" />
                              </button>
                              <button
                                onClick={() => handleDelete(f.id)}
                                title="Delete folder"
                                className="flex h-8 w-8 items-center justify-center rounded-lg text-slate-400 transition-all hover:bg-rose-50 hover:text-rose-600"
                              >
                                <Trash2 className="h-4 w-4" />
                              </button>
                            </div>
                          </td>
                        )}
                      </tr>

                      {/* Expanded documents row */}
                      {isExpanded && (
                        <tr key={`${f.id}-docs`}>
                          <td
                            colSpan={colSpan}
                            className="border-l-2 border-blue-500 bg-slate-50/50 px-5 py-5"
                          >
                            <div className="mb-4 flex items-center gap-2.5">
                              <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-blue-500 to-indigo-600 text-white shadow-sm">
                                <FolderOpen className="h-4 w-4" />
                              </span>
                              <div className="min-w-0">
                                <p className="text-[10px] font-bold uppercase tracking-wider text-slate-400">
                                  Documents in
                                </p>
                                <h4 className="truncate text-sm font-bold text-slate-900">
                                  {f.name}
                                </h4>
                              </div>
                            </div>

                            {/* Pass users + currentUser down so DocumentManager
                                can open the ShareModal from any document row. */}
                            <DocumentManager
  folderId={f.id}
  canUpload={true}
  canDelete={canDelete}
  allFolders={folders}
  users={users}
  userRole={currentUser?.role}
/>
                          </td>
                        </tr>
                      )}
                    </React.Fragment>
                  );
                })}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* ---------------------------- Modals ---------------------------- */}
      {showCreateModal && (
        <CreateFolderModal
          onClose={() => setShowCreateModal(false)}
          onSuccess={onFolderChange}
        />
      )}
      {editingFolder && (
        <EditFolderModal
          folderId={editingFolder.id}
          currentName={editingFolder.name}
          onClose={() => setEditingFolder(null)}
          onSuccess={() => {
            setEditingFolder(null);
            onFolderChange();
          }}
        />
      )}
    </div>
  );
}