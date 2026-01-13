import { Link, useRouterState } from '@tanstack/react-router'
import { Inbox, Settings, Users, Key, UserCircle } from 'lucide-react'

interface ConsoleNavProps {
  projectId: string
}

export function ConsoleNav({ projectId }: ConsoleNavProps) {
  const routerState = useRouterState()
  const currentPath = routerState.location.pathname

  const tabs = [
    {
      name: 'Inbox',
      href: `/projects/${projectId}/console`,
      icon: Inbox,
      exact: true,
    },
    {
      name: 'Users',
      href: `/projects/${projectId}/console/users`,
      icon: UserCircle,
    },
    {
      name: 'Settings',
      href: `/projects/${projectId}/console/settings`,
      icon: Settings,
    },
    {
      name: 'Members',
      href: `/projects/${projectId}/console/members`,
      icon: Users,
    },
    {
      name: 'SDK Tokens',
      href: `/projects/${projectId}/console/sdk-tokens`,
      icon: Key,
    },
  ]

  const isActive = (tab: (typeof tabs)[0]) => {
    if (tab.exact) {
      return currentPath === tab.href || currentPath === tab.href + '/'
    }
    return currentPath.startsWith(tab.href)
  }

  return (
    <div className="border-b border-gray-200 bg-white">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <nav className="-mb-px flex space-x-8" aria-label="Tabs">
          {tabs.map((tab) => {
            const active = isActive(tab)
            const Icon = tab.icon
            return (
              <Link
                key={tab.name}
                to={tab.href}
                className={`
                  flex items-center gap-2 whitespace-nowrap py-4 px-1 border-b-2 text-sm font-medium
                  ${
                    active
                      ? 'border-primary text-primary'
                      : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                  }
                `}
              >
                <Icon className="w-4 h-4" />
                {tab.name}
              </Link>
            )
          })}
        </nav>
      </div>
    </div>
  )
}
