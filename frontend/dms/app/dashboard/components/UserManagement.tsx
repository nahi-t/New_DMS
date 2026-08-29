'use client';

import { User } from '@/type';
import { Pencil, Trash2 } from 'lucide-react';
import { deleteUser } from '@/lib/api';
import toast from 'react-hot-toast';

interface UserManagementProps {
  users: User[];
  onUserChange: () => void;
  onEditUser: (user: User) => void;   // <-- callback to open modal
  currentUser?: User | null;          // for self‑edit context
}

export default function UserManagement({
  users,
  onUserChange,
  onEditUser,
  currentUser,
}: UserManagementProps) {
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

  // If the logged-in user is not admin, show only their own profile
  const displayUsers = currentUser?.role === 'admin' ? users : (currentUser ? [currentUser] : []);

  return (
    <div className="space-y-4">
      {displayUsers.length === 0 ? (
        <p className="text-gray-500 dark:text-gray-400 text-sm">No users found.</p>
      ) : (
        <ul className="divide-y divide-gray-200 dark:divide-gray-700">
          {displayUsers.map((user) => (
            <li key={user.id} className="py-3 flex items-center justify-between">
              <div>
                <p className="font-medium text-gray-900 dark:text-white">{user.username}</p>
                <p className="text-sm text-gray-500 dark:text-gray-400">{user.email}</p>
                <span className="inline-block mt-1 text-xs px-2 py-1 rounded bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300">
                  {user.role}
                </span>
              </div>
              <div className="flex items-center space-x-2">
                {/* Edit button – calls onEditUser */}
                <button
                  onClick={() => onEditUser(user)}
                  className="text-blue-600 hover:text-blue-800 transition"
                  aria-label="Edit user"
                >
                  <Pencil className="w-4 h-4" />
                </button>

                {/* Delete – only for admins and not self */}
                {currentUser?.role === 'admin' && user.id !== currentUser.id && (
                  <button
                    onClick={() => handleDelete(user.id)}
                    className="text-red-600 hover:text-red-800 transition"
                    aria-label="Delete user"
                  >
                    <Trash2 className="w-4 h-4" />
                  </button>
                )}
              </div>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}