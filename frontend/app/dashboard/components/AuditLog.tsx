'use client';

import { useCallback, useEffect, useMemo, useState } from 'react';
import { getAuditLogs, AUDIT_EVENTS } from '@/lib/api';
import type { AuditLog as AuditLogType } from '@/type';
import toast from 'react-hot-toast';

/* ------------------------------------------------------------------ */
/*  Helpers                                                            */
/* ------------------------------------------------------------------ */

function formatTime(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  return d.toLocaleString(undefined, {
    year: 'numeric',
    month: 'short',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  });
}

function eventTint(event: string): string {
  if (event === 'user.login.failed')
    return 'bg-rose-50 text-rose-700 ring-rose-100';
  if (event === 'user.login')
    return 'bg-emerald-50 text-emerald-700 ring-emerald-100';
  if (event === 'user.logout')
    return 'bg-slate-100 text-slate-700 ring-slate-200';

  if (event.includes('delete')) return 'bg-rose-50 text-rose-700 ring-rose-100';
  if (event.includes('create')) return 'bg-emerald-50 text-emerald-700 ring-emerald-100';
  if (event.includes('update') || event.includes('restore'))
    return 'bg-amber-50 text-amber-700 ring-amber-100';
  if (event.includes('download') || event.includes('view'))
    return 'bg-blue-50 text-blue-700 ring-blue-100';
  if (event.includes('permission')) return 'bg-purple-50 text-purple-700 ring-purple-100';
  return 'bg-slate-100 text-slate-700 ring-slate-200';
}

const PAGE_SIZE = 25;

/* ------------------------------------------------------------------ */
/*  Component                                                          */
/* ------------------------------------------------------------------ */

export default function AuditLog() {
  const [logs, setLogs] = useState<AuditLogType[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);

  const [page, setPage] = useState(0);
  const [userFilter, setUserFilter] = useState('');
  const [eventFilter, setEventFilter] = useState('');
  const [fromFilter, setFromFilter] = useState('');
  const [toFilter, setToFilter] = useState('');

  /* ---------------- fetch ---------------- */
  const fetchLogs = useCallback(async () => {
    setLoading(true);
    try {
      const result = await getAuditLogs({
        limit: PAGE_SIZE,
        offset: page * PAGE_SIZE,
        user_id: userFilter.trim() || undefined,
        event: eventFilter || undefined,
        from: fromFilter ? new Date(fromFilter).toISOString() : undefined,
        to: toFilter ? new Date(toFilter).toISOString() : undefined,
      });
      setLogs(result.data ?? []);
      setTotal(result.total ?? 0);
    } catch (err: any) {
      if (err.message?.toLowerCase().includes('forbidden') || err.message?.includes('403')) {
        toast.error('You do not have permission to view audit logs.');
      } else {
        toast.error(err.message || 'Failed to load audit logs');
      }
      setLogs([]);
      setTotal(0);
    } finally {
      setLoading(false);
    }
  }, [page, userFilter, eventFilter, fromFilter, toFilter]);

  useEffect(() => {
    fetchLogs();
  }, [fetchLogs]);

  useEffect(() => {
    setPage(0);
  }, [userFilter, eventFilter, fromFilter, toFilter]);

  /* ---------------- derived ---------------- */
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));
  const showingFrom = total === 0 ? 0 : page * PAGE_SIZE + 1;
  const showingTo = Math.min((page + 1) * PAGE_SIZE, total);

  const eventOptions = useMemo(
    () =>
      Object.entries(AUDIT_EVENTS).map(([key, value]) => ({
        label: key
          .replace(/([a-z0-9])([A-Z])/g, '$1 $2')
          .replaceAll('_', ' ')
          .toLowerCase()
          .trim(),
        value,
      })),
    []
  );

  const hasFilters =
    userFilter.trim() !== '' ||
    eventFilter !== '' ||
    fromFilter !== '' ||
    toFilter !== '';

  const clearFilters = () => {
    setUserFilter('');
    setEventFilter('');
    setFromFilter('');
    setToFilter('');
  };

  /* ---------------- render ---------------- */
  return (
    <div className="space-y-5">
      {/* Filters */}
      <div className="rounded-2xl border border-slate-200 bg-slate-50/60 p-4">
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-5">
          <div>
            <label className="mb-1 block text-[11px] font-bold uppercase tracking-wider text-slate-500">
              User ID
            </label>
            <input
              type="text"
              inputMode="numeric"
              value={userFilter}
              onChange={(e) => setUserFilter(e.target.value)}
              placeholder="e.g. 42"
              className="w-full rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm text-slate-800 outline-none transition focus:border-blue-400 focus:ring-2 focus:ring-blue-100"
            />
          </div>

          <div>
            <label className="mb-1 block text-[11px] font-bold uppercase tracking-wider text-slate-500">
              Event
            </label>
            <select
              value={eventFilter}
              onChange={(e) => setEventFilter(e.target.value)}
              className="w-full rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm text-slate-800 outline-none transition focus:border-blue-400 focus:ring-2 focus:ring-blue-100"
            >
              <option value="">All events</option>
              {eventOptions.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </select>
          </div>

          <div>
            <label className="mb-1 block text-[11px] font-bold uppercase tracking-wider text-slate-500">
              From
            </label>
            <input
              type="datetime-local"
              value={fromFilter}
              onChange={(e) => setFromFilter(e.target.value)}
              className="w-full rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm text-slate-800 outline-none transition focus:border-blue-400 focus:ring-2 focus:ring-blue-100"
            />
          </div>

          <div>
            <label className="mb-1 block text-[11px] font-bold uppercase tracking-wider text-slate-500">
              To
            </label>
            <input
              type="datetime-local"
              value={toFilter}
              onChange={(e) => setToFilter(e.target.value)}
              className="w-full rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm text-slate-800 outline-none transition focus:border-blue-400 focus:ring-2 focus:ring-blue-100"
            />
          </div>

          <div className="flex items-end gap-2">
            <button
              type="button"
              onClick={fetchLogs}
              className="flex-1 rounded-lg bg-blue-600 px-3 py-2 text-sm font-semibold text-white shadow-sm transition hover:bg-blue-700 disabled:opacity-50"
              disabled={loading}
            >
              {loading ? 'Loading…' : 'Refresh'}
            </button>
            <button
              type="button"
              onClick={clearFilters}
              disabled={!hasFilters}
              className="rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm font-medium text-slate-600 transition hover:bg-slate-100 disabled:opacity-40"
            >
              Clear
            </button>
          </div>
        </div>
      </div>

      {/* Table */}
      <div className="overflow-hidden rounded-2xl border border-slate-200 bg-white">
        <div className="overflow-x-auto">
          <table className="w-full border-collapse text-left text-sm">
            <thead>
              <tr className="border-b border-slate-200 bg-slate-50/70">
                <th className="px-4 py-3 text-[11px] font-bold uppercase tracking-wider text-slate-500">
                  Time
                </th>
                <th className="px-4 py-3 text-[11px] font-bold uppercase tracking-wider text-slate-500">
                  User
                </th>
                <th className="px-4 py-3 text-[11px] font-bold uppercase tracking-wider text-slate-500">
                  IP Address
                </th>
                <th className="px-4 py-3 text-[11px] font-bold uppercase tracking-wider text-slate-500">
                  Event
                </th>
                <th className="px-4 py-3 text-[11px] font-bold uppercase tracking-wider text-slate-500">
                  Status
                </th>
              </tr>
            </thead>
            <tbody>
              {loading ? (
                <tr>
                  <td colSpan={5} className="px-4 py-12 text-center text-slate-400">
                    <div className="flex items-center justify-center gap-3">
                      <div className="h-5 w-5 animate-spin rounded-full border-2 border-slate-200 border-t-blue-600" />
                      Loading audit events…
                    </div>
                  </td>
                </tr>
              ) : logs.length === 0 ? (
                <tr>
                  <td colSpan={5} className="px-4 py-12 text-center text-slate-400">
                    No audit events found.
                  </td>
                </tr>
              ) : (
                logs.map((log) => {
                  const isFailedLogin = log.event === 'user.login.failed';
                  const isFailed = log.success === false;
                  return (
                    <tr
                      key={log.id}
                      className={`border-b border-slate-100 transition-colors last:border-b-0 ${
                        isFailed ? 'bg-rose-50/40 hover:bg-rose-50/70' : 'hover:bg-slate-50/60'
                      }`}
                    >
                      <td className="whitespace-nowrap px-4 py-3 text-slate-600">
                        {formatTime(log.event_happened_time)}
                      </td>

                      <td className="px-4 py-3">
                        <div className="flex items-center gap-2.5">
                          <span
                            className={`flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-[11px] font-bold uppercase text-white ${
                              isFailedLogin
                                ? 'bg-gradient-to-br from-rose-500 to-rose-700'
                                : 'bg-gradient-to-br from-blue-600 to-indigo-600'
                            }`}
                          >
                            {log.user_name?.charAt(0) || '?'}
                          </span>
                          <div className="min-w-0">
                            <p className="truncate font-medium text-slate-800">
                              {log.user_name || 'Unknown'}
                            </p>
                            <p className="text-[11px] text-slate-400">
                              ID: {log.user_id ?? '—'}
                            </p>
                          </div>
                        </div>
                      </td>

                      <td className="whitespace-nowrap px-4 py-3">
                        {log.ip_address ? (
                          <code className="rounded bg-slate-100 px-2 py-1 font-mono text-[11px] text-slate-700">
                            {log.ip_address}
                          </code>
                        ) : (
                          <span className="text-slate-300">—</span>
                        )}
                      </td>

                      <td className="px-4 py-3">
                        <span
                          className={`inline-flex items-center rounded-full px-2.5 py-1 text-[11px] font-semibold ring-1 ring-inset ${eventTint(
                            log.event
                          )}`}
                        >
                          {log.event}
                        </span>
                      </td>

                      <td className="px-4 py-3">
                        {isFailed ? (
                          <span className="inline-flex items-center gap-1.5 rounded-full bg-rose-100 px-2.5 py-1 text-[11px] font-bold text-rose-700">
                            <span className="h-1.5 w-1.5 rounded-full bg-rose-500" />
                            Failed
                          </span>
                        ) : (
                          <span className="inline-flex items-center gap-1.5 rounded-full bg-emerald-100 px-2.5 py-1 text-[11px] font-bold text-emerald-700">
                            <span className="h-1.5 w-1.5 rounded-full bg-emerald-500" />
                            Success
                          </span>
                        )}
                      </td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>

        {/* Pagination */}
        <div className="flex flex-wrap items-center justify-between gap-3 border-t border-slate-100 bg-slate-50/70 px-4 py-3">
          <p className="text-xs text-slate-500">
            {total === 0
              ? 'No results'
              : `Showing ${showingFrom}–${showingTo} of ${total}`}
          </p>
          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={() => setPage((p) => Math.max(0, p - 1))}
              disabled={page === 0 || loading}
              className="rounded-lg border border-slate-200 bg-white px-3 py-1.5 text-xs font-medium text-slate-700 transition hover:bg-slate-100 disabled:opacity-40"
            >
              Previous
            </button>
            <span className="text-xs font-medium text-slate-500">
              Page {page + 1} of {totalPages}
            </span>
            <button
              type="button"
              onClick={() => setPage((p) => Math.min(totalPages - 1, p + 1))}
              disabled={page >= totalPages - 1 || loading}
              className="rounded-lg border border-slate-200 bg-white px-3 py-1.5 text-xs font-medium text-slate-700 transition hover:bg-slate-100 disabled:opacity-40"
            >
              Next
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}