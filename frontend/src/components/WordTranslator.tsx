import { useEffect, useMemo, useRef, useState } from 'react'
import { dictionaryService, WordInfo } from '../services/dictionaryService'
import { personalTranslationService, PersonalTranslation } from '../services/personalTranslationService'

interface TooltipPosition {
  x: number
  y: number
}

function normalizeSelectedWord(value: string) {
  return value.toLowerCase().replace(/[^a-z'-]/gi, '').trim()
}

export default function WordTranslator() {
  const [selectedWord, setSelectedWord] = useState('')
  const [translation, setTranslation] = useState<WordInfo | null>(null)
  const [personalTranslations, setPersonalTranslations] = useState<PersonalTranslation[]>([])
  const [position, setPosition] = useState<TooltipPosition | null>(null)
  const [loading, setLoading] = useState(false)
  const [newTranslation, setNewTranslation] = useState('')
  const [savingTranslation, setSavingTranslation] = useState(false)
  const [translationMessage, setTranslationMessage] = useState<string | null>(null)
  const [viewportWidth, setViewportWidth] = useState(() => (typeof window !== 'undefined' ? window.innerWidth : 1280))
  const [viewportHeight, setViewportHeight] = useState(() => (typeof window !== 'undefined' ? window.innerHeight : 800))
  const timeoutRef = useRef<number | null>(null)

  const hasPersonalTranslations = personalTranslations.length > 0
  const sortedPersonalTranslations = useMemo(
    () =>
      [...personalTranslations].sort((left, right) => {
        const leftDate = new Date(left.created_at).getTime()
        const rightDate = new Date(right.created_at).getTime()
        return rightDate - leftDate
      }),
    [personalTranslations]
  )

  const isMobileView = viewportWidth < 640
  const tooltipWidth = Math.min(384, Math.max(280, viewportWidth - 24))
  const tooltipLeft = position ? Math.min(Math.max(12, position.x + 10), Math.max(12, viewportWidth - tooltipWidth - 12)) : 12
  const tooltipTop = position ? Math.min(Math.max(12, position.y + 10), Math.max(12, viewportHeight - 220)) : 12

  const closeTooltip = () => {
    setPosition(null)
    setTranslation(null)
    setSelectedWord('')
    setPersonalTranslations([])
    setNewTranslation('')
    setTranslationMessage(null)
  }

  const loadWordDetails = async (word: string) => {
    setLoading(true)
    setTranslationMessage(null)

    try {
      const [dictionaryInfo, savedTranslations] = await Promise.all([
        dictionaryService.getWordInfo(word).catch(() => null),
        personalTranslationService.getAll(word).catch(() => []),
      ])

      setTranslation(dictionaryInfo)
      setPersonalTranslations(savedTranslations)
    } catch (error) {
      console.error('Translation error:', error)
      setTranslation(null)
      setPersonalTranslations([])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    const handleResize = () => {
      setViewportWidth(window.innerWidth)
      setViewportHeight(window.innerHeight)
    }

    handleResize()
    window.addEventListener('resize', handleResize)

    return () => {
      window.removeEventListener('resize', handleResize)
    }
  }, [])

  useEffect(() => {
    const handleMouseUp = async (event: MouseEvent) => {
      const selection = window.getSelection()
      const text = selection?.toString().trim()

      if (!text || text.length >= 50) {
        closeTooltip()
        return
      }

      const words = text.split(/\s+/)
      if (words.length !== 1) {
        closeTooltip()
        return
      }

      const word = normalizeSelectedWord(words[0])
      if (word.length < 2) {
        closeTooltip()
        return
      }

      setSelectedWord(word)
      setPosition({ x: event.clientX, y: event.clientY })
      setNewTranslation('')
      await loadWordDetails(word)
    }

    const handleClick = (event: MouseEvent) => {
      const target = event.target as HTMLElement
      if (!target.closest('.word-translator-tooltip')) {
        closeTooltip()
      }
    }

    document.addEventListener('mouseup', handleMouseUp)
    document.addEventListener('click', handleClick)

    return () => {
      document.removeEventListener('mouseup', handleMouseUp)
      document.removeEventListener('click', handleClick)
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current)
      }
    }
  }, [])

  const handlePlayAudio = () => {
    const audioUrl = translation?.audio_url
    if (!audioUrl) {
      return
    }

    const normalizedUrl = audioUrl.startsWith('//') ? `https:${audioUrl}` : audioUrl
    const audio = new Audio(normalizedUrl)
    audio.play().catch((error) => console.error('Error playing dictionary audio:', error))
  }

  const handleAddTranslation = async () => {
    if (!selectedWord || !newTranslation.trim()) {
      return
    }

    try {
      setSavingTranslation(true)
      const created = await personalTranslationService.create(selectedWord, newTranslation.trim())
      setPersonalTranslations((current) => [created, ...current])
      setNewTranslation('')
      setTranslationMessage('Ваш перевод сохранён.')
    } catch (error: any) {
      console.error('Error saving personal translation:', error)
      setTranslationMessage(error?.response?.data?.error || 'Не удалось сохранить перевод.')
    } finally {
      setSavingTranslation(false)
    }
  }

  const handleDeleteTranslation = async (translationId: number) => {
    try {
      await personalTranslationService.remove(translationId)
      setPersonalTranslations((current) => current.filter((item) => item.id !== translationId))
      setTranslationMessage('Перевод удалён.')
    } catch (error: any) {
      console.error('Error deleting personal translation:', error)
      setTranslationMessage(error?.response?.data?.error || 'Не удалось удалить перевод.')
    } finally {
      setSavingTranslation(false)
    }
  }

  if (!position || !selectedWord) {
    return null
  }

  return (
    <div
      className="word-translator-tooltip fixed z-50 rounded-3xl border border-slate-200 bg-white p-4 shadow-2xl"
      style={
        isMobileView
          ? {
              left: '12px',
              right: '12px',
              bottom: '12px',
            }
          : {
              width: `${tooltipWidth}px`,
              left: `${tooltipLeft}px`,
              top: `${tooltipTop}px`,
            }
      }
    >
      {loading ? (
        <div className="flex items-center space-x-2">
          <div className="h-4 w-4 animate-spin rounded-full border-b-2 border-link-light" />
          <span className="text-sm text-slate-600">Загружаю перевод...</span>
        </div>
      ) : (
        <div className="space-y-4">
          <div className="flex items-start justify-between gap-3 border-b border-slate-200 pb-3">
            <div className="min-w-0">
              <h3 className="break-words text-lg font-bold text-slate-800 sm:text-xl">{translation?.word || selectedWord}</h3>
              {translation?.phonetic && <p className="mt-1 text-sm text-slate-500">[{translation.phonetic}]</p>}
            </div>

            <div className="flex shrink-0 items-center gap-2">
              {translation?.audio_url && (
                <button
                  type="button"
                  onClick={handlePlayAudio}
                  className="rounded-full bg-emerald-100 px-3 py-1 text-sm font-semibold text-emerald-700 transition-colors hover:bg-emerald-200"
                  title="Прослушать"
                >
                  Аудио
                </button>
              )}
              <button
                type="button"
                onClick={closeTooltip}
                className="text-xl text-slate-400 transition-colors hover:text-slate-600"
              >
                ×
              </button>
            </div>
          </div>

          <div className="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3">
            <p className="text-xs font-semibold uppercase tracking-[0.18em] text-slate-400">Базовый перевод</p>
            <p className="mt-2 break-words text-base font-medium text-slate-700">
              {translation?.translation || 'Перевод не найден'}
            </p>
            {translation?.example && <p className="mt-2 text-sm italic text-slate-500">"{translation.example}"</p>}
          </div>

          <div className="space-y-3 rounded-2xl border border-emerald-200 bg-emerald-50 px-4 py-3">
            <div className="flex items-center justify-between gap-3">
              <p className="text-xs font-semibold uppercase tracking-[0.18em] text-emerald-700">Мои переводы</p>
              <span className="rounded-full bg-white px-2 py-0.5 text-[11px] font-semibold text-emerald-700">
                {personalTranslations.length}
              </span>
            </div>

            {hasPersonalTranslations ? (
              <div className="space-y-2">
                {sortedPersonalTranslations.map((item) => (
                  <div
                    key={item.id}
                    className="flex items-center justify-between gap-3 rounded-2xl border border-emerald-200 bg-white px-3 py-2"
                  >
                    <span className="min-w-0 break-words text-sm font-medium text-slate-700">{item.translation}</span>
                    <button
                      type="button"
                      onClick={() => void handleDeleteTranslation(item.id)}
                      className="shrink-0 text-sm font-semibold text-red-500 transition-colors hover:text-red-700"
                      title="Удалить перевод"
                    >
                      ×
                    </button>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-sm text-slate-500">Пока нет ваших личных переводов для этого слова.</p>
            )}

            <div className="flex flex-col gap-2 sm:flex-row">
              <input
                type="text"
                value={newTranslation}
                onChange={(event) => setNewTranslation(event.target.value)}
                placeholder="Добавить свой перевод"
                className="min-w-0 flex-1 rounded-2xl border border-emerald-200 bg-white px-3 py-2 text-sm focus:border-emerald-500 focus:outline-none"
              />
              <button
                type="button"
                onClick={() => void handleAddTranslation()}
                disabled={savingTranslation || !newTranslation.trim()}
                className="rounded-2xl bg-emerald-600 px-4 py-2 text-sm font-semibold text-white transition-colors hover:bg-emerald-700 disabled:cursor-not-allowed disabled:opacity-50"
              >
                +
              </button>
            </div>
          </div>

          {translationMessage && <p className="text-sm text-slate-500">{translationMessage}</p>}
        </div>
      )}
    </div>
  )
}
