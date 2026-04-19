import { useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { config } from '../config'
import { studyService, Course, UserCourse } from '../services/studyService'
import { useAuthStore } from '../store/authStore'
import { buildLatestUserCourseMap, normalizeCourseProgress } from '../utils/courseProgress'

function hasUsableAccessToken(token?: string | null) {
  if (!token) {
    return false
  }

  const tokenParts = token.split('.')
  if (tokenParts.length < 2) {
    return true
  }

  try {
    const normalizedPayload = tokenParts[1].replace(/-/g, '+').replace(/_/g, '/')
    const decodedPayload = JSON.parse(atob(normalizedPayload))
    if (typeof decodedPayload.exp === 'number') {
      return decodedPayload.exp * 1000 > Date.now() + 5000
    }
  } catch (error) {
    console.warn('Failed to parse access token expiration:', error)
  }

  return true
}

interface CourseCardProps {
  course: Course
  userCourse?: UserCourse
}

function CourseCard({ course, userCourse }: CourseCardProps) {
  const progress = normalizeCourseProgress(userCourse?.progress_percentage)

  return (
    <Link
      to={`/courses/${course.id}`}
      className="overflow-hidden rounded-lg border border-gray-200 bg-card-light shadow-md transition-all hover:border-link-light hover:shadow-lg"
    >
      {course.image_url && (
        <img
          src={config.getFullUrl(course.image_url)}
          alt={course.title}
          className="h-48 w-full object-cover"
          onError={(event) => {
            ;(event.target as HTMLImageElement).style.display = 'none'
          }}
        />
      )}

      <div className="p-6">
        <div className="mb-3 flex flex-wrap items-center gap-2">
          <h2 className="text-xl font-semibold text-text-light">{course.title}</h2>
          {progress >= 100 ? (
            <span className="rounded-full bg-green-100 px-2 py-1 text-xs font-semibold text-green-800">100% завершено</span>
          ) : progress > 0 ? (
            <span className="rounded-full bg-blue-100 px-2 py-1 text-xs font-semibold text-blue-800">{progress}% пройдено</span>
          ) : (
            <span className="rounded-full bg-slate-100 px-2 py-1 text-xs font-semibold text-slate-700">Доступ открыт</span>
          )}
        </div>

        {course.description && <p className="mb-4 line-clamp-2 text-sm text-text-light">{course.description}</p>}

        {userCourse && (
          <div className="mb-4 space-y-2">
            <div className="flex items-center justify-between text-xs uppercase tracking-[0.16em] text-gray-500">
              <span>Прогресс</span>
              <span>{progress}%</span>
            </div>
            <div className="h-2 rounded-full bg-gray-200">
              <div
                className={`h-2 rounded-full transition-all ${progress >= 100 ? 'bg-green-500' : 'bg-link-light'}`}
                style={{ width: `${progress}%` }}
              />
            </div>
          </div>
        )}

        <div className="flex items-center justify-between">
          <span className="text-xs text-gray-500">{new Date(course.created_at).toLocaleDateString('ru-RU')}</span>
          <span className="rounded-full bg-green-100 px-2 py-1 text-xs font-semibold text-green-800">Опубликован</span>
        </div>
      </div>
    </Link>
  )
}

export default function Courses() {
  const { user, accessToken, isAuthenticated } = useAuthStore()
  const [courses, setCourses] = useState<Course[]>([])
  const [userCourses, setUserCourses] = useState<UserCourse[]>([])
  const [loading, setLoading] = useState(true)

  const canUseProtectedStudyActions = Boolean(isAuthenticated && user?.id && hasUsableAccessToken(accessToken))
  const latestUserCourseMap = useMemo(() => buildLatestUserCourseMap(userCourses), [userCourses])

  useEffect(() => {
    void loadCourses()
  }, [user?.id, canUseProtectedStudyActions])

  const loadCourses = async () => {
    try {
      setLoading(true)

      const [courseList, userCourseList] = await Promise.all([
        studyService.getCourses(),
        canUseProtectedStudyActions && user?.id ? studyService.getUserCourses(user.id).catch(() => []) : Promise.resolve([]),
      ])

      setCourses(courseList || [])
      setUserCourses(userCourseList || [])
    } catch (error) {
      console.error('Error loading courses:', error)
      setCourses([])
      setUserCourses([])
    } finally {
      setLoading(false)
    }
  }

  if (loading) {
    return <div className="py-8 text-center text-text-light">Загрузка курсов...</div>
  }

  return (
    <div>
      <h1 className="mb-6 text-3xl font-bold text-text-light">Курсы</h1>

      {courses.length === 0 ? (
        <div className="rounded-lg border border-gray-200 bg-card-light p-6 shadow-md">
          <p className="text-center text-text-light">
            Вам пока не выдан доступ ни к одному курсу. Обратитесь к администратору, чтобы открыть нужные курсы.
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
          {courses.map((course) => (
            <CourseCard key={course.id} course={course} userCourse={latestUserCourseMap.get(course.id)} />
          ))}
        </div>
      )}
    </div>
  )
}
