import { createFileRoute, useParams } from '@tanstack/react-router'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState, useEffect } from 'react'
import { api } from '@/lib/api'
import type { PortalUserProfile, PortalNotificationPreferences } from '@/types'

export const Route = createFileRoute('/portal/$projectId/settings')({
  component: SettingsPage,
})

function SettingsPage() {
  const { projectId } = useParams({ from: '/portal/$projectId/settings' })
  const queryClient = useQueryClient()

  const { data: profile, isLoading } = useQuery<PortalUserProfile>({
    queryKey: ['portal', 'profile', projectId],
    queryFn: () => api.portal.getProfile(projectId) as Promise<PortalUserProfile>,
  })

  const [prefs, setPrefs] = useState<PortalNotificationPreferences>({
    status_changes: true,
    new_comments_on_my_feedback: true,
    new_comments_on_voted_feedback: false,
    weekly_digest: false,
  })

  const [saved, setSaved] = useState(false)

  useEffect(() => {
    if (profile?.notification_preferences) {
      setPrefs(profile.notification_preferences)
    }
  }, [profile])

  const mutation = useMutation({
    mutationFn: (newPrefs: PortalNotificationPreferences) =>
      api.portal.updateNotifications(projectId, newPrefs),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['portal', 'profile', projectId] })
      setSaved(true)
      setTimeout(() => setSaved(false), 2000)
    },
  })

  const handleToggle = (key: keyof PortalNotificationPreferences) => {
    const newPrefs = { ...prefs, [key]: !prefs[key] }
    setPrefs(newPrefs)
    mutation.mutate(newPrefs)
  }

  if (isLoading) {
    return (
      <div className="text-center py-12">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto" />
        <p className="mt-4 text-gray-500">Loading settings...</p>
      </div>
    )
  }

  return (
    <div>
      <h1 className="text-2xl font-bold text-gray-900 mb-6">Notification Settings</h1>

      <div className="bg-white rounded-lg shadow-sm border border-gray-200">
        <div className="p-6">
          <p className="text-sm text-gray-500 mb-6">
            Choose which notifications you&apos;d like to receive about your feedback.
          </p>

          <div className="space-y-4">
            <NotificationToggle
              label="Status changes"
              description="Get notified when the status of your feedback changes"
              checked={prefs.status_changes}
              onChange={() => handleToggle('status_changes')}
              disabled={mutation.isPending}
            />

            <NotificationToggle
              label="New comments on my feedback"
              description="Get notified when someone comments on your feedback"
              checked={prefs.new_comments_on_my_feedback}
              onChange={() => handleToggle('new_comments_on_my_feedback')}
              disabled={mutation.isPending}
            />

            <NotificationToggle
              label="New comments on voted feedback"
              description="Get notified when someone comments on feedback you voted for"
              checked={prefs.new_comments_on_voted_feedback ?? false}
              onChange={() => handleToggle('new_comments_on_voted_feedback')}
              disabled={mutation.isPending}
            />

            <NotificationToggle
              label="Weekly digest"
              description="Receive a weekly summary of activity on feedback you care about"
              checked={prefs.weekly_digest ?? false}
              onChange={() => handleToggle('weekly_digest')}
              disabled={mutation.isPending}
            />
          </div>

          {saved && (
            <p className="mt-4 text-sm text-green-600">Settings saved!</p>
          )}
        </div>
      </div>
    </div>
  )
}

interface NotificationToggleProps {
  label: string
  description: string
  checked: boolean
  onChange: () => void
  disabled?: boolean
}

function NotificationToggle({ label, description, checked, onChange, disabled }: NotificationToggleProps) {
  return (
    <div className="flex items-center justify-between py-3 border-b border-gray-100 last:border-0">
      <div>
        <p className="text-sm font-medium text-gray-900">{label}</p>
        <p className="text-sm text-gray-500">{description}</p>
      </div>
      <button
        type="button"
        role="switch"
        aria-checked={checked}
        onClick={onChange}
        disabled={disabled}
        className={`
          relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out
          focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2
          ${checked ? 'bg-primary' : 'bg-gray-200'}
          ${disabled ? 'opacity-50 cursor-not-allowed' : ''}
        `}
      >
        <span
          aria-hidden="true"
          className={`
            pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out
            ${checked ? 'translate-x-5' : 'translate-x-0'}
          `}
        />
      </button>
    </div>
  )
}
