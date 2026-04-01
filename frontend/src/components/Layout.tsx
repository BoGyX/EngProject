import { Outlet, Link, useNavigate } from 'react-router-dom'
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
    <div className="min-h-screen bg-bg-light">
      <nav className="bg-card-light shadow-md border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16">
            <div className="flex">
              <Link to="/courses" className="flex items-center px-2 py-2 text-xl font-bold text-logo-bright hover:text-logo-dark transition-colors">
                English Learning
              </Link>
              <div className="hidden sm:ml-6 sm:flex sm:space-x-8">
                <Link
                  to="/courses"
                  className="border-transparent text-text-light hover:text-link-light hover:border-link-light inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium transition-colors"
                >
                  Курсы
                </Link>
                <Link
                  to="/progress"
                  className="border-transparent text-text-light hover:text-link-light hover:border-link-light inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium transition-colors"
                >
                  Прогресс
                </Link>
                <Link
                  to="/vocabulary"
                  className="border-transparent text-text-light hover:text-link-light hover:border-link-light inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium transition-colors"
                >
                  Мой словарь
                </Link>
                <Link
                  to="/reader"
                  className="border-transparent text-text-light hover:text-link-light hover:border-link-light inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium transition-colors"
                >
                  Ридер
                </Link>
                {user?.role === 'admin' && (
                  <Link
                    to="/admin"
                    className="border-transparent text-text-light hover:text-link-light hover:border-link-light inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium transition-colors"
                  >
                    Админка
                  </Link>
                )}
              </div>
            </div>
            <div className="flex items-center space-x-4">
              <span className="text-text-light text-sm">{user?.name || user?.email}</span>
              <button
                onClick={handleLogout}
                className="bg-logo-bright hover:bg-logo-dark text-white px-4 py-2 rounded-lg text-sm font-medium transition-colors"
              >
                Выйти
              </button>
            </div>
          </div>
        </div>
      </nav>
      <main className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        <Outlet />
      </main>
      <WordTranslator />
    </div>
  )
}
