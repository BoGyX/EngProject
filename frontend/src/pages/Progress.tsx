import { useEffect, useMemo, useState } from 'react'
import { useAuthStore } from '../store/authStore'
import { Card, Deck, UserCard, UserDeck, studyService } from '../services/studyService'
import {
  ActiveTrainingMode,
  getCurrentTrainingModeForCard,
  getTrainingModesForCard,
  getTrainingProgressForCard,
  isTrainingModeCompleted,
  trainingModeMeta,
} from '../utils/trainingModes'

interface ModeProgressSummary {
  key: ActiveTrainingMode
  label: string
  availableCount: number
  completedCount: number
  percent: number
}

interface WordModeState {
  key: ActiveTrainingMode
  label: string
  shortLabel: string
  completed: boolean
  current: boolean
}

interface WordProgress {
  cardId: number
  word: string
  translation: string
  isCustom: boolean
  progressPercentage: number
  isCompleted: boolean
  statusLabel: string
  currentModeLabel: string
  modes: WordModeState[]
}

interface DeckProgress {
  deckId: number
  deckTitle: string
  totalCards: number
  completedWords: number
  averageProgress: number
  modeSummaries: ModeProgressSummary[]
  words: WordProgress[]
}

interface CourseProgress {
  courseId: number
  courseTitle: string
  totalDecks: number
  completedDecks: number
  averageProgress: number
  decks: DeckProgress[]
}

const trainingModeOrder: ActiveTrainingMode[] = ['view', 'choice', 'with_photo', 'russian', 'constructor']

function CircularProgress({ percent, size = 116 }: { percent: number; size?: number }) {
  const radius = (size - 10) / 2
  const circumference = 2 * Math.PI * radius
  const offset = circumference - (percent / 100) * circumference

  return (
    <div className="relative inline-flex items-center justify-center">
      <svg width={size} height={size} className="-rotate-90 transform">
        <circle cx={size / 2} cy={size / 2} r={radius} stroke="#f1f5f9" strokeWidth="8" fill="none" />
        <circle
          cx={size / 2}
          cy={size / 2}
          r={radius}
          stroke="#c62828"
          strokeWidth="8"
          fill="none"
          strokeDasharray={circumference}
          strokeDashoffset={offset}
          strokeLinecap="round"
          className="transition-all duration-500"
        />
      </svg>
      <div className="absolute inset-0 flex items-center justify-center">
        <span className="text-2xl font-bold text-text-light">{percent}%</span>
      </div>
    </div>
  )
}

function ProgressRow({
  label,
  value,
  extra,
  colorClass,
  textClass,
}: {
  label: string
  value: number
  extra: string
  colorClass: string
  textClass: string
}) {
  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between gap-4 text-sm">
        <div>
          <span className="font-medium text-slate-600">{label}</span>
          <span className="ml-2 text-xs text-slate-400">{extra}</span>
        </div>
        <span className={`font-semibold ${textClass}`}>{value}%</span>
      </div>
      <div className="h-2 rounded-full bg-slate-200">
        <div className={`h-2 rounded-full transition-all ${colorClass}`} style={{ width: `${value}%` }} />
      </div>
    </div>
  )
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

function buildUserCardsByUserDeckMap(userCards: UserCard[]) {
  const cardsMap = new Map<number, UserCard[]>()

  userCards.forEach((userCard) => {
    if (!userCard.user_deck_id) {
      return
    }

    const current = cardsMap.get(userCard.user_deck_id) || []
    current.push(userCard)
    cardsMap.set(userCard.user_deck_id, current)
  })

  return cardsMap
}

function buildWordProgress(card: Card, userCard?: UserCard): WordProgress {
  const progressPercentage = getTrainingProgressForCard(card, userCard)
  const currentMode = getCurrentTrainingModeForCard(card, userCard)
  const modes = getTrainingModesForCard(card).map((mode) => ({
    key: mode,
    label: trainingModeMeta[mode].label,
    shortLabel: trainingModeMeta[mode].shortLabel,
    completed: isTrainingModeCompleted(userCard, mode),
    current: currentMode === mode,
  }))

  return {
    cardId: card.id,
    word: card.word,
    translation: card.translation,
    isCustom: card.is_custom,
    progressPercentage,
    isCompleted: progressPercentage >= 100,
    statusLabel: progressPercentage >= 100 ? 'Изучено' : progressPercentage > 0 ? 'В процессе' : 'Не начато',
    currentModeLabel: currentMode === 'completed' ? 'Все режимы пройдены' : trainingModeMeta[currentMode].label,
    modes,
  }
}

function buildDeckProgress(deck: Deck, cards: Card[], userCards: UserCard[]): DeckProgress {
  const userCardMap = new Map<number, UserCard>(userCards.map((userCard) => [userCard.card_id, userCard]))
  const words = cards.map((card) => buildWordProgress(card, userCardMap.get(card.id)))

  const modeSummaries = trainingModeOrder
    .map((mode) => {
      const eligibleWords = words.filter((word) => word.modes.some((item) => item.key === mode))
      const completedWords = eligibleWords.filter((word) => word.modes.some((item) => item.key === mode && item.completed))

      return {
        key: mode,
        label: trainingModeMeta[mode].label,
        availableCount: eligibleWords.length,
        completedCount: completedWords.length,
        percent: eligibleWords.length > 0 ? Math.round((completedWords.length / eligibleWords.length) * 100) : 0,
      }
    })
    .filter((summary) => summary.availableCount > 0)

  const averageProgress =
    words.length > 0 ? Math.round(words.reduce((sum, word) => sum + word.progressPercentage, 0) / words.length) : 0

  return {
    deckId: deck.id,
    deckTitle: deck.title,
    totalCards: cards.length,
    completedWords: words.filter((word) => word.isCompleted).length,
    averageProgress,
    modeSummaries,
    words,
  }
}

export default function Progress() {
  const { user } = useAuthStore()
  const [courseProgress, setCourseProgress] = useState<CourseProgress[]>([])
  const [expandedCourses, setExpandedCourses] = useState<Set<number>>(new Set())
  const [loading, setLoading] = useState(true)
  const [overallProgress, setOverallProgress] = useState(0)

  useEffect(() => {
    void loadProgress()
  }, [user?.id])

  const loadProgress = async () => {
    if (!user?.id) {
      setCourseProgress([])
      setOverallProgress(0)
      setLoading(false)
      return
    }

    try {
      setLoading(true)

      const [courses, userCards, userDecks] = await Promise.all([
        studyService.getCourses(),
        studyService.getUserCards(user.id),
        studyService.getUserDecks(user.id),
      ])

      const latestUserDeckMap = buildLatestUserDeckMap(userDecks)
      const userCardsByUserDeckMap = buildUserCardsByUserDeckMap(userCards)

      const progressData = (
        await Promise.all(
          courses.map(async (course) => {
            const decks = await studyService.getDecksByCourse(course.id)
            const deckProgressData = (
              await Promise.all(
                decks.map(async (deck) => {
                  const cards = await studyService.getCardsByDeck(deck.id)
                  if (cards.length === 0) {
                    return null
                  }

                  const latestUserDeck = latestUserDeckMap.get(deck.id)
                  const deckUserCards = latestUserDeck ? userCardsByUserDeckMap.get(latestUserDeck.id) || [] : []

                  return buildDeckProgress(deck, cards, deckUserCards)
                })
              )
            ).filter((deck): deck is DeckProgress => deck !== null)

            if (deckProgressData.length === 0) {
              return null
            }

            const averageProgress = Math.round(
              deckProgressData.reduce((sum, deck) => sum + deck.averageProgress, 0) / deckProgressData.length
            )

            return {
              courseId: course.id,
              courseTitle: course.title,
              totalDecks: deckProgressData.length,
              completedDecks: deckProgressData.filter((deck) => deck.averageProgress >= 100).length,
              averageProgress,
              decks: deckProgressData,
            }
          })
        )
      ).filter((course): course is CourseProgress => course !== null)

      setCourseProgress(progressData)
      setOverallProgress(
        progressData.length > 0
          ? Math.round(progressData.reduce((sum, course) => sum + course.averageProgress, 0) / progressData.length)
          : 0
      )
      setExpandedCourses(new Set(progressData.slice(0, 1).map((course) => course.courseId)))
    } catch (error) {
      console.error('Error loading progress:', error)
      setCourseProgress([])
      setOverallProgress(0)
    } finally {
      setLoading(false)
    }
  }

  const toggleCourse = (courseId: number) => {
    const nextExpanded = new Set(expandedCourses)
    if (nextExpanded.has(courseId)) {
      nextExpanded.delete(courseId)
    } else {
      nextExpanded.add(courseId)
    }
    setExpandedCourses(nextExpanded)
  }

  const summary = useMemo(() => {
    const totalDecks = courseProgress.reduce((sum, course) => sum + course.totalDecks, 0)
    const completedDecks = courseProgress.reduce((sum, course) => sum + course.completedDecks, 0)
    const totalWords = courseProgress.reduce(
      (sum, course) => sum + course.decks.reduce((deckSum, deck) => deckSum + deck.totalCards, 0),
      0
    )
    const completedWords = courseProgress.reduce(
      (sum, course) => sum + course.decks.reduce((deckSum, deck) => deckSum + deck.completedWords, 0),
      0
    )

    return {
      courses: courseProgress.length,
      totalDecks,
      completedDecks,
      totalWords,
      completedWords,
    }
  }, [courseProgress])

  if (loading) {
    return <div className="py-8 text-center text-text-light">Загрузка прогресса...</div>
  }

  return (
    <div className="space-y-6">
      <section className="overflow-hidden rounded-[28px] border border-rose-100 bg-gradient-to-br from-white via-rose-50 to-orange-50 shadow-xl">
        <div className="grid gap-8 p-6 lg:grid-cols-[minmax(0,1fr)_220px] lg:p-8">
          <div className="space-y-5">
            <div className="flex flex-wrap items-center gap-3">
              <span className="rounded-full bg-white/80 px-3 py-1 text-xs font-semibold uppercase tracking-[0.22em] text-link-light shadow-sm">
                Progress
              </span>
              <span className="rounded-full bg-rose-100 px-3 py-1 text-xs font-semibold text-rose-800">
                Слова и режимы по подкурсам
              </span>
            </div>

            <div>
              <h1 className="text-3xl font-bold text-text-light lg:text-4xl">Мой прогресс</h1>
              <p className="mt-2 max-w-3xl text-sm leading-7 text-slate-600">
                Внутри каждого подкурса видно каждое слово, его текущий режим, процент прохождения и то, какие этапы уже
                закрыты.
              </p>
            </div>

            <div className="grid gap-3 sm:grid-cols-4">
              <div className="rounded-2xl border border-white/70 bg-white/80 p-4 shadow-sm backdrop-blur">
                <p className="text-xs uppercase tracking-[0.18em] text-slate-400">Курсов в работе</p>
                <p className="mt-2 text-3xl font-bold text-text-light">{summary.courses}</p>
              </div>
              <div className="rounded-2xl border border-white/70 bg-white/80 p-4 shadow-sm backdrop-blur">
                <p className="text-xs uppercase tracking-[0.18em] text-slate-400">Подкурсов пройдено</p>
                <p className="mt-2 text-3xl font-bold text-text-light">
                  {summary.completedDecks}
                  <span className="ml-2 text-lg font-medium text-slate-400">/ {summary.totalDecks}</span>
                </p>
              </div>
              <div className="rounded-2xl border border-white/70 bg-white/80 p-4 shadow-sm backdrop-blur">
                <p className="text-xs uppercase tracking-[0.18em] text-slate-400">Слов изучено</p>
                <p className="mt-2 text-3xl font-bold text-text-light">
                  {summary.completedWords}
                  <span className="ml-2 text-lg font-medium text-slate-400">/ {summary.totalWords}</span>
                </p>
              </div>
              <div className="rounded-2xl border border-white/70 bg-white/80 p-4 shadow-sm backdrop-blur">
                <p className="text-xs uppercase tracking-[0.18em] text-slate-400">Средний прогресс</p>
                <p className="mt-2 text-3xl font-bold text-link-light">{overallProgress}%</p>
              </div>
            </div>
          </div>

          <div className="flex items-center justify-center rounded-[24px] bg-white/80 p-6 shadow-lg">
            <CircularProgress percent={overallProgress} size={154} />
          </div>
        </div>
      </section>

      {courseProgress.length === 0 ? (
        <div className="rounded-[28px] border border-gray-200 bg-card-light p-10 text-center shadow-md">
          <p className="text-lg font-medium text-text-light">Начните изучать курсы, чтобы здесь появился прогресс.</p>
        </div>
      ) : (
        <div className="space-y-5">
          {courseProgress.map((course) => {
            const isExpanded = expandedCourses.has(course.courseId)

            return (
              <section key={course.courseId} className="overflow-hidden rounded-[28px] border border-gray-200 bg-card-light shadow-md">
                <button
                  type="button"
                  onClick={() => toggleCourse(course.courseId)}
                  className="w-full px-6 py-5 text-left transition-colors hover:bg-rose-50/40"
                >
                  <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
                    <div className="flex items-center gap-5">
                      <CircularProgress percent={course.averageProgress} size={96} />
                      <div>
                        <h2 className="text-2xl font-bold text-text-light">{course.courseTitle}</h2>
                        <p className="mt-2 text-sm text-slate-500">
                          Подкурсов: {course.totalDecks} | Завершено: {course.completedDecks}
                        </p>
                      </div>
                    </div>

                    <div className="flex items-center gap-3">
                      <span className="rounded-full bg-rose-50 px-3 py-1.5 text-sm font-semibold text-rose-700">
                        {course.averageProgress}% по курсу
                      </span>
                      <span className="text-2xl text-slate-400">{isExpanded ? '▾' : '▸'}</span>
                    </div>
                  </div>
                </button>

                {isExpanded && (
                  <div className="space-y-4 border-t border-gray-200 bg-gradient-to-br from-white to-slate-50 p-6">
                    {course.decks.map((deck) => (
                      <article key={deck.deckId} className="rounded-[24px] border border-slate-200 bg-white p-5 shadow-sm">
                        <div className="flex flex-col gap-4 xl:flex-row xl:items-start xl:justify-between">
                          <div>
                            <p className="text-xs uppercase tracking-[0.18em] text-slate-400">Подкурс</p>
                            <h3 className="mt-2 text-xl font-semibold text-text-light">{deck.deckTitle}</h3>
                            <p className="mt-1 text-sm text-slate-500">
                              Слов: {deck.totalCards} | Изучено: {deck.completedWords}
                            </p>
                          </div>

                          <div className="flex items-center gap-3">
                            <span className="rounded-full bg-orange-50 px-3 py-1.5 text-sm font-semibold text-orange-700">
                              {deck.averageProgress}% готово
                            </span>
                          </div>
                        </div>

                        <div className="mt-5 grid gap-4 xl:grid-cols-[340px_minmax(0,1fr)]">
                          <div className="rounded-[22px] border border-rose-100 bg-rose-50/60 p-4">
                            <div className="mb-4 flex items-center justify-between gap-3">
                              <h4 className="text-sm font-semibold uppercase tracking-[0.18em] text-rose-700">По режимам</h4>
                              <span className="rounded-full bg-white px-3 py-1 text-xs font-semibold text-rose-700">
                                {deck.modeSummaries.length} актив.
                              </span>
                            </div>

                            <div className="space-y-4">
                              {deck.modeSummaries.map((modeSummary) => (
                                <ProgressRow
                                  key={modeSummary.key}
                                  label={modeSummary.label}
                                  value={modeSummary.percent}
                                  extra={`${modeSummary.completedCount}/${modeSummary.availableCount} слов`}
                                  colorClass={
                                    modeSummary.key === 'view'
                                      ? 'bg-rose-400'
                                      : modeSummary.key === 'choice'
                                        ? 'bg-rose-500'
                                        : modeSummary.key === 'with_photo'
                                          ? 'bg-orange-500'
                                          : modeSummary.key === 'russian'
                                            ? 'bg-amber-500'
                                            : 'bg-red-700'
                                  }
                                  textClass={
                                    modeSummary.key === 'view'
                                      ? 'text-rose-500'
                                      : modeSummary.key === 'choice'
                                        ? 'text-rose-600'
                                        : modeSummary.key === 'with_photo'
                                          ? 'text-orange-600'
                                          : modeSummary.key === 'russian'
                                            ? 'text-amber-600'
                                            : 'text-red-700'
                                  }
                                />
                              ))}
                            </div>
                          </div>

                          <div className="space-y-3">
                            {deck.words.map((word) => (
                              <div key={word.cardId} className="rounded-[22px] border border-slate-200 bg-slate-50/70 p-4">
                                <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
                                  <div>
                                    <div className="flex flex-wrap items-center gap-2">
                                      <h4 className="text-lg font-semibold text-text-light">{word.word}</h4>
                                      {word.isCustom && (
                                        <span className="rounded-full bg-rose-100 px-2.5 py-1 text-[11px] font-semibold text-rose-800">
                                          Моё слово
                                        </span>
                                      )}
                                      <span
                                        className={`rounded-full px-2.5 py-1 text-[11px] font-semibold ${
                                          word.isCompleted
                                            ? 'bg-green-100 text-green-800'
                                            : word.progressPercentage > 0
                                              ? 'bg-amber-100 text-amber-800'
                                              : 'bg-slate-200 text-slate-700'
                                        }`}
                                      >
                                        {word.statusLabel}
                                      </span>
                                    </div>
                                    <p className="mt-1 text-sm text-slate-600">{word.translation}</p>
                                    <p className="mt-2 text-xs uppercase tracking-[0.16em] text-slate-400">
                                      {word.isCompleted ? 'Слово полностью закрыто' : `Текущий режим: ${word.currentModeLabel}`}
                                    </p>
                                  </div>

                                  <div className="min-w-[110px] text-right">
                                    <p className="text-xs uppercase tracking-[0.16em] text-slate-400">Прогресс слова</p>
                                    <p className="mt-1 text-2xl font-bold text-text-light">{word.progressPercentage}%</p>
                                  </div>
                                </div>

                                <div className="mt-4 h-2 rounded-full bg-slate-200">
                                  <div
                                    className="h-2 rounded-full bg-link-light transition-all"
                                    style={{ width: `${word.progressPercentage}%` }}
                                  />
                                </div>

                                <div className="mt-4 grid gap-2 md:grid-cols-2 xl:grid-cols-5">
                                  {word.modes.map((mode) => (
                                    <div
                                      key={`${word.cardId}-${mode.key}`}
                                      className={`rounded-2xl border px-3 py-3 ${
                                        mode.completed
                                          ? 'border-green-200 bg-green-50'
                                          : mode.current
                                            ? 'border-rose-200 bg-rose-50'
                                            : 'border-slate-200 bg-white'
                                      }`}
                                    >
                                      <p className="text-sm font-semibold text-text-light">{mode.shortLabel}</p>
                                      <p
                                        className={`mt-1 text-xs ${
                                          mode.completed
                                            ? 'text-green-700'
                                            : mode.current
                                              ? 'text-rose-700'
                                              : 'text-slate-500'
                                        }`}
                                      >
                                        {mode.completed ? 'Пройден' : mode.current ? 'Текущий' : 'Ждет'}
                                      </p>
                                    </div>
                                  ))}
                                </div>
                              </div>
                            ))}
                          </div>
                        </div>

                        {deck.averageProgress >= 100 && (
                          <div className="mt-5 rounded-2xl border border-green-200 bg-green-50 px-4 py-3 text-sm font-semibold text-green-700">
                            Этот подкурс завершен на 100%.
                          </div>
                        )}
                      </article>
                    ))}
                  </div>
                )}
              </section>
            )
          })}
        </div>
      )}
    </div>
  )
}
