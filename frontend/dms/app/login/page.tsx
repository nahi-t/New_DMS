'use client';

import { useState } from 'react';
import { useAuth } from '@/hooks/useAuth';
import {
  Mail,
  Lock,
  Loader2,
  Eye,
  EyeOff,
  FileText,
  FolderOpen,
  ShieldCheck,
  Users,
  ArrowRight,
  Sparkles,
  CheckCircle2,
} from 'lucide-react';
import CreateUserModal from '../dashboard/components/CreateUserModal';

export default function LoginPage() {
  const { login } = useAuth();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [showPassword, setShowPassword] = useState(false);
  const [remember, setRemember] = useState(true);
  const [showModal, setShowModal] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      await login(email, password);
    } catch (error) {
      // Error already handled in context with toast
    } finally {
      setLoading(false);
    }
  };

  const fillDemo = (role: 'admin' | 'manager' | 'user') => {
    setEmail(`${role}@example.com`);
    setPassword('password');
  };

  return (
    <div className="relative min-h-screen w-full overflow-hidden bg-slate-50">
      {/* ---------------- Background image + overlay ---------------- */}
      <div
        className="absolute inset-0 bg-cover bg-center"
        style={{
          backgroundImage:
            "url('https://images.unsplash.com/photo-1450101499163-c8848c66ca85?auto=format&fit=crop&w=2000&q=80')",
        }}
      />
      <div className="absolute inset-0 bg-gradient-to-br from-slate-900/90 via-blue-900/85 to-indigo-900/90" />

      {/* Decorative grid */}
      <div
        className="pointer-events-none absolute inset-0 opacity-[0.07]"
        style={{
          backgroundImage:
            'linear-gradient(to right, #fff 1px, transparent 1px), linear-gradient(to bottom, #fff 1px, transparent 1px)',
          backgroundSize: '40px 40px',
        }}
      />

      {/* ---------------- Content ---------------- */}
      <div className="relative z-10 flex min-h-screen items-center justify-center p-4 lg:p-8">
        <div className="grid w-full max-w-6xl overflow-hidden rounded-3xl bg-white shadow-2xl shadow-slate-900/40 lg:grid-cols-2">
          {/* ============= LEFT PANEL — Branding ============= */}
          <div className="relative hidden flex-col justify-between overflow-hidden bg-gradient-to-br from-blue-700 via-indigo-700 to-blue-800 p-10 lg:flex">
            {/* Glow blobs */}
            <div className="pointer-events-none absolute -right-20 -top-20 h-72 w-72 rounded-full bg-white/10 blur-3xl" />
            <div className="pointer-events-none absolute -bottom-24 -left-16 h-72 w-72 rounded-full bg-indigo-400/20 blur-3xl" />

            {/* Brand */}
            <div className="relative flex items-center gap-3">
              <div className="flex h-11 w-11 items-center justify-center rounded-2xl bg-white/15 text-white backdrop-blur-md ring-1 ring-white/20">
                <FileText className="h-5.5 w-5.5" />
              </div>
              <div className="leading-tight">
                <span className="block text-lg font-bold text-white">DocManager</span>
                <span className="text-[10px] font-semibold uppercase tracking-[0.16em] text-blue-200">
                  Document System
                </span>
              </div>
            </div>

            {/* Center content */}
            <div className="relative space-y-6">
              <div className="inline-flex items-center gap-2 rounded-full border border-white/15 bg-white/10 px-3.5 py-1.5 text-[11px] font-medium text-blue-50 backdrop-blur-md">
                <Sparkles className="h-3.5 w-3.5 text-blue-200" />
                Trusted by teams of every size
              </div>

              <h1 className="text-3xl font-bold leading-tight text-white xl:text-4xl">
                Organize, manage, and share your documents securely.
              </h1>
              <p className="max-w-md text-sm leading-relaxed text-blue-100/90">
                A modern document management workspace built for collaboration,
                access control, and effortless file organization.
              </p>

              {/* Feature bullets */}
              <ul className="space-y-3 pt-2">
                {[
                  { icon: <FolderOpen className="h-4 w-4" />, text: 'Unlimited folder organization' },
                  { icon: <ShieldCheck className="h-4 w-4" />, text: 'Role-based access control' },
                  { icon: <Users className="h-4 w-4" />, text: 'Real-time team collaboration' },
                ].map((f, i) => (
                  <li key={i} className="flex items-center gap-3 text-sm text-white/90">
                    <span className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-white/10 text-blue-100 ring-1 ring-white/15">
                      {f.icon}
                    </span>
                    {f.text}
                  </li>
                ))}
              </ul>
            </div>

            {/* Footer */}
            {/* <div className="relative flex items-center gap-3 text-xs text-blue-100/80">
              <div className="flex -space-x-2">
                {['A', 'M', 'U'].map((l, i) => (
                  <div
                    key={i}
                    className="flex h-7 w-7 items-center justify-center rounded-full border-2 border-blue-700 bg-gradient-to-br from-blue-400 to-indigo-500 text-[10px] font-bold text-white"
                  >
                    {l}
                  </div>
                ))}
              </div>
              <p>
                Join <span className="font-semibold text-white"></span> teams
                already using DocManager
              </p>
            </div> */}
          </div>

          {/* ============= RIGHT PANEL — Form ============= */}
          <div className="flex flex-col justify-center bg-white px-6 py-10 sm:px-10 lg:px-12 lg:py-14">
            {/* Mobile brand */}
            <div className="mb-8 flex items-center gap-3 lg:hidden">
              <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-gradient-to-br from-blue-600 to-indigo-600 text-white shadow-lg shadow-blue-500/25">
                <FileText className="h-5 w-5" />
              </div>
              <div className="leading-tight">
                <span className="block text-base font-bold text-slate-900">DocManager</span>
                <span className="text-[10px] font-semibold uppercase tracking-[0.14em] text-slate-400">
                  Document System
                </span>
              </div>
            </div>

            {/* Heading */}
            <div className="mb-8">
              <h2 className="text-2xl font-bold tracking-tight text-slate-900 sm:text-3xl">
                Welcome back 👋
              </h2>
              <p className="mt-2 text-sm text-slate-500">
                Sign in to continue to your workspace
              </p>
            </div>

            {/* Form */}
            <form onSubmit={handleSubmit} className="space-y-5">
              {/* Email */}
              <div>
                <label className="mb-1.5 block text-xs font-bold uppercase tracking-wider text-slate-500">
                  Email address
                </label>
                <div className="group flex items-center gap-2 rounded-xl border border-slate-200 bg-slate-50/60 px-3.5 py-3 transition-all focus-within:border-blue-400 focus-within:bg-white focus-within:ring-4 focus-within:ring-blue-100">
                  <Mail className="h-4 w-4 shrink-0 text-slate-400 transition-colors group-focus-within:text-blue-500" />
                  <input
                    type="email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    className="w-full bg-transparent text-sm font-medium text-slate-900 outline-none placeholder:text-slate-400"
                    placeholder="you@example.com"
                    required
                    autoComplete="email"
                  />
                </div>
              </div>

              {/* Password */}
              <div>
                <div className="mb-1.5 flex items-center justify-between">
                  <label className="text-xs font-bold uppercase tracking-wider text-slate-500">
                    Password
                  </label>
                  <button
                    type="button"
                    className="text-[11px] font-semibold text-blue-600 transition-colors hover:text-blue-800"
                    onClick={() => alert('Contact your administrator to reset.')}
                  >
                    Forgot password?
                  </button>
                </div>
                <div className="group flex items-center gap-2 rounded-xl border border-slate-200 bg-slate-50/60 px-3.5 py-3 transition-all focus-within:border-blue-400 focus-within:bg-white focus-within:ring-4 focus-within:ring-blue-100">
                  <Lock className="h-4 w-4 shrink-0 text-slate-400 transition-colors group-focus-within:text-blue-500" />
                  <input
                    type={showPassword ? 'text' : 'password'}
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    className="w-full bg-transparent text-sm font-medium text-slate-900 outline-none placeholder:text-slate-400"
                    placeholder="••••••••"
                    required
                    autoComplete="current-password"
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword((v) => !v)}
                    className="shrink-0 rounded-md p-1 text-slate-400 transition-colors hover:text-slate-600"
                    tabIndex={-1}
                    aria-label={showPassword ? 'Hide password' : 'Show password'}
                  >
                    {showPassword ? (
                      <EyeOff className="h-4 w-4" />
                    ) : (
                      <Eye className="h-4 w-4" />
                    )}
                  </button>
                </div>
              </div>

              {/* Remember me */}
              <label className="flex cursor-pointer items-center gap-2.5 text-sm text-slate-600 select-none">
                <span className="relative flex items-center">
                  <input
                    type="checkbox"
                    checked={remember}
                    onChange={(e) => setRemember(e.target.checked)}
                    className="peer h-4 w-4 cursor-pointer appearance-none rounded-md border border-slate-300 bg-white transition-all checked:border-blue-600 checked:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-200"
                  />
                  <CheckCircle2 className="pointer-events-none absolute left-0 h-4 w-4 scale-0 text-white transition-transform peer-checked:scale-100" />
                </span>
                Keep me signed in
              </label>

              {/* Submit */}
              <button
                type="submit"
                disabled={loading}
                className="group relative flex w-full items-center justify-center gap-2 overflow-hidden rounded-xl bg-gradient-to-br from-blue-600 to-indigo-600 px-4 py-3 text-sm font-semibold text-white shadow-lg shadow-blue-500/25 transition-all hover:-translate-y-0.5 hover:shadow-xl hover:shadow-blue-500/30 active:translate-y-0 disabled:cursor-not-allowed disabled:opacity-60 disabled:hover:translate-y-0"
              >
                {loading ? (
                  <>
                    <Loader2 className="h-4 w-4 animate-spin" />
                    Signing in…
                  </>
                ) : (
                  <>
                    Sign In
                    <ArrowRight className="h-4 w-4 transition-transform group-hover:translate-x-0.5" />
                  </>
                )}
              </button>
            </form>

            {/* Sign up link */}
            <p className="mt-6 text-center text-sm text-slate-500">
              Don't have an account?{' '}
              <button
                onClick={() => setShowModal(true)}
                className="font-semibold text-blue-600 transition-colors hover:text-blue-800 hover:underline"
              >
                Create one
              </button>
            </p>

            {/* Demo credentials */}
            <div className="mt-7 rounded-2xl border border-dashed border-slate-200 bg-slate-50/70 p-4">
              <div className="mb-2.5 flex items-center gap-2">
                <Sparkles className="h-3.5 w-3.5 text-amber-500" />
                <p className="text-[11px] font-bold uppercase tracking-wider text-slate-500">
                  Demo accounts
                </p>
              </div>
              <div className="flex flex-wrap gap-2">
                {(['admin', 'manager', 'user'] as const).map((r) => (
                  <button
                    key={r}
                    type="button"
                    onClick={() => fillDemo(r)}
                    className="inline-flex items-center gap-1.5 rounded-lg border border-slate-200 bg-white px-2.5 py-1.5 text-[11px] font-semibold capitalize text-slate-600 transition-all hover:-translate-y-0.5 hover:border-blue-300 hover:bg-blue-50 hover:text-blue-700"
                  >
                    {r}
                  </button>
                ))}
              </div>
              <p className="mt-2.5 text-[11px] text-slate-400">
                Click a role to auto-fill · password: <span className="font-mono">password</span>
              </p>
            </div>
          </div>
        </div>
      </div>

      {/* Create account modal */}
      {showModal && <CreateUserModal onClose={() => setShowModal(false)} />}
    </div>
  );
}