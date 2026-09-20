'use client';

import { useAuth } from '@/hooks/useAuth';
import {
  LogOut,
  FileText,
  Crown,
  ShieldCheck,
  User as UserIcon,
  Briefcase,
  Eye,
  UserCog,
  Bell,
  Search,
} from 'lucide-react';
import toast from 'react-hot-toast';

/* ------------------------------------------------------------------ */
/*  Role config                                                        */
/* ------------------------------------------------------------------ */
const ROLE_CONFIG: Record<
  string,
  { label: string; icon: React.ReactNode; badge: string; avatar: string }
> = {
  admin: {
    label: 'Admin',
    icon: <Crown className="h-3 w-3" />,
    badge: 'bg-gradient-to-r from-amber-100 to-orange-100 text-amber-700 ring-1 ring-amber-200',
    avatar: 'from-amber-500 to-orange-600',
  },
  manager: {
    label: 'Manager',
    icon: <ShieldCheck className="h-3 w-3" />,
    badge: 'bg-gradient-to-r from-blue-100 to-indigo-100 text-blue-700 ring-1 ring-blue-200',
    avatar: 'from-blue-500 to-indigo-600',
  },
  editor: {
    label: 'Editor',
    icon: <Briefcase className="h-3 w-3" />,
    badge: 'bg-gradient-to-r from-emerald-100 to-teal-100 text-emerald-700 ring-1 ring-emerald-200',
    avatar: 'from-emerald-500 to-teal-600',
  },
  viewer: {
    label: 'Viewer',
    icon: <Eye className="h-3 w-3" />,
    badge: 'bg-slate-100 text-slate-600 ring-1 ring-slate-200',
    avatar: 'from-slate-500 to-slate-700',
  },
  user: {
    label: 'Member',
    icon: <UserCog className="h-3 w-3" />,
    badge: 'bg-violet-100 text-violet-700 ring-1 ring-violet-200',
    avatar: 'from-violet-500 to-purple-600',
  },
};

const getRole = (role?: string) =>
  ROLE_CONFIG[(role ?? 'viewer').toLowerCase()] ?? ROLE_CONFIG.viewer;

/* ------------------------------------------------------------------ */
/*  Header                                                             */
/* ------------------------------------------------------------------ */
export default function Header() {
  const { user, logout } = useAuth();

  const role = getRole(user?.role);
  const initial = user?.username?.charAt(0)?.toUpperCase() ?? 'U';

  const handleLogout = async () => {
    try {
      if (logout) await logout();
      toast.success('Logged out successfully');
    } catch {
      toast.error('Logout failed');
    }
  };

  return (
    <header className="sticky top-0 z-30 border-b border-slate-200/80 bg-white/80 backdrop-blur-xl">
      <div className="mx-auto flex max-w-7xl items-center justify-between gap-4 px-4 py-3 sm:px-6 lg:px-8">
        {/* --------------------------- Brand --------------------------- */}
        <div className="flex items-center gap-3">
          <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-gradient-to-br from-blue-600 to-indigo-600 text-white shadow-lg shadow-blue-500/25">
            <FileText className="h-5 w-5" />
          </div>
          <div className="leading-tight">
            <span className="block bg-gradient-to-r from-slate-900 to-slate-600 bg-clip-text text-base font-bold text-transparent sm:text-lg">
              DocManager
            </span>
            <span className="hidden text-[10px] font-semibold uppercase tracking-[0.14em] text-slate-400 sm:block">
              Document System
            </span>
          </div>
        </div>

        {/* --------------------------- Right side --------------------------- */}
        <div className="flex items-center gap-2 sm:gap-3">
          {/* Optional search (hidden on tiny screens) */}
          <div className="relative hidden lg:block">
            <span className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-slate-400">
              <Search className="h-4 w-4" />
            </span>
            <input
              type="text"
              placeholder="Search…"
              className="w-56 rounded-xl border border-slate-200 bg-slate-50/70 py-2 pl-9 pr-3 text-sm text-slate-700 outline-none transition-all placeholder:text-slate-400 focus:border-blue-400 focus:bg-white focus:ring-4 focus:ring-blue-100"
            />
          </div>

          {/* Notifications */}
          <button
            className="relative flex h-10 w-10 items-center justify-center rounded-xl text-slate-500 transition-colors hover:bg-slate-100 hover:text-slate-900"
            aria-label="Notifications"
          >
            <Bell className="h-4.5 w-4.5" />
            <span className="absolute right-2 top-2 h-2 w-2 rounded-full bg-rose-500 ring-2 ring-white" />
          </button>

          {/* Divider */}
          <span className="hidden h-6 w-px bg-slate-200 sm:block" />

          {/* User chip */}
          <div className="flex items-center gap-3 rounded-xl border border-slate-200 bg-white px-2 py-1.5 shadow-sm transition-all hover:border-slate-300 hover:shadow-md sm:px-3">
            {/* Avatar */}
            <div className="relative shrink-0">
              <div
                className={`flex h-8 w-8 items-center justify-center rounded-full bg-gradient-to-br ${role.avatar} text-xs font-bold uppercase text-white shadow-sm`}
              >
                {initial}
              </div>
              <span className="absolute -bottom-0.5 -right-0.5 h-2.5 w-2.5 rounded-full border-2 border-white bg-emerald-500" />
            </div>

            {/* Name + role */}
            <div className="hidden min-w-0 sm:block">
              <p className="truncate text-[13px] font-semibold leading-tight text-slate-900">
                {user?.username || 'User'}
              </p>
              <span
                className={`mt-0.5 inline-flex items-center gap-1 rounded-full px-1.5 py-0.5 text-[9px] font-bold uppercase tracking-wider ${role.badge}`}
              >
                {role.icon}
                {role.label}
              </span>
            </div>
          </div>

          {/* Logout */}
          <button
            onClick={handleLogout}
            className="group flex items-center gap-1.5 rounded-xl border border-slate-200 bg-white px-2.5 py-2 text-sm font-medium text-slate-600 shadow-sm transition-all hover:border-rose-200 hover:bg-rose-50 hover:text-rose-600 sm:px-3"
            aria-label="Logout"
          >
            <LogOut className="h-4 w-4 transition-transform group-hover:translate-x-0.5" />
            <span className="hidden sm:inline">Logout</span>
          </button>
        </div>
      </div>
    </header>
  );
}