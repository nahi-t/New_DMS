'use client';

import { useCallback, useEffect, useState } from 'react';
import { getSharedWithMe, downloadDocument } from '@/lib/api';
import type { DocumentShare } from '@/type';
import toast from 'react-hot-toast';

function formatDate(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  return d.toLocaleDateString(undefined, {
    year: 'numeric',
    month: 'short',
    day: '2-digit',
  });
}

const SharedIcons = {
  doc: (
    <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
    </svg>
  ),
  download: (
    <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M4 16v2a2 2 0 002 2h12a2 2 0 002-2v-2M7 10l5 5m0 0l5-5m-5 5V4" />
    </svg>
  ),
  inbox: (
    <svg className="w-12 h-12" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
    </svg>
  ),
};

export default function SharedWithMe() {
  const [shares, setShares] = useState<DocumentShare[]>([]);
  const [loading, setLoading] = useState(true);
  const [downloading, setDownloading] = useState<number | null>(null);
  const [filter, setFilter] = useState<'all' | 'viewer' | 'editor'>('all');

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const data = await getSharedWithMe();
      setShares(data ?? []);
    } catch (e: any) {
      toast.error(e.message || 'Failed to load shared documents');
      setShares([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  const handleDownload = async (share: DocumentShare) => {
    setDownloading(share.document_id);
    try {
      await downloadDocument(
        share.document_id,
        share.document_name || `document_${share.document_id}`
      );
      toast.success('Download started');
    } catch (e: any) {
      toast.error(e.message || 'Download failed');
    } finally {
      setDownloading(null);
    }
  };

  const filtered =
    filter === 'all' ? shares : shares.filter((s) => s.permission === filter);

  const counts = {
    all: shares.length,
    viewer: shares.filter((s) => s.permission === 'viewer').length,
    editor: shares.filter((s) => s.permission === 'editor').length,
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center py-16">
        <div className="flex items-center gap-3 text-slate-400">
          <div className="h-5 w-5 animate-spin rounded-full border-2 border-slate-200 border-t-blue-600" />
          <span className="text-sm">Loading shared documents…</span>
        </div>
      </div>
    );
  }

  if (shares.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center gap-3 py-16 text-center">
        <span className="text-slate-300">{SharedIcons.inbox}</span>
        <p className="text-base font-semibold text-slate-700">
          Nothing shared with you yet
        </p>
        <p className="max-w-sm text-sm text-slate-400">
          When someone shares a document with you, it will appear here with the
          permission level they granted.
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-5">
      {/* Filter chips */}
      <div className="flex flex-wrap items-center gap-2">
        {(['all', 'viewer', 'editor'] as const).map((key) => {
          const active = filter === key;
          return (
            <button
              key={key}
              onClick={() => setFilter(key)}
              className={`inline-flex items-center gap-2 rounded-full px-3.5 py-1.5 text-xs font-semibold transition ${
                active
                  ? 'bg-blue-600 text-white shadow-sm shadow-blue-500/20'
                  : 'border border-slate-200 bg-white text-slate-600 hover:bg-slate-50'
              }`}
            >
              <span className="capitalize">{key === 'all' ? 'All' : key}</span>
              <span
                className={`rounded-full px-1.5 py-0.5 text-[10px] font-bold ${
                  active ? 'bg-white/20 text-white' : 'bg-slate-100 text-slate-500'
                }`}
              >
                {counts[key]}
              </span>
            </button>
          );
        })}

        <button
          onClick={load}
          className="ml-auto rounded-lg border border-slate-200 bg-white px-3 py-1.5 text-xs font-medium text-slate-600 transition hover:bg-slate-50"
        >
          Refresh
        </button>
      </div>

      {/* Cards */}
      <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
        {filtered.map((s) => (
          <div
            key={s.id}
            className="group flex flex-col rounded-2xl border border-slate-200 bg-white p-4 shadow-sm transition-all hover:-translate-y-0.5 hover:border-slate-300 hover:shadow-lg hover:shadow-slate-900/5"
          >
            <div className="flex items-start gap-3">
              <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-blue-50 to-indigo-50 text-blue-600 ring-1 ring-blue-100">
                {SharedIcons.doc}
              </span>
              <div className="min-w-0 flex-1">
                <p className="truncate text-sm font-semibold text-slate-900">
                  {s.document_name || `Document #${s.document_id}`}
                </p>
                <p className="mt-0.5 text-[11px] text-slate-400">
                  v{s.version} · Shared {formatDate(s.shared_at)}
                </p>
              </div>
              <span
                className={`shrink-0 rounded-full px-2 py-0.5 text-[10px] font-bold uppercase tracking-wide ${
                  s.permission === 'editor'
                    ? 'bg-amber-100 text-amber-700 ring-1 ring-amber-200'
                    : 'bg-blue-100 text-blue-700 ring-1 ring-blue-200'
                }`}
              >
                {s.permission}
              </span>
            </div>

            {s.document_description && (
              <p className="mt-3 line-clamp-2 text-xs text-slate-500">
                {s.document_description}
              </p>
            )}

            <div className="mt-4 flex items-center gap-2 border-t border-slate-100 pt-3">
              <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-gradient-to-br from-slate-600 to-slate-800 text-[9px] font-bold uppercase text-white">
                {s.shared_by_name?.charAt(0) || '?'}
              </span>
              <p className="truncate text-[11px] text-slate-500">
                Shared by{' '}
                <span className="font-semibold text-slate-700">
                  {s.shared_by_name || `User #${s.shared_by}`}
                </span>
              </p>
            </div>

            <div className="mt-3 flex items-center gap-2">
              <button
                onClick={() => handleDownload(s)}
                disabled={downloading === s.document_id}
                className="inline-flex flex-1 items-center justify-center gap-2 rounded-lg bg-blue-600 px-3 py-2 text-xs font-semibold text-white shadow-sm transition hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-50"
              >
                {downloading === s.document_id ? (
                  <>
                    <span className="h-3.5 w-3.5 animate-spin rounded-full border-2 border-white/40 border-t-white" />
                    Downloading…
                  </>
                ) : (
                  <>
                    {SharedIcons.download}
                    Download
                  </>
                )}
              </button>
            </div>
          </div>
        ))}
      </div>

      {filtered.length === 0 && (
        <p className="py-10 text-center text-sm text-slate-400">
          No documents match this filter.
        </p>
      )}
    </div>
  );
}