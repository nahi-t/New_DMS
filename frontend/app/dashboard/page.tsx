'use client';

import { useAuth } from '@/hooks/useAuth';
import UserManagement from './components/UserManagement';
import FolderManagement from './components/FolderManagement';
import EditUserModal from './components/EditUserModal';
import AuditLog from './components/AuditLog';
import { useEffect, useState } from 'react';
import type { ReactNode } from 'react';
import { getFolders, getUsers } from '@/lib/api';
import { Folder, User } from '@/type';
import toast from 'react-hot-toast';

type TabType = 'dashboard' | 'users' | 'folders' | 'audit';

/* ------------------------------------------------------------------ */
/*  Icons                                                              */
/* ------------------------------------------------------------------ */
const Icons = {
  logo: (
    <svg className="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
    </svg>
  ),
  grid: (
    <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zm10 0a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zm10 0a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
    </svg>
  ),
  users: (
    <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
    </svg>
  ),
  folder: (
    <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
    </svg>
  ),
  shield: (
    <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
    </svg>
  ),
  logout: (
    <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
    </svg>
  ),
  menu: (
    <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M4 6h16M4 12h16M4 18h16" />
    </svg>
  ),
  chevron: (
    <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M9 5l7 7-7 7" />
    </svg>
  ),
  arrowRight: (
    <svg className="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M5 12h14m0 0l-6-6m6 6l-6 6" />
    </svg>
  ),
  settings: (
    <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
      <path strokeLinecap="round" strokeLinejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
    </svg>
  ),
  sparkle: (
    <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M5 3v4M3 5h4M6 17v4m-2-2h4m5-16l2.286 6.857L21 12l-5.714 2.143L13 21l-2.286-6.857L5 12l5.714-2.143L13 3z" />
    </svg>
  ),
  fileDoc: (
    <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
    </svg>
  ),
  crown: (
    <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M3 7l4.5 4L12 5l4.5 6L21 7l-1.5 12h-15L3 7z" />
    </svg>
  ),
  trendUp: (
    <svg className="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
    </svg>
  ),
};

/* ------------------------------------------------------------------ */
/*  Page                                                               */
/* ------------------------------------------------------------------ */
export default function DashboardPage() {
  const { user, logout } = useAuth();

  const [folders, setFolders] = useState<Folder[]>([]);
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);

  const [activeTab, setActiveTab] = useState<TabType>('dashboard');
  const [isSidebarOpen, setIsSidebarOpen] = useState(false);

  const [selectedUser, setSelectedUser] = useState<User | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);

  const isAdmin = user?.role === 'admin';
  const canCreate = user?.role === 'admin' || user?.role === 'manager';

  const roleLabel = user?.role
    ? user.role.charAt(0).toUpperCase() + user.role.slice(1)
    : 'Guest';

  /* ----------------------------- data ----------------------------- */
  const fetchData = async () => {
    try {
      const [foldersData, usersData] = await Promise.all([
        getFolders(),
        user?.role === 'admin' ? getUsers() : Promise.resolve([]),
      ]);
      setFolders(foldersData);
      if (user?.role === 'admin') setUsers(usersData);
    } catch (error: any) {
      toast.error(error.message || 'Failed to load data');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, [user]);

  // Safety: a non-admin must never see the audit tab, even if
  // they try to set local state manually.
  useEffect(() => {
    if (!isAdmin && activeTab === 'audit') {
      setActiveTab('dashboard');
    }
  }, [isAdmin, activeTab]);

  /* --------------------------- derived stats ---------------------- */
  const stats = {
    totalFolders: folders.length,
    totalUsers: users.length,
    admins: users.filter((u) => u.role === 'admin').length,
    managers: users.filter((u) => u.role === 'manager').length,
    members: users.filter((u) => u.role !== 'admin' && u.role !== 'manager').length,
    recentFolders: [...folders]
      .sort((a, b) => {
        const da = new Date((a as any).created_at ?? 0).getTime();
        const db = new Date((b as any).created_at ?? 0).getTime();
        return db - da;
      })
      .slice(0, 5),
  };

  /* --------------------------- handlers --------------------------- */
  const handleEditUser = (userToEdit: User) => {
    setSelectedUser(userToEdit);
    setIsModalOpen(true);
  };

  const handleCloseModal = () => {
    setIsModalOpen(false);
    setSelectedUser(null);
  };

  const handleSuccess = () => {
    fetchData();
    handleCloseModal();
  };

  const handleLogout = async () => {
    try {
      if (logout) await logout();
      toast.success('Logged out successfully');
    } catch {
      toast.error('Logout failed');
    }
  };

  const goToTab = (tab: TabType) => {
    setActiveTab(tab);
    setIsSidebarOpen(false);
  };

  /* ---------------------------- loading --------------------------- */
  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-slate-50">
        <div className="flex flex-col items-center gap-4">
          <div className="relative">
            <div className="h-14 w-14 rounded-full border-4 border-slate-200" />
            <div className="absolute inset-0 h-14 w-14 animate-spin rounded-full border-4 border-transparent border-t-blue-600 border-r-blue-600" />
          </div>
          <p className="animate-pulse text-sm font-medium text-slate-500">
            Loading DMS Workspace…
          </p>
        </div>
      </div>
    );
  }

  const navItems: { id: TabType; label: string; icon: ReactNode }[] = [
    { id: 'dashboard', label: 'Dashboard', icon: Icons.grid },
    { id: 'users', label: isAdmin ? 'User Directory' : 'My Profile', icon: Icons.users },
    { id: 'folders', label: 'Folders & Documents', icon: Icons.folder },
    ...(isAdmin
      ? [{ id: 'audit' as const, label: 'Audit Log', icon: Icons.shield }]
      : []),
  ];

  return (
    <div className="flex h-screen overflow-hidden bg-slate-50 font-sans antialiased">
      {/* Mobile overlay */}
      {isSidebarOpen && (
        <div
          className="fixed inset-0 z-40 bg-slate-900/50 backdrop-blur-sm lg:hidden"
          onClick={() => setIsSidebarOpen(false)}
        />
      )}

      {/* ------------------------------ Sidebar ------------------------------ */}
      <aside
        className={`fixed inset-y-0 left-0 z-50 flex w-72 flex-col border-r border-slate-200 bg-white shadow-xl shadow-slate-900/5 transition-transform duration-300 ease-out lg:static lg:translate-x-0 lg:shadow-none ${
          isSidebarOpen ? 'translate-x-0' : '-translate-x-full'
        }`}
      >
        {/* Brand */}
        <div className="flex h-20 shrink-0 items-center gap-3 border-b border-slate-100 px-6">
          <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-gradient-to-tr from-blue-700 to-blue-500 shadow-lg shadow-blue-500/25">
            {Icons.logo}
          </div>
          <div className="leading-tight">
            <span className="block bg-gradient-to-r from-slate-900 to-slate-600 bg-clip-text text-lg font-bold text-transparent">
              DMS Portal
            </span>
            <span className="text-[10px] font-semibold uppercase tracking-[0.14em] text-slate-400">
              Document System
            </span>
          </div>
        </div>

        {/* Nav */}
        <nav className="flex-1 space-y-1 overflow-y-auto px-3 py-6">
          <p className="px-3 pb-2 text-[11px] font-bold uppercase tracking-[0.14em] text-slate-400">
            Navigation
          </p>

          {navItems.map((item) => {
            const active = activeTab === item.id;
            return (
              <button
                key={item.id}
                onClick={() => goToTab(item.id)}
                className={`group relative flex w-full items-center gap-3 rounded-xl px-3.5 py-3 text-sm font-medium transition-all duration-200 ${
                  active
                    ? 'bg-blue-50 text-blue-700 shadow-sm shadow-blue-500/5'
                    : 'text-slate-600 hover:bg-slate-50 hover:text-slate-900'
                }`}
              >
                <span
                  className={`absolute left-0 top-1/2 h-6 w-1 -translate-y-1/2 rounded-r-full bg-blue-600 transition-all duration-200 ${
                    active ? 'opacity-100' : 'opacity-0'
                  }`}
                />
                <span className={active ? 'text-blue-600' : 'text-slate-400 group-hover:text-slate-600'}>
                  {item.icon}
                </span>
                {item.label}
              </button>
            );
          })}
        </nav>

        {/* User card + logout */}
        <div className="space-y-2 border-t border-slate-100 p-4">
          <div className="flex items-center gap-3 rounded-xl border border-slate-200 bg-slate-50 p-3">
            <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-gradient-to-br from-blue-600 to-indigo-600 text-sm font-bold uppercase text-white shadow-sm">
              {user?.username?.charAt(0) || 'U'}
            </div>
            <div className="min-w-0 flex-1">
              <p className="truncate text-sm font-semibold text-slate-900">
                {user?.username || 'User'}
              </p>
              <p className="truncate text-xs capitalize text-slate-500">{roleLabel}</p>
            </div>
            <span className="h-2 w-2 shrink-0 rounded-full bg-emerald-500 ring-2 ring-emerald-100" />
          </div>

          <button
            onClick={handleLogout}
            className="flex w-full items-center justify-center gap-2 rounded-xl border border-rose-200 bg-rose-50/60 px-4 py-2.5 text-sm font-medium text-rose-600 transition-colors hover:bg-rose-100"
          >
            {Icons.logout}
            Sign Out
          </button>
        </div>
      </aside>

      {/* --------------------------- Main column --------------------------- */}
      <div className="flex h-screen flex-1 flex-col overflow-hidden">
        {/* Mobile header */}
        <header className="flex h-20 shrink-0 items-center justify-between border-b border-slate-200 bg-white px-6 lg:hidden">
          <div className="flex items-center gap-3">
            <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-gradient-to-tr from-blue-700 to-blue-500">
              {Icons.logo}
            </div>
            <span className="text-lg font-bold text-slate-900">DMS Portal</span>
          </div>
          <button
            onClick={() => setIsSidebarOpen(true)}
            className="-mr-2 rounded-lg p-2 text-slate-600 transition-colors hover:bg-slate-100"
            aria-label="Open menu"
          >
            {Icons.menu}
          </button>
        </header>

        {/* Scroll area */}
        <main className="flex-1 overflow-y-auto">
          <div className="mx-auto max-w-7xl space-y-7 p-6 lg:p-10">
            {/* Breadcrumb */}
            <nav className="flex flex-wrap items-center gap-1.5 text-sm">
              <button
                onClick={() => setActiveTab('dashboard')}
                className="font-medium text-slate-400 transition-colors hover:text-slate-700"
              >
                Workspace
              </button>
              <span className="text-slate-300">{Icons.chevron}</span>
              <span className="font-medium capitalize text-slate-700">
                {activeTab === 'folders'
                  ? 'Folders & Documents'
                  : activeTab === 'audit'
                  ? 'Audit Log'
                  : activeTab}
              </span>
            </nav>

            {/* ============ DASHBOARD TAB ============ */}
            {activeTab === 'dashboard' && (
              <>
                {/* ---- Welcome hero ---- */}
                <section className="relative overflow-hidden rounded-3xl bg-gradient-to-br from-blue-700 via-indigo-600 to-blue-600 px-8 py-14 shadow-2xl shadow-blue-500/20 lg:px-16 lg:py-20">
                  <div
                    className="pointer-events-none absolute inset-0 opacity-[0.12]"
                    style={{
                      backgroundImage:
                        'linear-gradient(to right, #fff 1px, transparent 1px), linear-gradient(to bottom, #fff 1px, transparent 1px)',
                      backgroundSize: '34px 34px',
                    }}
                  />
                  <div className="pointer-events-none absolute -right-16 -top-20 h-72 w-72 rounded-full bg-white/10 blur-3xl" />
                  <div className="pointer-events-none absolute -bottom-24 right-24 h-64 w-64 rounded-full bg-indigo-400/20 blur-3xl" />
                  <div className="pointer-events-none absolute -left-20 bottom-10 h-56 w-56 rounded-full bg-blue-400/20 blur-3xl" />

                  <div className="relative flex flex-col items-center text-center">
                    <div className="mb-6 inline-flex items-center gap-2 rounded-full border border-white/15 bg-white/10 px-4 py-1.5 text-xs font-medium text-blue-50 backdrop-blur-md">
                      <span className="h-2 w-2 animate-pulse rounded-full bg-emerald-400" />
                      Document Management System
                    </div>

                    <div className="mb-6 flex h-20 w-20 items-center justify-center rounded-3xl bg-white/15 backdrop-blur-md ring-1 ring-white/20">
                      <span className="scale-[2.2] text-white/90">{Icons.sparkle}</span>
                    </div>

                    <h1 className="text-3xl font-bold tracking-tight text-white sm:text-4xl lg:text-5xl">
                      Welcome back, {user?.username || 'User'} 👋
                    </h1>

                    <p className="mt-4 max-w-2xl text-sm text-blue-100 sm:text-base lg:text-lg">
                      You're signed in as{' '}
                      <span className="font-semibold text-white capitalize">{roleLabel}</span>.
                      Use the sidebar to manage users, organize folders, and access your documents.
                    </p>

                    <div className="mt-10 flex flex-wrap items-center justify-center gap-3">
                      <button
                        onClick={() => goToTab('users')}
                        className="group inline-flex items-center gap-2 rounded-xl border border-white/20 bg-white/10 px-4 py-2.5 text-sm font-semibold text-white backdrop-blur-md transition-all hover:-translate-y-0.5 hover:bg-white/20"
                      >
                        <span className="text-white/80">{Icons.users}</span>
                        {isAdmin ? 'User Directory' : 'My Profile'}
                        <span className="text-white/60 transition-transform group-hover:translate-x-0.5">
                          {Icons.arrowRight}
                        </span>
                      </button>

                      <button
                        onClick={() => goToTab('folders')}
                        className="group inline-flex items-center gap-2 rounded-xl bg-white px-4 py-2.5 text-sm font-semibold text-blue-700 shadow-lg transition-all hover:-translate-y-0.5 hover:bg-blue-50"
                      >
                        <span className="text-blue-600">{Icons.folder}</span>
                        Folders & Documents
                        <span className="text-blue-400 transition-transform group-hover:translate-x-0.5">
                          {Icons.arrowRight}
                        </span>
                      </button>

                      {isAdmin && (
                        <button
                          onClick={() => goToTab('audit')}
                          className="group inline-flex items-center gap-2 rounded-xl border border-white/20 bg-white/10 px-4 py-2.5 text-sm font-semibold text-white backdrop-blur-md transition-all hover:-translate-y-0.5 hover:bg-white/20"
                        >
                          <span className="text-white/80">{Icons.shield}</span>
                          Audit Log
                          <span className="text-white/60 transition-transform group-hover:translate-x-0.5">
                            {Icons.arrowRight}
                          </span>
                        </button>
                      )}
                    </div>
                  </div>
                </section>

                {/* ---- ADMIN-ONLY STATISTICS ---- */}
                {isAdmin && (
                  <>
                    <div className="flex items-center gap-2.5 pt-2">
                      <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-amber-400 to-orange-500 text-white shadow-md shadow-orange-500/20">
                        {Icons.crown}
                      </span>
                      <div>
                        <h2 className="text-sm font-bold text-slate-900">Admin Overview</h2>
                        <p className="text-[11px] text-slate-500">
                          Live statistics for your workspace
                        </p>
                      </div>
                    </div>

                    <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4">
                      <StatCard
                        label="Total Folders"
                        value={stats.totalFolders}
                        tint="blue"
                        icon={Icons.folder}
                        hint="Workspace folders"
                      />
                      <StatCard
                        label="Team Members"
                        value={stats.totalUsers}
                        tint="indigo"
                        icon={Icons.users}
                        hint={`${stats.admins} admin · ${stats.managers} manager`}
                      />
                      <StatCard
                        label="Admins"
                        value={stats.admins}
                        tint="amber"
                        icon={Icons.crown}
                        hint="Full access"
                      />
                      <StatCard
                        label="Members"
                        value={stats.members}
                        tint="emerald"
                        icon={Icons.shield}
                        hint="Standard access"
                      />
                    </div>

                    <div className="grid grid-cols-1 gap-7 xl:grid-cols-2">
                      <section className="flex flex-col overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
                        <header className="flex items-center justify-between gap-3 border-b border-slate-100 bg-slate-50/70 px-6 py-4">
                          <div className="flex items-center gap-2.5">
                            <span className="text-blue-600">{Icons.folder}</span>
                            <h2 className="text-[15px] font-semibold text-slate-900">
                              Recent Folders
                            </h2>
                          </div>
                          <button
                            onClick={() => goToTab('folders')}
                            className="inline-flex items-center gap-1 text-xs font-bold text-blue-600 transition-colors hover:text-blue-800"
                          >
                            View all {Icons.arrowRight}
                          </button>
                        </header>
                        <div className="flex-1 p-3">
                          {stats.recentFolders.length === 0 ? (
                            <p className="py-8 text-center text-sm text-slate-400">
                              No folders yet.
                            </p>
                          ) : (
                            <ul className="space-y-1">
                              {stats.recentFolders.map((f) => (
                                <li
                                  key={f.id}
                                  className="group flex items-center gap-3 rounded-xl px-3 py-2.5 transition-colors hover:bg-slate-50"
                                >
                                  <span className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-blue-50 text-blue-600 transition-colors group-hover:bg-blue-100">
                                    <svg className="h-4 w-4" viewBox="0 0 24 24" fill="currentColor">
                                      <path d="M3 7a2 2 0 012-2h4l2 2h8a2 2 0 012 2v8a2 2 0 01-2 2H5a2 2 0 01-2-2V7z" />
                                    </svg>
                                  </span>
                                  <div className="min-w-0 flex-1">
                                    <p className="truncate text-sm font-semibold text-slate-800">
                                      {f.name}
                                    </p>
                                    <p className="text-[11px] text-slate-400">
                                      by {(f as any).created_by || 'Unknown'}
                                    </p>
                                  </div>
                                </li>
                              ))}
                            </ul>
                          )}
                        </div>
                      </section>

                      <section className="flex flex-col overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
                        <header className="flex items-center gap-2.5 border-b border-slate-100 bg-slate-50/70 px-6 py-4">
                          <span className="text-indigo-600">{Icons.users}</span>
                          <h2 className="text-[15px] font-semibold text-slate-900">
                            Role Distribution
                          </h2>
                        </header>
                        <div className="flex-1 space-y-5 p-6">
                          <RoleBar
                            label="Admins"
                            count={stats.admins}
                            total={stats.totalUsers || 1}
                            color="from-amber-400 to-orange-500"
                          />
                          <RoleBar
                            label="Managers"
                            count={stats.managers}
                            total={stats.totalUsers || 1}
                            color="from-blue-500 to-indigo-600"
                          />
                          <RoleBar
                            label="Members"
                            count={stats.members}
                            total={stats.totalUsers || 1}
                            color="from-emerald-500 to-teal-600"
                          />

                          <div className="flex items-center gap-2 rounded-xl border border-slate-200 bg-slate-50/70 px-3.5 py-3 text-xs text-slate-500">
                            <span className="text-emerald-500">{Icons.trendUp}</span>
                            <p>
                              <span className="font-semibold text-slate-700">
                                {stats.totalUsers}
                              </span>{' '}
                              total accounts across{' '}
                              <span className="font-semibold text-slate-700">
                                {stats.totalFolders}
                              </span>{' '}
                              folders
                            </p>
                          </div>
                        </div>
                      </section>
                    </div>
                  </>
                )}
              </>
            )}

            {/* ============ USERS TAB ============ */}
            {activeTab === 'users' && (
              <Panel
                title={isAdmin ? 'Team Directory & User Management' : 'My User Profile'}
                icon={Icons.users}
                iconClass="text-blue-600"
              >
                <UserManagement
                  users={users}
                  onUserChange={fetchData}
                  onEditUser={handleEditUser}
                  currentUser={user}
                />
              </Panel>
            )}

            {/* ============ FOLDERS TAB ============ */}
            {activeTab === 'folders' && (
              <Panel
                title="Manage Folders & Documents"
                icon={Icons.settings}
                iconClass="text-indigo-500"
              >
                <FolderManagement
                  folders={folders}
                  onFolderChange={fetchData}
                  canCreate={canCreate}
                  canDelete={isAdmin}
                />
              </Panel>
            )}

            {/* ============ AUDIT LOG TAB (admin only) ============ */}
            {activeTab === 'audit' && isAdmin && (
              <Panel
                title="Audit Log — System Activity"
                icon={Icons.shield}
                iconClass="text-emerald-600"
              >
                <AuditLog />
              </Panel>
            )}
          </div>
        </main>
      </div>

      {/* Edit user modal */}
      {isModalOpen && selectedUser && (
        <EditUserModal
          user={selectedUser}
          currentUser={user}
          isOwnProfile={selectedUser.id === user?.id}
          onClose={handleCloseModal}
          onSuccess={handleSuccess}
        />
      )}
    </div>
  );
}

/* ------------------------------------------------------------------ */
/*  Sub-components                                                     */
/* ------------------------------------------------------------------ */
function Panel({
  title,
  icon,
  iconClass = 'text-blue-600',
  children,
}: {
  title: string;
  icon: ReactNode;
  iconClass?: string;
  children: ReactNode;
}) {
  return (
    <section className="flex flex-col overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm transition-shadow hover:shadow-md">
      <header className="flex items-center gap-3 border-b border-slate-100 bg-slate-50/70 px-6 py-4">
        <span className={iconClass}>{icon}</span>
        <h2 className="truncate text-[15px] font-semibold text-slate-900">{title}</h2>
      </header>
      <div className="flex-1 p-6">{children}</div>
    </section>
  );
}

const TINTS = {
  blue: {
    bg: 'bg-blue-50',
    text: 'text-blue-600',
    ring: 'ring-blue-100',
    bar: 'from-blue-500 to-indigo-600',
  },
  indigo: {
    bg: 'bg-indigo-50',
    text: 'text-indigo-600',
    ring: 'ring-indigo-100',
    bar: 'from-indigo-500 to-purple-600',
  },
  emerald: {
    bg: 'bg-emerald-50',
    text: 'text-emerald-600',
    ring: 'ring-emerald-100',
    bar: 'from-emerald-500 to-teal-600',
  },
  amber: {
    bg: 'bg-amber-50',
    text: 'text-amber-600',
    ring: 'ring-amber-100',
    bar: 'from-amber-400 to-orange-500',
  },
} as const;

function StatCard({
  label,
  value,
  icon,
  tint,
  hint,
}: {
  label: string;
  value: string | number;
  icon: ReactNode;
  tint: keyof typeof TINTS;
  hint?: string;
}) {
  const t = TINTS[tint];
  return (
    <div className="group relative overflow-hidden rounded-2xl border border-slate-200 bg-white p-5 shadow-sm transition-all hover:-translate-y-0.5 hover:border-slate-300 hover:shadow-lg hover:shadow-slate-900/5">
      <div
        className={`absolute inset-x-0 top-0 h-1 bg-gradient-to-r ${t.bar} opacity-0 transition-opacity group-hover:opacity-100`}
      />
      <div className="flex items-start justify-between">
        <div
          className={`flex h-11 w-11 items-center justify-center rounded-xl ring-4 ${t.bg} ${t.text} ${t.ring}`}
        >
          {icon}
        </div>
        <span className={`text-[10px] font-bold uppercase tracking-wider ${t.text}`}>
          Live
        </span>
      </div>
      <p className="mt-4 text-3xl font-bold tracking-tight text-slate-900">{value}</p>
      <p className="mt-0.5 text-[11px] font-bold uppercase tracking-[0.12em] text-slate-400">
        {label}
      </p>
      {hint && <p className="mt-1.5 text-[11px] text-slate-400">{hint}</p>}
    </div>
  );
}

function RoleBar({
  label,
  count,
  total,
  color,
}: {
  label: string;
  count: number;
  total: number;
  color: string;
}) {
  const pct = total > 0 ? Math.round((count / total) * 100) : 0;
  return (
    <div>
      <div className="mb-1.5 flex items-center justify-between text-xs">
        <span className="font-semibold text-slate-700">{label}</span>
        <span className="font-mono text-slate-500">
          {count} · {pct}%
        </span>
      </div>
      <div className="h-2 overflow-hidden rounded-full bg-slate-100">
        <div
          className={`h-full rounded-full bg-gradient-to-r ${color} transition-all duration-500`}
          style={{ width: `${pct}%` }}
        />
      </div>
    </div>
  );
}