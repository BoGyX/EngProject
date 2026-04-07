import { useEffect, useMemo, useState } from 'react'
import { config } from '../config'
import { adminService, Course, Podcast } from '../services/adminService'
import { useAuthStore } from '../store/authStore'

const ALL_COURSES_VALUE = 'all'

interface PodcastGroup {
  key: string
  title: string
  subtitle?: string
  items: Podcast[]
}

export default function AdminPodcasts() {
  const { user } = useAuthStore()
  const [courses, setCourses] = useState<Course[]>([])
  const [podcasts, setPodcasts] = useState<Podcast[]>([])
  const [selectedCourseId, setSelectedCourseId] = useState(ALL_COURSES_VALUE)
  const [loadingCourses, setLoadingCourses] = useState(true)
  const [loadingPodcasts, setLoadingPodcasts] = useState(true)

  useEffect(() => {
    if (user?.role !== 'admin') return
    void loadCourses()
  }, [user])

  useEffect(() => {
    if (user?.role !== 'admin') return
    const courseId = selectedCourseId === ALL_COURSES_VALUE ? undefined : Number(selectedCourseId)
    void loadPodcasts(courseId)
  }, [selectedCourseId, user])

  const loadCourses = async () => {
    try {
      setLoadingCourses(true)
      const data = await adminService.getAllCourses()
      setCourses(data || [])
    } catch (error) {
      console.error('Error loading podcast courses:', error)
      setCourses([])
    } finally {
      setLoadingCourses(false)
    }
  }

  const loadPodcasts = async (courseId?: number) => {
    try {
      setLoadingPodcasts(true)
      const data = await adminService.getPodcasts(courseId)
      setPodcasts(data || [])
    } catch (error) {
      console.error('Error loading podcasts:', error)
      setPodcasts([])
    } finally {
      setLoadingPodcasts(false)
    }
  }

  const groupedPodcasts = useMemo<PodcastGroup[]>(() => {
    const groups = new Map<string, PodcastGroup>()

    for (const podcast of podcasts) {
      const key = podcast.course_id > 0 ? String(podcast.course_id) : 'uncategorized'
      const title = podcast.course_title || 'Без курса'
      const subtitle = podcast.course_slug ? `/deck/${podcast.course_slug}` : 'Reader text without linked course'

      if (!groups.has(key)) {
        groups.set(key, {
          key,
          title,
          subtitle,
          items: [],
        })
      }

      groups.get(key)?.items.push(podcast)
    }

    return Array.from(groups.values())
  }, [podcasts])

  const selectedCourseTitle = useMemo(() => {
    if (selectedCourseId === ALL_COURSES_VALUE) {
      return 'Все курсы'
    }

    const matchedCourse = courses.find((course) => course.id === Number(selectedCourseId))
    return matchedCourse?.title || 'Выбранный курс'
  }, [courses, selectedCourseId])

  if (user?.role !== 'admin') {
    return <div className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-red-700">У вас нет доступа к этой странице</div>
  }

  if (loadingCourses || loadingPodcasts) {
    return <div className="py-8 text-center text-text-light">Загрузка подкастов...</div>
  }

  return (
    <div className="space-y-6">
      <section className="overflow-hidden rounded-[28px] border border-rose-100 bg-gradient-to-br from-white via-rose-50 to-orange-50 shadow-xl">
        <div className="grid gap-6 p-6 lg:grid-cols-[minmax(0,1fr)_260px] lg:p-8">
          <div className="space-y-4">
            <div className="flex flex-wrap items-center gap-3">
              <span className="rounded-full bg-white/80 px-3 py-1 text-xs font-semibold uppercase tracking-[0.22em] text-link-light shadow-sm">
                Admin Podcasts
              </span>
              <span className="rounded-full bg-rose-100 px-3 py-1 text-xs font-semibold text-rose-800">
                Reader audio library
              </span>
            </div>
            <div>
              <h1 className="text-3xl font-bold text-text-light lg:text-4xl">Подкасты по курсам</h1>
              <p className="mt-2 max-w-3xl text-sm leading-7 text-slate-600">
                Здесь собраны все mp3, которые прикреплены к reader-текстам. Можно выбрать курс и сразу прослушать все
                его аудиозаписи прямо из админ-панели.
              </p>
            </div>
          </div>

          <div className="rounded-[24px] border border-rose-100 bg-white/80 p-6 shadow-lg">
            <p className="text-xs uppercase tracking-[0.18em] text-rose-500">Текущий фильтр</p>
            <p className="mt-3 text-2xl font-bold text-text-light">{selectedCourseTitle}</p>
            <p className="mt-2 text-sm text-slate-600">Найдено аудио: {podcasts.length}</p>
          </div>
        </div>
      </section>

      <section className="rounded-[28px] border border-gray-200 bg-card-light p-6 shadow-md">
        <label className="mb-2 block text-sm font-semibold text-text-light">Курс</label>
        <select
          value={selectedCourseId}
          onChange={(event) => setSelectedCourseId(event.target.value)}
          className="w-full rounded-2xl border border-gray-300 px-4 py-3 focus:border-link-light focus:outline-none"
        >
          <option value={ALL_COURSES_VALUE}>Все курсы</option>
          {courses.map((course) => (
            <option key={course.id} value={course.id}>
              {course.title}
            </option>
          ))}
        </select>
      </section>

      {groupedPodcasts.length === 0 ? (
        <section className="rounded-[28px] border border-gray-200 bg-card-light p-10 text-center shadow-md">
          <p className="text-lg font-medium text-text-light">Для выбранного курса пока нет reader-аудио.</p>
          <p className="mt-2 text-sm text-slate-500">Добавьте mp3 при создании текста в reader, и запись сразу появится здесь.</p>
        </section>
      ) : (
        <div className="space-y-6">
          {groupedPodcasts.map((group) => (
            <section key={group.key} className="rounded-[28px] border border-gray-200 bg-card-light p-6 shadow-md">
              <div className="flex flex-col gap-3 border-b border-gray-200 pb-5 lg:flex-row lg:items-end lg:justify-between">
                <div>
                  <h2 className="text-2xl font-semibold text-text-light">{group.title}</h2>
                  {group.subtitle && <p className="mt-1 text-sm text-slate-500">{group.subtitle}</p>}
                </div>
                <span className="rounded-full bg-rose-50 px-3 py-1 text-xs font-semibold uppercase tracking-[0.18em] text-rose-700">
                  Аудио: {group.items.length}
                </span>
              </div>

              <div className="mt-6 grid gap-4">
                {group.items.map((podcast) => (
                  <article key={podcast.id} className="rounded-[24px] border border-slate-200 bg-white p-5 shadow-sm">
                    <div className="flex flex-col gap-4 xl:flex-row xl:items-start xl:justify-between">
                      <div className="space-y-2">
                        <div className="flex flex-wrap items-center gap-2">
                          <span className="rounded-full bg-emerald-100 px-2.5 py-1 text-[11px] font-semibold uppercase tracking-[0.14em] text-emerald-700">
                            MP3
                          </span>
                          <span className="text-xs uppercase tracking-[0.18em] text-slate-400">
                            {new Date(podcast.created_at).toLocaleDateString('ru-RU', {
                              year: 'numeric',
                              month: 'long',
                              day: 'numeric',
                            })}
                          </span>
                        </div>
                        <h3 className="text-xl font-semibold text-text-light">{podcast.title}</h3>
                        <p className="text-sm text-slate-600">
                          Автор: <span className="font-medium text-text-light">{podcast.user_name || podcast.user_email}</span>
                        </p>
                      </div>

                      <audio controls preload="none" className="w-full max-w-xl" src={config.getFullUrl(podcast.audio_url)}>
                        Ваш браузер не поддерживает аудио.
                      </audio>
                    </div>
                  </article>
                ))}
              </div>
            </section>
          ))}
        </div>
      )}
    </div>
  )
}
