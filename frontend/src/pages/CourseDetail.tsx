import { useState, useEffect } from 'react'
import { useParams, Link } from 'react-router-dom'
import api from '../services/api'
import StudyCards from '../components/StudyCards'
import { config } from '../config'

interface Course {
  id: number
  title: string
  description?: string
  image_url?: string
  is_published: boolean
  created_at: string
}

interface Deck {
  id: number
  course_id: number
  title: string
  description?: string
  position: number
}

interface Card {
  id: number
  deck_id: number
  word: string
  translation: string
  phonetic?: string
  audio_url?: string
  image_url?: string
  example?: string
}

export default function CourseDetail() {
  const { id } = useParams<{ id: string }>()
  const [course, setCourse] = useState<Course | null>(null)
  const [decks, setDecks] = useState<Deck[]>([])
  const [selectedDeck, setSelectedDeck] = useState<Deck | null>(null)
  const [cards, setCards] = useState<Card[]>([])
  const [loading, setLoading] = useState(true)
  const [loadingCards, setLoadingCards] = useState(false)
  const [showStudyModal, setShowStudyModal] = useState(false)

  useEffect(() => {
    if (id) {
      loadCourse()
      loadDecks()
    }
  }, [id])

  const loadCourse = async () => {
    try {
      const response = await api.get<Course>(`/courses/${id}`)
      setCourse(response.data)
    } catch (error: any) {
      console.error('Error loading course:', error)
    }
  }

  const loadDecks = async () => {
    try {
      setLoading(true)
      const response = await api.get<Deck[]>(`/decks?course_id=${id}`)
      setDecks(response.data || [])
    } catch (error: any) {
      console.error('Error loading decks:', error)
      setDecks([])
    } finally {
      setLoading(false)
    }
  }

  const loadCards = async (deckId: number) => {
    try {
      setLoadingCards(true)
      const response = await api.get<Card[]>(`/cards?deck_id=${deckId}`)
      const cardsData = response.data || []
      console.log('Loaded cards:', cardsData)
      console.log('Cards with images:', cardsData.filter(c => c.image_url))
      setCards(cardsData)
    } catch (error: any) {
      console.error('Error loading cards:', error)
      setCards([])
    } finally {
      setLoadingCards(false)
    }
  }

  const handleDeckClick = async (deck: Deck) => {
    setSelectedDeck(deck)
    await loadCards(deck.id)
  }

  const handleStartStudy = () => {
    if (cards.length > 0) {
      setShowStudyModal(true)
    }
  }

  const playAudio = (audioUrl: string) => {
    // Если URL относительный (начинается с //), добавляем протокол https:
    let url = audioUrl
    if (url.startsWith('//')) {
      url = `https:${url}`
    }
    const audio = new Audio(url)
    audio.play().catch(err => {
      console.error('Error playing audio:', err)
    })
  }

  if (loading) {
    return (
      <div className="text-center py-8 text-text-light">
        Загрузка курса...
      </div>
    )
  }

  if (!course) {
    return (
      <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg">
        Курс не найден
      </div>
    )
  }

  return (
    <div>
      <Link to="/courses" className="text-link-light hover:text-link-dark mb-4 inline-block transition-colors">
        ← Назад к курсам
      </Link>

      {/* Информация о курсе */}
      <div className="bg-card-light shadow-md rounded-lg p-6 mb-6 border border-gray-200">
        <div className="flex items-start space-x-6">
          {course.image_url && (
            <img
              src={config.getFullUrl(course.image_url)}
              alt={course.title}
              className="w-32 h-32 object-cover rounded-lg"
              onError={(e) => {
                (e.target as HTMLImageElement).style.display = 'none'
              }}
            />
          )}
          <div className="flex-1">
            <h1 className="text-3xl font-bold text-text-light mb-2">{course.title}</h1>
            {course.description && (
              <p className="text-text-light mb-4">{course.description}</p>
            )}
            <div className="flex items-center space-x-4 text-sm text-gray-500">
              <span>Создан: {new Date(course.created_at).toLocaleDateString('ru-RU')}</span>
              {course.is_published && (
                <span className="px-2 py-1 bg-green-100 text-green-800 text-xs font-semibold rounded-full">
                  Опубликован
                </span>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* Деки и карточки */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* Список дек */}
        <div className="bg-card-light shadow-md rounded-lg p-6 border border-gray-200">
          <h2 className="text-xl font-semibold mb-4 text-text-light">Уроки (Деки)</h2>
          {decks.length === 0 ? (
            <p className="text-text-light">Уроков пока нет в этом курсе</p>
          ) : (
            <div className="space-y-3">
              {decks.map((deck) => (
                <div
                  key={deck.id}
                  className={`p-4 border-2 rounded-lg cursor-pointer transition-all ${
                    selectedDeck?.id === deck.id
                      ? 'border-link-light bg-link-light bg-opacity-10 shadow-md'
                      : 'border-gray-200 hover:border-link-light hover:shadow-sm'
                  }`}
                  onClick={() => handleDeckClick(deck)}
                >
                  <div className="font-semibold text-text-light mb-1">{deck.title}</div>
                  {deck.description && (
                    <div className="text-sm text-gray-500">{deck.description}</div>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Карточки выбранной деки */}
        <div className="bg-card-light shadow-md rounded-lg p-6 border border-gray-200">
          <h2 className="text-xl font-semibold mb-4 text-text-light">
            Карточки {selectedDeck && `(${selectedDeck.title})`}
          </h2>

          {!selectedDeck ? (
            <div className="text-center py-8">
              <p className="text-text-light mb-2">Выберите урок для просмотра карточек</p>
              <p className="text-xs text-gray-400">Кликните на урок слева</p>
            </div>
          ) : loadingCards ? (
            <div className="text-center py-8 text-text-light">Загрузка карточек...</div>
          ) : cards.length === 0 ? (
            <div className="text-center py-8">
              <p className="text-text-light">В этом уроке пока нет карточек</p>
            </div>
          ) : (
            <>
              <div className="mb-4">
                <button
                  onClick={handleStartStudy}
                  className="w-full bg-link-light hover:bg-link-dark text-white px-6 py-3 rounded-lg font-semibold text-lg transition-colors shadow-md"
                >
                  🎓 Начать изучение ({cards.length} карточек)
                </button>
              </div>
              <div className="space-y-4">
                {cards.map((card) => (
                  <div
                    key={card.id}
                    className="p-4 border-2 border-gray-200 rounded-lg hover:border-link-light hover:shadow-sm transition-all bg-white"
                  >
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <div className="flex items-center space-x-2 mb-2">
                          <div className="font-bold text-xl text-text-light">{card.word}</div>
                          {card.phonetic && (
                            <div className="text-sm text-gray-400">[{card.phonetic}]</div>
                          )}
                          {card.audio_url && (
                            <button
                              onClick={() => playAudio(card.audio_url!)}
                              className="text-green-600 hover:text-green-800 text-lg transition-colors"
                              title="Прослушать произношение"
                            >
                              🔊
                            </button>
                          )}
                        </div>
                        <div className="text-lg text-gray-700 font-medium mb-2">{card.translation}</div>
                        {card.example && (
                          <div className="text-sm text-gray-500 italic border-l-2 border-link-light pl-2 mt-2">
                            "{card.example}"
                          </div>
                        )}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </>
          )}
        </div>
      </div>

      {/* Модальное окно для изучения карточек */}
      {showStudyModal && selectedDeck && (
        <StudyCards
          cards={cards}
          deckTitle={selectedDeck.title}
          onClose={() => setShowStudyModal(false)}
        />
      )}
    </div>
  )
}
