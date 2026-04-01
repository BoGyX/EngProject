import { useState, useEffect } from 'react'
import api from '../services/api'
import { vocabularyService, VocabularyWord } from '../services/vocabularyService'

interface Course {
  id: number
  title: string
  description?: string
}

interface Deck {
  id: number
  course_id: number
  title: string
  description?: string
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

interface LearnedWord extends Card {
  learned: boolean
}

interface CourseWithDecks {
  course: Course
  decks: Deck[]
  learnedWordsCount: number
}

export default function Vocabulary() {
  const [coursesData, setCoursesData] = useState<CourseWithDecks[]>([])
  const [expandedCourses, setExpandedCourses] = useState<Set<number>>(new Set())
  const [selectedDeck, setSelectedDeck] = useState<{ courseId: number; deckId: number } | null>(null)
  const [learnedWords, setLearnedWords] = useState<LearnedWord[]>([])
  const [loading, setLoading] = useState(true)
  const [loadingWords, setLoadingWords] = useState(false)
  
  // Форма добавления слова в текущий deck
  const [showAddWordForm, setShowAddWordForm] = useState(false)
  const [newWord, setNewWord] = useState('')
  const [newTranslation, setNewTranslation] = useState('')
  const [addingWord, setAddingWord] = useState(false)

  const getUserId = () => {
    const authStorage = localStorage.getItem('auth-storage')
    if (!authStorage) return null
    const parsed = JSON.parse(authStorage)
    return parsed?.state?.user?.id
  }

  useEffect(() => {
    loadCoursesWithDecks()
  }, [])

  const loadCoursesWithDecks = async () => {
    try {
      setLoading(true)
      
      const userId = getUserId()
      if (!userId) {
        setLoading(false)
        return
      }
      
      const coursesResponse = await api.get<Course[]>('/courses')
      const allCourses = coursesResponse.data || []

      // Загружаем личный словарь пользователя
      const vocabResponse = await api.get(`/vocabulary?user_id=${userId}`)
      const personalVocab = vocabResponse.data || []
      
      // Создаем Set изученных слов для быстрого поиска
      const learnedWordsSet = new Set(personalVocab.map((v: any) => v.word.toLowerCase()))

      const coursesWithDecks: CourseWithDecks[] = []

      for (const course of allCourses) {
        const decksResponse = await api.get<Deck[]>(`/decks?course_id=${course.id}`)
        const decks = decksResponse.data || []

        // Считаем общее количество изученных слов в курсе
        let totalLearnedWords = 0
        for (const deck of decks) {
          const cardsResponse = await api.get<Card[]>(`/cards?deck_id=${deck.id}`)
          const cards = cardsResponse.data || []

          // Проверяем, какие слова из этого deck есть в личном словаре
          cards.forEach(card => {
            if (learnedWordsSet.has(card.word.toLowerCase())) {
              totalLearnedWords++
            }
          })
        }

        coursesWithDecks.push({
          course,
          decks,
          learnedWordsCount: totalLearnedWords
        })
      }

      setCoursesData(coursesWithDecks)
    } catch (error) {
      console.error('Error loading courses:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleAddWord = async () => {
    if (!newWord.trim() || !newTranslation.trim()) {
      alert('Заполните слово и перевод')
      return
    }

    const userId = getUserId()
    if (!userId || !selectedDeck) {
      alert('Ошибка: не выбран подкурс или пользователь не авторизован')
      return
    }

    try {
      setAddingWord(true)
      // Добавляем слово в личный словарь как изученное
      await vocabularyService.addWord({
        user_id: userId,
        word: newWord.trim(),
        translation: newTranslation.trim(),
        status: 'learned'
      })
      
      setNewWord('')
      setNewTranslation('')
      setShowAddWordForm(false)
      
      // Перезагружаем слова для текущего deck
      await loadLearnedWords(selectedDeck.courseId, selectedDeck.deckId)
      await loadCoursesWithDecks()
    } catch (error) {
      console.error('Error adding word:', error)
      alert('Ошибка при добавлении слова')
    } finally {
      setAddingWord(false)
    }
  }

  const handleDeleteWord = async (word: string) => {
    if (!confirm('Удалить это слово из изученных?')) return

    const userId = getUserId()
    if (!userId || !selectedDeck) return

    try {
      // Находим слово в личном словаре пользователя
      const vocabResponse = await api.get<VocabularyWord[]>(`/vocabulary?user_id=${userId}`)
      const personalVocab = vocabResponse.data || []
      const wordToDelete = personalVocab.find(v => v.word.toLowerCase() === word.toLowerCase())
      
      if (wordToDelete) {
        await vocabularyService.deleteWord(wordToDelete.id)
        // Перезагружаем слова
        await loadLearnedWords(selectedDeck.courseId, selectedDeck.deckId)
        await loadCoursesWithDecks()
      }
    } catch (error) {
      console.error('Error deleting word:', error)
      alert('Ошибка при удалении слова')
    }
  }

  const toggleCourse = (courseId: number) => {
    const newExpanded = new Set(expandedCourses)
    if (newExpanded.has(courseId)) {
      newExpanded.delete(courseId)
    } else {
      newExpanded.add(courseId)
    }
    setExpandedCourses(newExpanded)
  }

  const loadLearnedWords = async (courseId: number, deckId: number) => {
    try {
      setLoadingWords(true)
      setSelectedDeck({ courseId, deckId })
      setShowAddWordForm(false) // Закрываем форму при переключении deck

      const userId = getUserId()
      if (!userId) {
        setLoadingWords(false)
        return
      }

      const response = await api.get<Card[]>(`/cards?deck_id=${deckId}`)
      const allCards = response.data || []

      // Загружаем личный словарь пользователя
      const vocabResponse = await api.get(`/vocabulary?user_id=${userId}`)
      const personalVocab = vocabResponse.data || []
      
      // Создаем Map для быстрого поиска (слово -> объект из personal_vocabulary)
      const learnedWordsMap = new Map(
        personalVocab.map((v: any) => [v.word.toLowerCase(), v])
      )

      // Фильтруем карточки, которые есть в личном словаре
      const learned = allCards
        .filter(card => learnedWordsMap.has(card.word.toLowerCase()))
        .map(card => ({ ...card, learned: true }))

      // Добавляем личные слова пользователя, которых нет в cards
      personalVocab.forEach((vocabWord: any) => {
        const existsInCards = allCards.some(
          card => card.word.toLowerCase() === vocabWord.word.toLowerCase()
        )
        if (!existsInCards) {
          // Это личное слово пользователя, добавляем его
          learned.push({
            id: vocabWord.id,
            deck_id: deckId,
            word: vocabWord.word,
            translation: vocabWord.translation,
            phonetic: vocabWord.phonetic,
            audio_url: vocabWord.audio_url,
            example: vocabWord.example,
            learned: true
          })
        }
      })

      setLearnedWords(learned)
    } catch (error) {
      console.error('Error loading learned words:', error)
      setLearnedWords([])
    } finally {
      setLoadingWords(false)
    }
  }

  const playAudio = (audioUrl: string) => {
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
    return <div className="text-center py-8 text-text-light">Загрузка...</div>
  }

  return (
    <div>
      <h1 className="text-3xl font-bold text-text-light mb-6">Мой словарь</h1>

      {coursesData.length === 0 ? (
        <div className="bg-card-light shadow-md rounded-lg p-6 border border-gray-200 text-center">
          <p className="text-text-light">Начните изучать курсы, чтобы пополнить словарь</p>
        </div>
      ) : (
        <div className="space-y-4">
          {coursesData.map(({ course, decks, learnedWordsCount }) => (
            <div key={course.id} className="bg-card-light shadow-md rounded-lg border border-gray-200 overflow-hidden">
              {/* Заголовок курса */}
              <button
                onClick={() => toggleCourse(course.id)}
                className="w-full p-6 hover:bg-gray-50 transition-colors"
              >
                <div className="flex items-center justify-between">
                  <div className="text-left">
                    <h3 className="text-xl font-bold text-text-light mb-2">{course.title}</h3>
                    <p className="text-sm text-gray-500">
                      Уроков: {decks.length} | Изучено слов: {learnedWordsCount}
                    </p>
                  </div>
                  <span className="text-2xl text-gray-400">
                    {expandedCourses.has(course.id) ? '▼' : '▶'}
                  </span>
                </div>
              </button>

              {/* Список деков с изученными словами */}
              {expandedCourses.has(course.id) && (
                <div className="border-t border-gray-200 bg-gray-50 p-6">
                  {decks.length === 0 ? (
                    <p className="text-center text-gray-500">Уроков пока нет в этом курсе</p>
                  ) : (
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                      {decks.map((deck) => {
                        const isSelected = selectedDeck?.courseId === course.id && selectedDeck?.deckId === deck.id
                        return (
                          <button
                            key={deck.id}
                            onClick={() => loadLearnedWords(course.id, deck.id)}
                            className={`bg-white rounded-lg p-4 border-2 transition-all text-left ${
                              isSelected
                                ? 'border-link-light shadow-md'
                                : 'border-gray-200 hover:border-link-light hover:shadow-sm'
                            }`}
                          >
                            <h4 className="font-semibold text-text-light mb-1">{deck.title}</h4>
                            {deck.description && (
                              <p className="text-xs text-gray-500 mb-2">{deck.description}</p>
                            )}
                            <div className="text-xs text-link-light font-medium">
                              Нажмите для просмотра слов →
                            </div>
                          </button>
                        )
                      })}
                    </div>
                  )}
                </div>
              )}
            </div>
          ))}
        </div>
      )}

      {/* Модальное окно с изученными словами */}
      {selectedDeck && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="bg-card-light rounded-lg shadow-2xl max-w-4xl w-full max-h-[90vh] overflow-hidden">
            {/* Заголовок */}
            <div className="sticky top-0 bg-card-light border-b border-gray-200 p-6">
              <div className="flex items-center justify-between mb-4">
                <div>
                  <h2 className="text-2xl font-bold text-text-light">Изученные слова</h2>
                  <p className="text-sm text-gray-500 mt-1">
                    {learnedWords.length} {learnedWords.length === 1 ? 'слово' : 'слов'}
                  </p>
                </div>
                <div className="flex items-center space-x-3">
                  <button
                    onClick={() => setShowAddWordForm(!showAddWordForm)}
                    className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-semibold text-sm"
                  >
                    {showAddWordForm ? '✕ Отмена' : '+ Добавить слово'}
                  </button>
                  <button
                    onClick={() => {
                      setSelectedDeck(null)
                      setLearnedWords([])
                      setShowAddWordForm(false)
                    }}
                    className="text-3xl text-gray-400 hover:text-gray-600 transition-colors"
                  >
                    ×
                  </button>
                </div>
              </div>

              {/* Форма добавления слова */}
              {showAddWordForm && (
                <div className="bg-blue-50 p-4 rounded-lg border-2 border-blue-300">
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-3 mb-3">
                    <div>
                      <label className="block text-xs font-medium text-gray-700 mb-1">
                        Слово (English)
                      </label>
                      <input
                        type="text"
                        value={newWord}
                        onChange={(e) => setNewWord(e.target.value)}
                        placeholder="например: apple"
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-sm"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-medium text-gray-700 mb-1">
                        Перевод (Русский)
                      </label>
                      <input
                        type="text"
                        value={newTranslation}
                        onChange={(e) => setNewTranslation(e.target.value)}
                        placeholder="например: яблоко"
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-sm"
                      />
                    </div>
                  </div>
                  <button
                    onClick={handleAddWord}
                    disabled={addingWord}
                    className="w-full px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors font-semibold disabled:bg-gray-400 text-sm"
                  >
                    {addingWord ? 'Добавление...' : '✓ Добавить в изученные'}
                  </button>
                </div>
              )}
            </div>

            {/* Контент */}
            <div className="p-6 overflow-y-auto max-h-[calc(90vh-200px)]">
              {loadingWords ? (
                <div className="text-center py-12 text-text-light">Загрузка слов...</div>
              ) : learnedWords.length === 0 ? (
                <div className="text-center py-12">
                  <p className="text-text-light mb-2">В этом уроке пока нет изученных слов</p>
                  <p className="text-xs text-gray-400">
                    Пройдите обучение или добавьте свои слова
                  </p>
                </div>
              ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  {learnedWords.map((word) => (
                    <div
                      key={word.id}
                      className="p-4 border-2 border-green-200 bg-green-50 rounded-lg hover:shadow-md transition-all"
                    >
                      <div className="flex items-start justify-between mb-2">
                        <div className="flex items-center space-x-2 flex-1">
                          <span className="text-green-600 text-lg">✓</span>
                          <div className="font-bold text-lg text-text-light">{word.word}</div>
                        </div>
                        <div className="flex items-center space-x-2">
                          {word.audio_url && (
                            <button
                              onClick={() => playAudio(word.audio_url!)}
                              className="text-green-600 hover:text-green-800 text-xl transition-colors"
                              title="Прослушать произношение"
                            >
                              🔊
                            </button>
                          )}
                          <button
                            onClick={() => handleDeleteWord(word.word)}
                            className="text-red-500 hover:text-red-700 text-lg transition-colors"
                            title="Удалить из изученных"
                          >
                            🗑️
                          </button>
                        </div>
                      </div>
                      {word.phonetic && (
                        <div className="text-sm text-gray-400 mb-2">[{word.phonetic}]</div>
                      )}
                      <div className="text-base text-gray-700 font-medium mb-2">{word.translation}</div>
                      {word.example && (
                        <div className="text-sm text-gray-500 italic border-l-2 border-green-400 pl-2 mt-2">
                          "{word.example}"
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
