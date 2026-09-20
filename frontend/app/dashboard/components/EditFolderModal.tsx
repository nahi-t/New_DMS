'use client';

import { useState, useEffect } from 'react';
import { updateFolder } from '@/lib/api';
import {
  X,
  Folder as FolderIcon,
  FolderOpen,
  Save,
  AlertCircle,
  CheckCircle2,
} from 'lucide-react';
import toast from 'react-hot-toast';

interface Props {
  folderId: number;
  currentName: string;
  onClose: () => void;
  onSuccess: () => void;
}

export default function EditFolderModal({
  folderId,
  currentName,
  onClose,
  onSuccess,
}: Props) {
  const [name, setName] = useState('');
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    setName(currentName);
  }, [currentName]);

  const trimmedName = name.trim();
  const isUnchanged = trimmedName === currentName.trim();
  const isTooShort = trimmedName.length > 0 && trimmedName.length < 2;
  const isTooLong = trimmedName.length > 60;
  const isValid = trimmedName.length >= 2 && !isTooLong;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!isValid) {
      if (isTooShort) toast.error('Folder name must be at least 2 characters.');
      else if (isTooLong) toast.error('Folder name is too long (max 60 characters).');
      return;
    }

    setLoading(true);
    try {
      await updateFolder(folderId, trimmedName);
      toast.success('Folder updated');
      onSuccess();
      onClose();
    } catch (error: any) {
      toast.error(error.message || 'Failed to update folder');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/40 p-4 backdrop-blur-sm"
      onClick={onClose}
    >
      <div
        className="relative w-full max-w-md overflow-hidden rounded-3xl bg-white shadow-2xl shadow-slate-900/20"
        onClick={(e) => e.stopPropagation()}
      >
        {/* -------------------------- Header -------------------------- */}
        <div className="relative overflow-hidden bg-gradient-to-br from-indigo-600 via-blue-600 to-indigo-700 px-6 py-6">
          {/* Decorative blobs */}
          <div className="pointer-events-none absolute -right-12 -top-16 h-44 w-44 rounded-full bg-white/10 blur-3xl" />
          <div className="pointer-events-none absolute -bottom-20 left-8 h-36 w-36 rounded-full bg-blue-400/20 blur-3xl" />
          <div
            className="pointer-events-none absolute inset-0 opacity-[0.08]"
            style={{
              backgroundImage:
                'linear-gradient(to right, #fff 1px, transparent 1px), linear-gradient(to bottom, #fff 1px, transparent 1px)',
              backgroundSize: '28px 28px',
            }}
          />

          <div className="relative flex items-start justify-between gap-4">
            <div className="flex items-center gap-3.5">
              <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-white/15 text-white backdrop-blur-md ring-1 ring-white/20">
                <FolderOpen className="h-5.5 w-5.5" />
              </div>
              <div className="min-w-0">
                <h3 className="text-lg font-bold tracking-tight text-white">
                  Edit Folder
                </h3>
                <p className="mt-0.5 truncate text-xs text-blue-100">
                  Rename your workspace folder
                </p>
              </div>
            </div>

            <button
              onClick={onClose}
              className="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-white/10 text-white/80 transition-colors hover:bg-white/20 hover:text-white"
              aria-label="Close"
            >
              <X className="h-4.5 w-4.5" />
            </button>
          </div>
        </div>

        {/* -------------------------- Body -------------------------- */}
        <form onSubmit={handleSubmit} className="px-6 py-6">
          <div>
            <div className="mb-1.5 flex items-center justify-between">
              <label className="text-xs font-bold uppercase tracking-wider text-slate-500">
                Folder Name
              </label>
              <span
                className={`text-[10px] font-medium ${
                  isTooLong ? 'text-rose-500' : 'text-slate-400'
                }`}
              >
                {trimmedName.length}/60
              </span>
            </div>

            {/* Input */}
            <div
              className={`group flex items-center gap-2 rounded-xl border bg-slate-50/60 px-3.5 py-2.5 transition-all focus-within:bg-white focus-within:ring-4 ${
                isTooLong
                  ? 'border-rose-300 focus-within:border-rose-400 focus-within:ring-rose-100'
                  : 'border-slate-200 focus-within:border-blue-400 focus-within:ring-blue-100'
              }`}
            >
              <span
                className={`shrink-0 transition-colors ${
                  isTooLong
                    ? 'text-rose-400'
                    : 'text-slate-400 group-focus-within:text-blue-500'
                }`}
              >
                <FolderIcon className="h-4 w-4" />
              </span>
              <input
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                className="w-full bg-transparent text-sm font-medium text-slate-900 outline-none placeholder:text-slate-400"
                placeholder="Enter folder name"
                required
                autoFocus
                maxLength={80}
              />
            </div>

            {/* Feedback row */}
            <div className="mt-2 min-h-[16px] pl-1">
              {isTooLong ? (
                <p className="flex items-center gap-1.5 text-[11px] font-medium text-rose-500">
                  <AlertCircle className="h-3.5 w-3.5" />
                  Folder name is too long
                </p>
              ) : isUnchanged && trimmedName.length > 0 ? (
                <p className="flex items-center gap-1.5 text-[11px] font-medium text-slate-400">
                  <AlertCircle className="h-3.5 w-3.5" />
                  No changes yet
                </p>
              ) : isValid ? (
                <p className="flex items-center gap-1.5 text-[11px] font-medium text-emerald-600">
                  <CheckCircle2 className="h-3.5 w-3.5" />
                  Ready to save
                </p>
              ) : null}
            </div>
          </div>

          {/* -------------------------- Footer -------------------------- */}
          <div className="mt-5 flex items-center justify-end gap-2 border-t border-slate-100 pt-5">
            <button
              type="button"
              onClick={onClose}
              className="rounded-xl border border-slate-200 bg-white px-4 py-2.5 text-sm font-semibold text-slate-600 transition-colors hover:bg-slate-50 hover:text-slate-900"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={loading || !isValid || isUnchanged}
              className="group inline-flex items-center gap-2 rounded-xl bg-gradient-to-br from-blue-600 to-indigo-600 px-5 py-2.5 text-sm font-semibold text-white shadow-md shadow-blue-500/25 transition-all hover:-translate-y-0.5 hover:shadow-lg hover:shadow-blue-500/30 active:translate-y-0 disabled:cursor-not-allowed disabled:opacity-60 disabled:hover:translate-y-0"
            >
              {loading ? (
                <>
                  <span className="h-3.5 w-3.5 animate-spin rounded-full border-2 border-white/40 border-t-white" />
                  Saving…
                </>
              ) : (
                <>
                  <Save className="h-4 w-4 transition-transform group-hover:scale-110" />
                  Save Changes
                </>
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}