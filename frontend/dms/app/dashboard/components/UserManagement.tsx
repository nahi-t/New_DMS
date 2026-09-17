'use client';

import { User } from '@/type';
import { deleteUser } from '@/lib/api';
import {
  Pencil,
  Trash2,
  Mail,
  Shield,
  ShieldCheck,
  User as UserIcon,
  Crown,
  Eye,
  Briefcase,
  UserCog,
  MoreVertical,
  Sparkles,
} from 'lucide-react';
import toast from 'react-hot-toast';
import { useState } from 'react';

interface UserManagementProps {
  users: User[];
  onUserChange: () => void;
  onEditUser: (user: User) => void;
  currentUser?: User | null;
}

/* ------------------------------------------------------------------ */
/*  Role config — colors, icons, labels                                */
/* ------------------------------------------------------------------ */
const ROLE_CONFIG: Record<
  string,
  {
    label: string;
    icon: React.ReactNode;
    badge: string;
    avatar: string;
    ring: string;
  }
> = {
  admin: {
    label: 'Admin',
    icon: <Crown className="h-3 w-3" />,
    badge: 'bg-gradient-to-r from-amber-100 to-orange-100 text-amber-700 ring-1 ring-amber-200',
    avatar: 'from-amber-500 to-orange-600',
    ring: 'ring-amber-200',
  },
  manager: {
    label: 'Manager',
    icon: <ShieldCheck className="h-3 w-3" />,
    badge: 'bg-gradient-to-r from-blue-100 to-indigo-100 text-blue-700 ring-1 ring-blue-200',
    avatar: 'from-blue-500 to-indigo-600',
    ring: 'ring-blue-200',
  },
  editor: {
    label: 'Editor',
    icon: <Briefcase className="h-3 w-3" />,
    badge: 'bg-gradient-to-r from-emerald-100 to-teal-100 text-emerald-700 ring-1 ring-emerald-200',
    avatar: 'from-emerald-500 to-teal-600',
    ring: 'ring-emerald-200',
  },
  viewer: {
    label: 'Viewer',
    icon: <Eye className="h-3 w-3" />,
    badge: 'bg-slate-100 text-slate-600 ring-1 ring-slate-200',
    avatar: 'from-slate-500 to-slate-700',
    ring: 'ring-slate-200',
  },
  user: {
    label: 'Member',
    icon: <UserCog className="h-3 w-3" />,
    badge: 'bg-violet-100 text-violet-700 ring-1 ring-violet-200',
    avatar: 'from-violet-500 to-purple-600',
    ring: 'ring-violet-200',
  },
};

const getRole = (role?: string) =>
  ROLE_CONFIG[(role ?? 'viewer').toLowerCase()] ?? ROLE_CONFIG.viewer;

/* ------------------------------------------------------------------ */
/*  Component                                                          */
/* ------------------------------------------------------------------ */
export default function UserManagement({
  users,
  onUserChange,
  onEditUser,
  currentUser,
}: UserManagementProps) {
  const [openMenuId, setOpenMenuId] = useState<number | null>(null);

  const handleDelete = async (id: number) => {
    if (!confirm('Are you sure you want to delete this user?')) return;
    try {
      await deleteUser(id);
      toast.success('User deleted');
      onUserChange();
    } catch (error: any) {
      toast.error(error.message || 'Failed to delete user');
    }
  };

  const displayUsers =
    currentUser?.role === 'admin' ? users : currentUser ? [currentUser] : [];

  /* ---------------------------- Empty state ---------------------------- */
  if (displayUsers.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center rounded-2xl border border-dashed border-slate-200 bg-slate-50/60 px-6 py-14 text-center">
        <div className="mb-3 flex h-14 w-14 items-center justify-center rounded-2xl bg-white text-slate-300 shadow-sm ring-1 ring-slate-200">
          <UserIcon className="h-7 w-7" />
        </div>
        <p className="text-sm font-semibold text-slate-700">No users found</p>
        <p className="mt-1 max-w-xs text-xs text-slate-400">
          There are no team members to display right now.
        </p>
      </div>
    );
  }

  /* ---------------------------- User list ---------------------------- */
  return (
    <div className="space-y-3">
      {/* Header strip */}
      <div className="flex flex-wrap items-center justify-between gap-3 pb-1">
        <div className="flex items-center gap-2.5">
          <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-gradient-to-br from-blue-500 to-indigo-600 text-white shadow-md shadow-blue-500/20">
            <UserIcon className="h-4.5 w-4.5" />
          </div>
          <div>
            <h3 className="text-sm font-bold text-slate-900">
              {currentUser?.role === 'admin' ? 'Team Members' : 'My Profile'}
            </h3>
            <p className="text-[11px] text-slate-500">
              {displayUsers.length}{' '}
              {displayUsers.length === 1 ? 'account' : 'accounts'} in workspace
            </p>
          </div>
        </div>

        <span className="inline-flex items-center gap-1 rounded-full bg-blue-50 px-2.5 py-1 text-[10px] font-bold uppercase tracking-wider text-blue-600 ring-1 ring-blue-100">
          <Sparkles className="h-3 w-3" />
          Active
        </span>
      </div>

      {/* User cards */}
      <ul className="space-y-2.5">
        {displayUsers.map((user) => {
          const role = getRole(user.role);
          const isSelf = user.id === currentUser?.id;
          const canDelete = currentUser?.role === 'admin' && !isSelf;
          const initial = user.username?.charAt(0)?.toUpperCase() ?? '?';

          return (
            <li
              key={user.id}
              className="group relative overflow-hidden rounded-2xl border border-slate-200 bg-white p-4 shadow-sm transition-all hover:-translate-y-0.5 hover:border-blue-200 hover:shadow-lg hover:shadow-blue-500/5"
            >
              {/* Left accent bar on hover */}
              <span className="pointer-events-none absolute inset-y-0 left-0 w-1 bg-gradient-to-b from-blue-500 to-indigo-600 opacity-0 transition-opacity group-hover:opacity-100" />

              <div className="flex items-center gap-4">
                {/* Avatar */}
                <div className="relative shrink-0">
                  <div
                    className={`flex h-12 w-12 items-center justify-center rounded-2xl bg-gradient-to-br ${role.avatar} text-base font-bold uppercase text-white shadow-md ring-4 ${role.ring}`}
                  >
                    {initial}
                  </div>
                  {/* Online dot */}
                  <span className="absolute -bottom-0.5 -right-0.5 h-3.5 w-3.5 rounded-full border-2 border-white bg-emerald-500" />
                </div>

                {/* Info */}
                <div className="min-w-0 flex-1">
                  <div className="flex flex-wrap items-center gap-2">
                    <p className="truncate text-sm font-bold text-slate-900">
                      {user.username || 'Unnamed user'}
                    </p>
                    {isSelf && (
                      <span className="inline-flex items-center rounded-full bg-slate-900 px-2 py-0.5 text-[9px] font-bold uppercase tracking-wider text-white">
                        You
                      </span>
                    )}
                  </div>

                  <div className="mt-0.5 flex items-center gap-1.5 text-xs text-slate-500">
                    <Mail className="h-3.5 w-3.5 shrink-0 text-slate-400" />
                    <span className="truncate">{user.email || '—'}</span>
                  </div>

                  {/* Role badge */}
                  <div className="mt-2">
                    <span
                      className={`inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-[10px] font-bold uppercase tracking-wider ${role.badge}`}
                    >
                      {role.icon}
                      {role.label}
                    </span>
                  </div>
                </div>

                {/* Actions */}
                <div className="flex shrink-0 items-center gap-1">
                  <button
                    onClick={() => onEditUser(user)}
                    title="Edit user"
                    className="flex h-9 w-9 items-center justify-center rounded-xl text-slate-400 transition-all hover:bg-blue-50 hover:text-blue-600 active:scale-95"
                  >
                    <Pencil className="h-4 w-4" />
                  </button>

                  {canDelete && (
                    <button
                      onClick={() => handleDelete(user.id)}
                      title="Delete user"
                      className="flex h-9 w-9 items-center justify-center rounded-xl text-slate-400 transition-all hover:bg-rose-50 hover:text-rose-600 active:scale-95"
                    >
                      <Trash2 className="h-4 w-4" />
                    </button>
                  )}
                </div>
              </div>
            </li>
          );
        })}
      </ul>

      {/* Footer hint for non-admins */}
      {currentUser?.role !== 'admin' && (
        <div className="flex items-start gap-2 rounded-xl border border-slate-200 bg-slate-50/70 px-3.5 py-2.5 text-[11px] text-slate-500">
          <Shield className="mt-0.5 h-3.5 w-3.5 shrink-0 text-slate-400" />
          <p>
            You are viewing your own profile. Contact an administrator to manage
            team access.
          </p>
        </div>
      )}
    </div>
  );
}