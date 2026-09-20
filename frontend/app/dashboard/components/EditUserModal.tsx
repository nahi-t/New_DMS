'use client';

import { useState, useEffect } from 'react';
import { updateUser, updatePassword } from '@/lib/api';
import {
  X,
  User as UserIcon,
  Mail,
  Shield,
  Lock,
  Eye,
  EyeOff,
  Save,
  KeyRound,
  AlertCircle,
  CheckCircle2,
} from 'lucide-react';
import toast from 'react-hot-toast';
import { User } from '@/type';

interface Props {
  user: User;
  currentUser?: User | null;
  isOwnProfile?: boolean;
  onClose: () => void;
  onSuccess: () => void;
}

/* Role options with labels */
const ROLE_OPTIONS = [
  { value: 'user', label: 'User' },
  { value: 'manager', label: 'Manager' },
  { value: 'admin', label: 'Admin' },
];

export default function EditUserModal({
  user,
  currentUser,
  isOwnProfile = false,
  onClose,
  onSuccess,
}: Props) {
  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');
  const [role, setRole] = useState('');
  const [loading, setLoading] = useState(false);

  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');

  const [showNewPassword, setShowNewPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);

  const canEditRole = !isOwnProfile || currentUser?.role === 'admin';

  useEffect(() => {
    setUsername(user.username);
    setEmail(user.email);
    setRole(user.role);
    setNewPassword('');
    setConfirmPassword('');
  }, [user]);

  /* -------------------- password strength meter -------------------- */
  const passwordStrength = (() => {
    if (!newPassword) return null;
    let score = 0;
    if (newPassword.length >= 6) score++;
    if (newPassword.length >= 10) score++;
    if (/[A-Z]/.test(newPassword)) score++;
    if (/[0-9]/.test(newPassword)) score++;
    if (/[^A-Za-z0-9]/.test(newPassword)) score++;

    if (score <= 2) return { level: 'Weak', color: 'bg-rose-500', text: 'text-rose-600', w: 'w-1/3' };
    if (score <= 3) return { level: 'Fair', color: 'bg-amber-500', text: 'text-amber-600', w: 'w-2/3' };
    return { level: 'Strong', color: 'bg-emerald-500', text: 'text-emerald-600', w: 'w-full' };
  })();

  const passwordsMatch =
    newPassword.length > 0 && confirmPassword.length > 0 && newPassword === confirmPassword;

  /* ------------------------------ submit ------------------------------ */
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);

    try {
      if (newPassword) {
        if (newPassword.length < 6) {
          toast.error('New password must be at least 6 characters.');
          setLoading(false);
          return;
        }
        if (newPassword !== confirmPassword) {
          toast.error('Passwords do not match.');
          setLoading(false);
          return;
        }
        await updatePassword(user.id, newPassword);
        toast.success('Password updated successfully');
      }

      const updateData: { username: string; email: string; role?: string } = {
        username,
        email,
      };
      if (canEditRole) {
        updateData.role = role;
      }

      await updateUser(user.id, updateData);
      toast.success('User updated successfully');

      onSuccess();
      onClose();
    } catch (error: any) {
      toast.error(error.message || 'Failed to update user');
    } finally {
      setLoading(false);
    }
  };

  const initial = user.username?.charAt(0)?.toUpperCase() ?? 'U';

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/40 p-4 backdrop-blur-sm"
      onClick={onClose}
    >
      <div
        className="relative w-full max-w-lg overflow-hidden rounded-3xl bg-white shadow-2xl shadow-slate-900/20"
        onClick={(e) => e.stopPropagation()}
      >
        {/* -------------------------- Header -------------------------- */}
        <div className="relative overflow-hidden bg-gradient-to-br from-blue-600 via-indigo-600 to-blue-700 px-6 py-6">
          {/* decorative blobs */}
          <div className="pointer-events-none absolute -right-12 -top-16 h-48 w-48 rounded-full bg-white/10 blur-3xl" />
          <div className="pointer-events-none absolute -bottom-20 left-10 h-40 w-40 rounded-full bg-indigo-400/20 blur-3xl" />
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
              <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-white/15 text-lg font-bold uppercase text-white backdrop-blur-md ring-1 ring-white/20">
                {initial}
              </div>
              <div className="min-w-0">
                <h3 className="text-lg font-bold tracking-tight text-white">
                  {isOwnProfile ? 'Edit My Profile' : 'Edit User'}
                </h3>
                <p className="mt-0.5 truncate text-xs text-blue-100">
                  {isOwnProfile
                    ? 'Update your account information'
                    : `Manage ${user.username || 'user'}'s account`}
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
        <form onSubmit={handleSubmit} className="max-h-[70vh] overflow-y-auto px-6 py-6">
          <div className="space-y-5">
            {/* Section: Account */}
            <div className="space-y-4">
              <p className="text-[10px] font-bold uppercase tracking-[0.14em] text-slate-400">
                Account Information
              </p>

              {/* Username */}
              <Field
                label="Username"
                icon={<UserIcon className="h-4 w-4" />}
              >
                <input
                  type="text"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  className="w-full bg-transparent text-sm font-medium text-slate-900 outline-none placeholder:text-slate-400"
                  placeholder="Enter username"
                  required
                />
              </Field>

              {/* Email */}
              <Field label="Email" icon={<Mail className="h-4 w-4" />}>
                <input
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  className="w-full bg-transparent text-sm font-medium text-slate-900 outline-none placeholder:text-slate-400"
                  placeholder="name@example.com"
                  required
                />
              </Field>

              {/* Role */}
              {canEditRole && (
                <Field label="Role" icon={<Shield className="h-4 w-4" />}>
                  <select
                    value={role}
                    onChange={(e) => setRole(e.target.value)}
                    className="w-full cursor-pointer bg-transparent text-sm font-medium text-slate-900 outline-none"
                  >
                    {ROLE_OPTIONS.map((opt) => (
                      <option key={opt.value} value={opt.value}>
                        {opt.label}
                      </option>
                    ))}
                  </select>
                </Field>
              )}

              {!canEditRole && (
                <div className="flex items-start gap-2 rounded-xl border border-slate-200 bg-slate-50/70 px-3.5 py-2.5 text-[11px] text-slate-500">
                  <AlertCircle className="mt-0.5 h-3.5 w-3.5 shrink-0 text-slate-400" />
                  <p>Only administrators can change roles.</p>
                </div>
              )}
            </div>

            {/* Divider */}
            <div className="flex items-center gap-3">
              <div className="h-px flex-1 bg-slate-200" />
              <span className="text-[10px] font-bold uppercase tracking-[0.14em] text-slate-400">
                Change Password
              </span>
              <div className="h-px flex-1 bg-slate-200" />
            </div>

            {/* Section: Password */}
            <div className="space-y-4">
              {/* New password */}
              <Field
                label="New Password"
                icon={<Lock className="h-4 w-4" />}
                hint="Leave blank to keep current"
              >
                <input
                  type={showNewPassword ? 'text' : 'password'}
                  value={newPassword}
                  onChange={(e) => setNewPassword(e.target.value)}
                  className="w-full bg-transparent text-sm font-medium text-slate-900 outline-none placeholder:text-slate-400"
                  placeholder="••••••••"
                  minLength={6}
                />
                <button
                  type="button"
                  onClick={() => setShowNewPassword((v) => !v)}
                  className="ml-2 shrink-0 rounded-md p-1 text-slate-400 transition-colors hover:text-slate-600"
                  tabIndex={-1}
                >
                  {showNewPassword ? (
                    <EyeOff className="h-4 w-4" />
                  ) : (
                    <Eye className="h-4 w-4" />
                  )}
                </button>
              </Field>

              {/* Strength meter */}
              {passwordStrength && (
                <div className="flex items-center gap-2 pl-11">
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

              {/* Confirm password */}
              <Field
                label="Confirm New Password"
                icon={<KeyRound className="h-4 w-4" />}
              >
                <input
                  type={showConfirmPassword ? 'text' : 'password'}
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                  className="w-full bg-transparent text-sm font-medium text-slate-900 outline-none placeholder:text-slate-400"
                  placeholder="Re-enter new password"
                />
                <button
                  type="button"
                  onClick={() => setShowConfirmPassword((v) => !v)}
                  className="ml-2 shrink-0 rounded-md p-1 text-slate-400 transition-colors hover:text-slate-600"
                  tabIndex={-1}
                >
                  {showConfirmPassword ? (
                    <EyeOff className="h-4 w-4" />
                  ) : (
                    <Eye className="h-4 w-4" />
                  )}
                </button>
              </Field>

              {/* Match indicator */}
              {confirmPassword.length > 0 && (
                <div
                  className={`flex items-center gap-1.5 pl-11 text-[11px] font-medium ${
                    passwordsMatch ? 'text-emerald-600' : 'text-rose-500'
                  }`}
                >
                  {passwordsMatch ? (
                    <>
                      <CheckCircle2 className="h-3.5 w-3.5" />
                      Passwords match
                    </>
                  ) : (
                    <>
                      <AlertCircle className="h-3.5 w-3.5" />
                      Passwords do not match
                    </>
                  )}
                </div>
              )}
            </div>
          </div>

          {/* -------------------------- Footer -------------------------- */}
          <div className="mt-7 flex items-center justify-end gap-2 border-t border-slate-100 pt-5">
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

/* ------------------------------------------------------------------ */
/*  Field wrapper                                                      */
/* ------------------------------------------------------------------ */
function Field({
  label,
  icon,
  hint,
  children,
}: {
  label: string;
  icon: React.ReactNode;
  hint?: string;
  children: React.ReactNode;
}) {
  return (
    <div>
      <div className="mb-1.5 flex items-center justify-between">
        <label className="text-xs font-bold uppercase tracking-wider text-slate-500">
          {label}
        </label>
        {hint && (
          <span className="text-[10px] font-medium text-slate-400">{hint}</span>
        )}
      </div>
      <div className="group flex items-center gap-2 rounded-xl border border-slate-200 bg-slate-50/60 px-3.5 py-2.5 transition-all focus-within:border-blue-400 focus-within:bg-white focus-within:ring-4 focus-within:ring-blue-100">
        <span className="shrink-0 text-slate-400 transition-colors group-focus-within:text-blue-500">
          {icon}
        </span>
        <div className="flex flex-1 items-center">{children}</div>
      </div>
    </div>
  );
}