import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import api from '../services/api'
import { uploadService } from '../services/uploadService'
import { Course, Deck, studyService } from '../services/studyService'

interface ReadingText {
  id: number
  user_id: string
  title: string
  content: string
  audio_url: string
  created_at: string
  updated_at: string
}

const normalizeBreaks = (value: string) => value.replace(/<br\s*\/?>/gi, '\n')

export default function Reader() {
  const navigate = useNavigate()
  const [texts, setTexts] = useState<ReadingText[]>([])
  const [loading, setLoading] = useState(true)
  const [showUploadForm, setShowUploadForm] = useState(false)
  const [title, setTitle] = useState('')
  const [content, setContent] = useState('')
  const [audioFile, setAudioFile] = useState<File | null>(null)
  const [audioInputKey, setAudioInputKey] = useState(0)
  const [uploading, setUploading] = useState(false)
  const [activeCourse, setActiveCourse] = useState<Course | null>(null)
  const [activeDeck, setActiveDeck] = useState<Deck | null>(null)

  const getUserId = () => {
    const authStorage = localStorage.getItem('auth-storage')
    if (!authStorage) return null
    const parsed = JSON.parse(authStorage)
    return parsed?.state?.user?.id
  }

  useEffect(() => {
    void Promise.all([loadTexts(), loadActiveTarget()])
  }, [])

  const loadActiveTarget = async () => {
    try {
      const [userCourse, userDeck] = await Promise.all([
        studyService.getActiveCourse().catch(() => null),
        studyService.getActiveDeck().catch(() => null),
      ])

      if (userCourse?.course_id) {
        setActiveCourse(await studyService.getCourse(userCourse.course_id))
      } else {
        setActiveCourse(null)
      }

      if (userDeck?.deck_id) {
        setActiveDeck(await studyService.getDeck(userDeck.deck_id))
      } else {
        setActiveDeck(null)
      }
    } catch (error) {
      console.error('Error loading active reader target:', error)
      setActiveCourse(null)
      setActiveDeck(null)
    }
  }

  const loadTexts = async () => {
    try {
      setLoading(true)
      const userId = getUserId()
      if (!userId) return

      const response = await api.get<ReadingText[]>(`/reading-texts?user_id=${userId}`)
      setTexts(response.data || [])
    } catch (error) {
      console.error('Error loading texts:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleFileUpload = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    if (!file) return

    const reader = new FileReader()
    reader.onload = (loadEvent) => {
      const text = loadEvent.target?.result as string
      setContent(text)
      setTitle(file.name.replace(/\.(txt|docx?)$/i, ''))
    }
    reader.readAsText(file)
  }

  const handleAudioFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    setAudioFile(file || null)
  }

  const resetForm = () => {
    setTitle('')
    setContent('')
    setAudioFile(null)
    setAudioInputKey((current) => current + 1)
    setShowUploadForm(false)
  }

  const handleSaveText = async () => {
    if (!title.trim() || !content.trim()) {
      alert('Заполните название и текст')
      return
    }

    const userId = getUserId()
    if (!userId) {
      alert('Пользователь не авторизован')
      return
    }

    let uploadedAudio: { url: string; filename: string } | null = null

    try {
      setUploading(true)

      if (audioFile) {
        uploadedAudio = await uploadService.uploadAudio(audioFile)
      }

      await api.post('/reading-texts', {
        user_id: userId,
        title: title.trim(),
        content: content.trim(),
        audio_url: uploadedAudio?.url || '',
      })

      resetForm()
      await loadTexts()
    } catch (error) {
      if (uploadedAudio?.filename) {
        void uploadService.deleteFile('audio', uploadedAudio.filename).catch((cleanupError) => {
          console.error('Error deleting uploaded audio after failed save:', cleanupError)
        })
      }

      console.error('Error saving text:', error)
      alert('Ошибка при сохранении текста')
    } finally {
      setUploading(false)
    }
  }

  const handleDeleteText = async (textId: number) => {
    if (!confirm('Удалить этот текст?')) return

    const userId = getUserId()
    if (!userId) return

    try {
      await api.delete(`/reading-texts/${textId}?user_id=${userId}`)
      await loadTexts()
    } catch (error) {
      console.error('Error deleting text:', error)
      alert('Ошибка при удалении текста')
    }
  }

  const openText = (textId: number) => {
    navigate(`/reader/${textId}`)
  }

  if (loading) {
    return <div className="py-8 text-center text-text-light">Загрузка...</div>
  }

  return (
    <div className="space-y-6">
      <section className="rounded-[28px] border border-rose-100 bg-gradient-to-br from-white via-rose-50 to-orange-50 p-6 shadow-xl">
        <div className="flex flex-col gap-5 lg:flex-row lg:items-start lg:justify-between">
          <div className="space-y-4">
            <div className="flex flex-wrap items-center gap-3">
              <span className="rounded-full bg-white/80 px-3 py-1 text-xs font-semibold uppercase tracking-[0.22em] text-link-light shadow-sm">
                Reader
              </span>
              {activeCourse && (
                <span className="rounded-full bg-rose-100 px-3 py-1 text-xs font-semibold text-rose-800">
                  Курс: {activeCourse.title}
                </span>
              )}
              {activeDeck && (
                <span className="rounded-full bg-orange-100 px-3 py-1 text-xs font-semibold text-orange-800">
                  Дек: {activeDeck.title}
                </span>
              )}
            </div>

            <div>
              <h1 className="text-3xl font-bold text-text-light">Reader и тексты для курса</h1>
              <p className="mt-2 max-w-3xl text-sm leading-7 text-slate-600">
                Нажимайте на слова внутри текста: reader покажет перевод, произношение и позволит сразу
                добавить слово в текущий активный дек.
              </p>
            </div>

            <div className="rounded-2xl border border-white/70 bg-white/80 p-4 text-sm text-slate-600 shadow-sm">
              {activeDeck ? (
                <p>
                  Сейчас новые слова будут добавляться в дек{' '}
                  <span className="font-semibold text-link-light">{activeDeck.title}</span>
                  {activeCourse ? (
                    <>
                      {' '}
                      курса <span className="font-semibold text-text-light">{activeCourse.title}</span>.
                    </>
                  ) : (
                    '.'
                  )}
                </p>
              ) : (
                <p>Сначала откройте нужный курс и активируйте дек, чтобы слова из reader добавлялись в правильное место.</p>
              )}
            </div>
          </div>

          <button
            onClick={() => setShowUploadForm(!showUploadForm)}
            className="rounded-2xl bg-link-light px-6 py-3 font-semibold text-white shadow-md transition-colors hover:bg-link-dark"
          >
            {showUploadForm ? 'Скрыть форму' : 'Добавить текст'}
          </button>
        </div>
      </section>

      {showUploadForm && (
        <section className="rounded-[28px] border border-rose-200 bg-white p-6 shadow-lg">
          <h2 className="text-xl font-bold text-text-light">Новый текст</h2>
          <p className="mt-2 text-sm text-slate-500">
            Заголовок поддерживает переносы через {'<br>'}, а для самого текста можно сразу прикрепить свою mp3-озвучку.
          </p>

          <div className="mt-5 grid gap-4">
            <div>
              <label className="mb-2 block text-sm font-medium text-slate-700">Загрузить из файла (.txt)</label>
              <input
                type="file"
                accept=".txt"
                onChange={handleFileUpload}
                className="w-full rounded-2xl border border-gray-300 px-4 py-3 focus:border-link-light focus:outline-none"
              />
            </div>

            <div>
              <label className="mb-2 block text-sm font-medium text-slate-700">Заголовок текста</label>
              <input
                type="text"
                value={title}
                onChange={(event) => setTitle(event.target.value)}
                placeholder="Например: My favorite book<br>Chapter 1"
                className="w-full rounded-2xl border border-gray-300 px-4 py-3 focus:border-link-light focus:outline-none"
              />
            </div>

            <div>
              <label className="mb-2 block text-sm font-medium text-slate-700">Текст на английском</label>
              <textarea
                value={content}
                onChange={(event) => setContent(event.target.value)}
                placeholder="Вставьте или введите текст..."
                rows={12}
                className="w-full rounded-2xl border border-gray-300 px-4 py-3 font-mono text-sm focus:border-link-light focus:outline-none"
              />
            </div>

            <div>
              <label className="mb-2 block text-sm font-medium text-slate-700">Своя озвучка (.mp3, необязательно)</label>
              <input
                key={audioInputKey}
                type="file"
                accept=".mp3,audio/*"
                onChange={handleAudioFileChange}
                className="w-full rounded-2xl border border-gray-300 px-4 py-3 focus:border-link-light focus:outline-none"
              />
              <p className="mt-2 text-sm text-slate-500">
                {audioFile ? `Выбран файл: ${audioFile.name}` : 'Можно прикрепить свой mp3 и потом запускать или ставить его на паузу в reader.'}
              </p>
            </div>

            <button
              onClick={handleSaveText}
              disabled={uploading}
              className="rounded-2xl bg-accent-light px-6 py-3 font-semibold text-white transition-colors hover:bg-accent-dark disabled:opacity-50"
            >
              {uploading ? 'Сохранение...' : 'Сохранить текст'}
            </button>
          </div>
        </section>
      )}

      <section className="grid grid-cols-1 gap-6 md:grid-cols-2 xl:grid-cols-3">
        {texts.length === 0 ? (
          <div className="col-span-full rounded-[28px] border border-gray-200 bg-card-light p-12 text-center shadow-md">
            <p className="text-lg font-medium text-text-light">У вас пока нет текстов для reader.</p>
            <p className="mt-2 text-sm text-slate-500">Добавьте первый текст и используйте его для словаря текущего курса.</p>
          </div>
        ) : (
          texts.map((text) => (
            <article
              key={text.id}
              className="group cursor-pointer overflow-hidden rounded-[28px] border border-gray-200 bg-card-light shadow-md transition-all hover:-translate-y-1 hover:border-rose-300 hover:shadow-xl"
              onClick={() => openText(text.id)}
            >
              <div className="p-6">
                <div className="flex flex-wrap items-center gap-2">
                  <p className="text-xs font-semibold uppercase tracking-[0.18em] text-slate-400">Текст reader</p>
                  {text.audio_url && (
                    <span className="rounded-full bg-emerald-100 px-2.5 py-1 text-[11px] font-semibold uppercase tracking-[0.14em] text-emerald-700">
                      MP3
                    </span>
                  )}
                </div>
                <h3 className="mt-3 whitespace-pre-line text-2xl font-bold leading-tight text-text-light">{normalizeBreaks(text.title)}</h3>
                <p className="mt-3 text-sm text-slate-500">
                  {new Date(text.created_at).toLocaleDateString('ru-RU', {
                    year: 'numeric',
                    month: 'long',
                    day: 'numeric',
                  })}
                </p>
                <p className="mt-4 line-clamp-4 whitespace-pre-line text-sm leading-7 text-slate-600">
                  {normalizeBreaks(text.content).substring(0, 180)}
                  ...
                </p>
              </div>

              <div className="flex items-center justify-between border-t border-gray-200 bg-slate-50 px-6 py-4">
                <span className="text-sm font-semibold text-link-light transition-colors group-hover:text-link-dark">Открыть reader</span>
                <button
                  onClick={(event) => {
                    event.stopPropagation()
                    void handleDeleteText(text.id)
                  }}
                  className="rounded-full bg-white px-3 py-1.5 text-sm font-medium text-red-500 transition-colors hover:bg-red-50 hover:text-red-700"
                  title="Удалить текст"
                >
                  Удалить
                </button>
              </div>
            </article>
          ))
        )}
      </section>
    </div>
  )
}
