import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { adminService, Course, CreateCourseRequest } from '../services/adminService'
import { useAuthStore } from '../store/authStore'
import { uploadService } from '../services/uploadService'
import { config } from '../config'
import { slugify } from '../utils/slug'
import { dictionaryService } from '../services/dictionaryService'

interface ImportedSheetRow {
  lesson: number
  position: number
  word: string
  translation: string
}

interface ImportPreview {
  rows: ImportedSheetRow[]
  lessonsCount: number
  error: string | null
}

const splitImportColumns = (line: string) => {
  if (line.includes('\t')) {
    return line.split('\t')
  }

  if (line.includes(';')) {
    return line.split(';')
  }

  if (line.includes(',')) {
    return line.split(',')
  }

  return line.split(/\s{2,}/)
}

const normalizeHeaderValue = (value: string) =>
  value
    .trim()
    .toLowerCase()
    .replace(/\s+/g, '')

const isImportHeaderRow = (columns: string[]) => {
  const normalized = columns.slice(0, 4).map(normalizeHeaderValue)
  const [lessonValue, positionValue, englishValue, russianValue] = normalized

  return (
    (lessonValue === 'урок' || lessonValue === 'lesson') &&
    (positionValue === '№' || positionValue === 'no' || positionValue === 'n' || positionValue === 'номер') &&
    (englishValue === 'en' || englishValue === 'english') &&
    (russianValue === 'ru' || russianValue === 'russian' || russianValue === 'translation')
  )
}

const parseImportSheetData = (rawValue: string): ImportedSheetRow[] => {
  const lines = rawValue
    .split(/\r?\n/)
    .map((line) => line.trim())
    .filter(Boolean)

  if (lines.length === 0) {
    return []
  }

  const rows: ImportedSheetRow[] = []
  const seenPositions = new Set<string>()

  lines.forEach((line, index) => {
    const columns = splitImportColumns(line).map((column) => column.trim())

    if (columns.length < 4) {
      throw new Error(`Строка ${index + 1}: нужно 4 столбца - Урок, №, En, Ru.`)
    }

    if (index === 0 && isImportHeaderRow(columns)) {
      return
    }

    const lesson = Number(columns[0])
    const position = Number(columns[1])
    const word = columns[2]
    const translation = columns[3]

    if (!Number.isInteger(lesson) || lesson <= 0) {
      throw new Error(`Строка ${index + 1}: поле "Урок" должно быть целым числом больше 0.`)
    }

    if (!Number.isInteger(position) || position <= 0) {
      throw new Error(`Строка ${index + 1}: поле "№" должно быть целым числом больше 0.`)
    }

    if (!word || !translation) {
      throw new Error(`Строка ${index + 1}: нужны и английское слово, и перевод.`)
    }

    const uniqueKey = `${lesson}:${position}`
    if (seenPositions.has(uniqueKey)) {
      throw new Error(`Урок ${lesson}: номер ${position} повторяется. Для каждого урока "№" должен быть уникальным.`)
    }

    seenPositions.add(uniqueKey)
    rows.push({
      lesson,
      position,
      word,
      translation,
    })
  })

  return rows.sort((left, right) => {
    if (left.lesson !== right.lesson) {
      return left.lesson - right.lesson
    }

    return left.position - right.position
  })
}

const buildImportPreview = (rawValue: string): ImportPreview => {
  if (!rawValue.trim()) {
    return {
      rows: [],
      lessonsCount: 0,
      error: null,
    }
  }

  try {
    const rows = parseImportSheetData(rawValue)
    return {
      rows,
      lessonsCount: new Set(rows.map((row) => row.lesson)).size,
      error: null,
    }
  } catch (error) {
    return {
      rows: [],
      lessonsCount: 0,
      error: error instanceof Error ? error.message : 'Не удалось разобрать таблицу.',
    }
  }
}

export default function AdminCourses() {
  const navigate = useNavigate()
  const { user, isAuthenticated } = useAuthStore()
  const [courses, setCourses] = useState<Course[]>([])
  const [loading, setLoading] = useState(true)
  const [showForm, setShowForm] = useState(false)
  const [editingCourse, setEditingCourse] = useState<Course | null>(null)
  const [formData, setFormData] = useState<CreateCourseRequest>({
    title: '',
    slug: '',
    description: '',
    image_url: '',
  })
  const [imageMode, setImageMode] = useState<'url' | 'file'>('url')
  const [selectedFile, setSelectedFile] = useState<File | null>(null)
  const [uploading, setUploading] = useState(false)
  const [slugTouched, setSlugTouched] = useState(false)
  const [importSheetData, setImportSheetData] = useState('')
  const [submitError, setSubmitError] = useState<string | null>(null)
  const [importStatus, setImportStatus] = useState('')

  const importPreview: ImportPreview = !editingCourse
    ? buildImportPreview(importSheetData)
    : { rows: [], lessonsCount: 0, error: null }

  useEffect(() => {
    if (isAuthenticated && user?.role === 'admin') {
      void loadCourses()
    }
  }, [isAuthenticated, user])

  const loadCourses = async () => {
    try {
      setLoading(true)
      const data = await adminService.getAllCourses()
      setCourses(data || [])
    } catch (error) {
      console.error('Error loading courses:', error)
      setCourses([])
    } finally {
      setLoading(false)
    }
  }

  const resetForm = () => {
    setFormData({ title: '', slug: '', description: '', image_url: '' })
    setSelectedFile(null)
    setImageMode('url')
    setSlugTouched(false)
    setEditingCourse(null)
    setImportSheetData('')
    setSubmitError(null)
    setImportStatus('')
    setShowForm(false)
  }

  const importCourseRows = async (course: Course, rows: ImportedSheetRow[]) => {
    const lessonMap = new Map<number, ImportedSheetRow[]>()

    rows.forEach((row) => {
      const currentLessonRows = lessonMap.get(row.lesson) || []
      currentLessonRows.push(row)
      lessonMap.set(row.lesson, currentLessonRows)
    })

    const lessonNumbers = Array.from(lessonMap.keys()).sort((left, right) => left - right)

    for (let lessonIndex = 0; lessonIndex < lessonNumbers.length; lessonIndex += 1) {
      const lessonNumber = lessonNumbers[lessonIndex]
      const lessonRows = (lessonMap.get(lessonNumber) || []).sort((left, right) => left.position - right.position)

      setImportStatus(`Импорт урока ${lessonNumber} (${lessonIndex + 1}/${lessonNumbers.length})`)

      const deck = await adminService.createDeck({
        course_id: course.id,
        title: `Урок ${lessonNumber}`,
        slug: slugify(`lesson-${lessonNumber}`),
        description: `Импортировано из таблицы`,
        position: lessonNumber,
      })

      for (let rowIndex = 0; rowIndex < lessonRows.length; rowIndex += 1) {
        const row = lessonRows[rowIndex]
        let phonetic: string | undefined
        let audioUrl: string | undefined

        setImportStatus(
          `Импорт урока ${lessonNumber}: слово ${rowIndex + 1}/${lessonRows.length} (${row.word})`,
        )

        try {
          const wordInfo = await dictionaryService.getWordInfo(row.word.toLowerCase())
          phonetic = wordInfo.phonetic || undefined
          audioUrl = wordInfo.audio_url || undefined
        } catch (error) {
          console.warn(`Audio lookup failed for "${row.word}"`, error)
        }

        await adminService.createCard({
          deck_id: deck.id,
          position: row.position,
          word: row.word,
          translation: row.translation,
          phonetic,
          audio_url: audioUrl,
        })
      }
    }
  }

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault()

    let parsedImportRows: ImportedSheetRow[] = []
    if (!editingCourse && importSheetData.trim()) {
      if (importPreview.error) {
        setSubmitError(importPreview.error)
        return
      }

      parsedImportRows = importPreview.rows
    }

    let createdCourse: Course | null = null

    try {
      setUploading(true)
      setSubmitError(null)
      setImportStatus('')

      let imageUrl = formData.image_url

      if (imageMode === 'file' && selectedFile) {
        const uploadResult = await uploadService.uploadImage(selectedFile)
        imageUrl = uploadResult.url
      }

      if (editingCourse) {
        await adminService.updateCourse(editingCourse.id, {
          title: formData.title,
          slug: formData.slug || undefined,
          description: formData.description || undefined,
          image_url: imageUrl || undefined,
        })
      } else {
        createdCourse = await adminService.createCourse({
          title: formData.title,
          slug: formData.slug || undefined,
          description: formData.description || undefined,
          image_url: imageUrl || undefined,
          is_published: false,
        })

        if (createdCourse && parsedImportRows.length > 0) {
          await importCourseRows(createdCourse, parsedImportRows)
        }
      }

      await loadCourses()
      setImportStatus('')

      if (createdCourse && parsedImportRows.length > 0) {
        resetForm()
        navigate(`/admin/decks/${createdCourse.id}`)
        return
      }

      resetForm()
    } catch (error) {
      console.error(`Error ${editingCourse ? 'updating' : 'creating'} course:`, error)

      if (createdCourse && parsedImportRows.length > 0) {
        setSubmitError('Курс создан, но импорт прервался. Откройте деки курса и проверьте уже загруженные карточки.')
        await loadCourses()
      } else {
        setSubmitError('Не удалось сохранить курс. Проверьте данные и попробуйте снова.')
      }
    } finally {
      setUploading(false)
      setImportStatus('')
    }
  }

  const handleEdit = (course: Course) => {
    setEditingCourse(course)
    setFormData({
      title: course.title,
      slug: course.slug || '',
      description: course.description || '',
      image_url: course.image_url || '',
    })
    setSlugTouched(true)
    setImportSheetData('')
    setSubmitError(null)
    setImportStatus('')
    setShowForm(true)
  }

  const handleTitleChange = (title: string) => {
    setFormData((prev) => ({
      ...prev,
      title,
      slug: slugTouched ? prev.slug : slugify(title),
    }))
  }

  const handlePublish = async (id: number) => {
    try {
      await adminService.publishCourse(id)
      await loadCourses()
    } catch (error) {
      console.error('Error publishing course:', error)
    }
  }

  const handleDelete = async (id: number) => {
    if (!confirm('Удалить курс? Это действие нельзя отменить.')) return
    try {
      await adminService.deleteCourse(id)
      await loadCourses()
    } catch (error) {
      console.error('Error deleting course:', error)
    }
  }

  if (!isAuthenticated || !user) {
    return <div className="py-8 text-center text-text-light">Проверка доступа...</div>
  }

  if (user.role !== 'admin') {
    return (
      <div className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-red-700">
        У вас нет доступа к этой странице. Требуется роль администратора.
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <section className="overflow-hidden rounded-[28px] border border-rose-100 bg-gradient-to-br from-white via-rose-50 to-orange-50 shadow-xl">
        <div className="grid gap-6 p-6 lg:grid-cols-[minmax(0,1fr)_240px] lg:p-8">
          <div className="space-y-4">
            <div className="flex flex-wrap items-center gap-3">
              <span className="rounded-full bg-white/80 px-3 py-1 text-xs font-semibold uppercase tracking-[0.22em] text-link-light shadow-sm">
                Admin Courses
              </span>
              <span className="rounded-full bg-rose-100 px-3 py-1 text-xs font-semibold text-rose-800">
                Контент и публикация
              </span>
            </div>
            <div>
              <h1 className="text-3xl font-bold text-text-light lg:text-4xl">Управление курсами</h1>
              <p className="mt-2 max-w-3xl text-sm leading-7 text-slate-600">
                Создавайте новые курсы, редактируйте оформление и сразу импортируйте слова из Google Sheets в готовые уроки.
              </p>
            </div>
          </div>

          <div className="flex items-center justify-end">
            <button
              onClick={() => {
                if (showForm) {
                  resetForm()
                } else {
                  setShowForm(true)
                  setEditingCourse(null)
                  setFormData({ title: '', slug: '', description: '', image_url: '' })
                  setSelectedFile(null)
                  setImageMode('url')
                  setSlugTouched(false)
                  setImportSheetData('')
                  setSubmitError(null)
                  setImportStatus('')
                }
              }}
              className="rounded-2xl bg-link-light px-5 py-3 font-semibold text-white transition-colors hover:bg-link-dark"
            >
              {showForm ? 'Скрыть форму' : 'Создать курс'}
            </button>
          </div>
        </div>
      </section>

      {showForm && (
        <section className="rounded-[28px] border border-gray-200 bg-card-light p-6 shadow-md">
          <h2 className="text-xl font-semibold text-text-light">
            {editingCourse ? 'Редактировать курс' : 'Создать новый курс'}
          </h2>

          <form onSubmit={handleSubmit} className="mt-5 space-y-4">
            <input
              type="text"
              placeholder="Название курса *"
              required
              value={formData.title}
              onChange={(event) => handleTitleChange(event.target.value)}
              className="w-full rounded-2xl border border-gray-300 px-4 py-3 focus:border-link-light focus:outline-none"
            />
            <input
              type="text"
              placeholder="slug"
              required
              value={formData.slug || ''}
              onChange={(event) => {
                setSlugTouched(true)
                setFormData({ ...formData, slug: slugify(event.target.value) })
              }}
              className="w-full rounded-2xl border border-gray-300 px-4 py-3 focus:border-link-light focus:outline-none"
            />
            <textarea
              placeholder="Описание"
              value={formData.description}
              onChange={(event) => setFormData({ ...formData, description: event.target.value })}
              className="w-full rounded-2xl border border-gray-300 px-4 py-3 focus:border-link-light focus:outline-none"
              rows={4}
            />

            <div className="rounded-[24px] border border-slate-200 bg-slate-50 p-5">
              <p className="text-sm font-semibold text-text-light">Изображение курса</p>
              <div className="mt-3 flex flex-wrap gap-4">
                <label className="flex items-center gap-2 text-sm text-slate-600">
                  <input
                    type="radio"
                    name="imageMode"
                    value="url"
                    checked={imageMode === 'url'}
                    onChange={() => {
                      setImageMode('url')
                      setSelectedFile(null)
                    }}
                  />
                  URL изображения
                </label>
                <label className="flex items-center gap-2 text-sm text-slate-600">
                  <input
                    type="radio"
                    name="imageMode"
                    value="file"
                    checked={imageMode === 'file'}
                    onChange={() => {
                      setImageMode('file')
                      setFormData({ ...formData, image_url: '' })
                    }}
                  />
                  Загрузить файл
                </label>
              </div>

              <div className="mt-4">
                {imageMode === 'url' ? (
                  <input
                    type="url"
                    placeholder="https://example.com/image.jpg"
                    value={formData.image_url}
                    onChange={(event) => setFormData({ ...formData, image_url: event.target.value })}
                    className="w-full rounded-2xl border border-gray-300 px-4 py-3 focus:border-link-light focus:outline-none"
                  />
                ) : (
                  <div className="space-y-2">
                    <input
                      type="file"
                      accept="image/*"
                      onChange={(event) => {
                        const file = event.target.files?.[0]
                        if (file) setSelectedFile(file)
                      }}
                      className="w-full rounded-2xl border border-gray-300 px-4 py-3 focus:border-link-light focus:outline-none file:mr-4 file:rounded-xl file:border-0 file:bg-link-light file:px-4 file:py-2 file:text-sm file:font-semibold file:text-white hover:file:bg-link-dark"
                    />
                    {selectedFile && <p className="text-sm text-slate-500">Выбран файл: {selectedFile.name}</p>}
                  </div>
                )}
              </div>
            </div>

            {!editingCourse && (
              <div className="rounded-[24px] border border-emerald-200 bg-emerald-50/60 p-5">
                <div className="flex flex-col gap-2">
                  <p className="text-sm font-semibold text-text-light">Импорт слов из Google Sheets</p>
                  <p className="text-sm leading-6 text-slate-600">
                    Вставьте диапазон из таблицы со столбцами <span className="font-semibold">Урок</span>, <span className="font-semibold">№</span>, <span className="font-semibold">En</span>, <span className="font-semibold">Ru</span>.
                    На каждый номер урока создастся свой подкурс, карточки будут отсортированы по полю "№", а озвучка подтянется автоматически так же, как при ручном создании карточки.
                  </p>
                </div>

                <textarea
                  placeholder={'Урок\t№\tEn\tRu\n1\t1\tHello\tПривет\n1\t2\tHi\tПривет'}
                  value={importSheetData}
                  onChange={(event) => {
                    setImportSheetData(event.target.value)
                    setSubmitError(null)
                  }}
                  className="mt-4 min-h-[220px] w-full rounded-2xl border border-emerald-200 bg-white px-4 py-3 font-mono text-sm focus:border-emerald-500 focus:outline-none"
                />

                <div className="mt-4 space-y-2 text-sm">
                  {importPreview.error ? (
                    <div className="rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-red-700">
                      {importPreview.error}
                    </div>
                  ) : importPreview.rows.length > 0 ? (
                    <div className="rounded-2xl border border-emerald-200 bg-white px-4 py-3 text-emerald-800">
                      Распознано карточек: {importPreview.rows.length}. Уроков: {importPreview.lessonsCount}. Карточки будут созданы в порядке возрастания урока и номера.
                    </div>
                  ) : (
                    <div className="rounded-2xl border border-dashed border-emerald-200 bg-white px-4 py-3 text-slate-500">
                      Импорт необязателен. Если оставить поле пустым, создастся только сам курс.
                    </div>
                  )}
                </div>
              </div>
            )}

            {submitError && (
              <div className="rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
                {submitError}
              </div>
            )}

            {importStatus && (
              <div className="rounded-2xl border border-blue-200 bg-blue-50 px-4 py-3 text-sm text-blue-700">
                {importStatus}
              </div>
            )}

            <div className="flex flex-wrap gap-3">
              <button
                type="submit"
                disabled={uploading || Boolean(importPreview.error)}
                className="rounded-2xl bg-accent-light px-5 py-3 font-semibold text-white transition-colors hover:bg-accent-dark disabled:cursor-not-allowed disabled:opacity-50"
              >
                {uploading
                  ? parsedButtonLabel(Boolean(editingCourse), importPreview.rows.length > 0)
                  : editingCourse
                    ? 'Сохранить курс'
                    : 'Создать курс'}
              </button>
              <button
                type="button"
                onClick={resetForm}
                className="rounded-2xl bg-slate-200 px-5 py-3 font-semibold text-slate-700 transition-colors hover:bg-slate-300"
              >
                Отмена
              </button>
            </div>
          </form>
        </section>
      )}

      {loading ? (
        <div className="py-8 text-center text-text-light">Загрузка курсов...</div>
      ) : (
        <section className="grid gap-5 md:grid-cols-2">
          {courses.length === 0 ? (
            <div className="rounded-[28px] border border-gray-200 bg-card-light p-10 text-center shadow-md md:col-span-2">
              <p className="text-lg font-medium text-text-light">Курсов пока нет. Создайте первый курс.</p>
            </div>
          ) : (
            courses.map((course) => (
              <article key={course.id} className="overflow-hidden rounded-[28px] border border-gray-200 bg-card-light shadow-md">
                {course.image_url && (
                  <img
                    src={config.getFullUrl(course.image_url)}
                    alt={course.title}
                    className="h-52 w-full object-cover"
                    onError={(event) => {
                      ;(event.target as HTMLImageElement).style.display = 'none'
                    }}
                  />
                )}

                <div className="space-y-4 p-6">
                  <div className="flex flex-wrap items-start justify-between gap-3">
                    <div>
                      <h2 className="text-2xl font-semibold text-text-light">{course.title}</h2>
                      {course.slug && <p className="mt-2 text-xs uppercase tracking-[0.18em] text-slate-400">/deck/{course.slug}</p>}
                    </div>
                    <span
                      className={`rounded-full px-3 py-1 text-xs font-semibold ${
                        course.is_published ? 'bg-green-100 text-green-800' : 'bg-slate-100 text-slate-700'
                      }`}
                    >
                      {course.is_published ? 'Опубликован' : 'Черновик'}
                    </span>
                  </div>

                  {course.description && <p className="text-sm leading-7 text-slate-600">{course.description}</p>}

                  <div className="flex flex-wrap gap-2">
                    <Link
                      to={`/admin/decks/${course.id}`}
                      className="rounded-2xl border border-rose-200 bg-rose-50 px-4 py-2 text-sm font-semibold text-rose-700 transition-colors hover:bg-rose-100"
                    >
                      Деки
                    </Link>
                    <button
                      onClick={() => handleEdit(course)}
                      className="rounded-2xl border border-slate-300 bg-slate-50 px-4 py-2 text-sm font-semibold text-slate-700 transition-colors hover:bg-slate-100"
                    >
                      Редактировать
                    </button>
                    <button
                      onClick={() => handlePublish(course.id)}
                      className="rounded-2xl border border-orange-200 bg-orange-50 px-4 py-2 text-sm font-semibold text-orange-700 transition-colors hover:bg-orange-100"
                    >
                      {course.is_published ? 'Снять с публикации' : 'Опубликовать'}
                    </button>
                    <button
                      onClick={() => handleDelete(course.id)}
                      className="rounded-2xl border border-red-200 bg-red-50 px-4 py-2 text-sm font-semibold text-red-700 transition-colors hover:bg-red-100"
                    >
                      Удалить
                    </button>
                  </div>
                </div>
              </article>
            ))
          )}
        </section>
      )}
    </div>
  )
}

function parsedButtonLabel(isEditing: boolean, hasImportRows: boolean) {
  if (isEditing) {
    return 'Сохраняю...'
  }

  if (hasImportRows) {
    return 'Создаю курс и импортирую...'
  }

  return 'Создаю курс...'
}
