import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { config } from '../config'
import FileUpload from '../components/FileUpload'
import { adminService, Card, CreateDeckRequest, Deck } from '../services/adminService'
import { dictionaryService, WordInfo } from '../services/dictionaryService'
import { useAuthStore } from '../store/authStore'
import { slugify } from '../utils/slug'

interface CardEditorState {
  word: string
  translation: string
  audio_url: string
  image_url: string
}

const emptyCardForm: CardEditorState = {
  word: '',
  translation: '',
  audio_url: '',
  image_url: '',
}

export default function AdminDecks() {
  const { courseId } = useParams<{ courseId: string }>()
  const { user, isAuthenticated } = useAuthStore()

  const [course, setCourse] = useState<{ id: number; title: string } | null>(null)
  const [decks, setDecks] = useState<Deck[]>([])
  const [selectedDeck, setSelectedDeck] = useState<Deck | null>(null)
  const [cards, setCards] = useState<Card[]>([])
  const [loading, setLoading] = useState(true)

  const [showDeckForm, setShowDeckForm] = useState(false)
  const [showCardForm, setShowCardForm] = useState(false)
  const [editingDeck, setEditingDeck] = useState<Deck | null>(null)
  const [editingCard, setEditingCard] = useState<Card | null>(null)

  const [deckForm, setDeckForm] = useState<CreateDeckRequest>({
    course_id: Number(courseId || 0),
    title: '',
    slug: '',
    description: '',
    position: 0,
  })
  const [cardForm, setCardForm] = useState<CardEditorState>(emptyCardForm)
  const [slugTouched, setSlugTouched] = useState(false)
  const [searchingWord, setSearchingWord] = useState(false)
  const [autoSuggestions, setAutoSuggestions] = useState(true)
  const [searchTimeout, setSearchTimeout] = useState<ReturnType<typeof setTimeout> | null>(null)

  useEffect(() => {
    if (isAuthenticated && user?.role === 'admin' && courseId) {
      void loadCourse()
      void loadDecks()
    }
  }, [courseId, isAuthenticated, user])

  useEffect(() => {
    return () => {
      if (searchTimeout) {
        clearTimeout(searchTimeout)
      }
    }
  }, [searchTimeout])

  const loadCourse = async () => {
    try {
      const courseData = await adminService.getCourse(Number(courseId || 0))
      setCourse({ id: courseData.id, title: courseData.title })
    } catch (error) {
      console.error('Error loading course:', error)
    }
  }

  const loadDecks = async () => {
    if (!courseId) {
      setLoading(false)
      return
    }

    try {
      setLoading(true)
      const data = await adminService.getDecksByCourse(Number(courseId))
      setDecks(data || [])
    } catch (error) {
      console.error('Error loading decks:', error)
      setDecks([])
    } finally {
      setLoading(false)
    }
  }

  const loadCards = async (deckId: number) => {
    try {
      const data = await adminService.getCardsByDeck(deckId)
      setCards(data || [])
    } catch (error) {
      console.error('Error loading cards:', error)
      setCards([])
    }
  }

  const resetDeckForm = () => {
    setDeckForm({
      course_id: Number(courseId || 0),
      title: '',
      slug: '',
      description: '',
      position: 0,
    })
    setEditingDeck(null)
    setSlugTouched(false)
  }

  const resetCardForm = () => {
    setCardForm(emptyCardForm)
    setEditingCard(null)
  }

  const handleDeckTitleChange = (title: string) => {
    setDeckForm((current) => ({
      ...current,
      title,
      slug: slugTouched ? current.slug : slugify(title),
    }))
  }

  const handleCreateDeck = async (event: React.FormEvent) => {
    event.preventDefault()

    try {
      if (editingDeck) {
        await adminService.updateDeck(editingDeck.id, {
          title: deckForm.title,
          slug: deckForm.slug || undefined,
          description: deckForm.description || undefined,
          position: deckForm.position,
        })
      } else {
        await adminService.createDeck(deckForm)
      }

      resetDeckForm()
      setShowDeckForm(false)
      await loadDecks()
    } catch (error) {
      console.error('Error saving deck:', error)
    }
  }

  const handleEditDeck = (deck: Deck) => {
    setEditingDeck(deck)
    setDeckForm({
      course_id: deck.course_id,
      title: deck.title,
      slug: deck.slug || '',
      description: deck.description || '',
      position: deck.position,
    })
    setSlugTouched(true)
    setShowDeckForm(true)
  }

  const applyWordInfo = (info: WordInfo, overwrite: boolean) => {
    setCardForm((current) => ({
      word: info.word || current.word,
      translation: overwrite || !current.translation ? info.translation || current.translation : current.translation,
      audio_url: overwrite || !current.audio_url ? info.audio_url || current.audio_url : current.audio_url,
      image_url: overwrite || !current.image_url ? info.image_url || current.image_url : current.image_url,
    }))
  }

  const handleSearchWord = async (wordToSearch?: string, overwrite = false) => {
    const normalizedWord = (wordToSearch || cardForm.word).trim().toLowerCase()
    if (normalizedWord.length < 2) {
      return
    }

    try {
      setSearchingWord(true)
      const info = await dictionaryService.getWordInfo(normalizedWord)
      applyWordInfo(info, overwrite)
    } catch (error) {
      console.error('Error searching word:', error)
    } finally {
      setSearchingWord(false)
    }
  }

  const handleWordChange = (value: string) => {
    setCardForm((current) => ({ ...current, word: value }))

    if (searchTimeout) {
      clearTimeout(searchTimeout)
    }

    if (autoSuggestions && value.trim().length >= 2) {
      const timeout = setTimeout(() => {
        void handleSearchWord(value, false)
      }, 700)
      setSearchTimeout(timeout)
    }
  }

  const getMediaUrl = (value?: string) => {
    if (!value) {
      return ''
    }
    return config.getFullUrl(value)
  }

  const playAudio = (value?: string) => {
    const audioUrl = getMediaUrl(value)
    if (!audioUrl) {
      return
    }

    const audio = new Audio(audioUrl)
    audio.play().catch((error) => console.error('Error playing audio:', error))
  }

  const normalizeOptionalValue = (value: string) => {
    const trimmed = value.trim()
    return trimmed === '' ? undefined : trimmed
  }

  const handleCreateCard = async (event: React.FormEvent) => {
    event.preventDefault()
    if (!selectedDeck) {
      return
    }

    const payload = {
      deck_id: selectedDeck.id,
      word: cardForm.word.trim(),
      translation: cardForm.translation.trim(),
      audio_url: normalizeOptionalValue(cardForm.audio_url),
      image_url: normalizeOptionalValue(cardForm.image_url),
    }

    try {
      if (editingCard) {
        await adminService.updateCard(editingCard.id, payload)
      } else {
        await adminService.createCard(payload)
      }

      resetCardForm()
      setShowCardForm(false)
      await loadCards(selectedDeck.id)
    } catch (error) {
      console.error('Error saving card:', error)
    }
  }

  const handleEditCard = (card: Card) => {
    setEditingCard(card)
    setCardForm({
      word: card.word,
      translation: card.translation,
      audio_url: card.audio_url || '',
      image_url: card.image_url || '',
    })
    setShowCardForm(true)
  }

  const handleDeleteDeck = async (deckId: number) => {
    if (!window.confirm('Удалить подкурс? Это действие нельзя отменить.')) {
      return
    }

    try {
      await adminService.deleteDeck(deckId)
      if (selectedDeck?.id === deckId) {
        setSelectedDeck(null)
        setCards([])
        setShowCardForm(false)
        resetCardForm()
      }
      await loadDecks()
    } catch (error) {
      console.error('Error deleting deck:', error)
    }
  }

  const handleDeleteCard = async (cardId: number) => {
    if (!window.confirm('Удалить карточку?')) {
      return
    }

    try {
      await adminService.deleteCard(cardId)
      if (selectedDeck) {
        await loadCards(selectedDeck.id)
      }
    } catch (error) {
      console.error('Error deleting card:', error)
    }
  }

  if (!isAuthenticated || !user) {
    return <div className="py-8 text-center text-text-light">Проверка доступа...</div>
  }

  if (user.role !== 'admin') {
    return (
      <div className="rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-red-700">
        У вас нет доступа к этой странице. Нужна роль администратора.
      </div>
    )
  }

  if (!courseId) {
    return (
      <div className="rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-red-700">
        Не указан ID курса.
      </div>
    )
  }

  return (
    <div className="w-full">
      {loading && decks.length === 0 ? (
        <div className="py-8 text-center text-text-light">Загрузка...</div>
      ) : (
        <>
          <div className="mb-6 flex flex-col gap-5 rounded-[28px] border border-rose-100 bg-gradient-to-br from-white via-rose-50 to-orange-50 p-6 shadow-xl lg:flex-row lg:items-center lg:justify-between lg:p-8">
            <div>
              <Link to="/admin/courses" className="mb-2 inline-block text-link-light transition-colors hover:text-link-dark">
                ← Назад к курсам
              </Link>
              <h1 className="text-3xl font-bold text-text-light">
                Карточки и подкурсы
                {course && <span className="ml-2 text-lg font-normal text-gray-500">· {course.title}</span>}
              </h1>
              <p className="mt-2 max-w-2xl text-sm leading-6 text-slate-600">
                При вводе слова форма автоматически подсказывает перевод, озвучку и картинку. Если нужен свой файл, просто загрузите его ниже.
              </p>
            </div>

            <button
              type="button"
              onClick={() => {
                if (showDeckForm) {
                  setShowDeckForm(false)
                  resetDeckForm()
                  return
                }
                resetDeckForm()
                setShowDeckForm(true)
              }}
              className="rounded-2xl bg-link-light px-5 py-3 font-semibold text-white transition-colors hover:bg-link-dark"
            >
              {showDeckForm ? 'Скрыть форму' : '+ Создать подкурс'}
            </button>
          </div>

          {showDeckForm && (
            <div className="mb-6 rounded-[28px] border border-gray-200 bg-card-light p-6 shadow-md">
              <h2 className="mb-4 text-xl font-semibold text-text-light">
                {editingDeck ? 'Редактировать подкурс' : 'Новый подкурс'}
              </h2>
              <form onSubmit={handleCreateDeck} className="space-y-4">
                <input
                  type="text"
                  placeholder="Название подкурса"
                  required
                  value={deckForm.title}
                  onChange={(event) => handleDeckTitleChange(event.target.value)}
                  className="w-full rounded-2xl border border-gray-300 px-4 py-3 focus:border-link-light focus:outline-none focus:ring-2 focus:ring-link-light"
                />
                <input
                  type="text"
                  placeholder="Slug"
                  required
                  value={deckForm.slug || ''}
                  onChange={(event) => {
                    setSlugTouched(true)
                    setDeckForm((current) => ({ ...current, slug: slugify(event.target.value) }))
                  }}
                  className="w-full rounded-2xl border border-gray-300 px-4 py-3 focus:border-link-light focus:outline-none focus:ring-2 focus:ring-link-light"
                />
                <textarea
                  placeholder="Короткое описание"
                  value={deckForm.description || ''}
                  onChange={(event) => setDeckForm((current) => ({ ...current, description: event.target.value }))}
                  rows={2}
                  className="w-full rounded-2xl border border-gray-300 px-4 py-3 focus:border-link-light focus:outline-none focus:ring-2 focus:ring-link-light"
                />
                <input
                  type="number"
                  min={0}
                  placeholder="Позиция"
                  value={deckForm.position || 0}
                  onChange={(event) =>
                    setDeckForm((current) => ({
                      ...current,
                      position: Number(event.target.value || 0),
                    }))
                  }
                  className="w-full rounded-2xl border border-gray-300 px-4 py-3 focus:border-link-light focus:outline-none focus:ring-2 focus:ring-link-light"
                />
                <div className="flex gap-3">
                  <button type="submit" className="rounded-2xl bg-accent-light px-5 py-3 text-white transition-colors hover:bg-accent-dark">
                    {editingDeck ? 'Сохранить' : 'Создать'}
                  </button>
                  <button
                    type="button"
                    onClick={() => {
                      setShowDeckForm(false)
                      resetDeckForm()
                    }}
                    className="rounded-2xl bg-gray-200 px-5 py-3 text-text-light transition-colors hover:bg-gray-300"
                  >
                    Отмена
                  </button>
                </div>
              </form>
            </div>
          )}

          <div className="grid grid-cols-1 gap-6 lg:grid-cols-[0.95fr_1.05fr]">
            <section className="rounded-[28px] border border-gray-200 bg-card-light p-6 shadow-md">
              <h2 className="mb-4 text-xl font-semibold text-text-light">Подкурсы</h2>

              {decks.length === 0 ? (
                <p className="text-text-light">Подкурсов пока нет.</p>
              ) : (
                <div className="space-y-3">
                  {decks.map((deck) => (
                    <button
                      key={deck.id}
                      type="button"
                      onClick={() => {
                        setSelectedDeck(deck)
                        setShowCardForm(false)
                        resetCardForm()
                        void loadCards(deck.id)
                      }}
                      className={`w-full rounded-[24px] border-2 p-4 text-left transition-all ${
                        selectedDeck?.id === deck.id
                          ? 'border-link-light bg-rose-50 shadow-md'
                          : 'border-gray-200 bg-white hover:border-link-light hover:shadow-sm'
                      }`}
                    >
                      <div className="flex items-start justify-between gap-3">
                        <div>
                          <div className="font-semibold text-text-light">{deck.title}</div>
                          {deck.description && <div className="mt-1 text-sm text-slate-500">{deck.description}</div>}
                          <div className="mt-2 text-xs text-slate-400">
                            slug: {deck.slug} · позиция: {deck.position}
                          </div>
                        </div>

                        <div className="flex gap-2">
                          <span
                            onClick={(event) => {
                              event.stopPropagation()
                              handleEditDeck(deck)
                            }}
                            className="cursor-pointer rounded-xl px-2 py-1 text-sm text-link-light transition-colors hover:bg-rose-100 hover:text-link-dark"
                            title="Редактировать"
                          >
                            ✏️
                          </span>
                          <span
                            onClick={(event) => {
                              event.stopPropagation()
                              void handleDeleteDeck(deck.id)
                            }}
                            className="cursor-pointer rounded-xl px-2 py-1 text-sm text-logo-bright transition-colors hover:bg-red-50 hover:text-logo-dark"
                            title="Удалить"
                          >
                            🗑️
                          </span>
                        </div>
                      </div>
                    </button>
                  ))}
                </div>
              )}
            </section>

            <section className="rounded-[28px] border border-gray-200 bg-card-light p-6 shadow-md">
              <div className="mb-4 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                <div>
                  <h2 className="text-xl font-semibold text-text-light">
                    Карточки {selectedDeck ? `· ${selectedDeck.title}` : ''}
                  </h2>
                  {selectedDeck && (
                    <p className="mt-1 text-sm text-slate-500">
                      Подкурс ID: {selectedDeck.id} · карточек: {cards.length}
                    </p>
                  )}
                </div>

                {selectedDeck && (
                  <button
                    type="button"
                    onClick={() => {
                      if (showCardForm) {
                        setShowCardForm(false)
                        resetCardForm()
                        return
                      }
                      resetCardForm()
                      setShowCardForm(true)
                    }}
                    className="rounded-2xl bg-link-light px-4 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-link-dark"
                  >
                    {showCardForm ? 'Скрыть форму' : '+ Добавить карточку'}
                  </button>
                )}
              </div>

              {!selectedDeck ? (
                <div className="rounded-[24px] border border-dashed border-slate-300 bg-slate-50 px-6 py-10 text-center text-slate-500">
                  Выберите подкурс слева, чтобы редактировать его карточки.
                </div>
              ) : showCardForm ? (
                <form onSubmit={handleCreateCard} className="space-y-4 rounded-[24px] border border-slate-200 bg-white p-5">
                  <div className="flex items-center justify-between gap-3">
                    <h3 className="text-lg font-semibold text-text-light">
                      {editingCard ? 'Редактирование карточки' : 'Новая карточка'}
                    </h3>
                    <label className="flex items-center gap-2 text-sm text-slate-600">
                      <input
                        type="checkbox"
                        checked={autoSuggestions}
                        onChange={(event) => setAutoSuggestions(event.target.checked)}
                      />
                      Автоподсказки
                    </label>
                  </div>

                  <div className="flex gap-2">
                    <div className="relative flex-1">
                      <input
                        type="text"
                        placeholder="Английское слово"
                        required
                        value={cardForm.word}
                        onChange={(event) => handleWordChange(event.target.value)}
                        className="w-full rounded-2xl border border-gray-300 px-4 py-3 pr-10 text-sm focus:border-link-light focus:outline-none focus:ring-2 focus:ring-link-light"
                      />
                      {searchingWord && (
                        <span className="absolute right-3 top-1/2 -translate-y-1/2 text-link-light">⏳</span>
                      )}
                    </div>
                    <button
                      type="button"
                      onClick={() => void handleSearchWord(undefined, true)}
                      disabled={!cardForm.word.trim() || searchingWord}
                      className="rounded-2xl bg-link-light px-4 py-2 text-sm font-semibold text-white transition-colors hover:bg-link-dark disabled:cursor-not-allowed disabled:bg-gray-400"
                    >
                      Из API
                    </button>
                  </div>

                  <p className="text-xs leading-5 text-slate-500">
                    По слову подтягиваем перевод, озвучку и картинку. Если нужен свой вариант, просто замените URL или загрузите файл.
                  </p>

                  <input
                    type="text"
                    placeholder="Перевод"
                    required
                    value={cardForm.translation}
                    onChange={(event) => setCardForm((current) => ({ ...current, translation: event.target.value }))}
                    className="w-full rounded-2xl border border-gray-300 px-4 py-3 text-sm focus:border-link-light focus:outline-none focus:ring-2 focus:ring-link-light"
                  />

                  <div className="rounded-[22px] border border-rose-100 bg-rose-50/70 p-4">
                    <div className="mb-3 flex items-center justify-between gap-3">
                      <div>
                        <div className="text-sm font-semibold text-text-light">Озвучка</div>
                        <p className="text-xs text-slate-500">Можно оставить URL из API или загрузить свой аудиофайл.</p>
                      </div>
                      {cardForm.audio_url && (
                        <button
                          type="button"
                          onClick={() => playAudio(cardForm.audio_url)}
                          className="rounded-xl bg-emerald-600 px-3 py-2 text-sm font-semibold text-white transition-colors hover:bg-emerald-700"
                        >
                          Прослушать
                        </button>
                      )}
                    </div>

                    <FileUpload
                      type="audio"
                      currentUrl={cardForm.audio_url}
                      onUrlChange={(url) => setCardForm((current) => ({ ...current, audio_url: url }))}
                      label="Файл или URL озвучки"
                      placeholder="https://...mp3"
                    />
                  </div>

                  <div className="rounded-[22px] border border-orange-100 bg-orange-50/70 p-4">
                    <div className="mb-3">
                      <div className="text-sm font-semibold text-text-light">Картинка</div>
                      <p className="text-xs text-slate-500">Автоподбор подставляет внешний URL, но вы всегда можете загрузить своё изображение.</p>
                    </div>

                    <FileUpload
                      type="image"
                      currentUrl={cardForm.image_url}
                      onUrlChange={(url) => setCardForm((current) => ({ ...current, image_url: url }))}
                      label="Файл или URL изображения"
                      placeholder="https://example.com/image.jpg"
                    />
                  </div>

                  <div className="flex gap-3">
                    <button type="submit" className="rounded-2xl bg-accent-light px-5 py-3 text-sm font-semibold text-white transition-colors hover:bg-accent-dark">
                      {editingCard ? 'Сохранить карточку' : 'Создать карточку'}
                    </button>
                    <button
                      type="button"
                      onClick={() => {
                        setShowCardForm(false)
                        resetCardForm()
                      }}
                      className="rounded-2xl bg-gray-200 px-5 py-3 text-sm font-semibold text-text-light transition-colors hover:bg-gray-300"
                    >
                      Отмена
                    </button>
                  </div>
                </form>
              ) : cards.length === 0 ? (
                <div className="rounded-[24px] border border-dashed border-slate-300 bg-slate-50 px-6 py-10 text-center">
                  <p className="text-text-light">В этом подкурсе пока нет карточек.</p>
                  <button
                    type="button"
                    onClick={() => setShowCardForm(true)}
                    className="mt-4 rounded-2xl bg-link-light px-5 py-3 text-white transition-colors hover:bg-link-dark"
                  >
                    + Добавить первую карточку
                  </button>
                </div>
              ) : (
                <div className="space-y-3">
                  {cards.map((card) => (
                    <div
                      key={card.id}
                      className="rounded-[24px] border-2 border-gray-200 bg-white p-4 transition-all hover:border-link-light hover:shadow-sm"
                    >
                      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
                        <div className="flex gap-4">
                          {card.image_url && (
                            <img
                              src={getMediaUrl(card.image_url)}
                              alt={card.word}
                              className="h-20 w-20 rounded-2xl object-cover"
                              onError={(event) => {
                                ;(event.target as HTMLImageElement).style.display = 'none'
                              }}
                            />
                          )}

                          <div>
                            <div className="flex flex-wrap items-center gap-2">
                              <div className="text-lg font-bold text-text-light">{card.word}</div>
                              {card.audio_url && (
                                <button
                                  type="button"
                                  onClick={() => playAudio(card.audio_url)}
                                  className="rounded-xl bg-emerald-50 px-3 py-1 text-xs font-semibold text-emerald-700 transition-colors hover:bg-emerald-100"
                                >
                                  Озвучка
                                </button>
                              )}
                              {card.image_url && (
                                <span className="rounded-xl bg-orange-50 px-3 py-1 text-xs font-semibold text-orange-700">
                                  Картинка
                                </span>
                              )}
                            </div>
                            <div className="mt-2 text-base font-medium text-slate-700">{card.translation}</div>
                          </div>
                        </div>

                        <div className="flex gap-2">
                          <button
                            type="button"
                            onClick={() => handleEditCard(card)}
                            className="rounded-xl px-3 py-2 text-sm text-link-light transition-colors hover:bg-rose-50 hover:text-link-dark"
                          >
                            Редактировать
                          </button>
                          <button
                            type="button"
                            onClick={() => void handleDeleteCard(card.id)}
                            className="rounded-xl px-3 py-2 text-sm text-logo-bright transition-colors hover:bg-red-50 hover:text-logo-dark"
                          >
                            Удалить
                          </button>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </section>
          </div>
        </>
      )}
    </div>
  )
}
