'use client';

import { useState } from 'react';
import { register } from '@/lib/api';
import {
  X,
  User as UserIcon,
  Mail,
  Lock,
  Eye,
  EyeOff,
  UserPlus,
  UserCog,
  Sparkles,
} from 'lucide-react';
import toast from 'react-hot-toast';

interface Props {
  onClose: () => void;
  onSuccess?: () => void;
}

export default function CreateUserModal({ onClose, onSuccess }: Props) {
  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [showPassword, setShowPassword] = useState(false);

  /* -------------------- password strength -------------------- */
  const passwordStrength = (() => {
    if (!password) return null;
    let score = 0;
    if (password.length >= 6) score++;
    if (password.length >= 10) score++;
    if (/[A-Z]/.test(password)) score++;
    if (/[0-9]/.test(password)) score++;
    if (/[^A-Za-z0-9]/.test(password)) score++;
    if (score <= 2) return { level: 'Weak', color: 'bg-rose-500', text: 'text-rose-600', w: 'w-1/3' };
    if (score <= 3) return { level: 'Fair', color: 'bg-amber-500', text: 'text-amber-600', w: 'w-2/3' };
    return { level: 'Strong', color: 'bg-emerald-500', text: 'text-emerald-600', w: 'w-full' };
  })();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      // Always register as a standard user
      await register(username, email, password, 'user');
      toast.success('User created successfully');
      onSuccess?.();
      onClose();
    } catch (error: any) {
      toast.error(error.message || 'Failed to create user');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 p-4 backdrop-blur-sm"
      onClick={onClose}
    >
      <div
        className="relative w-full max-w-md overflow-hidden rounded-3xl bg-white shadow-2xl shadow-slate-900/30"
        onClick={(e) => e.stopPropagation()}
      >
        {/* ------------------------ Gradient header ------------------------ */}
        <div className="relative overflow-hidden bg-gradient-to-br from-blue-600 via-indigo-600 to-blue-700 px-6 py-6">
          <div className="pointer-events-none absolute -right-12 -top-16 h-48 w-48 rounded-full bg-white/10 blur-3xl" />
          <div className="pointer-events-none absolute -bottom-20 left-8 h-40 w-40 rounded-full bg-indigo-400/20 blur-3xl" />
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
                <UserPlus className="h-5.5 w-5.5" />
              </div>
              <div className="min-w-0">
                <h3 className="text-lg font-bold tracking-tight text-white">
                  Create New User
                </h3>
                <p className="mt-0.5 truncate text-xs text-blue-100">
                  Add a new member to your workspace
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

        {/* ------------------------ Body ------------------------ */}
        <form onSubmit={handleSubmit} className="max-h-[70vh] overflow-y-auto px-6 py-6">
          <div className="space-y-5">
            <p className="text-[10px] font-bold uppercase tracking-[0.14em] text-slate-400">
              Account Information
            </p>

            {/* Username */}
            <Field label="Username" icon={<UserIcon className="h-4 w-4" />}>
              <input
                type="text"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                placeholder="e.g. john_doe"
                className="w-full bg-transparent text-sm font-medium text-slate-900 outline-none placeholder:text-slate-400"
                required
                autoFocus
              />
            </Field>

            {/* Email */}
            <Field label="Email address" icon={<Mail className="h-4 w-4" />}>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="john@example.com"
                className="w-full bg-transparent text-sm font-medium text-slate-900 outline-none placeholder:text-slate-400"
                required
                autoComplete="email"
              />
            </Field>

            {/* Password */}
            <div>
              <Field label="Password" icon={<Lock className="h-4 w-4" />}>
                <input
                  type={showPassword ? 'text' : 'password'}
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="••••••••"
                  className="w-full bg-transparent text-sm font-medium text-slate-900 outline-none placeholder:text-slate-400"
                  required
                  minLength={6}
                  autoComplete="new-password"
                />
                <button
                  type="button"
                  onClick={() => setShowPassword((v) => !v)}
                  className="ml-2 shrink-0 rounded-md p-1 text-slate-400 transition-colors hover:text-slate-600"
                  tabIndex={-1}
                  aria-label={showPassword ? 'Hide password' : 'Show password'}
                >
                  {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                </button>
              </Field>

              {/* Strength meter */}
              {passwordStrength && (
                <div className="mt-2 flex items-center gap-2 pl-11">
                  <div className="h-1.5 flex-1 overflow-hidden rounded-full bg-slate-100">
                    <div
                      className={`h-full rounded-full transition-all duration-300 ${passwordStrength.color} ${passwordStrength.w}`}
                    />
                  </div>
                  <span
                    className={`text-[10px] font-bold uppercase tracking-wider ${passwordStrength.text}`}
                  >
                    {passwordStrength.level}
                  </span>
                </div>
              )}
            </div>

            {/* Fixed role — User only */}
            <div>
              <label className="mb-2 block text-xs font-bold uppercase tracking-wider text-slate-500">
                Role
              </label>
              <div className="flex items-center gap-3 rounded-xl border border-blue-400 bg-blue-50 px-3.5 py-3 shadow-sm shadow-blue-500/10">
                <span className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-gradient-to-br from-blue-500 to-indigo-600 text-white shadow-sm">
                  <UserCog className="h-4 w-4" />
                </span>
                <div className="min-w-0 flex-1">
                  <p className="text-[13px] font-bold text-blue-700">User</p>
                  <p className="text-[10px] text-blue-600/70">
                    Standard workspace access
                  </p>
                </div>
                <span className="rounded-full bg-blue-100 px-2 py-0.5 text-[9px] font-bold uppercase tracking-wider text-blue-700">
                  Default
                </span>
              </div>
            </div>

            {/* Info hint */}
            <div className="flex items-start gap-2 rounded-xl border border-slate-200 bg-slate-50/70 px-3.5 py-2.5 text-[11px] text-slate-500">
              <Sparkles className="mt-0.5 h-3.5 w-3.5 shrink-0 text-amber-500" />
              <p>
                New accounts are created with standard User access. Only an
                administrator can promote them later.
              </p>
            </div>
          </div>

          {/* ------------------------ Footer ------------------------ */}
          <div className="mt-6 flex items-center justify-end gap-2 border-t border-slate-100 pt-5">
            <button
              type="button"
              onClick={onClose}
              className="rounded-xl border border-slate-200 bg-white px-4 py-2.5 text-sm font-semibold text-slate-600 transition-colors hover:bg-slate-50 hover:text-slate-900"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={loading}
              className="group inline-flex items-center gap-2 rounded-xl bg-gradient-to-br from-blue-600 to-indigo-600 px-5 py-2.5 text-sm font-semibold text-white shadow-md shadow-blue-500/25 transition-all hover:-translate-y-0.5 hover:shadow-lg hover:shadow-blue-500/30 active:translate-y-0 disabled:cursor-not-allowed disabled:opacity-60 disabled:hover:translate-y-0"
            >
              {loading ? (
                <>
                  <span className="h-3.5 w-3.5 animate-spin rounded-full border-2 border-white/40 border-t-white" />
                  Creating…
                </>
              ) : (
                <>
                  <UserPlus className="h-4 w-4 transition-transform group-hover:scale-110" />
                  Create User
                </>
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

/* ------------------------------------------------------------------ */
/*  Field wrapper — matches the login page style                       */
/* ------------------------------------------------------------------ */
function Field({
  label,
  icon,
  children,
}: {
  label: string;
  icon: React.ReactNode;
  children: React.ReactNode;
}) {
  return (
    <div>
      <label className="mb-1.5 block text-xs font-bold uppercase tracking-wider text-slate-500">
        {label}
      </label>
      <div className="group flex items-center gap-2 rounded-xl border border-slate-200 bg-slate-50/60 px-3.5 py-3 transition-all focus-within:border-blue-400 focus-within:bg-white focus-within:ring-4 focus-within:ring-blue-100">
        <span className="shrink-0 text-slate-400 transition-colors group-focus-within:text-blue-500">
          {icon}
        </span>
        <div className="flex flex-1 items-center">{children}</div>
      </div>
    </div>
  );
}