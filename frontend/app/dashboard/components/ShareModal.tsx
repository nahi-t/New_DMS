'use client';

import { useCallback, useEffect, useState } from 'react';
import {
  shareDocument,
  unshareDocument,
  getDocumentShares,
} from '@/lib/api';
import type { DocumentShare, User } from '@/type';
import toast from 'react-hot-toast';

/* ------------------------------------------------------------------ */
/*  Icons                                                              */
/* ------------------------------------------------------------------ */
const ShareIcons = {
  close: (
    <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
    </svg>
  ),
  trash: (
    <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
    </svg>
  ),
};

/* ------------------------------------------------------------------ */
/*  Component                                                          */
/* ------------------------------------------------------------------ */

export default function ShareModal({
  docId,
  docName,
  allUsers,
  currentUserId,
  onClose,
}: {
  docId: number;
  docName: string;
  allUsers: User[];
  currentUserId: number;
  onClose: () => void;
}) {
  const [shares, setShares] = useState<DocumentShare[]>([]);
  const [selectedUser, setSelectedUser] = useState('');
  const [permission, setPermission] = useState<'viewer' | 'editor'>('viewer');
  const [loading, setLoading] = useState(false);
  const [initialLoading, setInitialLoading] = useState(true);
  const [forbidden, setForbidden] = useState(false);

  const loadShares = useCallback(async () => {
    try {
      const data = await getDocumentShares(docId);
      setShares(data ?? []);
      setForbidden(false);
    } catch (e: any) {
      const msg = String(e.message || '');
      if (msg.toLowerCase().includes('forbidden') || msg.includes('403')) {
        setForbidden(true);
      } else {
        toast.error(msg || 'Failed to load shares');
      }
    } finally {
      setInitialLoading(false);
    }
  }, [docId]);

  useEffect(() => {
    loadShares();
  }, [loadShares]);

  const handleShare = async () => {
    if (!selectedUser) {
      toast.error('Select a member first');
      return;
    }
    setLoading(true);
    try {
      await shareDocument(docId, Number(selectedUser), permission);
      toast.success('Document shared');
      setSelectedUser('');
      setPermission('viewer');
      await loadShares();
    } catch (e: any) {
      const msg = String(e.message || '');
      if (msg.toLowerCase().includes('forbidden')) {
        toast.error('You do not have permission to share this document');
      } else {
        toast.error(msg || 'Failed to share');
      }
    } finally {
      setLoading(false);
    }
  };

  const handleUnshare = async (userId: number) => {
    try {
      await unshareDocument(docId, userId);
      toast.success('Access removed');
      await loadShares();
    } catch (e: any) {
      const msg = String(e.message || '');
      if (msg.toLowerCase().includes('forbidden')) {
        toast.error('You do not have permission to remove access');
      } else {
        toast.error(msg || 'Failed to remove');
      }
    }
  };

  const availableUsers = allUsers.filter(
    (u) => u.id !== currentUserId && !shares.some((s) => s.user_id === u.id)
  );

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/60 p-4 backdrop-blur-sm"
      onClick={onClose}
    >
      <div
        className="w-full max-w-lg overflow-hidden rounded-2xl bg-white shadow-2xl"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className="flex items-start justify-between gap-3 border-b border-slate-100 px-6 py-4">
          <div className="min-w-0">
            <h2 className="text-base font-bold text-slate-900">Share document</h2>
            <p className="mt-0.5 truncate text-xs text-slate-500">{docName}</p>
          </div>
          <button
            onClick={onClose}
            className="-mr-1 rounded-lg p-1.5 text-slate-400 transition hover:bg-slate-100 hover:text-slate-700"
            aria-label="Close"
          >
            {ShareIcons.close}
          </button>
        </div>

        {/* Forbidden notice */}
        {forbidden ? (
          <div className="px-6 py-10 text-center">
            <p className="text-sm font-semibold text-rose-700">
              You do not have permission to manage sharing for this document.
            </p>
            <p className="mt-1 text-xs text-slate-500">
              Only admins, managers, or the folder owner can share.
            </p>
          </div>
        ) : (
          <>
            {/* Add share form */}
            <div className="border-b border-slate-100 bg-slate-50/60 px-6 py-4">
              <p className="mb-2 text-[11px] font-bold uppercase tracking-wider text-slate-500">
                Add people
              </p>
              <div className="flex gap-2">
                <select
                  value={selectedUser}
                  onChange={(e) => setSelectedUser(e.target.value)}
                  className="min-w-0 flex-1 rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm text-slate-800 outline-none transition focus:border-blue-400 focus:ring-2 focus:ring-blue-100"
                >
                  <option value="">Select a member…</option>
                  {availableUsers.map((u) => (
                    <option key={u.id} value={u.id}>
                      {u.username} — {u.email}
                    </option>
                  ))}
                </select>

                <select
                  value={permission}
                  onChange={(e) => setPermission(e.target.value as 'viewer' | 'editor')}
                  className="rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm text-slate-800 outline-none transition focus:border-blue-400 focus:ring-2 focus:ring-blue-100"
                >
                  <option value="viewer">Viewer</option>
                  <option value="editor">Editor</option>
                </select>

                <button
                  onClick={handleShare}
                  disabled={loading || !selectedUser}
                  className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-semibold text-white shadow-sm transition hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-50"
                >
                  {loading ? 'Sharing…' : 'Share'}
                </button>
              </div>
            </div>

            {/* Current shares */}
            <div className="max-h-80 overflow-y-auto px-6 py-4">
              <p className="mb-3 text-[11px] font-bold uppercase tracking-wider text-slate-500">
                People with access ({shares.length})
              </p>

              {initialLoading ? (
                <div className="flex items-center justify-center gap-2 py-8 text-slate-400">
                  <div className="h-4 w-4 animate-spin rounded-full border-2 border-slate-200 border-t-blue-600" />
                  <span className="text-sm">Loading…</span>
                </div>
              ) : shares.length === 0 ? (
                <p className="py-8 text-center text-sm text-slate-400">
                  Not shared with anyone yet.
                </p>
              ) : (
                <ul className="space-y-2">
                  {shares.map((s) => (
                    <li
                      key={s.id}
                      className="flex items-center justify-between gap-3 rounded-xl border border-slate-100 bg-white px-3 py-2.5"
                    >
                      <div className="flex min-w-0 items-center gap-2.5">
                        <span className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-gradient-to-br from-blue-600 to-indigo-600 text-[11px] font-bold uppercase text-white">
                          {s.user_name?.charAt(0) || '?'}
                        </span>
                        <div className="min-w-0">
                          <p className="truncate text-sm font-medium text-slate-800">
                            {s.user_name}
                          </p>
                          <p className="truncate text-[11px] text-slate-400">
                            {s.user_email}
                          </p>
                        </div>
                      </div>

                      <div className="flex shrink-0 items-center gap-2">
                        <span
                          className={`rounded-full px-2.5 py-1 text-[10px] font-bold uppercase tracking-wide ${
                            s.permission === 'editor'
                              ? 'bg-amber-100 text-amber-700'
                              : 'bg-blue-100 text-blue-700'
                          }`}
                        >
                          {s.permission}
                        </span>
                        <button
                          onClick={() => handleUnshare(s.user_id)}
                          className="rounded-lg p-1.5 text-slate-400 transition hover:bg-rose-50 hover:text-rose-600"
                          title="Remove access"
                        >
                          {ShareIcons.trash}
                        </button>
                      </div>
                    </li>
                  ))}
                </ul>
              )}
            </div>
          </>
        )}

        {/* Footer */}
        <div className="flex justify-end gap-2 border-t border-slate-100 bg-slate-50/60 px-6 py-3">
          <button
            onClick={onClose}
            className="rounded-lg border border-slate-200 bg-white px-4 py-2 text-sm font-medium text-slate-600 transition hover:bg-slate-50"
          >
            Done
          </button>
        </div>
      </div>
    </div>
  );
}