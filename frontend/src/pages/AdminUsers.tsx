import { useEffect, useMemo, useState } from 'react'
import { adminService, Course } from '../services/adminService'
import { useAuthStore } from '../store/authStore'
import api from '../services/api'

interface User {
  id: string
  email: string
  name?: string
  role: string
  created_at: string
}

const USERS_PER_PAGE = 10

export default function AdminUsers() {
  const { user } = useAuthStore()
  const [users, setUsers] = useState<User[]>([])
  const [loading, setLoading] = useState(true)
  const [currentPage, setCurrentPage] = useState(1)
  const [searchQuery, setSearchQuery] = useState('')
  const [availableCourses, setAvailableCourses] = useState<Course[]>([])
  const [expandedAccessUserId, setExpandedAccessUserId] = useState<string | null>(null)
  const [userCourseAccess, setUserCourseAccess] = useState<Record<string, number[]>>({})
  const [loadingAccessUserId, setLoadingAccessUserId] = useState<string | null>(null)
  const [savingRoleUserId, setSavingRoleUserId] = useState<string | null>(null)
  const [savingAccessUserId, setSavingAccessUserId] = useState<string | null>(null)

  useEffect(() => {
    if (user?.role !== 'admin') return
    void loadUsersAndCourses()
  }, [user])

  useEffect(() => {
    setCurrentPage(1)
  }, [searchQuery])

  const loadUsersAndCourses = async () => {
    try {
      setLoading(true)
      const [usersResponse, courses] = await Promise.all([api.get<User[]>('/users'), adminService.getAllCourses()])
      setUsers(usersResponse.data)
      setAvailableCourses(courses || [])
    } catch (error) {
      console.error('Error loading users:', error)
      setAvailableCourses([])
    } finally {
      setLoading(false)
    }
  }

  const handleChangeRole = async (targetUserId: string, newRole: string) => {
    if (!confirm(`Изменить роль пользователя на "${newRole}"?`)) return
    try {
      setSavingRoleUserId(targetUserId)
      const updatedUser = await adminService.updateUserRole(targetUserId, newRole as 'admin' | 'user')
      setUsers((current) => current.map((item) => (item.id === targetUserId ? { ...item, role: updatedUser.role } : item)))
    } catch (error) {
      console.error('Error changing role:', error)
      alert('Не удалось изменить роль пользователя.')
    } finally {
      setSavingRoleUserId(null)
    }
  }

  const loadUserCourseAccess = async (targetUserId: string) => {
    try {
      setLoadingAccessUserId(targetUserId)
      const response = await adminService.getUserCourseAccess(targetUserId)
      setUserCourseAccess((current) => ({
        ...current,
        [targetUserId]: response.course_ids || [],
      }))
    } catch (error) {
      console.error('Error loading user course access:', error)
      alert('Не удалось загрузить доступы к курсам.')
    } finally {
      setLoadingAccessUserId(null)
    }
  }

  const toggleAccessEditor = async (targetUserId: string) => {
    if (expandedAccessUserId === targetUserId) {
      setExpandedAccessUserId(null)
      return
    }

    setExpandedAccessUserId(targetUserId)
    if (!(targetUserId in userCourseAccess)) {
      await loadUserCourseAccess(targetUserId)
    }
  }

  const handleToggleCourse = (targetUserId: string, courseId: number) => {
    setUserCourseAccess((current) => {
      const selected = new Set(current[targetUserId] || [])
      if (selected.has(courseId)) {
        selected.delete(courseId)
      } else {
        selected.add(courseId)
      }

      return {
        ...current,
        [targetUserId]: Array.from(selected),
      }
    })
  }

  const handleSaveCourseAccess = async (targetUserId: string) => {
    try {
      setSavingAccessUserId(targetUserId)
      const response = await adminService.updateUserCourseAccess(targetUserId, userCourseAccess[targetUserId] || [])
      setUserCourseAccess((current) => ({
        ...current,
        [targetUserId]: response.course_ids || [],
      }))
    } catch (error) {
      console.error('Error saving user course access:', error)
      alert('Не удалось сохранить доступ к курсам.')
    } finally {
      setSavingAccessUserId(null)
    }
  }

  const filteredUsers = useMemo(() => {
    const query = searchQuery.toLowerCase().trim()
    if (!query) return users

    return users.filter((item) => {
      const email = item.email.toLowerCase()
      const name = (item.name || '').toLowerCase()
      return email.includes(query) || name.includes(query)
    })
  }, [searchQuery, users])

  const sortedCourses = useMemo(
    () => [...availableCourses].sort((left, right) => left.title.localeCompare(right.title, 'ru-RU')),
    [availableCourses]
  )

  const totalPages = Math.ceil(filteredUsers.length / USERS_PER_PAGE)
  const startIndex = (currentPage - 1) * USERS_PER_PAGE
  const endIndex = startIndex + USERS_PER_PAGE
  const currentUsers = filteredUsers.slice(startIndex, endIndex)

  const goToPage = (page: number) => {
    setCurrentPage(page)
    window.scrollTo({ top: 0, behavior: 'smooth' })
  }

  const renderPagination = () => {
    if (totalPages <= 1) return null

    const pages = []
    const maxVisiblePages = 5

    let startPage = Math.max(1, currentPage - Math.floor(maxVisiblePages / 2))
    let endPage = Math.min(totalPages, startPage + maxVisiblePages - 1)

    if (endPage - startPage < maxVisiblePages - 1) {
      startPage = Math.max(1, endPage - maxVisiblePages + 1)
    }

    if (startPage > 1) {
      pages.push(
        <button
          key="first"
          onClick={() => goToPage(1)}
          className="rounded-2xl border border-slate-300 px-4 py-2 text-sm font-semibold text-text-light transition-colors hover:border-rose-200 hover:bg-rose-50"
        >
          1
        </button>
      )
      if (startPage > 2) {
        pages.push(
          <span key="dots-start" className="px-2 text-slate-400">
            ...
          </span>
        )
      }
    }

    for (let page = startPage; page <= endPage; page++) {
      pages.push(
        <button
          key={page}
          onClick={() => goToPage(page)}
          className={`rounded-2xl border px-4 py-2 text-sm font-semibold transition-colors ${
            currentPage === page
              ? 'border-link-light bg-link-light text-white'
              : 'border-slate-300 text-text-light hover:border-rose-200 hover:bg-rose-50'
          }`}
        >
          {page}
        </button>
      )
    }

    if (endPage < totalPages) {
      if (endPage < totalPages - 1) {
        pages.push(
          <span key="dots-end" className="px-2 text-slate-400">
            ...
          </span>
        )
      }
      pages.push(
        <button
          key="last"
          onClick={() => goToPage(totalPages)}
          className="rounded-2xl border border-slate-300 px-4 py-2 text-sm font-semibold text-text-light transition-colors hover:border-rose-200 hover:bg-rose-50"
        >
          {totalPages}
        </button>
      )
    }

    return (
      <div className="flex flex-wrap items-center justify-center gap-2">
        <button
          onClick={() => goToPage(currentPage - 1)}
          disabled={currentPage === 1}
          className="rounded-2xl border border-slate-300 px-4 py-2 text-sm font-semibold text-text-light transition-colors hover:border-rose-200 hover:bg-rose-50 disabled:cursor-not-allowed disabled:opacity-50"
        >
          Назад
        </button>
        {pages}
        <button
          onClick={() => goToPage(currentPage + 1)}
          disabled={currentPage === totalPages}
          className="rounded-2xl border border-slate-300 px-4 py-2 text-sm font-semibold text-text-light transition-colors hover:border-rose-200 hover:bg-rose-50 disabled:cursor-not-allowed disabled:opacity-50"
        >
          Вперед
        </button>
      </div>
    )
  }

  if (user?.role !== 'admin') {
    return <div className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-red-700">У вас нет доступа к этой странице</div>
  }

  if (loading) {
    return <div className="py-8 text-center text-text-light">Загрузка пользователей...</div>
  }

  return (
    <div className="space-y-6">
      <section className="overflow-hidden rounded-[28px] border border-rose-100 bg-gradient-to-br from-white via-rose-50 to-orange-50 shadow-xl">
        <div className="grid gap-6 p-6 lg:grid-cols-[minmax(0,1fr)_260px] lg:p-8">
          <div className="space-y-4">
            <div className="flex flex-wrap items-center gap-3">
              <span className="rounded-full bg-white/80 px-3 py-1 text-xs font-semibold uppercase tracking-[0.22em] text-link-light shadow-sm">
                Admin Users
              </span>
              <span className="rounded-full bg-rose-100 px-3 py-1 text-xs font-semibold text-rose-800">
                Поиск и роли
              </span>
            </div>
            <div>
              <h1 className="text-3xl font-bold text-text-light lg:text-4xl">Управление пользователями</h1>
              <p className="mt-2 max-w-3xl text-sm leading-7 text-slate-600">
                Здесь собраны все пользователи платформы. Можно быстро найти человека по имени или email и отследить текущую роль.
              </p>
            </div>
          </div>

          <div className="rounded-[24px] border border-rose-100 bg-white/80 p-6 shadow-lg">
            <p className="text-xs uppercase tracking-[0.18em] text-rose-500">Сводка</p>
            <p className="mt-3 text-3xl font-bold text-text-light">{users.length}</p>
            <p className="mt-2 text-sm text-slate-600">{searchQuery ? `Найдено по фильтру: ${filteredUsers.length}` : 'Всего пользователей в системе'}</p>
          </div>
        </div>
      </section>

      <section className="rounded-[28px] border border-gray-200 bg-card-light p-6 shadow-md">
        <div className="relative">
          <input
            type="text"
            placeholder="Поиск по имени или email..."
            value={searchQuery}
            onChange={(event) => setSearchQuery(event.target.value)}
            className="w-full rounded-2xl border border-gray-300 px-4 py-3 pl-12 focus:border-link-light focus:outline-none"
          />
          <div className="absolute left-4 top-1/2 -translate-y-1/2 transform text-slate-400">
            <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
          </div>
          {searchQuery && (
            <button
              onClick={() => setSearchQuery('')}
              className="absolute right-4 top-1/2 -translate-y-1/2 transform text-slate-400 transition-colors hover:text-slate-600"
              title="Очистить поиск"
            >
              ×
            </button>
          )}
        </div>
      </section>

      <section className="grid gap-4">
        {currentUsers.length === 0 ? (
          <div className="rounded-[28px] border border-gray-200 bg-card-light p-10 text-center shadow-md">
            <p className="text-lg font-medium text-text-light">{searchQuery ? 'Пользователи не найдены.' : 'Пользователей пока нет.'}</p>
          </div>
        ) : (
          currentUsers.map((item) => (
            <article key={item.id} className="rounded-[28px] border border-gray-200 bg-card-light p-6 shadow-md">
              <div className="flex flex-col gap-5 lg:flex-row lg:items-center lg:justify-between">
                <div className="space-y-2">
                  <div className="flex flex-wrap items-center gap-2">
                    <h2 className="text-xl font-semibold text-text-light">{item.name || 'Без имени'}</h2>
                    <span
                      className={`rounded-full px-3 py-1 text-xs font-semibold ${
                        item.role === 'admin' ? 'bg-rose-100 text-rose-800' : 'bg-slate-100 text-slate-700'
                      }`}
                    >
                      {item.role === 'admin' ? 'Администратор' : 'Пользователь'}
                    </span>
                  </div>
                  <p className="text-sm text-slate-600">{item.email}</p>
                  <p className="text-xs uppercase tracking-[0.18em] text-slate-400">
                    Регистрация: {new Date(item.created_at).toLocaleDateString('ru-RU')}
                  </p>
                  <p className="text-sm text-slate-500">
                    Доступно курсов: <span className="font-semibold text-text-light">{(userCourseAccess[item.id] || []).length}</span>
                  </p>
                </div>

                <div className="flex flex-wrap gap-2">
                  {item.role !== 'admin' && (
                    <button
                      onClick={() => handleChangeRole(item.id, 'admin')}
                      disabled={savingRoleUserId === item.id}
                      className="rounded-2xl border border-orange-200 bg-orange-50 px-4 py-2 text-sm font-semibold text-orange-700 transition-colors hover:bg-orange-100 disabled:cursor-not-allowed disabled:opacity-50"
                    >
                      {savingRoleUserId === item.id ? 'Сохраняю...' : 'Сделать админом'}
                    </button>
                  )}
                  {item.role === 'admin' && item.id !== user?.id && (
                    <button
                      onClick={() => handleChangeRole(item.id, 'user')}
                      disabled={savingRoleUserId === item.id}
                      className="rounded-2xl border border-slate-300 bg-slate-50 px-4 py-2 text-sm font-semibold text-slate-700 transition-colors hover:bg-slate-100 disabled:cursor-not-allowed disabled:opacity-50"
                    >
                      {savingRoleUserId === item.id ? 'Сохраняю...' : 'Убрать админа'}
                    </button>
                  )}
                  <button
                    onClick={() => void toggleAccessEditor(item.id)}
                    className="rounded-2xl border border-rose-200 bg-rose-50 px-4 py-2 text-sm font-semibold text-rose-700 transition-colors hover:bg-rose-100"
                  >
                    {expandedAccessUserId === item.id ? 'Скрыть курсы' : 'Доступ к курсам'}
                  </button>
                </div>
              </div>

              {expandedAccessUserId === item.id && (
                <div className="mt-5 rounded-[24px] border border-rose-100 bg-rose-50/60 p-5">
                  <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                    <div>
                      <h3 className="text-lg font-semibold text-text-light">Назначение курсов</h3>
                      <p className="text-sm text-slate-600">Пользователь увидит только опубликованные курсы из этого списка.</p>
                    </div>
                    <button
                      onClick={() => void handleSaveCourseAccess(item.id)}
                      disabled={savingAccessUserId === item.id || loadingAccessUserId === item.id}
                      className="rounded-2xl bg-link-light px-4 py-2 text-sm font-semibold text-white transition-colors hover:bg-link-dark disabled:cursor-not-allowed disabled:opacity-50"
                    >
                      {savingAccessUserId === item.id ? 'Сохраняю...' : 'Сохранить доступ'}
                    </button>
                  </div>

                  {loadingAccessUserId === item.id ? (
                    <p className="mt-4 text-sm text-slate-500">Загрузка курсов...</p>
                  ) : sortedCourses.length === 0 ? (
                    <p className="mt-4 text-sm text-slate-500">Курсов пока нет.</p>
                  ) : (
                    <div className="mt-4 grid gap-3 md:grid-cols-2">
                      {sortedCourses.map((course) => {
                        const selected = (userCourseAccess[item.id] || []).includes(course.id)

                        return (
                          <label
                            key={course.id}
                            className={`flex cursor-pointer items-start gap-3 rounded-2xl border px-4 py-3 transition-colors ${
                              selected ? 'border-rose-200 bg-white' : 'border-slate-200 bg-white/80'
                            }`}
                          >
                            <input
                              type="checkbox"
                              checked={selected}
                              onChange={() => handleToggleCourse(item.id, course.id)}
                              className="mt-1 h-4 w-4 rounded border-slate-300 text-link-light focus:ring-link-light"
                            />
                            <div>
                              <p className="font-semibold text-text-light">{course.title}</p>
                              <p className="mt-1 text-xs uppercase tracking-[0.16em] text-slate-400">
                                {course.is_published ? 'Опубликован' : 'Черновик'}
                              </p>
                            </div>
                          </label>
                        )
                      })}
                    </div>
                  )}
                </div>
              )}
            </article>
          ))
        )}
      </section>

      <section className="space-y-4 rounded-[28px] border border-gray-200 bg-card-light p-6 shadow-md">
        {renderPagination()}
        {filteredUsers.length > 0 && (
          <p className="text-center text-sm text-slate-500">
            Показано {startIndex + 1}-{Math.min(endIndex, filteredUsers.length)} из {filteredUsers.length} | Страница{' '}
            <strong>{currentPage}</strong> из <strong>{totalPages}</strong>
          </p>
        )}
      </section>
    </div>
  )
}
