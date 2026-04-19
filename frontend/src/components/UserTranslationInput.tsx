import React, { useEffect, useState } from 'react'
import { wordTranslationService } from '../services/wordTranslationService'

interface Props {
  word: string
  autoTranslation?: string
}

export default function UserTranslationInput({ word, autoTranslation }: Props) {
  const [myTranslation, setMyTranslation] = useState('')
  const [saved, setSaved] = useState(false)
  const [loading, setLoading] = useState(false)
  const [editing, setEditing] = useState(false)

  useEffect(() => {
    if (!word) return
    setSaved(false)
    setEditing(false)
    setMyTranslation('')
    void loadExisting()
  }, [word])

  const loadExisting = async () => {
    try {
      const existing = await wordTranslationService.get(word)
      if (existing) {
        setMyTranslation(existing.translation)
        setSaved(true)
      }
    } catch {
      // тихо игнорируем
    }
  }

  const handleSave = async () => {
    if (!myTranslation.trim()) return
    try {
      setLoading(true)
      await wordTranslationService.upsert(word, myTranslation.trim())
      setSaved(true)
      setEditing(false)
    } catch {
      // тихо игнорируем
    } finally {
      setLoading(false)
    }
  }

  const handleDelete = async () => {
    try {
      setLoading(true)
      await wordTranslationService.delete(word)
      setMyTranslation('')
      setSaved(false)
      setEditing(false)
    } catch {
      // тихо игнорируем
    } finally {
      setLoading(false)
    }
  }

  if (!word) return null

  return (
    <div className="rounded-2xl border border-blue-100 bg-blue-50/60 px-4 py-3">
      <p className="text-xs font-semibold uppercase tracking-[0.16em] text-blue-500">Мой перевод</p>

      {saved && !editing ? (
        <div className="mt-2 flex items-center justify-between gap-3">
          <p className="text-base font-medium text-slate-700">{myTranslation}</p>
          <div className="flex gap-2">
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
              Авто: <span className="italic">{autoTranslation}</span>
              {' '}
              <button
                type="button"
                onClick={() => setMyTranslation(autoTranslation)}
                className="text-blue-500 underline hover:text-blue-700"
              >
                использовать
              </button>
            </p>
          )}
          <div className="flex gap-2">
            <input
              type="text"
              value={myTranslation}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => setMyTranslation(e.target.value)}
              onKeyDown={(e: React.KeyboardEvent<HTMLInputElement>) => e.key === 'Enter' && void handleSave()}
              placeholder="Введите свой перевод..."
              className="flex-1 rounded-xl border border-blue-200 px-3 py-1.5 text-sm focus:border-blue-400 focus:outline-none"
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
                onClick={() => { setEditing(false); void loadExisting() }}
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
