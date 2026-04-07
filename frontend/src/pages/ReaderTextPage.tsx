import { useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { config } from '../config'
import api from '../services/api'
import { dictionaryService } from '../services/dictionaryService'
import { Course, Deck, studyService } from '../services/studyService'

interface ReadingText {
  id: number
  user_id: string
  course_id: number
  course_title: string
  course_slug: string
  title: string
  content: string
  audio_url: string
  created_at: string
  updated_at: string
}

interface WordTranslation {
  word: string
  translation: string
  phonetic?: string
  audio_url?: string
}

function normalizeBreaks(value: string): string {
  return value.replace(/<br\s*\/?>/gi, '\n')
}

function cleanWord(value: string): string {
  return value.replace(/[.,!?;:"'()[\]{}]/g, '').toLowerCase()
}

const TEXT_SPEECH_CHUNK_LIMIT = 260

function normalizeTextForSpeech(value: string): string {
  return value
    .replace(/<br\s*\/?>/gi, ' ')
    .replace(/\s+/g, ' ')
    .trim()
}

function splitLongSpeechPart(value: string, maxLength: number): string[] {
  const words = value.split(/\s+/).filter(Boolean)
  const chunks: string[] = []
  let current = ''

  for (const word of words) {
    const next = current ? `${current} ${word}` : word
    if (next.length <= maxLength) {
      current = next
      continue
    }

    if (current) {
      chunks.push(current)
      current = word
      continue
    }

    chunks.push(word.slice(0, maxLength))
    current = word.slice(maxLength)
  }

  if (current) {
    chunks.push(current)
  }

  return chunks
}

function splitTextForSpeech(value: string): string[] {
  const normalized = normalizeTextForSpeech(value)
  if (!normalized) {
    return []
  }

  const sentenceParts = normalized.split(/(?<=[.!?])\s+/).filter(Boolean)
  const chunks: string[] = []
  let current = ''

  for (const sentence of sentenceParts) {
    if (sentence.length > TEXT_SPEECH_CHUNK_LIMIT) {
      if (current) {
        chunks.push(current)
        current = ''
      }

      const smallerParts = sentence
        .split(/(?<=[,;:])\s+/)
        .flatMap((part) =>
          part.length > TEXT_SPEECH_CHUNK_LIMIT ? splitLongSpeechPart(part, TEXT_SPEECH_CHUNK_LIMIT) : [part]
        )

      chunks.push(...smallerParts.filter(Boolean))
      continue
    }

    const next = current ? `${current} ${sentence}` : sentence
    if (next.length <= TEXT_SPEECH_CHUNK_LIMIT) {
      current = next
      continue
    }

    if (current) {
      chunks.push(current)
    }
    current = sentence
  }

  if (current) {
    chunks.push(current)
  }

  return chunks
}

function pickPreferredVoice(voices: SpeechSynthesisVoice[]): SpeechSynthesisVoice | null {
  if (voices.length === 0) {
    return null
  }

  const englishVoices = voices.filter((voice) => voice.lang.toLowerCase().startsWith('en'))
  const voicePool = englishVoices.length > 0 ? englishVoices : voices
  const preferredPatterns = [
    /microsoft.*(aria|guy|jenny|davis|ana).*(natural|online)/i,
    /google.*english/i,
    /microsoft.*(natural|online)/i,
    /samantha/i,
    /alex/i,
    /daniel/i,
    /english/i,
  ]

  for (const pattern of preferredPatterns) {
    const match = voicePool.find((voice) => pattern.test(voice.name))
    if (match) {
      return match
    }
  }

  return voicePool[0] || null
}

export default function ReaderTextPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [text, setText] = useState<ReadingText | null>(null)
  const [loading, setLoading] = useState(true)
  const [wordTranslation, setWordTranslation] = useState<WordTranslation | null>(null)
  const [loadingTranslation, setLoadingTranslation] = useState(false)
  const [addingWord, setAddingWord] = useState(false)
  const [isSpeaking, setIsSpeaking] = useState(false)
  const [isUploadedAudioPlaying, setIsUploadedAudioPlaying] = useState(false)
  const [activeCourse, setActiveCourse] = useState<Course | null>(null)
  const [activeDeck, setActiveDeck] = useState<Deck | null>(null)
  const [readerMessage, setReaderMessage] = useState<string | null>(null)
  const [addedWords, setAddedWords] = useState<Set<string>>(new Set())
  const [availableVoices, setAvailableVoices] = useState<SpeechSynthesisVoice[]>([])
  const playbackTokenRef = useRef(0)
  const uploadedAudioRef = useRef<HTMLAudioElement | null>(null)

  const normalizedTitle = useMemo(() => normalizeBreaks(text?.title || ''), [text?.title])
  const normalizedContent = useMemo(() => normalizeBreaks(text?.content || ''), [text?.content])
  const uploadedAudioUrl = useMemo(() => (text?.audio_url ? config.getFullUrl(text.audio_url) : ''), [text?.audio_url])
  const speechSupported = typeof window !== 'undefined' && 'speechSynthesis' in window
  const preferredVoice = useMemo(() => pickPreferredVoice(availableVoices), [availableVoices])

  const getUserId = () => {
    const authStorage = localStorage.getItem('auth-storage')
    if (!authStorage) return null
    const parsed = JSON.parse(authStorage)
    return parsed?.state?.user?.id
  }

  const stopTextPlayback = (updateState = true) => {
    playbackTokenRef.current += 1

    if (speechSupported) {
      window.speechSynthesis.cancel()
    }

    if (updateState) {
      setIsSpeaking(false)
    }
  }

  const stopUploadedAudio = (reset = false, updateState = true) => {
    const audio = uploadedAudioRef.current
    if (!audio) {
      return
    }

    audio.pause()
    if (reset) {
      audio.currentTime = 0
    }
    if (updateState) {
      setIsUploadedAudioPlaying(false)
    }
  }

  useEffect(() => {
    void Promise.all([loadText(), loadActiveTarget()])

    return () => {
      stopTextPlayback(false)
      stopUploadedAudio(true, false)
    }
  }, [id])

  useEffect(() => {
    if (!speechSupported) {
      return
    }

    const loadVoices = () => {
      const voices = window.speechSynthesis.getVoices()
      if (voices.length > 0) {
        setAvailableVoices(voices)
      }
    }

    loadVoices()
    window.speechSynthesis.onvoiceschanged = loadVoices

    return () => {
      window.speechSynthesis.onvoiceschanged = null
    }
  }, [speechSupported])

  useEffect(() => {
    if (!uploadedAudioUrl) {
      stopUploadedAudio(true)
      uploadedAudioRef.current = null
      return
    }

    const audio = new Audio(uploadedAudioUrl)
    audio.preload = 'metadata'

    const handlePlay = () => setIsUploadedAudioPlaying(true)
    const handlePause = () => setIsUploadedAudioPlaying(false)
    const handleEnded = () => {
      setIsUploadedAudioPlaying(false)
      audio.currentTime = 0
    }

    audio.addEventListener('play', handlePlay)
    audio.addEventListener('pause', handlePause)
    audio.addEventListener('ended', handleEnded)
    uploadedAudioRef.current = audio

    return () => {
      audio.removeEventListener('play', handlePlay)
      audio.removeEventListener('pause', handlePause)
      audio.removeEventListener('ended', handleEnded)
      audio.pause()

      if (uploadedAudioRef.current === audio) {
        uploadedAudioRef.current = null
      }
    }
  }, [uploadedAudioUrl])

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
      console.error('Error loading reader target:', error)
      setActiveCourse(null)
      setActiveDeck(null)
    }
  }

  const loadText = async () => {
    try {
      setLoading(true)
      const userId = getUserId()
      if (!userId || !id) {
        navigate('/reader')
        return
      }

      const response = await api.get<ReadingText>(`/reading-texts/${id}?user_id=${userId}`)
      setText(response.data)
    } catch (error) {
      console.error('Error loading text:', error)
      navigate('/reader')
    } finally {
      setLoading(false)
    }
  }

  const handleWordClick = async (value: string) => {
    const word = cleanWord(value)
    if (!word) return

    setReaderMessage(null)

    try {
      setLoadingTranslation(true)
      const result = await dictionaryService.getWordInfo(word)
      setWordTranslation({
        word,
        translation: result.translation || 'Перевод не найден',
        phonetic: result.phonetic,
        audio_url: result.audio_url,
      })
    } catch (error) {
      console.error('Error translating word:', error)
      setWordTranslation({
        word,
        translation: 'Ошибка при получении перевода',
      })
    } finally {
      setLoadingTranslation(false)
    }
  }

  const speakText = (value: string, onEnd?: () => void) => {
    if (!value.trim() || !speechSupported) return

    const utterance = new SpeechSynthesisUtterance(value)
    utterance.lang = preferredVoice?.lang || 'en-US'
    if (preferredVoice) {
      utterance.voice = preferredVoice
    }
    utterance.rate = preferredVoice && /(natural|online|google)/i.test(preferredVoice.name) ? 1.02 : 0.94
    utterance.pitch = 1
    utterance.onend = () => onEnd?.()
    utterance.onerror = () => onEnd?.()

    window.speechSynthesis.cancel()
    window.speechSynthesis.speak(utterance)
  }

  const speakTextChunks = (chunks: string[], index: number, playbackToken: number) => {
    if (!speechSupported) {
      setIsSpeaking(false)
      return
    }

    if (playbackToken !== playbackTokenRef.current) {
      return
    }

    if (index >= chunks.length) {
      setIsSpeaking(false)
      return
    }

    const utterance = new SpeechSynthesisUtterance(chunks[index])
    utterance.lang = preferredVoice?.lang || 'en-US'
    if (preferredVoice) {
      utterance.voice = preferredVoice
    }
    utterance.rate = preferredVoice && /(natural|online|google)/i.test(preferredVoice.name) ? 1.02 : 0.94
    utterance.pitch = 1
    utterance.onend = () => {
      speakTextChunks(chunks, index + 1, playbackToken)
    }
    utterance.onerror = (error) => {
      console.error('Reader speech synthesis error:', error)
      if (playbackToken === playbackTokenRef.current) {
        setIsSpeaking(false)
      }
    }

    window.speechSynthesis.speak(utterance)
  }

  const handleSpeakText = () => {
    const chunks = splitTextForSpeech(normalizedContent)
    if (chunks.length === 0) {
      return
    }

    if (!speechSupported) {
      setReaderMessage('В этом браузере недоступна озвучка текста.')
      return
    }

    if (isSpeaking) {
      stopTextPlayback()
      return
    }

    stopUploadedAudio()
    stopTextPlayback(false)
    const playbackToken = playbackTokenRef.current + 1
    playbackTokenRef.current = playbackToken
    setIsSpeaking(true)
    speakTextChunks(chunks, 0, playbackToken)
  }

  const handleToggleUploadedAudio = async () => {
    const audio = uploadedAudioRef.current
    if (!audio) {
      return
    }

    setReaderMessage(null)

    if (!audio.paused) {
      audio.pause()
      return
    }

    stopTextPlayback()

    try {
      await audio.play()
    } catch (error) {
      console.error('Error playing uploaded audio:', error)
      setReaderMessage('Не удалось запустить загруженную озвучку.')
    }
  }

  const handleRestartUploadedAudio = async () => {
    const audio = uploadedAudioRef.current
    if (!audio) {
      return
    }

    setReaderMessage(null)
    stopTextPlayback()
    audio.currentTime = 0

    try {
      await audio.play()
    } catch (error) {
      console.error('Error restarting uploaded audio:', error)
      setReaderMessage('Не удалось перезапустить загруженную озвучку.')
    }
  }

  const handleSpeakWord = () => {
    if (!wordTranslation) return

    stopUploadedAudio()

    if (wordTranslation.audio_url) {
      const audio = new Audio(wordTranslation.audio_url)
      audio.play().catch((error) => console.error('Error playing dictionary audio:', error))
      return
    }

    speakText(wordTranslation.word)
  }

  const handleAddToDictionary = async () => {
    if (!wordTranslation) return

    try {
      setAddingWord(true)
      const result = await studyService.createCustomCard({
        word: wordTranslation.word,
        translation: wordTranslation.translation,
        phonetic: wordTranslation.phonetic,
        audio_url: wordTranslation.audio_url,
      })
      const targetDeck = activeDeck || (await studyService.getDeck(result.user_deck.deck_id).catch(() => null))

      setAddedWords((current) => new Set(current).add(wordTranslation.word))
      setReaderMessage(`Слово "${wordTranslation.word}" добавлено в дек "${targetDeck?.title || 'текущий дек'}".`)
      await loadActiveTarget()
    } catch (error: any) {
      console.error('Error adding word to custom dictionary:', error)
      setReaderMessage(error?.response?.data?.error || 'Не удалось добавить слово в словарь.')
    } finally {
      setAddingWord(false)
    }
  }

  const renderInteractiveText = (content: string) => {
    return content.split(/\n+/).map((line, lineIndex) => {
      const parts = line.split(/(\s+|[.,!?;:"'()[\]{}—-])/g)
      return (
        <p key={lineIndex} className="mb-4">
          {parts.map((part, partIndex) => {
            if (!part) return null
            if (/^\s+$/.test(part) || /^[.,!?;:"'()[\]{}—-]$/.test(part)) {
              return <span key={partIndex}>{part}</span>
            }

            return (
              <button
                key={partIndex}
                type="button"
                className="inline rounded-lg px-1.5 py-0.5 text-left transition-colors hover:bg-yellow-200"
                onClick={() => void handleWordClick(part)}
              >
                {part}
              </button>
            )
          })}
        </p>
      )
    })
  }

  if (loading) {
    return <div className="py-10 text-center text-text-light">Загрузка текста...</div>
  }

  if (!text) {
    return null
  }

  const alreadyAdded = wordTranslation ? addedWords.has(wordTranslation.word) : false
  const linkedCourseTitle = text.course_title?.trim()

  return (
    <div className="min-h-screen bg-gradient-to-br from-rose-50 to-orange-100">
      <div className="sticky top-0 z-10 border-b border-gray-200 bg-white/95 shadow-md backdrop-blur">
        <div className="mx-auto flex max-w-6xl flex-col gap-4 px-6 py-4 lg:flex-row lg:items-center lg:justify-between">
          <div className="flex items-center gap-4">
            <button
              type="button"
              onClick={() => navigate('/reader')}
              className="font-semibold text-link-light transition-colors hover:text-link-dark"
            >
              Назад
            </button>
            <div>
              <div className="flex flex-wrap items-center gap-2">
                {linkedCourseTitle && (
                  <span className="rounded-full bg-rose-100 px-3 py-1 text-xs font-semibold text-rose-800">
                    Курс текста: {linkedCourseTitle}
                  </span>
                )}
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
                {uploadedAudioUrl && (
                  <span className="rounded-full bg-emerald-100 px-3 py-1 text-xs font-semibold text-emerald-800">
                    Есть mp3
                  </span>
                )}
              </div>
              <h1 className="mt-2 whitespace-pre-line text-2xl font-bold text-gray-800">{normalizedTitle}</h1>
            </div>
          </div>

          <div className="flex flex-wrap items-center gap-3">
            {uploadedAudioUrl && (
              <>
                <button
                  type="button"
                  onClick={() => void handleToggleUploadedAudio()}
                  className="rounded-2xl bg-emerald-600 px-4 py-2 text-sm font-semibold text-white transition-colors hover:bg-emerald-700"
                >
                  {isUploadedAudioPlaying ? 'Пауза mp3' : 'Запустить mp3'}
                </button>
                <button
                  type="button"
                  onClick={() => void handleRestartUploadedAudio()}
                  className="rounded-2xl border border-emerald-200 bg-white px-4 py-2 text-sm font-semibold text-emerald-700 transition-colors hover:bg-emerald-50"
                >
                  С начала
                </button>
              </>
            )}

            <button
              type="button"
              onClick={handleSpeakText}
              className="rounded-2xl bg-link-light px-4 py-2 text-sm font-semibold text-white transition-colors hover:bg-link-dark"
            >
              {isSpeaking ? 'Остановить озвучку' : uploadedAudioUrl ? 'Озвучить браузером' : 'Озвучить текст'}
            </button>
          </div>
        </div>
      </div>

      <div className="mx-auto grid max-w-6xl gap-6 px-6 py-8 xl:grid-cols-[minmax(0,1fr)_360px]">
        <div className="space-y-6">
          <div className="rounded-[28px] bg-white p-8 shadow-xl">
            <div className="mb-6 rounded-2xl border border-rose-100 bg-rose-50/70 px-4 py-3 text-sm text-slate-600">
              Нажимайте по словам, чтобы увидеть перевод, прослушать произношение и добавить слово в текущий активный дек.
            </div>
            <div className="prose max-w-none text-lg leading-10 text-gray-800">{renderInteractiveText(normalizedContent)}</div>
          </div>

          {readerMessage && (
            <div className="rounded-2xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-800">
              {readerMessage}
            </div>
          )}
        </div>

        <aside className="space-y-4">
          <div className="rounded-[28px] border border-gray-200 bg-white p-6 shadow-xl">
            <h2 className="text-lg font-semibold text-text-light">Куда добавится слово</h2>
            <div className="mt-4 space-y-3 text-sm text-slate-600">
              {linkedCourseTitle && (
                <div className="rounded-2xl border border-rose-100 bg-rose-50 px-4 py-3">
                  <p className="text-xs uppercase tracking-[0.18em] text-rose-500">Курс текста</p>
                  <p className="mt-2 font-semibold text-text-light">{linkedCourseTitle}</p>
                </div>
              )}
              {activeDeck ? (
                <>
                  <div className="rounded-2xl border border-rose-100 bg-rose-50 px-4 py-3">
                    <p className="text-xs uppercase tracking-[0.18em] text-rose-500">Активный курс</p>
                    <p className="mt-2 font-semibold text-text-light">{activeCourse?.title || 'Курс выбран'}</p>
                  </div>
                  <div className="rounded-2xl border border-orange-100 bg-orange-50 px-4 py-3">
                    <p className="text-xs uppercase tracking-[0.18em] text-orange-500">Активный дек</p>
                    <p className="mt-2 font-semibold text-text-light">{activeDeck.title}</p>
                  </div>
                </>
              ) : (
                <div className="rounded-2xl border border-amber-200 bg-amber-50 px-4 py-3 text-amber-800">
                  Нет активного дека. Откройте нужный курс и выберите дек, тогда reader будет добавлять слова в него.
                </div>
              )}
            </div>
          </div>

          <div className="rounded-[28px] border border-gray-200 bg-white p-6 shadow-xl">
            <h2 className="text-lg font-semibold text-text-light">Своя озвучка</h2>
            {uploadedAudioUrl ? (
              <div className="mt-4 space-y-3 text-sm text-slate-600">
                <div className="rounded-2xl border border-emerald-100 bg-emerald-50 px-4 py-3 text-emerald-800">
                  Для этого текста загружен собственный аудиофайл.
                </div>
                <button
                  type="button"
                  onClick={() => void handleToggleUploadedAudio()}
                  className="w-full rounded-2xl bg-emerald-600 px-4 py-3 font-semibold text-white transition-colors hover:bg-emerald-700"
                >
                  {isUploadedAudioPlaying ? 'Поставить на паузу' : 'Включить mp3'}
                </button>
                <button
                  type="button"
                  onClick={() => void handleRestartUploadedAudio()}
                  className="w-full rounded-2xl border border-emerald-200 bg-white px-4 py-3 font-semibold text-emerald-700 transition-colors hover:bg-emerald-50"
                >
                  Запустить сначала
                </button>
              </div>
            ) : (
              <div className="mt-4 rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm text-slate-600">
                Для этого текста пока не загружена своя mp3-озвучка. Добавить ее можно при создании текста в reader.
              </div>
            )}
          </div>

          <div className="rounded-[28px] border border-gray-200 bg-white p-6 shadow-xl">
            {loadingTranslation ? (
              <p className="text-center text-sm text-slate-500">Загрузка перевода...</p>
            ) : wordTranslation ? (
              <div className="space-y-4">
                <div>
                  <p className="text-xs font-semibold uppercase tracking-[0.18em] text-slate-400">Выбранное слово</p>
                  <div className="mt-2 flex items-start justify-between gap-4">
                    <div>
                      <div className="flex flex-wrap items-center gap-3">
                        <h3 className="text-3xl font-bold text-gray-800">{wordTranslation.word}</h3>
                        <button
                          type="button"
                          onClick={handleSpeakWord}
                          className="rounded-full bg-rose-100 px-3 py-1 text-sm font-semibold text-rose-700 transition-colors hover:bg-rose-200"
                        >
                          Произношение
                        </button>
                      </div>
                      {wordTranslation.phonetic && <p className="mt-2 text-sm text-slate-500">[{wordTranslation.phonetic}]</p>}
                    </div>
                    <button
                      type="button"
                      onClick={() => setWordTranslation(null)}
                      className="text-2xl text-slate-400 transition-colors hover:text-slate-600"
                    >
                      x
                    </button>
                  </div>
                </div>

                <div className="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3">
                  <p className="text-xs uppercase tracking-[0.18em] text-slate-400">Перевод</p>
                  <p className="mt-2 text-lg font-medium text-slate-700">{wordTranslation.translation}</p>
                </div>

                {alreadyAdded && (
                  <div className="rounded-2xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-800">
                    Это слово уже было добавлено в словарь в текущей сессии reader.
                  </div>
                )}

                <button
                  type="button"
                  onClick={() => void handleAddToDictionary()}
                  disabled={addingWord || !activeDeck || alreadyAdded}
                  className="w-full rounded-2xl bg-link-light px-4 py-3 text-sm font-semibold text-white transition-colors hover:bg-link-dark disabled:cursor-not-allowed disabled:opacity-50"
                >
                  {addingWord
                    ? 'Добавляю...'
                    : activeDeck
                      ? `Добавить в дек "${activeDeck.title}"`
                      : 'Сначала выберите активный дек'}
                </button>
              </div>
            ) : (
              <div className="text-center text-sm text-slate-500">
                Нажмите на слово в тексте, и здесь появятся перевод, произношение и кнопка добавления в словарь.
              </div>
            )}
          </div>
        </aside>
      </div>
    </div>
  )
}
