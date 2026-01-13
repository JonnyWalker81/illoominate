import { useState, useEffect } from 'react'
import { Link, useNavigate } from '@tanstack/react-router'
import { ArrowLeft, ChevronRight, Trash2 } from 'lucide-react'
import { api } from '../lib/api'
import type { Project } from '../types'

interface ProjectHeaderProps {
  projectId: string
}

export function ProjectHeader({ projectId }: ProjectHeaderProps) {
  const navigate = useNavigate()
  const [project, setProject] = useState<Project | null>(null)
  const [showDeleteModal, setShowDeleteModal] = useState(false)
  const [isDeleting, setIsDeleting] = useState(false)

  useEffect(() => {
    const fetchProject = async () => {
      try {
        const data = (await api.projects.get(projectId)) as Project
        setProject(data)
      } catch (err) {
        console.error('Failed to fetch project:', err)
      }
    }
    fetchProject()
  }, [projectId])

  const handleDelete = async () => {
    setIsDeleting(true)
    try {
      await api.projects.delete(projectId)
      navigate({ to: '/projects' })
    } catch (err) {
      console.error('Failed to delete project:', err)
      setIsDeleting(false)
    }
  }

  return (
    <>
      <header className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            {/* Breadcrumb */}
            <div className="flex items-center space-x-2">
              <Link
                to="/projects"
                className="flex items-center text-gray-500 hover:text-gray-700"
              >
                <ArrowLeft className="w-4 h-4 mr-1" />
                Projects
              </Link>
              <ChevronRight className="w-4 h-4 text-gray-400" />
              <span className="text-lg font-semibold text-gray-900">
                {project?.name || 'Loading...'}
              </span>
            </div>

            <div className="flex items-center space-x-4">
              <nav className="flex space-x-4">
                <Link
                  to="/projects/$projectId/community"
                  params={{ projectId }}
                  className="px-3 py-2 text-sm font-medium text-gray-600 hover:text-gray-900 [&.active]:text-primary [&.active]:border-b-2 [&.active]:border-primary"
                >
                  Community
                </Link>
                <Link
                  to="/projects/$projectId/console"
                  params={{ projectId }}
                  className="px-3 py-2 text-sm font-medium text-gray-600 hover:text-gray-900 [&.active]:text-primary [&.active]:border-b-2 [&.active]:border-primary"
                >
                  Console
                </Link>
              </nav>
              <button
                onClick={() => setShowDeleteModal(true)}
                className="p-2 text-gray-400 hover:text-red-600 transition-colors"
                title="Delete project"
              >
                <Trash2 className="w-5 h-5" />
              </button>
            </div>
          </div>
        </div>
      </header>

      {/* Delete Confirmation Modal */}
      {showDeleteModal && (
        <div className="fixed inset-0 z-50 overflow-y-auto">
          <div className="flex min-h-full items-center justify-center p-4">
            <div
              className="fixed inset-0 bg-black/50 transition-opacity"
              onClick={() => setShowDeleteModal(false)}
            />
            <div className="relative bg-white rounded-lg shadow-xl w-full max-w-md p-6">
              <h2 className="text-lg font-semibold text-gray-900 mb-4">
                Delete Project
              </h2>
              <p className="text-gray-600 mb-6">
                Are you sure you want to delete{' '}
                <strong>{project?.name}</strong>? This action cannot be undone
                and all project data will be permanently deleted.
              </p>
              <div className="flex justify-end gap-3">
                <button
                  onClick={() => setShowDeleteModal(false)}
                  className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
                >
                  Cancel
                </button>
                <button
                  onClick={handleDelete}
                  disabled={isDeleting}
                  className="px-4 py-2 text-sm font-medium text-white bg-red-600 rounded-md hover:bg-red-700 disabled:opacity-50"
                >
                  {isDeleting ? 'Deleting...' : 'Delete Project'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </>
  )
}
