import { FormEvent, useState } from 'react'
import { Navigate, useNavigate } from 'react-router-dom'
import { useAuthStore } from '../store/authStore'

export default function Login() {
  const navigate = useNavigate()
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated)
  const loginMoodle = useAuthStore((state) => state.loginMoodle)

  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  if (isAuthenticated) {
    return <Navigate to="/" replace />
  }

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    setError('')
    setLoading(true)

    try {
      await loginMoodle(username.trim(), password)
      navigate('/')
    } catch (err: any) {
      setError(err?.response?.data?.message || err?.response?.data?.error || 'Не удалось войти через Moodle')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="relative flex min-h-screen items-center justify-center overflow-hidden bg-[#faf4f2] px-4 py-8 sm:px-6 lg:px-8">
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_top_left,_rgba(220,38,38,0.14),_transparent_34%),radial-gradient(circle_at_bottom_right,_rgba(249,115,22,0.12),_transparent_28%)]" />
      <div className="absolute left-[-10%] top-[-8%] h-72 w-72 rounded-full bg-rose-200/30 blur-3xl" />
      <div className="absolute bottom-[-10%] right-[-6%] h-80 w-80 rounded-full bg-orange-200/30 blur-3xl" />

      <section className="relative w-full max-w-3xl rounded-[34px] border border-[#ffd8d2] bg-white/95 p-8 shadow-[0_24px_90px_rgba(185,28,28,0.10)] sm:p-10 lg:p-12">
        <div className="text-sm font-semibold uppercase tracking-[0.32em] text-[#ff2d3d]">Moodle Access</div>
        <h1 className="mt-6 text-4xl font-bold text-text-light sm:text-5xl">Авторизация</h1>
        <p className="mt-6 max-w-2xl text-base leading-8 text-slate-600 sm:text-lg">
          Используйте логин и пароль от Moodle. Обычный локальный вход и регистрация для этого проекта отключены.
        </p>

        <form className="mt-12 space-y-7" onSubmit={handleSubmit}>
          {error && (
            <div className="rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
              {error}
            </div>
          )}

          <div>
            <label htmlFor="username" className="mb-3 block text-lg font-medium text-text-light">
              Логин Moodle
            </label>
            <input
              id="username"
              type="text"
              autoComplete="username"
              required
              value={username}
              onChange={(event) => setUsername(event.target.value)}
              className="w-full rounded-[20px] border border-[#d6e5f8] bg-[#eaf3ff] px-5 py-4 text-base text-text-light outline-none transition focus:border-[#f05252] focus:bg-white focus:ring-4 focus:ring-red-100 sm:text-lg"
            />
          </div>

          <div>
            <label htmlFor="password" className="mb-3 block text-lg font-medium text-text-light">
              Пароль
            </label>
            <input
              id="password"
              type="password"
              autoComplete="current-password"
              required
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              className="w-full rounded-[20px] border border-[#d6e5f8] bg-[#eaf3ff] px-5 py-4 text-base text-text-light outline-none transition focus:border-[#f05252] focus:bg-white focus:ring-4 focus:ring-red-100 sm:text-lg"
            />
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full rounded-[20px] bg-[#d12a2a] px-6 py-4 text-base font-semibold text-white transition hover:bg-[#b72222] disabled:cursor-not-allowed disabled:opacity-70 sm:text-lg"
          >
            {loading ? 'Проверяем Moodle...' : 'Войти через Moodle'}
          </button>
        </form>
      </section>
    </div>
  )
}
