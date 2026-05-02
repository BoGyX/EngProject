import { isAxiosError } from 'axios'
import { useEffect, useMemo, useState } from 'react'
import { config } from '../config'
import { Course, Deck, TrainingCardState, TrainingSessionState, studyService } from '../services/studyService'
import { getTrainingModeStepLabel } from '../utils/trainingModes'

interface StudySessionModalProps {
  course: Course
  deck: Deck
  onClose: () => void
}

interface FeedbackState {
  correct: boolean
  text: string
  submittedAnswer?: string
  correctAnswer?: string
}

function shuffleLetters(value: string): string[] {
  return value
    .split('')
    .sort(() => Math.random() - 0.5)
}

function getErrorMessage(error: unknown, fallbackMessage: string) {
  if (isAxiosError(error)) {
    const apiMessage = error.response?.data?.error
    if (typeof apiMessage === 'string' && apiMessage.trim()) {
      return apiMessage
    }
  }

  if (error instanceof Error && error.message.trim()) {
    return error.message
  }

  return fallbackMessage
}

function getCorrectAnswer(card: TrainingCardState) {
  switch (card.current_mode) {
    case 'choice':
      return card.translation
    case 'with_photo':
    case 'russian':
    case 'constructor':
      return card.word
    default:
      return undefined
  }
}

export default function StudySessionModal({ course, deck, onClose }: StudySessionModalProps) {
  const [session, setSession] = useState<TrainingSessionState | null>(null)
  const [pendingSession, setPendingSession] = useState<TrainingSessionState | null>(null)
  const [answeredCard, setAnsweredCard] = useState<TrainingCardState | null>(null)
  const [loading, setLoading] = useState(true)
  const [submitting, setSubmitting] = useState(false)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)
  const [textAnswer, setTextAnswer] = useState('')
  const [constructorPool, setConstructorPool] = useState<string[]>([])
  const [constructorAnswer, setConstructorAnswer] = useState<string[]>([])
  const [feedback, setFeedback] = useState<FeedbackState | null>(null)

  const currentCard = answeredCard || session?.current_card || null
  const sessionForProgress = pendingSession || session
  const hasPendingAdvance = Boolean(pendingSession && answeredCard && feedback)

  useEffect(() => {
    void startSession()
  }, [deck.id])

  useEffect(() => {
    if (!currentCard) {
      return
    }

    setTextAnswer('')

    if (currentCard.current_mode === 'constructor') {
      setConstructorAnswer([])
      setConstructorPool(shuffleLetters(currentCard.word))
      return
    }

    setConstructorAnswer([])
    setConstructorPool([])
  }, [currentCard?.session_card_id, currentCard?.current_mode])

  const progressValue = useMemo(() => {
    if (!sessionForProgress || sessionForProgress.cards.length === 0) {
      return 0
    }

    const completedSteps = sessionForProgress.cards.filter((card) => card.is_completed).length
    return Math.round((completedSteps / sessionForProgress.cards.length) * 100)
  }, [sessionForProgress])

  const startSession = async () => {
    try {
      setLoading(true)
      setErrorMessage(null)
      setFeedback(null)
      setPendingSession(null)
      setAnsweredCard(null)

      const nextSession = await studyService.startTraining(deck.id, course.id)
      setSession(nextSession)
    } catch (error) {
      console.error('Error starting training session:', error)
      setSession(null)
      setPendingSession(null)
      setAnsweredCard(null)
      setErrorMessage(getErrorMessage(error, 'Не удалось запустить обучение. Попробуйте еще раз.'))
    } finally {
      setLoading(false)
    }
  }

  const submitAnswer = async (answer = '') => {
    if (!currentCard || !session || hasPendingAdvance) {
      return
    }

    try {
      setSubmitting(true)
      setFeedback(null)

      const response = await studyService.answerTraining(session.session.id, currentCard.session_card_id, answer)

      setPendingSession(response.session)
      setAnsweredCard(currentCard)
      setFeedback({
        correct: response.is_correct,
        text: response.is_correct
          ? 'Вы ответили верно. Нажмите «Далее», чтобы перейти к следующей карточке.'
          : 'Вы ответили неправильно. Попробуйте этот режим снова в следующей сессии, а сейчас перейдите далее.',
        submittedAnswer: answer === 'viewed' ? undefined : answer,
        correctAnswer: getCorrectAnswer(currentCard),
      })
    } catch (error) {
      console.error('Error submitting training answer:', error)
      setFeedback({
        correct: false,
        text: getErrorMessage(error, 'Не удалось проверить ответ. Попробуйте еще раз.'),
      })
    } finally {
      setSubmitting(false)
    }
  }

  const playAudio = (audioUrl?: string) => {
    if (!audioUrl) {
      return
    }

    const normalizedUrl = config.getFullUrl(audioUrl)
    const audio = new Audio(normalizedUrl)
    audio.play().catch((error) => console.error('Error playing audio:', error))
  }

  const advanceToNextCard = () => {
    if (!pendingSession) {
      return
    }

    setSession(pendingSession)
    setPendingSession(null)
    setAnsweredCard(null)
    setFeedback(null)
  }

  const renderAnswerReview = (card: TrainingCardState) => {
    if (!feedback || !hasPendingAdvance) {
      return null
    }

    return (
      <div
        className={`space-y-4 rounded-2xl border px-5 py-4 text-left ${
          feedback.correct ? 'border-green-200 bg-green-50 text-green-800' : 'border-red-200 bg-red-50 text-red-800'
        }`}
      >
        <p className="text-sm font-semibold leading-6">{feedback.text}</p>

        {feedback.submittedAnswer && (
          <div className="space-y-1 text-sm">
            <p>
              <span className="font-semibold">Ваш ответ:</span> {feedback.submittedAnswer}
            </p>
            {!feedback.correct && feedback.correctAnswer && (
              <p>
                <span className="font-semibold">Правильный ответ:</span> {feedback.correctAnswer}
              </p>
            )}
          </div>
        )}

        {!feedback.submittedAnswer && card.current_mode === 'view' && (
          <p className="text-sm">Карточка просмотра засчитана.</p>
        )}

        <div className="flex justify-center pt-1">
          <button
            type="button"
            onClick={advanceToNextCard}
            className="rounded-lg bg-link-light px-6 py-3 font-semibold text-white transition-colors hover:bg-link-dark"
          >
            Далее
          </button>
        </div>
      </div>
    )
  }

  const renderMode = (card: TrainingCardState) => {
    switch (card.current_mode) {
      case 'view':
        return (
          <div className="space-y-6 text-center">
            <div className="space-y-3">
              <div className="flex flex-wrap items-center justify-center gap-3">
                <h2 className="text-4xl font-bold text-text-light">{card.word}</h2>
                {card.audio_url && (
                  <button
                    type="button"
                    onClick={() => playAudio(card.audio_url)}
                    className="text-3xl"
                    title="Прослушать слово"
                  >
                    🔊
                  </button>
                )}
              </div>
              {card.phonetic && <p className="text-lg text-gray-400">[{card.phonetic}]</p>}
              <p className="text-2xl font-medium text-gray-700">{card.translation}</p>
              {card.example && <p className="text-sm italic text-gray-500">{card.example}</p>}
            </div>

            {!hasPendingAdvance && (
              <button
                type="button"
                onClick={() => void submitAnswer('viewed')}
                disabled={submitting}
                className="rounded-lg bg-link-light px-6 py-3 font-semibold text-white transition-colors hover:bg-link-dark disabled:opacity-50"
              >
                Понял, дальше
              </button>
            )}

            {renderAnswerReview(card)}
          </div>
        )

      case 'choice':
        return (
          <div className="space-y-6 text-center">
            <div className="space-y-3">
              <div className="flex flex-wrap items-center justify-center gap-3">
                <h2 className="text-4xl font-bold text-text-light">{card.word}</h2>
                {card.audio_url && (
                  <button
                    type="button"
                    onClick={() => playAudio(card.audio_url)}
                    className="text-3xl"
                    title="Прослушать слово"
                  >
                    🔊
                  </button>
                )}
              </div>
              {card.phonetic && <p className="text-lg text-gray-400">[{card.phonetic}]</p>}
            </div>

            <div className="grid gap-3">
              {(card.options || []).map((option) => {
                const isCorrectOption = hasPendingAdvance && option === feedback?.correctAnswer
                const isSelectedWrong = hasPendingAdvance && option === feedback?.submittedAnswer && !feedback?.correct

                return (
                  <button
                    key={option}
                    type="button"
                    onClick={() => void submitAnswer(option)}
                    disabled={submitting || hasPendingAdvance}
                    className={`rounded-lg border-2 px-5 py-4 text-left text-lg font-medium transition-all disabled:opacity-70 ${
                      isCorrectOption
                        ? 'border-green-300 bg-green-50 text-green-800'
                        : isSelectedWrong
                          ? 'border-red-300 bg-red-50 text-red-800'
                          : 'border-gray-200 bg-white text-text-light hover:border-link-light hover:shadow-sm'
                    }`}
                  >
                    {option}
                  </button>
                )
              })}
            </div>

            {renderAnswerReview(card)}
          </div>
        )

      case 'with_photo':
        return (
          <div className="space-y-6 text-center">
            <div className="relative mx-auto w-full max-w-md">
              {card.image_url ? (
                <img
                  src={config.getFullUrl(card.image_url)}
                  alt={card.word}
                  className="h-64 w-full rounded-xl object-cover"
                />
              ) : (
                <div className="flex h-64 w-full items-center justify-center rounded-xl bg-gray-100 text-gray-400">
                  Нет картинки
                </div>
              )}

              {card.audio_url && (
                <button
                  type="button"
                  onClick={() => playAudio(card.audio_url)}
                  className="absolute right-3 top-3 inline-flex h-12 w-12 items-center justify-center rounded-full border border-white/80 bg-white/90 text-2xl text-link-light shadow-lg transition-transform hover:scale-105"
                  title="Прослушать слово"
                >
                  <span aria-hidden="true">🔊</span>
                  <span className="sr-only">Прослушать слово</span>
                </button>
              )}
            </div>

            <div className="space-y-3">
              <p className="text-lg text-gray-500">Введите слово по картинке</p>
              <input
                type="text"
                value={textAnswer}
                onChange={(event) => setTextAnswer(event.target.value)}
                onKeyDown={(event) => {
                  if (event.key === 'Enter' && textAnswer.trim() && !hasPendingAdvance) {
                    void submitAnswer(textAnswer)
                  }
                }}
                disabled={hasPendingAdvance}
                className="w-full rounded-lg border-2 border-gray-300 px-4 py-3 text-xl focus:border-link-light focus:outline-none disabled:bg-gray-100"
                placeholder="Введите английское слово"
                autoFocus
              />
              {!hasPendingAdvance && (
                <button
                  type="button"
                  onClick={() => void submitAnswer(textAnswer)}
                  disabled={!textAnswer.trim() || submitting}
                  className="rounded-lg bg-link-light px-6 py-3 font-semibold text-white transition-colors hover:bg-link-dark disabled:opacity-50"
                >
                  Проверить
                </button>
              )}
            </div>

            {renderAnswerReview(card)}
          </div>
        )

      case 'russian':
        return (
          <div className="space-y-6 text-center">
            <div className="space-y-3">
              <p className="text-3xl font-medium text-gray-700">{card.translation}</p>
              {card.example && <p className="text-sm italic text-gray-500">{card.example}</p>}
            </div>

            <input
              type="text"
              value={textAnswer}
              onChange={(event) => setTextAnswer(event.target.value)}
              onKeyDown={(event) => {
                if (event.key === 'Enter' && textAnswer.trim() && !hasPendingAdvance) {
                  void submitAnswer(textAnswer)
                }
              }}
              disabled={hasPendingAdvance}
              className="w-full rounded-lg border-2 border-gray-300 px-4 py-3 text-xl focus:border-link-light focus:outline-none disabled:bg-gray-100"
              placeholder="Введите слово на английском"
              autoFocus
            />

            {!hasPendingAdvance && (
              <button
                type="button"
                onClick={() => void submitAnswer(textAnswer)}
                disabled={!textAnswer.trim() || submitting}
                className="rounded-lg bg-link-light px-6 py-3 font-semibold text-white transition-colors hover:bg-link-dark disabled:opacity-50"
              >
                Проверить
              </button>
            )}

            {renderAnswerReview(card)}
          </div>
        )

      case 'constructor':
        return (
          <div className="space-y-6 text-center">
            <div className="space-y-2">
              <p className="text-3xl font-medium text-gray-700">{card.translation}</p>
              {card.phonetic && <p className="text-sm text-gray-400">[{card.phonetic}]</p>}
            </div>

            <div className="min-h-20 rounded-xl bg-gray-100 p-4">
              <div className="flex flex-wrap justify-center gap-2">
                {constructorAnswer.length === 0 ? (
                  <span className="text-gray-400">Соберите слово из букв</span>
                ) : (
                  constructorAnswer.map((letter, index) => (
                    <button
                      key={`${letter}-${index}`}
                      type="button"
                      onClick={() => {
                        const nextAnswer = [...constructorAnswer]
                        const [removed] = nextAnswer.splice(index, 1)
                        setConstructorAnswer(nextAnswer)
                        setConstructorPool([...constructorPool, removed])
                      }}
                      disabled={hasPendingAdvance}
                      className="rounded-lg border-2 border-link-light bg-white px-4 py-2 text-2xl font-bold text-link-light disabled:opacity-70"
                    >
                      {letter}
                    </button>
                  ))
                )}
              </div>
            </div>

            <div className="flex flex-wrap justify-center gap-2">
              {constructorPool.map((letter, index) => (
                <button
                  key={`${letter}-${index}`}
                  type="button"
                  onClick={() => {
                    setConstructorAnswer([...constructorAnswer, letter])
                    const nextPool = [...constructorPool]
                    nextPool.splice(index, 1)
                    setConstructorPool(nextPool)
                  }}
                  disabled={hasPendingAdvance}
                  className="rounded-lg border-2 border-gray-300 bg-white px-4 py-2 text-2xl font-bold text-text-light disabled:opacity-70"
                >
                  {letter}
                </button>
              ))}
            </div>

            <div className="flex items-center justify-center gap-3">
              <button
                type="button"
                onClick={() => {
                  setConstructorPool([...constructorPool, ...constructorAnswer])
                  setConstructorAnswer([])
                }}
                disabled={hasPendingAdvance}
                className="rounded-lg bg-gray-200 px-4 py-2 text-sm font-semibold text-gray-700 disabled:opacity-70"
              >
                Очистить
              </button>

              {!hasPendingAdvance && (
                <button
                  type="button"
                  onClick={() => void submitAnswer(constructorAnswer.join(''))}
                  disabled={constructorAnswer.length === 0 || submitting}
                  className="rounded-lg bg-link-light px-6 py-3 font-semibold text-white transition-colors hover:bg-link-dark disabled:opacity-50"
                >
                  Проверить
                </button>
              )}
            </div>

            {renderAnswerReview(card)}
          </div>
        )

      default:
        return <div className="text-center text-text-light">Сессия завершена</div>
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div className="max-h-[92vh] w-full max-w-4xl overflow-hidden rounded-2xl bg-card-light shadow-2xl">
        <div className="border-b border-gray-200 p-5">
          <div className="flex items-start justify-between gap-4">
            <div className="space-y-2">
              <h2 className="text-2xl font-bold text-text-light">{course.title}</h2>
              <p className="text-sm text-gray-500">{deck.title}</p>
              <div className="space-y-1">
                <div className="flex items-center justify-between gap-4 text-sm text-gray-500">
                  <span>Общий прогресс сессии</span>
                  <span className="font-semibold text-link-light">{progressValue}%</span>
                </div>
                <div className="h-2 rounded-full bg-gray-200">
                  <div className="h-2 rounded-full bg-link-light transition-all" style={{ width: `${progressValue}%` }} />
                </div>
              </div>
            </div>
            <button type="button" onClick={onClose} className="text-3xl text-gray-400 transition-colors hover:text-gray-600">
              ×
            </button>
          </div>
        </div>

        <div className="max-h-[calc(92vh-160px)] overflow-y-auto p-8">
          {loading ? (
            <div className="py-20 text-center text-text-light">Подготавливаю сессию...</div>
          ) : !session ? (
            <div className="space-y-4 py-16 text-center">
              <p className="text-lg font-medium text-text-light">Не удалось запустить обучение.</p>
              {errorMessage && (
                <div className="mx-auto max-w-xl rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
                  {errorMessage}
                </div>
              )}
              <button
                type="button"
                onClick={() => void startSession()}
                className="rounded-lg bg-link-light px-5 py-3 font-semibold text-white transition-colors hover:bg-link-dark"
              >
                Попробовать снова
              </button>
            </div>
          ) : !currentCard ? (
            <div className="space-y-4 py-16 text-center">
              <p className="text-2xl font-bold text-text-light">Сессия завершена</p>
              <p className="text-gray-500">Все слова в этой подборке пройдены.</p>
              <button
                type="button"
                onClick={onClose}
                className="rounded-lg bg-link-light px-5 py-3 font-semibold text-white transition-colors hover:bg-link-dark"
              >
                Закрыть
              </button>
            </div>
          ) : (
            <div className="space-y-6">
              <div className="hidden flex flex-wrap items-center justify-between gap-3 rounded-xl bg-white p-4">
                <div>
                  <p className="text-xs uppercase tracking-wide text-gray-400">
                    Шаг {currentCard.sequence_number} из {sessionForProgress?.cards.length || 0}
                  </p>
                  <p className="text-lg font-semibold text-link-light">
                    {getTrainingModeStepLabel(currentCard, currentCard.current_mode)}
                  </p>
                </div>
                <div className="text-right">
                  <p className="text-xs uppercase tracking-wide text-gray-400">Прогресс слова</p>
                  <p className="text-2xl font-bold text-text-light">{currentCard.progress_percentage}%</p>
                </div>
              </div>

              {feedback && !pendingSession && (
                <div
                  className={`rounded-xl border px-4 py-3 text-sm font-medium ${
                    feedback.correct ? 'border-green-200 bg-green-50 text-green-700' : 'border-red-200 bg-red-50 text-red-700'
                  }`}
                >
                  {feedback.text}
                </div>
              )}

              {renderMode(currentCard)}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
