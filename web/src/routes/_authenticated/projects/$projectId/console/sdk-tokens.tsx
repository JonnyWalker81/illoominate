import { useState, useEffect } from 'react'
import { createFileRoute, useParams } from '@tanstack/react-router'
import { ProjectHeader } from '../../../../../components/ProjectHeader'
import { ConsoleNav } from '../../../../../components/ConsoleNav'
import { api } from '../../../../../lib/api'
import type { SDKToken } from '../../../../../types'
import { Key, Plus, Copy, Check, Trash2, AlertTriangle, X } from 'lucide-react'

export const Route = createFileRoute(
  '/_authenticated/projects/$projectId/console/sdk-tokens'
)({
  component: SDKTokensPage,
})

function SDKTokensPage() {
  const { projectId } = useParams({
    from: '/_authenticated/projects/$projectId/console/sdk-tokens',
  })

  const [tokens, setTokens] = useState<SDKToken[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // Create modal state
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [creating, setCreating] = useState(false)
  const [newTokenName, setNewTokenName] = useState('')
  const [newAllowedOrigins, setNewAllowedOrigins] = useState('')
  const [newRateLimit, setNewRateLimit] = useState('60')

  // Token created modal state
  const [createdToken, setCreatedToken] = useState<SDKToken | null>(null)
  const [copied, setCopied] = useState(false)

  // Revoke modal state
  const [tokenToRevoke, setTokenToRevoke] = useState<SDKToken | null>(null)
  const [revoking, setRevoking] = useState(false)

  useEffect(() => {
    fetchTokens()
  }, [projectId])

  const fetchTokens = async () => {
    try {
      setLoading(true)
      const data = (await api.creator.listSDKTokens(projectId)) as SDKToken[]
      setTokens(data)
      setError(null)
    } catch (err) {
      setError('Failed to load SDK tokens')
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!newTokenName.trim()) return

    setCreating(true)
    try {
      const origins = newAllowedOrigins
        .split(',')
        .map((o) => o.trim())
        .filter((o) => o.length > 0)

      const token = (await api.creator.createSDKToken(projectId, {
        name: newTokenName.trim(),
        allowed_origins: origins.length > 0 ? origins : undefined,
        rate_limit_per_minute: parseInt(newRateLimit) || 60,
      })) as SDKToken

      setCreatedToken(token)
      setShowCreateModal(false)
      setNewTokenName('')
      setNewAllowedOrigins('')
      setNewRateLimit('60')
      fetchTokens()
    } catch (err) {
      console.error('Failed to create token:', err)
      alert('Failed to create token')
    } finally {
      setCreating(false)
    }
  }

  const handleCopy = async () => {
    if (!createdToken?.token) return
    try {
      await navigator.clipboard.writeText(createdToken.token)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch (err) {
      console.error('Failed to copy:', err)
    }
  }

  const handleRevoke = async () => {
    if (!tokenToRevoke) return
    setRevoking(true)
    try {
      await api.creator.revokeSDKToken(projectId, tokenToRevoke.id)
      setTokenToRevoke(null)
      fetchTokens()
    } catch (err) {
      console.error('Failed to revoke token:', err)
      alert('Failed to revoke token')
    } finally {
      setRevoking(false)
    }
  }

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    })
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <ProjectHeader projectId={projectId} />
      <ConsoleNav projectId={projectId} />

      <div className="max-w-4xl mx-auto py-8 px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between items-center mb-8">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">SDK Tokens</h1>
            <p className="mt-1 text-sm text-gray-500">
              Manage API tokens for your iOS and Android SDKs
            </p>
          </div>
          <button
            onClick={() => setShowCreateModal(true)}
            className="flex items-center gap-2 px-4 py-2 bg-primary text-white rounded-md hover:bg-primary/90 transition-colors"
          >
            <Plus className="w-4 h-4" />
            Create Token
          </button>
        </div>

        {loading ? (
          <div className="bg-white shadow rounded-lg p-6 text-center text-gray-500">
            Loading tokens...
          </div>
        ) : error ? (
          <div className="bg-white shadow rounded-lg p-6 text-center text-red-500">
            {error}
          </div>
        ) : tokens.length === 0 ? (
          <div className="bg-white shadow rounded-lg p-12 text-center">
            <Key className="w-12 h-12 mx-auto text-gray-400 mb-4" />
            <h3 className="text-lg font-medium text-gray-900 mb-2">
              No SDK tokens yet
            </h3>
            <p className="text-gray-500 mb-6">
              Create a token to integrate your mobile apps with FullDisclosure.
            </p>
            <button
              onClick={() => setShowCreateModal(true)}
              className="inline-flex items-center gap-2 px-4 py-2 bg-primary text-white rounded-md hover:bg-primary/90"
            >
              <Plus className="w-4 h-4" />
              Create Your First Token
            </button>
          </div>
        ) : (
          <div className="bg-white shadow rounded-lg overflow-hidden">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Name
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Token
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Last Used
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Created
                  </th>
                  <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {tokens.map((token) => (
                  <tr key={token.id}>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center">
                        <Key className="w-4 h-4 text-gray-400 mr-2" />
                        <span className="text-sm font-medium text-gray-900">
                          {token.name}
                        </span>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <code className="text-sm text-gray-500 bg-gray-100 px-2 py-1 rounded">
                        {token.token_prefix}...
                      </code>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {token.last_used_at
                        ? formatDate(token.last_used_at)
                        : 'Never'}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {formatDate(token.created_at)}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right">
                      <button
                        onClick={() => setTokenToRevoke(token)}
                        className="text-red-600 hover:text-red-800 transition-colors"
                        title="Revoke token"
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Create Token Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 z-50 overflow-y-auto">
          <div className="flex min-h-full items-center justify-center p-4">
            <div
              className="fixed inset-0 bg-black/50 transition-opacity"
              onClick={() => setShowCreateModal(false)}
            />
            <div className="relative bg-white rounded-lg shadow-xl w-full max-w-md p-6">
              <div className="flex justify-between items-center mb-4">
                <h2 className="text-lg font-semibold text-gray-900">
                  Create SDK Token
                </h2>
                <button
                  onClick={() => setShowCreateModal(false)}
                  className="text-gray-400 hover:text-gray-500"
                >
                  <X className="w-5 h-5" />
                </button>
              </div>

              <form onSubmit={handleCreate}>
                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Token Name *
                    </label>
                    <input
                      type="text"
                      value={newTokenName}
                      onChange={(e) => setNewTokenName(e.target.value)}
                      className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent"
                      placeholder="e.g., iOS Production"
                      required
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Allowed Origins (optional)
                    </label>
                    <input
                      type="text"
                      value={newAllowedOrigins}
                      onChange={(e) => setNewAllowedOrigins(e.target.value)}
                      className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent"
                      placeholder="https://example.com, https://app.example.com"
                    />
                    <p className="mt-1 text-xs text-gray-500">
                      Comma-separated list of allowed origins. Leave empty to
                      allow all.
                    </p>
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Rate Limit (requests/minute)
                    </label>
                    <input
                      type="number"
                      value={newRateLimit}
                      onChange={(e) => setNewRateLimit(e.target.value)}
                      className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent"
                      min="1"
                      max="1000"
                    />
                  </div>
                </div>

                <div className="mt-6 flex justify-end gap-3">
                  <button
                    type="button"
                    onClick={() => setShowCreateModal(false)}
                    className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
                  >
                    Cancel
                  </button>
                  <button
                    type="submit"
                    disabled={creating || !newTokenName.trim()}
                    className="px-4 py-2 text-sm font-medium text-white bg-primary rounded-md hover:bg-primary/90 disabled:opacity-50"
                  >
                    {creating ? 'Creating...' : 'Create Token'}
                  </button>
                </div>
              </form>
            </div>
          </div>
        </div>
      )}

      {/* Token Created Modal */}
      {createdToken && (
        <div className="fixed inset-0 z-50 overflow-y-auto">
          <div className="flex min-h-full items-center justify-center p-4">
            <div className="fixed inset-0 bg-black/50 transition-opacity" />
            <div className="relative bg-white rounded-lg shadow-xl w-full max-w-lg p-6">
              <div className="flex items-center gap-3 mb-4">
                <div className="flex-shrink-0 w-10 h-10 bg-yellow-100 rounded-full flex items-center justify-center">
                  <AlertTriangle className="w-5 h-5 text-yellow-600" />
                </div>
                <div>
                  <h2 className="text-lg font-semibold text-gray-900">
                    Token Created Successfully
                  </h2>
                  <p className="text-sm text-gray-500">
                    Copy your token now. You won't be able to see it again!
                  </p>
                </div>
              </div>

              <div className="bg-gray-900 rounded-lg p-4 mb-4">
                <div className="flex items-center justify-between">
                  <code className="text-green-400 text-sm break-all font-mono">
                    {createdToken.token}
                  </code>
                  <button
                    onClick={handleCopy}
                    className="ml-3 flex-shrink-0 p-2 text-gray-400 hover:text-white transition-colors"
                    title="Copy to clipboard"
                  >
                    {copied ? (
                      <Check className="w-5 h-5 text-green-400" />
                    ) : (
                      <Copy className="w-5 h-5" />
                    )}
                  </button>
                </div>
              </div>

              <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4 mb-6">
                <p className="text-sm text-yellow-800">
                  <strong>Important:</strong> This token will only be displayed
                  once. Store it securely in your app's configuration. If you
                  lose it, you'll need to create a new token.
                </p>
              </div>

              <div className="bg-gray-50 rounded-lg p-4 mb-6">
                <p className="text-sm font-medium text-gray-700 mb-2">
                  Use this token in your iOS app:
                </p>
                <pre className="text-xs bg-gray-800 text-gray-100 rounded p-3 overflow-x-auto">
                  {`FullDisclosure.shared.initialize(
    token: "${createdToken.token}"
)`}
                </pre>
              </div>

              <div className="flex justify-end">
                <button
                  onClick={() => {
                    setCreatedToken(null)
                    setCopied(false)
                  }}
                  className="px-4 py-2 text-sm font-medium text-white bg-primary rounded-md hover:bg-primary/90"
                >
                  Done
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Revoke Confirmation Modal */}
      {tokenToRevoke && (
        <div className="fixed inset-0 z-50 overflow-y-auto">
          <div className="flex min-h-full items-center justify-center p-4">
            <div
              className="fixed inset-0 bg-black/50 transition-opacity"
              onClick={() => setTokenToRevoke(null)}
            />
            <div className="relative bg-white rounded-lg shadow-xl w-full max-w-md p-6">
              <h2 className="text-lg font-semibold text-gray-900 mb-4">
                Revoke Token
              </h2>
              <p className="text-gray-600 mb-6">
                Are you sure you want to revoke{' '}
                <strong>{tokenToRevoke.name}</strong>? Apps using this token
                will no longer be able to submit feedback.
              </p>
              <div className="flex justify-end gap-3">
                <button
                  onClick={() => setTokenToRevoke(null)}
                  className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
                >
                  Cancel
                </button>
                <button
                  onClick={handleRevoke}
                  disabled={revoking}
                  className="px-4 py-2 text-sm font-medium text-white bg-red-600 rounded-md hover:bg-red-700 disabled:opacity-50"
                >
                  {revoking ? 'Revoking...' : 'Revoke Token'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
