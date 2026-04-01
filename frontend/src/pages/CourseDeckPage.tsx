import { useEffect, useMemo, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import StudySessionModal from '../components/StudySessionModal'
import { config } from '../config'
import { useAuthStore } from '../store/authStore'
import { Card, Course, Deck, UserCourse, UserDeck, studyService } from '../services/studyService'

type DeckState = 'completed' | 'active' | 'available' | 'locked'

interface DeckPresentation {
  deck: Deck
  userDeck?: UserDeck
  progress: number
  learnedCards: number
  totalCards: number
  state: DeckState
  sharePath: string
  lockReason?: string
}

function normalizeProgress(value?: number) {
  return Math.max(0, Math.min(100, Math.round(Number(value || 0))))
}

function sortUserDecks(userDecks: UserDeck[]) {
  return [...userDecks].sort((left, right) => {
    const leftValue = new Date(left.updated_at || left.created_at || left.last_opened_at || 0).getTime()
    const rightValue = new Date(right.updated_at || right.created_at || right.last_opened_at || 0).getTime()
    return rightValue - leftValue
  })
}

function buildLatestUserDeckMap(userDecks: UserDeck[]) {
  const latestDeckMap = new Map<number, UserDeck>()

  sortUserDecks(userDecks).forEach((userDeck) => {
    if (!latestDeckMap.has(userDeck.deck_id)) {
      latestDeckMap.set(userDeck.deck_id, userDeck)
    }
  })

  return latestDeckMap
}

function buildDeckPresentations(
  course: Course | null,
  decks: Deck[],
  userDecks: UserDeck[],
  activeDeckId?: number
): DeckPresentation[] {
  const latestUserDeckMap = buildLatestUserDeckMap(userDecks)
  const firstIncompleteIndex = decks.findIndex((deck) => normalizeProgress(latestUserDeckMap.get(deck.id)?.progress_percentage) < 100)
  const activeDeckIndex = activeDeckId ? decks.findIndex((deck) => deck.id === activeDeckId) : -1

  let unlockedBoundary = firstIncompleteIndex === -1 ? decks.length - 1 : firstIncompleteIndex
  if (activeDeckIndex > unlockedBoundary) {
    unlockedBoundary = activeDeckIndex
  }

  return decks.map((deck, index) => {
    const userDeck = latestUserDeckMap.get(deck.id)
    const progress = normalizeProgress(userDeck?.progress_percentage)
    const learnedCards = userDeck?.learned_cards_count || 0
    const totalCards = userDeck?.total_cards_count || 0
    const isCompleted = progress >= 100 || userDeck?.status === 'completed'
    const isActive = activeDeckId === deck.id
    const isUnlocked = isCompleted || isActive || index <= unlockedBoundary

    let state: DeckState = 'locked'
    if (isCompleted) {
      state = 'completed'
    } else if (isActive) {
      state = 'active'
    } else if (isUnlocked) {
      state = 'available'
    }

    return {
      deck,
      userDeck,
      progress,
      learnedCards,
      totalCards,
      state,
      sharePath: course?.slug ? `/deck/${course.slug}/${deck.slug}` : `/courses/${course?.id ?? deck.course_id}`,
      lockReason:
        state === 'locked' && index > 0
          ? `Сначала завершите "${decks[index - 1].title}".`
          : 'Этот дек откроется после предыдущего.',
    }
  })
}

export default function CourseDeckPage() {
  const navigate = useNavigate()
  const { user } = useAuthStore()
  const { id, courseSlug, deckSlug } = useParams<{
    id?: string
    courseSlug?: string
    deckSlug?: string
  }>()

  const [course, setCourse] = useState<Course | null>(null)
  const [decks, setDecks] = useState<Deck[]>([])
  const [userDecks, setUserDecks] = useState<UserDeck[]>([])
  const [selectedDeck, setSelectedDeck] = useState<Deck | null>(null)
  const [cards, setCards] = useState<Card[]>([])
  const [activeDeck, setActiveDeck] = useState<UserDeck | null>(null)
  const [activeCourse, setActiveCourse] = useState<UserCourse | null>(null)
  const [loading, setLoading] = useState(true)
  const [loadingCards, setLoadingCards] = useState(false)
  const [showStudyModal, setShowStudyModal] = useState(false)
  const [selectionMessage, setSelectionMessage] = useState<string | null>(null)

  const deckPresentations = useMemo(
    () => buildDeckPresentations(course, decks, userDecks, activeDeck?.deck_id),
    [course, decks, userDecks, activeDeck?.deck_id]
  )

  const selectedDeckPresentation = useMemo(
    () => deckPresentations.find((item) => item.deck.id === selectedDeck?.id) || null,
    [deckPresentations, selectedDeck?.id]
  )

  const selectedDeckCards = useMemo(() => {
    const regular = cards.filter((card) => !card.is_custom || card.created_by !== user?.id)
    const custom = cards.filter((card) => card.is_custom && card.created_by === user?.id)
    return { regular, custom }
  }, [cards, user?.id])

  const courseStats = useMemo(() => {
    const completedDecks = deckPresentations.filter((deck) => deck.state === 'completed').length
    const learnedWords = deckPresentations.reduce((sum, deck) => sum + deck.learnedCards, 0)
    const totalWords = deckPresentations.reduce((sum, deck) => sum + deck.totalCards, 0)
    const averageProgress =
      deckPresentations.length > 0
        ? Math.round(deckPresentations.reduce((sum, deck) => sum + deck.progress, 0) / deckPresentations.length)
        : 0

    return {
      completedDecks,
      totalDecks: deckPresentations.length,
      learnedWords,
      totalWords,
      averageProgress,
    }
  }, [deckPresentations])

  useEffect(() => {
    void loadCourseData()
  }, [id, courseSlug, deckSlug, user?.id])

  const loadCards = async (deckId: number) => {
    try {
      setLoadingCards(true)
      const loadedCards = await studyService.getCardsByDeck(deckId)
      setCards(loadedCards)
    } catch (error) {
      console.error('Error loading cards:', error)
      setCards([])
    } finally {
      setLoadingCards(false)
    }
  }

  const loadCourseData = async () => {
    try {
      setLoading(true)
      setSelectionMessage(null)

      let loadedCourse: Course
      if (courseSlug) {
        loadedCourse = await studyService.getCourseBySlug(courseSlug)
      } else if (id) {
        loadedCourse = await studyService.getCourse(Number(id))
      } else {
        throw new Error('Course route params are missing')
      }

      setCourse(loadedCourse)

      if (id && loadedCourse.slug) {
        navigate(`/deck/${loadedCourse.slug}`, { replace: true })
      }

      try {
        await studyService.activateCourse(loadedCourse.id)
      } catch (error) {
        console.error('Error activating course:', error)
      }

      const loadedDecks = await studyService.getDecksByCourse(loadedCourse.id)
      setDecks(loadedDecks)

      const [nextActiveCourse, nextActiveDeck, nextUserDecks] = await Promise.all([
        studyService.getActiveCourse().catch(() => null),
        studyService.getActiveDeck().catch(() => null),
        user?.id ? studyService.getUserDecks(user.id).catch(() => []) : Promise.resolve([]),
      ])

      setActiveCourse(nextActiveCourse)
      setActiveDeck(nextActiveDeck)

      const scopedUserDecks = nextUserDecks.filter((userDeck) => loadedDecks.some((deck) => deck.id === userDeck.deck_id))
      setUserDecks(scopedUserDecks)

      const availableDecks = buildDeckPresentations(loadedCourse, loadedDecks, scopedUserDecks, nextActiveDeck?.deck_id)

      let targetDeck =
        (deckSlug ? availableDecks.find((item) => item.deck.slug === deckSlug)?.deck : null) ||
        (nextActiveDeck ? availableDecks.find((item) => item.deck.id === nextActiveDeck.deck_id)?.deck : null) ||
        availableDecks.find((item) => item.state !== 'locked')?.deck ||
        loadedDecks[0] ||
        null

      if (targetDeck) {
        const selectedState = availableDecks.find((item) => item.deck.id === targetDeck!.id)
        if (selectedState?.state === 'locked') {
          setSelectionMessage(selectedState.lockReason || 'Этот дек пока закрыт.')
          targetDeck = availableDecks.find((item) => item.state !== 'locked')?.deck || targetDeck
        }

        if (deckSlug && loadedCourse.slug && targetDeck.slug !== deckSlug) {
          navigate(`/deck/${loadedCourse.slug}/${targetDeck.slug}`, { replace: true })
        }

        await handleDeckSelect(targetDeck, loadedCourse, false, loadedDecks, scopedUserDecks, nextActiveDeck?.deck_id)
      } else {
        setSelectedDeck(null)
        setCards([])
      }
    } catch (error) {
      console.error('Error loading course detail:', error)
      setCourse(null)
      setDecks([])
      setUserDecks([])
      setSelectedDeck(null)
      setCards([])
      setSelectionMessage('Не удалось загрузить курс.')
    } finally {
      setLoading(false)
    }
  }

  const refreshScopedUserDecks = async (courseDecks: Deck[]) => {
    if (!user?.id) return

    try {
      const nextUserDecks = await studyService.getUserDecks(user.id)
      setUserDecks(nextUserDecks.filter((userDeck) => courseDecks.some((deck) => deck.id === userDeck.deck_id)))
    } catch (error) {
      console.error('Error refreshing deck progress:', error)
    }
  }

  const handleDeckSelect = async (
    deck: Deck,
    loadedCourse = course,
    pushHistory = true,
    availableDecks = decks,
    availableUserDecks = userDecks,
    activeDeckId = activeDeck?.deck_id
  ) => {
    const scopedDecks = availableDecks.length ? availableDecks : [deck]
    const deckState = buildDeckPresentations(loadedCourse || null, scopedDecks, availableUserDecks, activeDeckId).find((item) => item.deck.id === deck.id)

    if (deckState?.state === 'locked') {
      setSelectionMessage(deckState.lockReason || 'Этот дек пока закрыт.')
      return
    }

    setSelectionMessage(null)
    setSelectedDeck(deck)

    try {
      const nextActiveDeck = await studyService.activateDeck(deck.id)
      setActiveDeck(nextActiveDeck)
      if (loadedCourse?.id) {
        try {
          setActiveCourse(await studyService.getActiveCourse())
        } catch {
          setActiveCourse(null)
        }
      }
      await refreshScopedUserDecks(scopedDecks)
    } catch (error: any) {
      console.error('Error activating deck:', error)
      setSelectionMessage(error?.response?.data?.error || 'Не удалось открыть дек.')
      return
    }

    await loadCards(deck.id)

    if (pushHistory && loadedCourse?.slug && deck.slug) {
      navigate(`/deck/${loadedCourse.slug}/${deck.slug}`)
    }
  }

  const playAudio = (audioUrl: string) => {
    const normalizedUrl = config.getFullUrl(audioUrl)
    const audio = new Audio(normalizedUrl)
    audio.play().catch((error) => console.error('Error playing audio:', error))
  }

  if (loading) {
    return <div className="py-8 text-center text-text-light">Загрузка курса...</div>
  }

  if (!course) {
    return (
      <div className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-red-700">
        Курс не найден
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <Link to="/courses" className="inline-flex items-center gap-2 text-sm font-semibold text-link-light transition-colors hover:text-link-dark">
        <span>←</span>
        <span>Назад к курсам</span>
      </Link>

      <section className="overflow-hidden rounded-[28px] border border-rose-100 bg-gradient-to-br from-white via-rose-50 to-orange-50 shadow-xl">
        <div className="grid gap-8 p-6 lg:grid-cols-[220px_minmax(0,1fr)] lg:p-8">
          <div className="relative">
            {course.image_url ? (
              <img
                src={config.getFullUrl(course.image_url)}
                alt={course.title}
                className="h-52 w-full rounded-[24px] object-cover shadow-lg lg:h-full"
                onError={(event) => {
                  ;(event.target as HTMLImageElement).style.display = 'none'
                }}
              />
            ) : (
              <div className="flex h-52 w-full items-center justify-center rounded-[24px] bg-gradient-to-br from-rose-100 to-orange-100 text-6xl shadow-lg">
                📘
              </div>
            )}
          </div>

          <div className="space-y-6">
            <div className="flex flex-wrap items-center gap-3">
              <span className="rounded-full bg-white/80 px-3 py-1 text-xs font-semibold uppercase tracking-[0.22em] text-link-light shadow-sm">
                Маршрут курса
              </span>
              {course.is_published && (
                <span className="rounded-full bg-green-100 px-3 py-1 text-xs font-semibold text-green-800">
                  Опубликован
                </span>
              )}
              {activeCourse?.course_id === course.id && (
                <span className="rounded-full bg-rose-100 px-3 py-1 text-xs font-semibold text-rose-800">
                  Текущий курс
                </span>
              )}
            </div>

            <div className="space-y-3">
              <h1 className="text-3xl font-bold text-text-light lg:text-4xl">{course.title}</h1>
              {course.description && <p className="max-w-3xl text-base leading-7 text-slate-600">{course.description}</p>}
            </div>

            <div className="grid gap-3 sm:grid-cols-3">
              <div className="rounded-2xl border border-white/70 bg-white/80 p-4 shadow-sm backdrop-blur">
                <p className="text-xs uppercase tracking-[0.18em] text-slate-400">Пройдено деков</p>
                <p className="mt-2 text-3xl font-bold text-text-light">
                  {courseStats.completedDecks}
                  <span className="ml-2 text-lg font-medium text-slate-400">/ {courseStats.totalDecks}</span>
                </p>
              </div>
              <div className="rounded-2xl border border-white/70 bg-white/80 p-4 shadow-sm backdrop-blur">
                <p className="text-xs uppercase tracking-[0.18em] text-slate-400">Слова изучены</p>
                <p className="mt-2 text-3xl font-bold text-text-light">
                  {courseStats.learnedWords}
                  <span className="ml-2 text-lg font-medium text-slate-400">/ {courseStats.totalWords || 0}</span>
                </p>
              </div>
              <div className="rounded-2xl border border-white/70 bg-white/80 p-4 shadow-sm backdrop-blur">
                <p className="text-xs uppercase tracking-[0.18em] text-slate-400">Средний прогресс</p>
                <p className="mt-2 text-3xl font-bold text-link-light">{courseStats.averageProgress}%</p>
              </div>
            </div>

            <div className="rounded-2xl border border-rose-100 bg-white/85 p-4 text-sm text-slate-600 shadow-sm">
              <p className="font-semibold text-text-light">Куда добавляются новые слова</p>
              <p className="mt-1">
                Слова из reader и ручного словаря попадают в текущий активный дек.
                {' '}
                <span className="font-semibold text-link-light">
                  {activeDeck ? deckPresentations.find((item) => item.deck.id === activeDeck.deck_id)?.deck.title || 'Дек не найден' : 'Дек пока не выбран'}
                </span>
              </p>
            </div>
          </div>
        </div>
      </section>

      {selectionMessage && (
        <div className="rounded-2xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-800">
          {selectionMessage}
        </div>
      )}

      <div className="grid gap-6 lg:grid-cols-[360px_minmax(0,1fr)]">
        <aside className="rounded-[28px] border border-gray-200 bg-card-light p-5 shadow-md">
          <div className="mb-5 flex items-center justify-between">
            <div>
              <h2 className="text-xl font-semibold text-text-light">Деки курса</h2>
              <p className="text-sm text-gray-500">У каждого дека есть своя ссылка и своё состояние.</p>
            </div>
            <span className="rounded-full bg-rose-50 px-3 py-1 text-xs font-semibold text-rose-700">
              {deckPresentations.length} шт.
            </span>
          </div>

          <div className="space-y-3">
            {deckPresentations.map((item, index) => {
              const isSelected = selectedDeck?.id === item.deck.id
              const stateStyles =
                item.state === 'completed'
                  ? 'border-green-200 bg-green-50'
                  : item.state === 'active'
                    ? 'border-link-light bg-rose-50 shadow-md'
                    : item.state === 'available'
                      ? 'border-slate-200 bg-white hover:border-link-light hover:shadow-sm'
                      : 'border-slate-200 bg-slate-50 opacity-80'

              const badgeText =
                item.state === 'completed'
                  ? 'Пройден'
                  : item.state === 'active'
                    ? 'Активный'
                    : item.state === 'available'
                      ? 'Открыт'
                      : '🔒 Замок'

              const badgeColor =
                item.state === 'completed'
                  ? 'bg-green-100 text-green-800'
                  : item.state === 'active'
                    ? 'bg-rose-100 text-rose-800'
                    : item.state === 'available'
                      ? 'bg-slate-100 text-slate-700'
                      : 'bg-amber-100 text-amber-800'

              return (
                <div
                  key={item.deck.id}
                  className={`rounded-2xl border p-4 transition-all ${stateStyles} ${isSelected ? 'ring-2 ring-rose-100' : ''}`}
                >
                  <div className="mb-3 flex items-start justify-between gap-3">
                    <div>
                      <p className="text-xs font-semibold uppercase tracking-[0.18em] text-slate-400">Дек {index + 1}</p>
                      <h3 className="mt-1 text-lg font-semibold text-text-light">{item.deck.title}</h3>
                    </div>
                    <span className={`rounded-full px-2.5 py-1 text-[11px] font-semibold ${badgeColor}`}>{badgeText}</span>
                  </div>

                  {item.deck.description && <p className="text-sm leading-6 text-slate-500">{item.deck.description}</p>}

                  <div className="mt-4 space-y-2">
                    <div className="flex items-center justify-between text-xs uppercase tracking-[0.14em] text-slate-400">
                      <span>Прогресс</span>
                      <span>{item.progress}%</span>
                    </div>
                    <div className="h-2 rounded-full bg-slate-200">
                      <div
                        className={`h-2 rounded-full transition-all ${
                          item.state === 'completed' ? 'bg-green-500' : item.state === 'active' ? 'bg-rose-500' : 'bg-slate-400'
                        }`}
                        style={{ width: `${item.progress}%` }}
                      />
                    </div>
                    <p className="text-xs text-slate-500">
                      Изучено слов: {item.learnedCards} из {item.totalCards || 0}
                    </p>
                  </div>

                  <div className="mt-4 flex flex-wrap items-center gap-3 text-sm">
                    {item.state === 'locked' ? (
                      <span className="font-medium text-amber-700">{item.lockReason}</span>
                    ) : (
                      <>
                        <button
                          type="button"
                          onClick={() => void handleDeckSelect(item.deck)}
                          className="font-semibold text-link-light transition-colors hover:text-link-dark"
                        >
                          Открыть дек
                        </button>
                        <Link to={item.sharePath} className="text-slate-500 transition-colors hover:text-link-light">
                          Своя ссылка
                        </Link>
                      </>
                    )}
                  </div>
                </div>
              )
            })}
          </div>
        </aside>

        <section className="space-y-6">
          {!selectedDeck || !selectedDeckPresentation ? (
            <div className="rounded-[28px] border border-gray-200 bg-card-light p-10 text-center shadow-md">
              <p className="text-lg font-medium text-text-light">Выберите дек слева, чтобы открыть слова и тренировку.</p>
            </div>
          ) : (
            <>
              <div className="rounded-[28px] border border-gray-200 bg-card-light p-6 shadow-md">
                <div className="flex flex-col gap-5 xl:flex-row xl:items-start xl:justify-between">
                  <div className="space-y-4">
                    <div className="flex flex-wrap items-center gap-3">
                      <h2 className="text-3xl font-bold text-text-light">{selectedDeck.title}</h2>
                      <span
                        className={`rounded-full px-3 py-1 text-xs font-semibold ${
                          selectedDeckPresentation.state === 'completed'
                            ? 'bg-green-100 text-green-800'
                            : selectedDeckPresentation.state === 'active'
                              ? 'bg-rose-100 text-rose-800'
                              : 'bg-slate-100 text-slate-700'
                        }`}
                      >
                        {selectedDeckPresentation.state === 'completed'
                          ? 'Пройден'
                          : selectedDeckPresentation.state === 'active'
                            ? 'Текущий дек'
                            : 'Открыт'}
                      </span>
                    </div>

                    {selectedDeck.description && <p className="max-w-3xl text-base leading-7 text-slate-600">{selectedDeck.description}</p>}

                    <div className="flex flex-wrap items-center gap-3 text-sm">
                      <Link to={selectedDeckPresentation.sharePath} className="rounded-full bg-slate-100 px-3 py-1.5 font-medium text-slate-600 transition-colors hover:bg-rose-100 hover:text-link-light">
                        {selectedDeckPresentation.sharePath}
                      </Link>
                      {selectedDeckPresentation.state === 'active' && (
                        <span className="rounded-full bg-rose-50 px-3 py-1.5 font-medium text-rose-700">
                          Слова из reader попадут сюда
                        </span>
                      )}
                    </div>
                  </div>

                  <div className="grid gap-3 sm:grid-cols-3 xl:min-w-[360px]">
                    <div className="rounded-2xl border border-slate-200 bg-white p-4 text-center shadow-sm">
                      <p className="text-xs uppercase tracking-[0.16em] text-slate-400">Прогресс</p>
                      <p className="mt-2 text-3xl font-bold text-link-light">{selectedDeckPresentation.progress}%</p>
                    </div>
                    <div className="rounded-2xl border border-slate-200 bg-white p-4 text-center shadow-sm">
                      <p className="text-xs uppercase tracking-[0.16em] text-slate-400">Выучено</p>
                      <p className="mt-2 text-3xl font-bold text-text-light">{selectedDeckPresentation.learnedCards}</p>
                    </div>
                    <div className="rounded-2xl border border-slate-200 bg-white p-4 text-center shadow-sm">
                      <p className="text-xs uppercase tracking-[0.16em] text-slate-400">Карточек</p>
                      <p className="mt-2 text-3xl font-bold text-text-light">{selectedDeckPresentation.totalCards || cards.length}</p>
                    </div>
                  </div>
                </div>

                <div className="mt-6 flex flex-col gap-4 rounded-2xl border border-rose-100 bg-gradient-to-r from-rose-50 to-orange-50 p-5 lg:flex-row lg:items-center lg:justify-between">
                  <div>
                    <p className="text-sm font-semibold uppercase tracking-[0.18em] text-link-light">Режим прохождения</p>
                    <p className="mt-2 text-base text-slate-600">
                      В тренировке будет
                      {' '}
                      <span className="font-semibold text-text-light">до 10 случайных слов</span>
                      {' '}
                      из этого дека. Слово не перейдёт к следующему виду, пока не пройдёт текущий.
                    </p>
                  </div>
                  <button
                    type="button"
                    onClick={() => setShowStudyModal(true)}
                    disabled={loadingCards || cards.length === 0}
                    className="rounded-2xl bg-link-light px-5 py-3 font-semibold text-white transition-colors hover:bg-link-dark disabled:cursor-not-allowed disabled:opacity-50"
                  >
                    Начать обучение
                  </button>
                </div>
              </div>

              {loadingCards ? (
                <div className="rounded-[28px] border border-gray-200 bg-card-light p-10 text-center shadow-md">
                  <p className="text-text-light">Загрузка слов...</p>
                </div>
              ) : cards.length === 0 ? (
                <div className="rounded-[28px] border border-gray-200 bg-card-light p-10 text-center shadow-md">
                  <p className="text-text-light">В этом деке пока нет слов.</p>
                </div>
              ) : (
                <div className="space-y-6">
                  {selectedDeckCards.custom.length > 0 && (
                  <div className="rounded-[28px] border border-rose-200 bg-rose-50/70 p-6 shadow-sm">
                      <div className="mb-4 flex items-center justify-between gap-4">
                        <div>
                          <h3 className="text-xl font-semibold text-text-light">Вы добавили в словарь</h3>
                          <p className="text-sm text-slate-600">Эти слова прикреплены именно к текущему деку.</p>
                        </div>
                        <span className="rounded-full bg-white px-3 py-1 text-xs font-semibold text-rose-700">
                          {selectedDeckCards.custom.length} слов
                        </span>
                      </div>

                      <div className="grid gap-4 md:grid-cols-2">
                        {selectedDeckCards.custom.map((card) => (
                          <div key={card.id} className="rounded-2xl border border-rose-200 bg-white p-4 shadow-sm">
                            <div className="flex items-start justify-between gap-4">
                              <div className="space-y-2">
                                <div className="flex flex-wrap items-center gap-2">
                                  <span className="text-xl font-bold text-text-light">{card.word}</span>
                                  <span className="rounded-full bg-rose-100 px-2 py-1 text-[11px] font-semibold text-rose-800">
                                    Мое слово
                                  </span>
                                  {card.audio_url && (
                                    <button
                                      type="button"
                                      onClick={() => playAudio(card.audio_url!)}
                                      className="text-lg text-green-600 transition-colors hover:text-green-800"
                                      title="Прослушать произношение"
                                    >
                                      🔊
                                    </button>
                                  )}
                                </div>
                                <p className="text-base font-medium text-slate-700">{card.translation}</p>
                                {card.example && <p className="border-l-2 border-link-light pl-3 text-sm italic text-slate-500">{card.example}</p>}
                              </div>
                            </div>
                          </div>
                        ))}
                      </div>
                    </div>
                  )}

                  <div className="rounded-[28px] border border-gray-200 bg-card-light p-6 shadow-md">
                    <div className="mb-4 flex items-center justify-between gap-4">
                      <div>
                        <h3 className="text-xl font-semibold text-text-light">Слова дека</h3>
                        <p className="text-sm text-slate-500">Базовые слова, которые участвуют в прохождении.</p>
                      </div>
                      <span className="rounded-full bg-slate-100 px-3 py-1 text-xs font-semibold text-slate-700">
                        {selectedDeckCards.regular.length} слов
                      </span>
                    </div>

                    <div className="grid gap-4 xl:grid-cols-2">
                      {selectedDeckCards.regular.map((card) => (
                        <div key={card.id} className="rounded-2xl border border-slate-200 bg-white p-4 transition-all hover:border-link-light hover:shadow-sm">
                          <div className="flex items-start justify-between gap-4">
                            <div className="space-y-2">
                              <div className="flex flex-wrap items-center gap-2">
                                <span className="text-xl font-bold text-text-light">{card.word}</span>
                                {card.phonetic && <span className="text-sm text-slate-400">[{card.phonetic}]</span>}
                                {card.audio_url && (
                                  <button
                                    type="button"
                                    onClick={() => playAudio(card.audio_url!)}
                                    className="text-lg text-green-600 transition-colors hover:text-green-800"
                                    title="Прослушать произношение"
                                  >
                                    🔊
                                  </button>
                                )}
                              </div>
                              <p className="text-base font-medium text-slate-700">{card.translation}</p>
                              {card.example && <p className="border-l-2 border-link-light pl-3 text-sm italic text-slate-500">{card.example}</p>}
                            </div>
                            {card.image_url && (
                              <img
                                src={config.getFullUrl(card.image_url)}
                                alt={card.word}
                                className="h-20 w-20 rounded-2xl object-cover"
                                onError={(event) => {
                                  ;(event.target as HTMLImageElement).style.display = 'none'
                                }}
                              />
                            )}
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
              )}
            </>
          )}
        </section>
      </div>

      {showStudyModal && selectedDeck && (
        <StudySessionModal
          course={course}
          deck={selectedDeck}
          onClose={() => {
            setShowStudyModal(false)
            void refreshScopedUserDecks(decks)
          }}
        />
      )}
    </div>
  )
}
