'use client';

import { useAuth } from '@/hooks/useAuth';
import { User } from '@/type';  // make sure your types file is named 'types'
import { deleteUser } from '@/lib/api';
import { Trash2, Plus, Pencil } from 'lucide-react';
import toast from 'react-hot-toast';
import { useState } from 'react';
import CreateUserModal from './CreateUserModal';
import EditUserModal from './EditUserModal';

interface Props {
  users: User[];
  onUserChange: () => void;
}

export default function UserManagement({ users, onUserChange }: Props) {
  const { user } = useAuth();
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [editingUser, setEditingUser] = useState<User | null>(null);

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

  // Profile card for non‑admin users
  if (user?.role !== 'admin') {
    return (
      <div className="space-y-3">
        <div className="bg-gray-50 dark:bg-gray-700 p-4 rounded-lg">
          <p className="text-sm text-gray-500 dark:text-gray-400">Username</p>
          <p className="font-medium text-gray-900 dark:text-white">{user?.username}</p>
        </div>
        <div className="bg-gray-50 dark:bg-gray-700 p-4 rounded-lg">
          <p className="text-sm text-gray-500 dark:text-gray-400">Email</p>
          <p className="font-medium text-gray-900 dark:text-white">{user?.email}</p>
        </div>
        <div className="bg-gray-50 dark:bg-gray-700 p-4 rounded-lg">
          <p className="text-sm text-gray-500 dark:text-gray-400">Role</p>
          <p className="font-medium text-gray-900 dark:text-white capitalize">{user?.role}</p>
        </div>
        <div className="bg-gray-50 dark:bg-gray-700 p-4 rounded-lg">
          <p className="text-sm text-gray-500 dark:text-gray-400">Joined</p>
          <p className="font-medium text-gray-900 dark:text-white">
            {user?.created_at ? new Date(user.created_at).toLocaleDateString() : 'N/A'}
          </p>
        </div>
      </div>
    );
  }

  // Admin view – full user table with Edit + Delete
  return (
    <div>
      <div className="flex justify-end mb-4">
        <button
          onClick={() => setShowCreateModal(true)}
          className="flex items-center space-x-1 bg-blue-600 hover:bg-blue-700 text-white text-sm px-3 py-1.5 rounded-lg transition"
        >
          <Plus className="h-4 w-4" />
          <span>Add User</span>
        </button>
      </div>

      <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
          <thead>
            <tr>
              <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">ID</th>
              <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Username</th>
              <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Email</th>
              <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Role</th>
              <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Created</th>
              <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
            {users.map((u) => (
              <tr key={u.id}>
                <td className="px-3 py-2 whitespace-nowrap text-sm text-gray-900 dark:text-white">{u.id}</td>
                <td className="px-3 py-2 whitespace-nowrap text-sm text-gray-900 dark:text-white">{u.username}</td>
                <td className="px-3 py-2 whitespace-nowrap text-sm text-gray-900 dark:text-white">{u.email}</td>
                <td className="px-3 py-2 whitespace-nowrap text-sm capitalize">{u.role}</td>
                <td className="px-3 py-2 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                  {new Date(u.created_at).toLocaleDateString()}
                </td>
                <td className="px-3 py-2 whitespace-nowrap text-right space-x-2">
                  {/* Edit button – Admin only, and optionally prevent self‑edit */}
                  <button
                    onClick={() => setEditingUser(u)}
                    className="text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300"
                    disabled={u.id === user?.id} // optional: prevent editing yourself
                  >
                    <Pencil className="h-5 w-5 inline" />
                  </button>
                  <button
                    onClick={() => handleDelete(u.id)}
                    className="text-red-600 hover:text-red-800 dark:text-red-400 dark:hover:text-red-300"
                    disabled={u.id === user?.id}
                  >
                    <Trash2 className="h-5 w-5 inline" />
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Create User Modal */}
      {showCreateModal && (
        <CreateUserModal
          onClose={() => setShowCreateModal(false)}
          onSuccess={onUserChange}
        />
      )}

      {/* Edit User Modal */}
      {editingUser && (
        <EditUserModal
          user={editingUser}
          onClose={() => setEditingUser(null)}
          onSuccess={() => {
            setEditingUser(null);
            onUserChange();
          }}
        />
      )}
    </div>
  );
}