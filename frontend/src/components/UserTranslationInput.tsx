import React, { useEffect, useState } from 'react'
import { PersonalTranslation, personalTranslationService } from '../services/personalTranslationService'

interface Props {
  word: string
  autoTranslation?: string
  onTranslationChange?: (translation: string | null) => void
}

function getLatestTranslation(translations: PersonalTranslation[]): PersonalTranslation | null {
  if (!translations.length) {
    return null
  }

  return [...translations].sort((left, right) => {
    const leftTime = new Date(left.updated_at || left.created_at).getTime()
    const rightTime = new Date(right.updated_at || right.created_at).getTime()
    return rightTime - leftTime
  })[0]
}

export default function UserTranslationInput({ word, autoTranslation, onTranslationChange }: Props) {
  const [myTranslation, setMyTranslation] = useState('')
  const [savedTranslation, setSavedTranslation] = useState<PersonalTranslation | null>(null)
  const [loading, setLoading] = useState(false)
  const [editing, setEditing] = useState(false)

  useEffect(() => {
    if (!word) {
      onTranslationChange?.(null)
      return
    }

    setEditing(false)
    setMyTranslation('')
    setSavedTranslation(null)
    void loadExisting()
  }, [word])

  const loadExisting = async () => {
    try {
      const existingTranslations = await personalTranslationService.getAll(word)
      const latestTranslation = getLatestTranslation(existingTranslations)

      setSavedTranslation(latestTranslation)
      setMyTranslation(latestTranslation?.translation || '')
      onTranslationChange?.(latestTranslation?.translation || null)
    } catch {
      onTranslationChange?.(null)
    }
  }

  const handleSave = async () => {
    const trimmedTranslation = myTranslation.trim()
    if (!trimmedTranslation) return

    try {
      setLoading(true)
      const saved = await personalTranslationService.create(word, trimmedTranslation)
      setSavedTranslation(saved)
      setMyTranslation(saved.translation)
      setEditing(false)
      onTranslationChange?.(saved.translation)
    } catch {
      // Игнорируем ошибку и оставляем текущее состояние формы
    } finally {
      setLoading(false)
    }
  }

  const handleDelete = async () => {
    if (!savedTranslation) return

    try {
      setLoading(true)
      await personalTranslationService.remove(savedTranslation.id)
      setSavedTranslation(null)
      setMyTranslation('')
      setEditing(false)
      onTranslationChange?.(null)
    } catch {
      // Игнорируем ошибку удаления
    } finally {
      setLoading(false)
    }
  }

  if (!word) return null

  const isSaved = Boolean(savedTranslation)

  return (
    <div className="rounded-2xl border border-blue-100 bg-blue-50/60 px-4 py-3">
      <p className="text-xs font-semibold uppercase tracking-[0.16em] text-blue-500">Мой перевод</p>

      {isSaved && !editing ? (
        <div className="mt-2 flex flex-wrap items-center justify-between gap-3">
          <p className="min-w-0 text-base font-medium text-slate-700">{savedTranslation?.translation}</p>
          <div className="flex flex-wrap gap-2">
            <button
              type="button"
              onClick={() => setEditing(true)}
              className="rounded-full bg-white px-3 py-1 text-xs font-semibold text-blue-600 transition-colors hover:bg-blue-100"
            >
              Изменить
            </button>
            <button
              type="button"
              onClick={() => void handleDelete()}
              disabled={loading}
              className="rounded-full bg-white px-3 py-1 text-xs font-semibold text-red-500 transition-colors hover:bg-red-50 disabled:opacity-50"
            >
              Удалить
            </button>
          </div>
        </div>
      ) : (
        <div className="mt-2 space-y-2">
          {autoTranslation && (
            <p className="text-xs text-slate-500">
              Авто: <span className="italic">{autoTranslation}</span>{' '}
              <button
                type="button"
                onClick={() => setMyTranslation(autoTranslation)}
                className="text-blue-500 underline hover:text-blue-700"
              >
                использовать
              </button>
            </p>
          )}
          <div className="flex flex-wrap gap-2">
            <input
              type="text"
              value={myTranslation}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => setMyTranslation(e.target.value)}
              onKeyDown={(e: React.KeyboardEvent<HTMLInputElement>) => e.key === 'Enter' && void handleSave()}
              placeholder="Введите свой перевод..."
              className="min-w-0 flex-1 rounded-xl border border-blue-200 px-3 py-1.5 text-sm focus:border-blue-400 focus:outline-none"
            />
            <button
              type="button"
              onClick={() => void handleSave()}
              disabled={loading || !myTranslation.trim()}
              className="rounded-xl bg-blue-500 px-3 py-1.5 text-sm font-semibold text-white transition-colors hover:bg-blue-600 disabled:opacity-50"
            >
              {loading ? '...' : 'Сохранить'}
            </button>
            {editing && (
              <button
                type="button"
                onClick={() => {
                  setEditing(false)
                  void loadExisting()
                }}
                className="rounded-xl border border-slate-200 px-3 py-1.5 text-sm text-slate-500 hover:bg-slate-50"
              >
                Отмена
              </button>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
