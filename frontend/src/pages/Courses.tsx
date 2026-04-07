import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { studyService, Course } from '../services/studyService'
import { config } from '../config'

export default function Courses() {
  const [courses, setCourses] = useState<Course[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    loadCourses()
  }, [])

  const loadCourses = async () => {
    try {
      setLoading(true)
      const data = await studyService.getCourses()
      setCourses(data || [])
    } catch (error: any) {
      console.error('Error loading courses:', error)
      setCourses([])
    } finally {
      setLoading(false)
    }
  }

  if (loading) {
    return (
      <div className="text-center py-8 text-text-light">
        Загрузка курсов...
      </div>
    )
  }

  return (
    <div>
      <h1 className="text-3xl font-bold text-text-light mb-6">Курсы</h1>

      {courses.length === 0 ? (
        <div className="bg-card-light shadow-md rounded-lg p-6 border border-gray-200">
          <p className="text-text-light text-center">Курсов пока нет. Скоро здесь появятся новые курсы!</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {courses.map((course) => (
            <Link
              key={course.id}
              to={`/courses/${course.id}`}
              className="bg-card-light shadow-md rounded-lg overflow-hidden border border-gray-200 hover:shadow-lg hover:border-link-light transition-all"
            >
              {course.image_url && (
                <img
                  src={config.getFullUrl(course.image_url)}
                  alt={course.title}
                  className="w-full h-48 object-cover"
                  onError={(e) => {
                    (e.target as HTMLImageElement).style.display = 'none'
                  }}
                />
              )}
              <div className="p-6">
                <h2 className="text-xl font-semibold text-text-light mb-2">{course.title}</h2>
                {course.description && (
                  <p className="text-text-light text-sm mb-4 line-clamp-2">{course.description}</p>
                )}
                <div className="flex items-center justify-between">
                  <span className="text-xs text-gray-500">
                    {new Date(course.created_at).toLocaleDateString('ru-RU')}
                  </span>
                  <span className="px-2 py-1 bg-green-100 text-green-800 text-xs font-semibold rounded-full">
                    Опубликован
                  </span>
                </div>
              </div>
            </Link>
          ))}
        </div>
      )}
    </div>
  )
}
