import { Link, Outlet, useNavigate } from 'react-router-dom'
import { useAuthStore } from '../store/authStore'
import WordTranslator from './WordTranslator'

export default function Layout() {
  const { user, logout, isAuthenticated } = useAuthStore()
  const navigate = useNavigate()

  const handleLogout = async () => {
    await logout()
    navigate('/login')
  }

  if (!isAuthenticated) return null

  return (
    <div className="app-shell min-h-screen bg-bg-light">
      <nav className="border-b border-gray-200 bg-card-light shadow-md">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          <div className="flex h-16 justify-between">
            <div className="flex">
              <Link to="/courses" className="flex items-center px-2 py-2 text-xl font-bold text-logo-bright transition-colors hover:text-logo-dark">
                English Learning
              </Link>

              <div className="hidden sm:ml-6 sm:flex sm:space-x-8">
                <Link
                  to="/courses"
                  className="inline-flex items-center border-b-2 border-transparent px-1 pt-1 text-sm font-medium text-text-light transition-colors hover:border-link-light hover:text-link-light"
                >
                  Курсы
                </Link>
                <Link
                  to="/progress"
                  className="inline-flex items-center border-b-2 border-transparent px-1 pt-1 text-sm font-medium text-text-light transition-colors hover:border-link-light hover:text-link-light"
                >
                  Прогресс
                </Link>
                <Link
                  to="/vocabulary"
                  className="inline-flex items-center border-b-2 border-transparent px-1 pt-1 text-sm font-medium text-text-light transition-colors hover:border-link-light hover:text-link-light"
                >
                  Мой словарь
                </Link>
                <Link
                  to="/reader"
                  className="inline-flex items-center border-b-2 border-transparent px-1 pt-1 text-sm font-medium text-text-light transition-colors hover:border-link-light hover:text-link-light"
                >
                  Ридер
                </Link>
                {user?.role === 'admin' && (
                  <>
                    <Link
                      to="/admin"
                      className="inline-flex items-center border-b-2 border-transparent px-1 pt-1 text-sm font-medium text-text-light transition-colors hover:border-link-light hover:text-link-light"
                    >
                      Админка
                    </Link>
                    <Link
                      to="/admin/podcasts"
                      className="inline-flex items-center border-b-2 border-transparent px-1 pt-1 text-sm font-medium text-text-light transition-colors hover:border-link-light hover:text-link-light"
                    >
                      Подкасты
                    </Link>
                  </>
                )}
              </div>
            </div>

            <div className="flex items-center space-x-4">
              <span className="text-sm text-text-light">{user?.name || user?.email}</span>
              <button
                onClick={handleLogout}
                className="rounded-lg bg-logo-bright px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-logo-dark"
              >
                Выйти
              </button>
            </div>
          </div>
        </div>
      </nav>

      <main className="mx-auto flex-1 max-w-7xl py-6 sm:px-6 lg:px-8">
        <Outlet />
      </main>

      <WordTranslator />
    </div>
  )
}
