'use client';

import { useAuth } from '@/hooks/useAuth';
import Header from './components/Header';
import UserManagement from './components/UserManagement';
import FolderManagement from './components/FolderManagement';
import EditUserModal from './components/EditUserModal';
import { useEffect, useState } from 'react';
import { getFolders, getUsers } from '@/lib/api';
import { Folder, User } from '@/type';
import toast from 'react-hot-toast';

export default function DashboardPage() {
  const { user } = useAuth();
  const [folders, setFolders] = useState<Folder[]>([]);
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);

  // Modal state
  const [selectedUser, setSelectedUser] = useState<User | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);

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

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500"></div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      <Header />
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
          {/* User Management */}
          <div className="bg-white dark:bg-gray-800 rounded-xl shadow overflow-hidden">
            <div className="p-6 border-b border-gray-200 dark:border-gray-700">
              <h2 className="text-xl font-semibold text-gray-900 dark:text-white">
                {user?.role === 'admin' ? 'All Users' : 'My Profile'}
              </h2>
            </div>
            <div className="p-6">
              <UserManagement
                users={users}
                onUserChange={fetchData}
                onEditUser={handleEditUser}
                currentUser={user}
              />
            </div>
          </div>

          {/* Folder Management */}
          <div className="bg-white dark:bg-gray-800 rounded-xl shadow overflow-hidden">
            <div className="p-6 border-b border-gray-200 dark:border-gray-700 flex justify-between items-center">
              <h2 className="text-xl font-semibold text-gray-900 dark:text-white">Folders</h2>
            </div>
            <div className="p-6">
              <FolderManagement
                folders={folders}
                onFolderChange={fetchData}
                canCreate={user?.role === 'admin' || user?.role === 'manager'}
                canDelete={user?.role === 'admin'}
              />
            </div>
          </div>
        </div>
      </div>

      {/* Modal – with currentUser passed */}
      {isModalOpen && selectedUser && (
        <EditUserModal
          user={selectedUser}
          currentUser={user}                     // <-- pass logged‑in user
          isOwnProfile={selectedUser.id === user?.id}
          onClose={handleCloseModal}
          onSuccess={handleSuccess}
        />
      )}
    </div>
  );
}