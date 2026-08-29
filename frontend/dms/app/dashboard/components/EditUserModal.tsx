'use client';

import { useState, useEffect } from 'react';
import { updateUser, updatePassword } from '@/lib/api';
import { X } from 'lucide-react';
import toast from 'react-hot-toast';
import { User } from '@/type';

interface Props {
  user: User;
  currentUser?: User | null;
  isOwnProfile?: boolean;
  onClose: () => void;
  onSuccess: () => void;
}

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

  const canEditRole = !isOwnProfile || currentUser?.role === 'admin';

  useEffect(() => {
    setUsername(user.username);
    setEmail(user.email);
    setRole(user.role);
    setNewPassword('');
    setConfirmPassword('');
  }, [user]);

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

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-md w-full p-6">
        <div className="flex justify-between items-center mb-4">
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
            {isOwnProfile ? 'Edit My Profile' : 'Edit User'}
          </h3>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600">
            <X className="h-5 w-5" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
              Username
            </label>
            <input
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              className="mt-1 w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-gray-50 dark:bg-gray-700 p-2"
              required
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
              Email
            </label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="mt-1 w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-gray-50 dark:bg-gray-700 p-2"
              required
            />
          </div>

          {canEditRole && (
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
                Role
              </label>
              <select
                value={role}
                onChange={(e) => setRole(e.target.value)}
                className="mt-1 w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-gray-50 dark:bg-gray-700 p-2"
              >
                <option value="user">User</option>
                <option value="manager">Manager</option>
                <option value="admin">Admin</option>
              </select>
            </div>
          )}

          <hr className="my-2 border-gray-300 dark:border-gray-600" />
          <p className="text-sm font-medium text-gray-700 dark:text-gray-300">
            Change Password (optional)
          </p>

          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
              New Password
            </label>
            <input
              type="password"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              className="mt-1 w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-gray-50 dark:bg-gray-700 p-2"
              placeholder="Leave blank to keep current"
              minLength={6}
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
              Confirm New Password
            </label>
            <input
              type="password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              className="mt-1 w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-gray-50 dark:bg-gray-700 p-2"
              placeholder="Re‑enter new password"
            />
          </div>

          <div className="flex justify-end space-x-3 pt-2">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-lg transition"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={loading}
              className="px-4 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded-lg transition disabled:opacity-50"
            >
              {loading ? 'Saving...' : 'Save'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}