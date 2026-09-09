'use client';

import { Folder } from '@/type';
import { deleteFolder } from '@/lib/api';
import { Trash2, Plus, Pencil, ChevronDown, ChevronRight } from 'lucide-react';
import toast from 'react-hot-toast';
import React, { useState } from 'react';
import CreateFolderModal from './CreateFolderModal';
import EditFolderModal from './EditFolderModal';
import DocumentManager from './DocumentManager';

interface Props {
  folders: Folder[];
  onFolderChange: () => void;
  canCreate: boolean;
  canDelete: boolean;
}

export default function FolderManagement({ folders, onFolderChange, canCreate, canDelete }: Props) {
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [editingFolder, setEditingFolder] = useState<{ id: number; name: string } | null>(null);
  const [expandedFolders, setExpandedFolders] = useState<Set<number>>(new Set());

  const toggleExpand = (folderId: number) => {
    setExpandedFolders((prev) => {
      const newSet = new Set(prev);
      if (newSet.has(folderId)) {
        newSet.delete(folderId);
      } else {
        newSet.add(folderId);
      }
      return newSet;
    });
  };

  const handleDelete = async (id: number) => {
    if (!confirm('Delete this folder and all its documents?')) return;
    try {
      await deleteFolder(id);
      toast.success('Folder deleted');
      onFolderChange();
    } catch (error: any) {
      toast.error(error.message || 'Failed to delete folder');
    }
  };

  const colSpan = canDelete ? 5 : 4;

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
              <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Folder</th>
              <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Created By</th>
              <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Created</th>
              {canDelete && (
                <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
              )}
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
            {folders.map((f) => {
              const isExpanded = expandedFolders.has(f.id);
              return (
                <React.Fragment key={f.id}>
                  {/* Main folder row */}
                  <tr>
                    <td className="px-3 py-2 whitespace-nowrap text-sm text-gray-900 dark:text-white">
                      <button
                        onClick={() => toggleExpand(f.id)}
                        className="mr-2 text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300"
                        title={isExpanded ? 'Collapse' : 'Expand'}
                      >
                        {isExpanded ? (
                          <ChevronDown className="h-4 w-4 inline" />
                        ) : (
                          <ChevronRight className="h-4 w-4 inline" />
                        )}
                      </button>
                      {f.name}
                    </td>
                    <td className="px-3 py-2 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                      {f.created_by}
                    </td>
                    <td className="px-3 py-2 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                      {new Date(f.created_at).toLocaleDateString()}
                    </td>
                    {canDelete && (
                      <td className="px-3 py-2 whitespace-nowrap text-right space-x-2">
                        <button
                          onClick={() => setEditingFolder({ id: f.id, name: f.name })}
                          className="text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300"
                        >
                          <Pencil className="h-5 w-5 inline" />
                        </button>
                        <button
                          onClick={() => handleDelete(f.id)}
                          className="text-red-600 hover:text-red-800 dark:text-red-400 dark:hover:text-red-300"
                        >
                          <Trash2 className="h-5 w-5 inline" />
                        </button>
                      </td>
                    )}
                  </tr>

                  {/* Document expansion row - now with unique key */}
                  {isExpanded && (
                    <tr key={`${f.id}-docs`}>
                      <td colSpan={colSpan} className="px-3 py-2 bg-gray-50 dark:bg-gray-800/50">
                        <DocumentManager
                          folderId={f.id}
                          canUpload={true}   
                          canDelete={canDelete}
                           allFolders={folders} 
                        />
                      </td>
                    </tr>
                  )}
                </React.Fragment>
              );
            })}
          </tbody>
        </table>
        {folders.length === 0 && (
          <p className="text-center text-gray-500 dark:text-gray-400 py-4">No folders found.</p>
        )}
      </div>

      {/* Modals */}
      {showCreateModal && (
        <CreateFolderModal
          onClose={() => setShowCreateModal(false)}
          onSuccess={onFolderChange}
        />
      )}
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