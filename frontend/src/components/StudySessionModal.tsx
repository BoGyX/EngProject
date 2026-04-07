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

export default function StudySessionModal({ course, deck, onClose }: StudySessionModalProps) {
  const [session, setSession] = useState<TrainingSessionState | null>(null)
  const [loading, setLoading] = useState(true)
  const [submitting, setSubmitting] = useState(false)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)
  const [textAnswer, setTextAnswer] = useState('')
  const [constructorPool, setConstructorPool] = useState<string[]>([])
  const [constructorAnswer, setConstructorAnswer] = useState<string[]>([])
  const [feedback, setFeedback] = useState<{ correct: boolean; text: string } | null>(null)

  const currentCard = session?.current_card || null

  useEffect(() => {
    void startSession()
  }, [deck.id])

  useEffect(() => {
    if (!currentCard) {
      return
    }

    setTextAnswer('')
    setFeedback(null)

    if (currentCard.current_mode === 'constructor') {
      setConstructorAnswer([])
      setConstructorPool(shuffleLetters(currentCard.word))
    }
  }, [currentCard?.session_card_id, currentCard?.current_mode])

  const progressValue = useMemo(() => {
    if (!session || session.cards.length === 0) {
      return 0
    }

    const completedSteps = session.cards.filter((card) => card.is_completed).length
    return Math.round((completedSteps / session.cards.length) * 100)
  }, [session])

  const startSession = async () => {
    try {
      setLoading(true)
      setErrorMessage(null)
      setFeedback(null)

      const nextSession = await studyService.startTraining(deck.id, course.id)
      setSession(nextSession)
    } catch (error) {
      console.error('Error starting training session:', error)
      setSession(null)
      setErrorMessage(getErrorMessage(error, 'Не удалось запустить обучение. Попробуйте еще раз.'))
    } finally {
      setLoading(false)
    }
  }

  const submitAnswer = async (answer = '') => {
    if (!currentCard || !session) {
      return
    }

    try {
      setSubmitting(true)
      setFeedback(null)

      const response = await studyService.answerTraining(session.session.id, currentCard.session_card_id, answer)
      setSession(response.session)
      setFeedback({
        correct: response.is_correct,
        text: response.is_correct
          ? 'Верно. Переходим к следующему слову.'
          : 'Ошибка. Идем к следующему слову; этот режим вернется уже в новой сессии.',
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
            <button
              type="button"
              onClick={() => void submitAnswer('viewed')}
              disabled={submitting}
              className="rounded-lg bg-link-light px-6 py-3 font-semibold text-white transition-colors hover:bg-link-dark disabled:opacity-50"
            >
              Понял, дальше
            </button>
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
              {(card.options || []).map((option) => (
                <button
                  key={option}
                  type="button"
                  onClick={() => void submitAnswer(option)}
                  disabled={submitting}
                  className="rounded-lg border-2 border-gray-200 bg-white px-5 py-4 text-left text-lg font-medium text-text-light transition-all hover:border-link-light hover:shadow-sm disabled:opacity-50"
                >
                  {option}
                </button>
              ))}
            </div>
          </div>
        )

      case 'with_photo':
        return (
          <div className="space-y-6 text-center">
            {card.image_url ? (
              <img
                src={config.getFullUrl(card.image_url)}
                alt={card.word}
                className="mx-auto h-64 w-full max-w-md rounded-xl object-cover"
              />
            ) : (
              <div className="mx-auto flex h-64 w-full max-w-md items-center justify-center rounded-xl bg-gray-100 text-gray-400">
                Нет картинки
              </div>
            )}
            <div className="space-y-3">
              <p className="text-lg text-gray-500">Введите слово по картинке</p>
              <input
                type="text"
                value={textAnswer}
                onChange={(event) => setTextAnswer(event.target.value)}
                onKeyDown={(event) => {
                  if (event.key === 'Enter' && textAnswer.trim()) {
                    void submitAnswer(textAnswer)
                  }
                }}
                className="w-full rounded-lg border-2 border-gray-300 px-4 py-3 text-xl focus:border-link-light focus:outline-none"
                placeholder="Введите английское слово"
                autoFocus
              />
              <button
                type="button"
                onClick={() => void submitAnswer(textAnswer)}
                disabled={!textAnswer.trim() || submitting}
                className="rounded-lg bg-link-light px-6 py-3 font-semibold text-white transition-colors hover:bg-link-dark disabled:opacity-50"
              >
                Проверить
              </button>
            </div>
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
                if (event.key === 'Enter' && textAnswer.trim()) {
                  void submitAnswer(textAnswer)
                }
              }}
              className="w-full rounded-lg border-2 border-gray-300 px-4 py-3 text-xl focus:border-link-light focus:outline-none"
              placeholder="Введите слово на английском"
              autoFocus
            />
            <button
              type="button"
              onClick={() => void submitAnswer(textAnswer)}
              disabled={!textAnswer.trim() || submitting}
              className="rounded-lg bg-link-light px-6 py-3 font-semibold text-white transition-colors hover:bg-link-dark disabled:opacity-50"
            >
              Проверить
            </button>
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
                      className="rounded-lg border-2 border-link-light bg-white px-4 py-2 text-2xl font-bold text-link-light"
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
                  className="rounded-lg border-2 border-gray-300 bg-white px-4 py-2 text-2xl font-bold text-text-light"
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
                className="rounded-lg bg-gray-200 px-4 py-2 text-sm font-semibold text-gray-700"
              >
                Очистить
              </button>
              <button
                type="button"
                onClick={() => void submitAnswer(constructorAnswer.join(''))}
                disabled={constructorAnswer.length === 0 || submitting}
                className="rounded-lg bg-link-light px-6 py-3 font-semibold text-white transition-colors hover:bg-link-dark disabled:opacity-50"
              >
                Проверить
              </button>
            </div>
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
              <div className="flex flex-wrap items-center justify-between gap-3 rounded-xl bg-white p-4">
                <div>
                  <p className="text-xs uppercase tracking-wide text-gray-400">
                    Шаг {currentCard.sequence_number} из {session.cards.length}
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

              {feedback && (
                <div
                  className={`rounded-xl border px-4 py-3 text-sm font-medium ${
                    feedback.correct
                      ? 'border-green-200 bg-green-50 text-green-700'
                      : 'border-red-200 bg-red-50 text-red-700'
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
