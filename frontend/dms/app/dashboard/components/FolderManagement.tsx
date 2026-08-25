'use client';

import { Folder } from '@/type';  // make sure path matches your types
import { deleteFolder } from '@/lib/api';
import { Trash2, Plus, Pencil } from 'lucide-react';
import toast from 'react-hot-toast';
import { useState } from 'react';
import CreateFolderModal from './CreateFolderModal';
import EditFolderModal from './EditFolderModal';

interface Props {
  folders: Folder[];
  onFolderChange: () => void;
  canCreate: boolean;
  canDelete: boolean;  // admin only – gives both edit and delete permissions
}

export default function FolderManagement({ folders, onFolderChange, canCreate, canDelete }: Props) {
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [editingFolder, setEditingFolder] = useState<{ id: number; name: string } | null>(null);

  const handleDelete = async (id: number) => {
    if (!confirm('Delete this folder?')) return;
    try {
      await deleteFolder(id);
      toast.success('Folder deleted');
      onFolderChange();
    } catch (error: any) {
      toast.error(error.message || 'Failed to delete folder');
    }
  };

  return (
    <div>
      {canCreate && (
        <div className="flex justify-end mb-4">
          <button
            onClick={() => setShowCreateModal(true)}
            className="flex items-center space-x-1 bg-blue-600 hover:bg-blue-700 text-white text-sm px-3 py-1.5 rounded-lg transition"
          >
            <Plus className="h-4 w-4" />
            <span>New Folder</span>
          </button>
        </div>
      )}

      <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
          <thead>
            <tr>
              <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">ID</th>
              <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>
              <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Created By</th>
              <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Created</th>
              {canDelete && (
                <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
              )}
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
            {folders.map((f) => (
              <tr key={f.id}>
                <td className="px-3 py-2 whitespace-nowrap text-sm text-gray-900 dark:text-white">{f.id}</td>
                <td className="px-3 py-2 whitespace-nowrap text-sm text-gray-900 dark:text-white">{f.name}</td>
                <td className="px-3 py-2 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">{f.created_by}</td>
                <td className="px-3 py-2 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                  {new Date(f.created_at).toLocaleDateString()}
                </td>
                {canDelete && (
                  <td className="px-3 py-2 whitespace-nowrap text-right space-x-2">
                    {/* Edit button */}
                    <button
                      onClick={() => setEditingFolder({ id: f.id, name: f.name })}
                      className="text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300"
                    >
                      <Pencil className="h-5 w-5 inline" />
                    </button>
                    {/* Delete button */}
                    <button
                      onClick={() => handleDelete(f.id)}
                      className="text-red-600 hover:text-red-800 dark:text-red-400 dark:hover:text-red-300"
                    >
                      <Trash2 className="h-5 w-5 inline" />
                    </button>
                  </td>
                )}
              </tr>
            ))}
          </tbody>
        </table>
        {folders.length === 0 && (
          <p className="text-center text-gray-500 dark:text-gray-400 py-4">No folders found.</p>
        )}
      </div>

      {/* Create Folder Modal */}
      {showCreateModal && (
        <CreateFolderModal
          onClose={() => setShowCreateModal(false)}
          onSuccess={onFolderChange}
        />
      )}

      {/* Edit Folder Modal */}
      {editingFolder && (
        <EditFolderModal
          folderId={editingFolder.id}
          currentName={editingFolder.name}
          onClose={() => setEditingFolder(null)}
          onSuccess={() => {
            setEditingFolder(null);
            onFolderChange();
          }}
        />
      )}
    </div>
  );
}