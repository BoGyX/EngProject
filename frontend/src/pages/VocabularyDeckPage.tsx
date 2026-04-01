import { FormEvent, useEffect, useMemo, useState } from 'react'
import { useAuthStore } from '../store/authStore'
import { Card, Course, Deck, UserCard, UserDeck, studyService } from '../services/studyService'

interface DeckSection {
  deck: Deck
  learnedCards: Card[]
  customCards: Card[]
}

interface CourseSection {
  course: Course
  decks: DeckSection[]
}

export default function VocabularyDeckPage() {
  const { user } = useAuthStore()
  const [sections, setSections] = useState<CourseSection[]>([])
  const [activeDeck, setActiveDeck] = useState<UserDeck | null>(null)
  const [legacyVocabulary, setLegacyVocabulary] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [expandedCourses, setExpandedCourses] = useState<Set<number>>(new Set())
  const [selectedDeckId, setSelectedDeckId] = useState<number | null>(null)
  const [newWord, setNewWord] = useState('')
  const [newTranslation, setNewTranslation] = useState('')
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    void loadVocabulary()
  }, [user?.id])

  const loadVocabulary = async () => {
    if (!user?.id) return

    try {
      setLoading(true)
      const courses = await studyService.getCourses()
      const [userCards, legacyWords] = await Promise.all([
        studyService.getUserCards(user.id),
        studyService.getLegacyVocabulary(user.id),
      ])

      try {
        setActiveDeck(await studyService.getActiveDeck())
      } catch {
        setActiveDeck(null)
      }

      setLegacyVocabulary(legacyWords)

      const userCardMap = new Map<number, UserCard>(userCards.map((userCard) => [userCard.card_id, userCard]))
      const nextSections: CourseSection[] = []

      for (const course of courses) {
        const decks = await studyService.getDecksByCourse(course.id)
        const deckSections: DeckSection[] = []

        for (const deck of decks) {
          const cards = await studyService.getCardsByDeck(deck.id)
          const learnedCards = cards.filter((card) => (userCardMap.get(card.id)?.progress_percentage || 0) >= 100)
          const customCards = cards.filter((card) => card.is_custom && card.created_by === user.id)

          deckSections.push({
            deck,
            learnedCards,
            customCards,
          })
        }

        nextSections.push({ course, decks: deckSections })
      }

      setSections(nextSections)
      setExpandedCourses(new Set(nextSections.map((section) => section.course.id)))
    } catch (error) {
      console.error('Error loading vocabulary page:', error)
      setSections([])
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

  const handleAddWord = async (event: FormEvent) => {
    event.preventDefault()
    if (!newWord.trim() || !newTranslation.trim()) return

    try {
      setSubmitting(true)
      await studyService.createCustomCard({
        word: newWord.trim(),
        translation: newTranslation.trim(),
      })
      setNewWord('')
      setNewTranslation('')
      await loadVocabulary()
    } catch (error) {
      console.error('Error creating custom card:', error)
    } finally {
      setSubmitting(false)
    }
  }

  const activeDeckTitle = useMemo(() => {
    if (!activeDeck) return 'не выбран'

    return (
      sections
        .flatMap((section) => section.decks)
        .find((section) => section.deck.id === activeDeck.deck_id)?.deck.title || 'не найден'
    )
  }, [activeDeck, sections])

  const summary = useMemo(() => {
    const totalCustomWords = sections.reduce(
      (sum, section) => sum + section.decks.reduce((deckSum, deckSection) => deckSum + deckSection.customCards.length, 0),
      0
    )
    const totalLearnedWords = sections.reduce(
      (sum, section) => sum + section.decks.reduce((deckSum, deckSection) => deckSum + deckSection.learnedCards.length, 0),
      0
    )

    return {
      courses: sections.length,
      totalCustomWords,
      totalLearnedWords,
    }
  }, [sections])

  if (loading) {
    return <div className="py-8 text-center text-text-light">Загрузка словаря...</div>
  }

  return (
    <div className="space-y-6">
      <section className="overflow-hidden rounded-[28px] border border-rose-100 bg-gradient-to-br from-white via-rose-50 to-orange-50 shadow-xl">
        <div className="grid gap-8 p-6 lg:grid-cols-[minmax(0,1fr)_320px] lg:p-8">
          <div className="space-y-5">
            <div className="flex flex-wrap items-center gap-3">
              <span className="rounded-full bg-white/80 px-3 py-1 text-xs font-semibold uppercase tracking-[0.22em] text-link-light shadow-sm">
                Vocabulary
              </span>
              <span className="rounded-full bg-rose-100 px-3 py-1 text-xs font-semibold text-rose-800">
                Привязка к активному деку
              </span>
            </div>

            <div>
              <h1 className="text-3xl font-bold text-text-light lg:text-4xl">Словарь пользователя</h1>
              <p className="mt-2 max-w-3xl text-sm leading-7 text-slate-600">
                Новые слова из reader и ручного словаря добавляются в текущий активный дек. Выученные и пользовательские слова
                разложены по курсам и подкурсам отдельно.
              </p>
            </div>

            <div className="grid gap-3 sm:grid-cols-3">
              <div className="rounded-2xl border border-white/70 bg-white/80 p-4 shadow-sm backdrop-blur">
                <p className="text-xs uppercase tracking-[0.18em] text-slate-400">Курсов со словарём</p>
                <p className="mt-2 text-3xl font-bold text-text-light">{summary.courses}</p>
              </div>
              <div className="rounded-2xl border border-white/70 bg-white/80 p-4 shadow-sm backdrop-blur">
                <p className="text-xs uppercase tracking-[0.18em] text-slate-400">Добавлено вами</p>
                <p className="mt-2 text-3xl font-bold text-link-light">{summary.totalCustomWords}</p>
              </div>
              <div className="rounded-2xl border border-white/70 bg-white/80 p-4 shadow-sm backdrop-blur">
                <p className="text-xs uppercase tracking-[0.18em] text-slate-400">Выученных слов</p>
                <p className="mt-2 text-3xl font-bold text-text-light">{summary.totalLearnedWords}</p>
              </div>
            </div>
          </div>

          <div className="rounded-[24px] border border-rose-100 bg-white/80 p-6 shadow-lg">
            <p className="text-xs uppercase tracking-[0.18em] text-rose-500">Активный дек сейчас</p>
            <p className="mt-3 text-2xl font-bold text-text-light">{activeDeckTitle}</p>
            <p className="mt-3 text-sm leading-7 text-slate-600">
              Подкурс не закрывается после прохождения, но новые слова продолжают попадать именно в тот дек, который активен сейчас.
            </p>
          </div>
        </div>
      </section>

      <div className="space-y-5">
        {sections.map((section) => (
          <section key={section.course.id} className="overflow-hidden rounded-[28px] border border-gray-200 bg-card-light shadow-md">
            <button
              type="button"
              onClick={() => toggleCourse(section.course.id)}
              className="flex w-full items-center justify-between px-6 py-5 text-left transition-colors hover:bg-rose-50/40"
            >
              <div>
                <h2 className="text-2xl font-bold text-text-light">{section.course.title}</h2>
                <p className="mt-2 text-sm text-slate-500">
                  Подкурсов: {section.decks.length} | Выученных слов:{' '}
                  {section.decks.reduce((sum, deckSection) => sum + deckSection.learnedCards.length, 0)}
                </p>
              </div>
              <span className="text-2xl text-slate-400">{expandedCourses.has(section.course.id) ? '▾' : '▸'}</span>
            </button>

            {expandedCourses.has(section.course.id) && (
              <div className="space-y-4 border-t border-gray-200 bg-gradient-to-br from-white to-slate-50 p-5">
                {section.decks.map((deckSection) => {
                  const isActive = activeDeck?.deck_id === deckSection.deck.id
                  const isSelected = selectedDeckId === deckSection.deck.id

                  return (
                    <article key={deckSection.deck.id} className="rounded-[24px] border border-slate-200 bg-white p-5 shadow-sm">
                      <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
                        <div>
                          <div className="flex flex-wrap items-center gap-2">
                            <h3 className="text-xl font-semibold text-text-light">{deckSection.deck.title}</h3>
                            <span
                              className={`rounded-full px-2.5 py-1 text-[11px] font-semibold ${
                                isActive ? 'bg-rose-100 text-rose-800' : 'bg-amber-100 text-amber-800'
                              }`}
                            >
                              {isActive ? 'Активный' : 'Неактивный'}
                            </span>
                          </div>
                          {deckSection.deck.description && <p className="mt-2 text-sm text-slate-500">{deckSection.deck.description}</p>}
                        </div>

                        <button
                          type="button"
                          onClick={() => setSelectedDeckId(isSelected ? null : deckSection.deck.id)}
                          className="rounded-2xl border border-slate-300 px-4 py-2 text-sm font-semibold text-text-light transition-colors hover:border-rose-200 hover:bg-rose-50"
                        >
                          {isSelected ? 'Скрыть слова' : 'Показать слова'}
                        </button>
                      </div>

                      {isSelected && (
                        <div className="mt-5 space-y-5">
                          {isActive ? (
                            <form onSubmit={handleAddWord} className="rounded-[22px] border border-rose-200 bg-rose-50 p-5">
                              <h4 className="text-sm font-semibold uppercase tracking-[0.18em] text-rose-700">
                                Добавить слово в активный подкурс
                              </h4>
                              <div className="mt-4 grid gap-3 md:grid-cols-2">
                                <input
                                  type="text"
                                  value={newWord}
                                  onChange={(event) => setNewWord(event.target.value)}
                                  placeholder="word"
                                  className="rounded-2xl border border-gray-300 px-4 py-3 text-sm focus:border-link-light focus:outline-none"
                                />
                                <input
                                  type="text"
                                  value={newTranslation}
                                  onChange={(event) => setNewTranslation(event.target.value)}
                                  placeholder="перевод"
                                  className="rounded-2xl border border-gray-300 px-4 py-3 text-sm focus:border-link-light focus:outline-none"
                                />
                              </div>
                              <button
                                type="submit"
                                disabled={submitting}
                                className="mt-4 rounded-2xl bg-link-light px-5 py-3 text-sm font-semibold text-white transition-colors hover:bg-link-dark disabled:opacity-50"
                              >
                                {submitting ? 'Добавляю...' : 'Добавить слово'}
                              </button>
                            </form>
                          ) : (
                            <div className="rounded-[22px] border border-amber-200 bg-amber-50 px-5 py-4 text-sm text-amber-800">
                              Этот подкурс можно пересматривать и после прохождения, но новые слова сейчас добавляются только в активный дек.
                            </div>
                          )}

                          <div className="grid gap-5 xl:grid-cols-2">
                            <div className="rounded-[22px] border border-rose-200 bg-rose-50/70 p-5">
                              <div className="mb-4 flex items-center justify-between gap-3">
                                <div>
                                  <h4 className="text-lg font-semibold text-text-light">Вы добавили в словарь</h4>
                                  <p className="text-sm text-slate-600">Только ваши пользовательские слова в этом деке.</p>
                                </div>
                                <span className="rounded-full bg-white px-3 py-1 text-xs font-semibold text-rose-700">
                                  {deckSection.customCards.length}
                                </span>
                              </div>

                              {deckSection.customCards.length === 0 ? (
                                <p className="text-sm text-slate-500">Пользовательских слов пока нет.</p>
                              ) : (
                                <div className="grid gap-3">
                                  {deckSection.customCards.map((card) => (
                                    <div key={card.id} className="rounded-2xl border border-rose-200 bg-white p-4 shadow-sm">
                                      <p className="text-lg font-semibold text-text-light">{card.word}</p>
                                      <p className="mt-1 text-sm text-slate-600">{card.translation}</p>
                                    </div>
                                  ))}
                                </div>
                              )}
                            </div>

                            <div className="rounded-[22px] border border-green-200 bg-green-50/70 p-5">
                              <div className="mb-4 flex items-center justify-between gap-3">
                                <div>
                                  <h4 className="text-lg font-semibold text-text-light">Выученные слова</h4>
                                  <p className="text-sm text-slate-600">Слова с прогрессом 100% в этом деке.</p>
                                </div>
                                <span className="rounded-full bg-white px-3 py-1 text-xs font-semibold text-green-700">
                                  {deckSection.learnedCards.length}
                                </span>
                              </div>

                              {deckSection.learnedCards.length === 0 ? (
                                <p className="text-sm text-slate-500">Пока нет слов с прогрессом 100%.</p>
                              ) : (
                                <div className="grid gap-3">
                                  {deckSection.learnedCards.map((card) => (
                                    <div key={card.id} className="rounded-2xl border border-green-200 bg-white p-4 shadow-sm">
                                      <p className="text-lg font-semibold text-text-light">{card.word}</p>
                                      <p className="mt-1 text-sm text-slate-600">{card.translation}</p>
                                    </div>
                                  ))}
                                </div>
                              )}
                            </div>
                          </div>
                        </div>
                      )}
                    </article>
                  )
                })}
              </div>
            )}
          </section>
        ))}
      </div>

      <section className="rounded-[28px] border border-gray-200 bg-card-light p-6 shadow-md">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h2 className="text-xl font-semibold text-text-light">Legacy personal_vocabulary</h2>
            <p className="mt-2 text-sm text-slate-500">
              Старый словарь оставлен только для чтения, чтобы можно было безопасно откатить изменения без потери данных.
            </p>
          </div>
          <span className="rounded-full bg-slate-100 px-3 py-1 text-xs font-semibold text-slate-700">
            {legacyVocabulary.length} записей
          </span>
        </div>

        <div className="mt-5 flex flex-wrap gap-2">
          {legacyVocabulary.length === 0 ? (
            <span className="text-sm text-slate-500">Записей нет.</span>
          ) : (
            legacyVocabulary.map((word) => (
              <span key={word.id} className="rounded-full bg-slate-100 px-3 py-1 text-xs font-medium text-slate-700">
                {word.word}
              </span>
            ))
          )}
        </div>
      </section>
    </div>
  )
}
