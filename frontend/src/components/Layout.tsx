import { useEffect, useMemo, useState } from 'react'
import { Link, Outlet, useLocation, useNavigate } from 'react-router-dom'
import { useAuthStore } from '../store/authStore'
import WordTranslator from './WordTranslator'

function isIframeView() {
  try {
    return window.self !== window.top
  } catch {
    return true
  }
}

export default function Layout() {
  const { user, logout, isAuthenticated } = useAuthStore()
  const navigate = useNavigate()
  const location = useLocation()
  const embedded = isIframeView()
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false)

  const navigationItems = useMemo(() => {
    const items = [
      { to: '/courses', label: 'Курсы' },
      { to: '/progress', label: 'Прогресс' },
      { to: '/vocabulary', label: 'Мой словарь' },
      { to: '/reader', label: 'Ридер' },
    ]

    if (user?.role === 'admin') {
      items.push(
        { to: '/admin', label: 'Админка' },
        { to: '/admin/podcasts', label: 'Подкасты' },
      )
    }

    return items
  }, [user?.role])

  useEffect(() => {
    setMobileMenuOpen(false)
  }, [location.pathname])

  const handleLogout = async () => {
    await logout()
    navigate('/login')
  }

  if (!isAuthenticated) return null

  return (
    <div className="app-shell min-h-screen bg-bg-light">
      {!embedded && (
        <nav className="border-b border-gray-200 bg-card-light shadow-md">
          <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
            <div className="flex min-h-16 items-center justify-between gap-3 py-3 sm:h-16 sm:py-0">
              <div className="flex min-w-0 items-center gap-2 sm:gap-4">
                <Link
                  to="/courses"
                  className="truncate px-1 py-2 text-lg font-bold text-logo-bright transition-colors hover:text-logo-dark sm:px-2 sm:text-xl"
                >
                  English Learning
                </Link>

                <div className="hidden sm:ml-6 sm:flex sm:space-x-8">
                  {navigationItems.map((item) => (
                    <Link
                      key={item.to}
                      to={item.to}
                      className="inline-flex items-center border-b-2 border-transparent px-1 pt-1 text-sm font-medium text-text-light transition-colors hover:border-link-light hover:text-link-light"
                    >
                      {item.label}
                    </Link>
                  ))}
                </div>
              </div>

              <div className="hidden items-center space-x-4 sm:flex">
                <span className="text-sm text-text-light">{user?.name || user?.email}</span>
                <button
                  onClick={handleLogout}
                  className="rounded-lg bg-logo-bright px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-logo-dark"
                >
                  Выйти
                </button>
              </div>

              <div className="flex items-center gap-2 sm:hidden">
                <span className="max-w-[9rem] truncate text-xs text-slate-500">{user?.name || user?.email}</span>
                <button
                  type="button"
                  onClick={() => setMobileMenuOpen((current) => !current)}
                  className="rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm font-semibold text-text-light transition-colors hover:border-link-light hover:text-link-light"
                >
                  {mobileMenuOpen ? 'Закрыть' : 'Меню'}
                </button>
              </div>
            </div>

            {mobileMenuOpen && (
              <div className="border-t border-gray-100 py-3 sm:hidden">
                <div className="grid gap-2">
                  {navigationItems.map((item) => (
                    <Link
                      key={item.to}
                      to={item.to}
                      className="rounded-2xl bg-slate-50 px-4 py-3 text-sm font-semibold text-text-light transition-colors hover:bg-rose-50 hover:text-link-light"
                    >
                      {item.label}
                    </Link>
                  ))}
                </div>

                <div className="mt-3 flex items-center justify-between gap-3 rounded-2xl bg-slate-50 p-3">
                  <span className="min-w-0 truncate text-sm text-slate-600">{user?.name || user?.email}</span>
                  <button
                    onClick={handleLogout}
                    className="shrink-0 rounded-xl bg-logo-bright px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-logo-dark"
                  >
                    Выйти
                  </button>
                </div>
              </div>
            )}
          </div>
        </nav>
      )}

      <main className={`mx-auto w-full max-w-7xl flex-1 ${embedded ? 'px-0 py-0' : 'px-4 py-4 sm:px-6 sm:py-6 lg:px-8'}`}>
        <Outlet />
      </main>

      <WordTranslator />
    </div>
  )
}
