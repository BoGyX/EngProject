import { useEffect, useMemo, useState } from 'react'
import { useAuthStore } from '../store/authStore'
import api from '../services/api'

interface User {
  id: string
  email: string
  name?: string
  role: string
  created_at: string
}

interface Course {
  id: number
  title: string
  is_published: boolean
}

interface CourseAccess {
  user_id: string
  course_id: number
  created_at: string
}

const USERS_PER_PAGE = 10

function buildCourseAccessMap(accesses: CourseAccess[]) {
  const accessMap = new Map<string, number[]>()

  accesses.forEach((access) => {
    const current = accessMap.get(access.user_id) || []
    if (!current.includes(access.course_id)) {
      current.push(access.course_id)
    }
    accessMap.set(access.user_id, current)
  })

  return accessMap
}

export default function AdminUsers() {
  const { user } = useAuthStore()
  const [users, setUsers] = useState<User[]>([])
  const [courses, setCourses] = useState<Course[]>([])
  const [accesses, setAccesses] = useState<CourseAccess[]>([])
  const [loading, setLoading] = useState(true)
  const [busyUserId, setBusyUserId] = useState<string | null>(null)
  const [currentPage, setCurrentPage] = useState(1)
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedCourseByUser, setSelectedCourseByUser] = useState<Record<string, string>>({})

  useEffect(() => {
    if (user?.role !== 'admin') return
    void loadData()
  }, [user])

  useEffect(() => {
    setCurrentPage(1)
  }, [searchQuery])

  const loadData = async () => {
    try {
      setLoading(true)

      const [usersResponse, coursesResponse, accessesResponse] = await Promise.all([
        api.get<User[]>('/admin/users'),
        api.get<Course[]>('/admin/courses'),
        api.get<CourseAccess[]>('/admin/users/course-accesses'),
      ])

      setUsers(usersResponse.data || [])
      setCourses(coursesResponse.data || [])
      setAccesses(accessesResponse.data || [])
    } catch (error) {
      console.error('Error loading admin users data:', error)
    } finally {
      setLoading(false)
    }
  }

  const courseById = useMemo(() => new Map(courses.map((course) => [course.id, course])), [courses])
  const courseAccessMap = useMemo(() => buildCourseAccessMap(accesses), [accesses])

  const filteredUsers = useMemo(() => {
    const query = searchQuery.toLowerCase().trim()
    if (!query) return users

    return users.filter((item) => {
      const email = item.email.toLowerCase()
      const name = (item.name || '').toLowerCase()
      return email.includes(query) || name.includes(query)
    })
  }, [searchQuery, users])

  const totalPages = Math.max(1, Math.ceil(filteredUsers.length / USERS_PER_PAGE))
  const startIndex = (currentPage - 1) * USERS_PER_PAGE
  const currentUsers = filteredUsers.slice(startIndex, startIndex + USERS_PER_PAGE)

  const goToPage = (page: number) => {
    setCurrentPage(page)
    window.scrollTo({ top: 0, behavior: 'smooth' })
  }

  const handleChangeRole = async (targetUserId: string, newRole: string) => {
    if (!confirm(`Изменить роль пользователя на "${newRole}"?`)) {
      return
    }

    try {
      setBusyUserId(targetUserId)
      await api.put(`/admin/users/${targetUserId}/role`, { role: newRole })
      await loadData()
    } catch (error) {
      console.error('Error changing role:', error)
      alert('Не удалось изменить роль пользователя')
    } finally {
      setBusyUserId(null)
    }
  }

  const handleGrantCourseAccess = async (targetUserId: string) => {
    const selectedCourseId = Number(selectedCourseByUser[targetUserId])
    if (!selectedCourseId) {
      alert('Сначала выберите курс')
      return
    }

    try {
      setBusyUserId(targetUserId)
      await api.post(`/admin/users/${targetUserId}/course-accesses/${selectedCourseId}`)
      setSelectedCourseByUser((current) => ({ ...current, [targetUserId]: '' }))
      await loadData()
    } catch (error) {
      console.error('Error granting course access:', error)
      alert('Не удалось выдать доступ к курсу')
    } finally {
      setBusyUserId(null)
    }
  }

  const handleRevokeCourseAccess = async (targetUserId: string, courseId: number) => {
    if (!confirm('Убрать доступ к этому курсу?')) {
      return
    }

    try {
      setBusyUserId(targetUserId)
      await api.delete(`/admin/users/${targetUserId}/course-accesses/${courseId}`)
      await loadData()
    } catch (error) {
      console.error('Error revoking course access:', error)
      alert('Не удалось убрать доступ к курсу')
    } finally {
      setBusyUserId(null)
    }
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

    for (let page = startPage; page <= endPage; page += 1) {
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
          Вперёд
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
                Роли и доступ к курсам
              </span>
            </div>
            <div>
              <h1 className="text-3xl font-bold text-text-light lg:text-4xl">Пользователи и права</h1>
              <p className="mt-2 max-w-3xl text-sm leading-7 text-slate-600">
                Здесь администратор управляет ролями пользователей и вручную выдаёт доступ к конкретным курсам. Новый пользователь
                без назначений не увидит ни одного курса.
              </p>
            </div>
          </div>

          <div className="rounded-[24px] border border-rose-100 bg-white/80 p-6 shadow-lg">
            <p className="text-xs uppercase tracking-[0.18em] text-rose-500">Сводка</p>
            <p className="mt-3 text-3xl font-bold text-text-light">{users.length}</p>
            <p className="mt-2 text-sm text-slate-600">
              {searchQuery ? `Найдено по фильтру: ${filteredUsers.length}` : 'Всего пользователей в системе'}
            </p>
            <p className="mt-4 text-sm text-slate-600">Курсов для назначения: {courses.length}</p>
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
        </div>
      </section>

      <section className="grid gap-4">
        {currentUsers.length === 0 ? (
          <div className="rounded-[28px] border border-gray-200 bg-card-light p-10 text-center shadow-md">
            <p className="text-lg font-medium text-text-light">{searchQuery ? 'Пользователи не найдены.' : 'Пользователей пока нет.'}</p>
          </div>
        ) : (
          currentUsers.map((item) => {
            const assignedCourseIds = courseAccessMap.get(item.id) || []
            const assignedCourses = assignedCourseIds
              .map((courseId) => courseById.get(courseId))
              .filter((course): course is Course => Boolean(course))
              .sort((left, right) => left.title.localeCompare(right.title, 'ru'))
            const availableCourses = courses
              .filter((course) => !assignedCourseIds.includes(course.id))
              .sort((left, right) => left.title.localeCompare(right.title, 'ru'))
            const isBusy = busyUserId === item.id

            return (
              <article key={item.id} className="rounded-[28px] border border-gray-200 bg-card-light p-6 shadow-md">
                <div className="grid gap-6 xl:grid-cols-[minmax(0,1fr)_360px]">
                  <div className="space-y-4">
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

                    <div className="space-y-1">
                      <p className="text-sm text-slate-600">{item.email}</p>
                      <p className="text-xs uppercase tracking-[0.18em] text-slate-400">
                        Регистрация: {new Date(item.created_at).toLocaleDateString('ru-RU')}
                      </p>
                    </div>

                    <div className="flex flex-wrap gap-2">
                      {item.role !== 'admin' && (
                        <button
                          onClick={() => void handleChangeRole(item.id, 'admin')}
                          disabled={isBusy}
                          className="rounded-2xl border border-orange-200 bg-orange-50 px-4 py-2 text-sm font-semibold text-orange-700 transition-colors hover:bg-orange-100 disabled:cursor-not-allowed disabled:opacity-60"
                        >
                          Сделать админом
                        </button>
                      )}
                      {item.role === 'admin' && item.id !== user?.id && (
                        <button
                          onClick={() => void handleChangeRole(item.id, 'user')}
                          disabled={isBusy}
                          className="rounded-2xl border border-slate-300 bg-slate-50 px-4 py-2 text-sm font-semibold text-slate-700 transition-colors hover:bg-slate-100 disabled:cursor-not-allowed disabled:opacity-60"
                        >
                          Убрать админа
                        </button>
                      )}
                    </div>
                  </div>

                  <div className="rounded-[24px] border border-slate-200 bg-slate-50/70 p-5">
                    <div className="mb-4 flex items-center justify-between gap-3">
                      <div>
                        <h3 className="text-lg font-semibold text-text-light">Доступ к курсам</h3>
                        <p className="text-sm text-slate-500">
                          {item.role === 'admin'
                            ? 'Администратор видит все курсы без отдельных назначений.'
                            : 'Назначайте только те курсы, которые должны появиться у пользователя.'}
                        </p>
                      </div>
                      <span className="rounded-full bg-white px-3 py-1 text-xs font-semibold text-slate-700">
                        {item.role === 'admin' ? 'Все курсы' : `${assignedCourses.length} назначено`}
                      </span>
                    </div>

                    {item.role === 'admin' ? (
                      <div className="rounded-2xl border border-rose-100 bg-white px-4 py-3 text-sm text-slate-600">
                        Для администратора отдельные назначения не требуются.
                      </div>
                    ) : (
                      <div className="space-y-4">
                        <div className="flex flex-wrap gap-2">
                          {assignedCourses.length === 0 ? (
                            <div className="rounded-2xl border border-dashed border-slate-300 bg-white px-4 py-3 text-sm text-slate-500">
                              Пока не выдан ни один курс.
                            </div>
                          ) : (
                            assignedCourses.map((course) => (
                              <div
                                key={`${item.id}-${course.id}`}
                                className="flex items-center gap-2 rounded-full border border-rose-200 bg-white px-3 py-2 text-sm text-text-light"
                              >
                                <span>{course.title}</span>
                                {!course.is_published && (
                                  <span className="rounded-full bg-slate-100 px-2 py-0.5 text-[11px] font-semibold text-slate-600">
                                    Черновик
                                  </span>
                                )}
                                <button
                                  type="button"
                                  onClick={() => void handleRevokeCourseAccess(item.id, course.id)}
                                  disabled={isBusy}
                                  className="text-sm font-semibold text-red-500 transition-colors hover:text-red-700 disabled:cursor-not-allowed disabled:opacity-60"
                                  title="Убрать доступ"
                                >
                                  ×
                                </button>
                              </div>
                            ))
                          )}
                        </div>

                        <div className="flex flex-col gap-3 sm:flex-row">
                          <select
                            value={selectedCourseByUser[item.id] || ''}
                            onChange={(event) =>
                              setSelectedCourseByUser((current) => ({
                                ...current,
                                [item.id]: event.target.value,
                              }))
                            }
                            className="flex-1 rounded-2xl border border-gray-300 px-4 py-3 focus:border-link-light focus:outline-none"
                          >
                            <option value="">Выберите курс</option>
                            {availableCourses.map((course) => (
                              <option key={course.id} value={course.id}>
                                {course.title}
                                {course.is_published ? '' : ' (черновик)'}
                              </option>
                            ))}
                          </select>

                          <button
                            type="button"
                            onClick={() => void handleGrantCourseAccess(item.id)}
                            disabled={isBusy || availableCourses.length === 0}
                            className="rounded-2xl bg-link-light px-5 py-3 text-sm font-semibold text-white transition-colors hover:bg-link-dark disabled:cursor-not-allowed disabled:opacity-60"
                          >
                            Выдать доступ
                          </button>
                        </div>
                      </div>
                    )}
                  </div>
                </div>
              </article>
            )
          })
        )}
      </section>

      <section className="space-y-4 rounded-[28px] border border-gray-200 bg-card-light p-6 shadow-md">
        {renderPagination()}
        {filteredUsers.length > 0 && (
          <p className="text-center text-sm text-slate-500">
            Показано {startIndex + 1}-{Math.min(startIndex + USERS_PER_PAGE, filteredUsers.length)} из {filteredUsers.length} | Страница{' '}
            <strong>{currentPage}</strong> из <strong>{totalPages}</strong>
          </p>
        )}
      </section>
    </div>
  )
}
