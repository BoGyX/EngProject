import { useState, useEffect } from 'react'
import { vocabularyService } from '../services/vocabularyService'
import api from '../services/api'
import { config } from '../config'

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

interface StudyCardsProps {
  cards: Card[]
  deckTitle: string
  onClose: () => void
}

type StudyMode = 'view' | 'withPhoto' | 'withoutPhoto' | 'russian' | 'translate' | 'constructor'

export default function StudyCards({ cards, deckTitle, onClose }: StudyCardsProps) {
  const [currentIndex, setCurrentIndex] = useState(0)
  const [studyMode, setStudyMode] = useState<StudyMode>('view')
  const [showAnswer, setShowAnswer] = useState(false)
  const [userAnswer, setUserAnswer] = useState('')
  const [isCorrect, setIsCorrect] = useState<boolean | null>(null)
  const [learnedCards, setLearnedCards] = useState<Set<number>>(new Set())
  const [constructorLetters, setConstructorLetters] = useState<string[]>([])
  const [constructorAnswer, setConstructorAnswer] = useState<string[]>([])
  const [modeProgress, setModeProgress] = useState<Record<StudyMode, Set<number>>>({
    view: new Set(),
    withPhoto: new Set(),
    withoutPhoto: new Set(),
    russian: new Set(),
    translate: new Set(),
    constructor: new Set()
  })

  const currentCard = cards[currentIndex]
  
  // Вычисляем общий прогресс по всем режимам
  const calculateOverallProgress = () => {
    const modes: StudyMode[] = ['view', 'withPhoto', 'withoutPhoto', 'russian', 'constructor']
    let totalProgress = 0
    
    modes.forEach(mode => {
      const modeCardsLearned = modeProgress[mode].size
      const modeProgressPercent = (modeCardsLearned / cards.length) * 100
      totalProgress += modeProgressPercent
    })
    
    return totalProgress / modes.length
  }
  
  const overallProgress = calculateOverallProgress()

  // Загружаем прогресс из backend API
  useEffect(() => {
    const loadProgress = async () => {
      try {
        const authStorage = localStorage.getItem('auth-storage')
        if (!authStorage) {
          console.log('No auth storage found')
          return
        }
        
        const parsed = JSON.parse(authStorage)
        const userId = parsed?.state?.user?.id
        if (!userId) {
          console.log('No user ID found')
          return
        }

        console.log('Loading progress for user:', userId)
        console.log('User ID type:', typeof userId)
        console.log('User ID length:', userId.length)

        // Загружаем прогресс пользователя по карточкам
        const response = await api.get(`/user-cards/user/${userId}`)
        const userCards = response.data || []
        
        console.log('User cards loaded:', userCards)
        
        // Фильтруем карточки текущего deck
        const deckCards = userCards.filter((uc: any) => 
          cards.some(c => c.id === uc.card_id)
        )
        
        console.log('Deck cards filtered:', deckCards)
        
        // Формируем прогресс по режимам из backend
        const loadedProgress: Record<StudyMode, Set<number>> = {
          view: new Set(),
          withPhoto: new Set(),
          withoutPhoto: new Set(),
          russian: new Set(),
          translate: new Set(),
          constructor: new Set()
        }
        
        deckCards.forEach((uc: any) => {
          console.log(`Card ${uc.card_id} modes:`, {
            view: uc.mode_view,
            withPhoto: uc.mode_with_photo,
            withoutPhoto: uc.mode_without_photo,
            russian: uc.mode_russian,
            constructor: uc.mode_constructor
          })
          
          if (uc.mode_view) loadedProgress.view.add(uc.card_id)
          if (uc.mode_with_photo) loadedProgress.withPhoto.add(uc.card_id)
          if (uc.mode_without_photo) loadedProgress.withoutPhoto.add(uc.card_id)
          if (uc.mode_russian) loadedProgress.russian.add(uc.card_id)
          if (uc.mode_constructor) loadedProgress.constructor.add(uc.card_id)
        })
        
        console.log('Loaded progress:', {
          view: Array.from(loadedProgress.view),
          withPhoto: Array.from(loadedProgress.withPhoto),
          withoutPhoto: Array.from(loadedProgress.withoutPhoto),
          russian: Array.from(loadedProgress.russian),
          constructor: Array.from(loadedProgress.constructor)
        })
        
        setModeProgress(loadedProgress)
      } catch (e) {
        console.error('Error loading progress:', e)
      }
    }
    
    loadProgress()
  }, [cards])

  // Сохраняем прогресс в backend API
  const saveProgress = async (cardId: number, isCorrect: boolean) => {
    const newLearnedCards = new Set(learnedCards)
    if (isCorrect) {
      newLearnedCards.add(cardId)
    }
    setLearnedCards(newLearnedCards)
    
    // Обновляем прогресс для текущего режима
    const newModeProgress = { ...modeProgress }
    if (isCorrect) {
      newModeProgress[studyMode] = new Set([...newModeProgress[studyMode], cardId])
    }
    setModeProgress(newModeProgress)
    
    // Проверяем, изучено ли слово полностью (во всех режимах кроме view)
    const modesForCompletion: StudyMode[] = ['withPhoto', 'withoutPhoto', 'russian', 'constructor']
    const isFullyLearned = modesForCompletion.every(mode => 
      newModeProgress[mode].has(cardId)
    )
    
    // Если слово полностью изучено, добавляем в личный словарь
    if (isFullyLearned && currentCard) {
      addToPersonalVocabulary(currentCard)
    }
    
    // Сохраняем в backend API
    try {
      const authStorage = localStorage.getItem('auth-storage')
      if (!authStorage) {
        console.log('No auth storage for saving progress')
        return
      }
      
      const parsed = JSON.parse(authStorage)
      const userId = parsed?.state?.user?.id
      if (!userId) {
        console.log('No user ID for saving progress')
        return
      }

      console.log('Saving progress for card:', cardId, 'mode:', studyMode, 'isCorrect:', isCorrect)

      // Проверяем, существует ли запись user_card
      const existingCards = await api.get(`/user-cards/user/${userId}`)
      const existingCard = existingCards.data?.find((uc: any) => uc.card_id === cardId)
      
      console.log('Existing card:', existingCard)
      
      // Формируем данные для режимов
      const modeData: any = {
        status: isFullyLearned ? 'learned' : 'learning',
        last_seen: new Date().toISOString(),
        next_review: new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString()
      }
      
      // Добавляем флаги режимов ТОЛЬКО если ответ правильный
      if (isCorrect) {
        if (studyMode === 'view') modeData.mode_view = true
        if (studyMode === 'withPhoto') modeData.mode_with_photo = true
        if (studyMode === 'withoutPhoto') modeData.mode_without_photo = true
        if (studyMode === 'russian') modeData.mode_russian = true
        if (studyMode === 'constructor') modeData.mode_constructor = true
      }
      
      if (existingCard) {
        // Обновляем существующую запись, сохраняя предыдущие режимы
        modeData.correct_count = existingCard.correct_count + (isCorrect ? 1 : 0)
        modeData.wrong_count = existingCard.wrong_count + (isCorrect ? 0 : 1)
        
        // Сохраняем все предыдущие режимы и добавляем текущий (только если правильный ответ)
        if (existingCard.mode_view) modeData.mode_view = true
        if (existingCard.mode_with_photo) modeData.mode_with_photo = true
        if (existingCard.mode_without_photo) modeData.mode_without_photo = true
        if (existingCard.mode_russian) modeData.mode_russian = true
        if (existingCard.mode_constructor) modeData.mode_constructor = true
        
        // Добавляем текущий режим только если ответ правильный
        if (isCorrect) {
          if (studyMode === 'view') modeData.mode_view = true
          if (studyMode === 'withPhoto') modeData.mode_with_photo = true
          if (studyMode === 'withoutPhoto') modeData.mode_without_photo = true
          if (studyMode === 'russian') modeData.mode_russian = true
          if (studyMode === 'constructor') modeData.mode_constructor = true
        }
        
        console.log('Updating card with data:', modeData)
        const result = await api.put(`/user-cards/${existingCard.id}`, modeData)
        console.log('Update result:', result.data)
      } else {
        // Создаем новую запись
        const createData: any = {
          user_id: userId,
          card_id: cardId,
          status: isFullyLearned ? 'learned' : (isCorrect ? 'learning' : 'new'),
          correct_count: isCorrect ? 1 : 0,
          wrong_count: isCorrect ? 0 : 1,
          mode_view: false,
          mode_with_photo: false,
          mode_without_photo: false,
          mode_russian: false,
          mode_constructor: false
        }
        
        // Устанавливаем текущий режим только если ответ правильный
        if (isCorrect) {
          if (studyMode === 'view') createData.mode_view = true
          if (studyMode === 'withPhoto') createData.mode_with_photo = true
          if (studyMode === 'withoutPhoto') createData.mode_without_photo = true
          if (studyMode === 'russian') createData.mode_russian = true
          if (studyMode === 'constructor') createData.mode_constructor = true
        }
        
        console.log('Creating card with data:', createData)
        const result = await api.post('/user-cards', createData)
        console.log('Create result:', result.data)
      }
    } catch (error) {
      console.error('Error saving progress to backend:', error)
    }
  }

  // Добавляем слово в личный словарь
  const addToPersonalVocabulary = async (card: Card) => {
    try {
      // Проверяем, не добавлено ли уже это слово
      const alreadyAdded = localStorage.getItem(`vocab_added_${card.id}`)
      if (alreadyAdded) {
        console.log(`Слово "${card.word}" уже добавлено в словарь`)
        return
      }

      console.log(`🎉 Слово "${card.word}" полностью изучено! Добавляем в словарь...`)

      // Получаем user_id из auth store
      const authStorage = localStorage.getItem('auth-storage')
      if (!authStorage) {
        console.error('User not authenticated')
        return
      }
      
      const parsed = JSON.parse(authStorage)
      const userId = parsed?.state?.user?.id
      
      if (!userId) {
        console.error('User ID not found')
        return
      }

      await vocabularyService.addWord({
        user_id: userId,
        word: card.word,
        translation: card.translation,
        phonetic: card.phonetic || '',
        audio_url: card.audio_url || '',
        example: card.example || '',
        notes: `Изучено в курсе: ${deckTitle}`,
        tags: ['изучено', deckTitle.toLowerCase()],
        status: 'learned'
      })

      // Помечаем, что слово добавлено
      localStorage.setItem(`vocab_added_${card.id}`, 'true')
      console.log(`✅ Слово "${card.word}" успешно добавлено в личный словарь`)
    } catch (error) {
      console.error('Error adding to vocabulary:', error)
    }
  }

  useEffect(() => {
    setShowAnswer(false)
    setUserAnswer('')
    setIsCorrect(null)
    setConstructorAnswer([])
    
    // Инициализируем буквы для конструктора
    if (studyMode === 'constructor' && currentCard) {
      const letters = currentCard.word.split('')
      // Перемешиваем буквы
      const shuffled = [...letters].sort(() => Math.random() - 0.5)
      setConstructorLetters(shuffled)
    }
  }, [currentIndex, studyMode, currentCard])

  const playAudio = () => {
    if (currentCard?.audio_url) {
      // Если URL относительный, добавляем базовый URL для Dictionary API
      let audioUrl = currentCard.audio_url
      if (!audioUrl.startsWith('http')) {
        audioUrl = `https:${audioUrl}`
      }
      const audio = new Audio(audioUrl)
      audio.play().catch(err => {
        console.error('Error playing audio:', err)
      })
    } else {
      console.log('Аудио для этого слова недоступно')
    }
  }

  const handleNext = async () => {
    // В режиме просмотра автоматически засчитываем карточку как изученную
    if (studyMode === 'view' && currentCard) {
      const newModeProgress = { ...modeProgress }
      newModeProgress[studyMode] = new Set([...newModeProgress[studyMode], currentCard.id])
      setModeProgress(newModeProgress)
      
      // Сохраняем в backend
      try {
        const authStorage = localStorage.getItem('auth-storage')
        if (authStorage) {
          const parsed = JSON.parse(authStorage)
          const userId = parsed?.state?.user?.id
          if (userId) {
            const existingCards = await api.get(`/user-cards/user/${userId}`)
            const existingCard = existingCards.data?.find((uc: any) => uc.card_id === currentCard.id)
            
            console.log('Existing card for view mode:', existingCard)
            
            if (existingCard) {
              console.log('Updating existing card in view mode')
              await api.put(`/user-cards/${existingCard.id}`, {
                status: existingCard.status,
                correct_count: existingCard.correct_count,
                wrong_count: existingCard.wrong_count,
                mode_view: true,
                mode_with_photo: existingCard.mode_with_photo || false,
                mode_without_photo: existingCard.mode_without_photo || false,
                mode_russian: existingCard.mode_russian || false,
                mode_constructor: existingCard.mode_constructor || false
              })
              console.log('View mode updated successfully')
            } else {
              const createData = {
                user_id: userId,
                card_id: currentCard.id,
                status: 'new',
                correct_count: 0,
                wrong_count: 0,
                mode_view: true,
                mode_with_photo: false,
                mode_without_photo: false,
                mode_russian: false,
                mode_constructor: false
              }
              console.log('Creating new card in view mode with data:', createData)
              try {
                const result = await api.post('/user-cards', createData)
                console.log('Create result:', result.data)
              } catch (createError: any) {
                console.error('Error creating card:', createError)
                console.error('Error response:', createError.response?.data)
              }
            }
          }
        }
      } catch (error) {
        console.error('Error saving view progress:', error)
      }
    }
    
    if (currentIndex < cards.length - 1) {
      setCurrentIndex(currentIndex + 1)
    } else {
      // Завершили все карточки текущего режима, переходим к следующему
      const modeSequence: StudyMode[] = ['view', 'withPhoto', 'withoutPhoto', 'russian', 'constructor']
      const currentModeIndex = modeSequence.indexOf(studyMode)
      
      if (currentModeIndex < modeSequence.length - 1) {
        // Переходим к следующему режиму
        const nextMode = modeSequence[currentModeIndex + 1]
        setCurrentIndex(0)
        setStudyMode(nextMode)
      } else {
        // Прошли все режимы - закрываем окно и возвращаемся к курсам
        onClose()
      }
    }
  }

  const handlePrevious = () => {
    if (currentIndex > 0) {
      setCurrentIndex(currentIndex - 1)
    }
  }

  const handleCheckAnswer = () => {
    if (!currentCard) return

    let correct = false
    if (studyMode === 'russian') {
      // Пользователь должен ввести английское слово
      correct = userAnswer.trim().toLowerCase() === currentCard.word.toLowerCase()
    } else if (studyMode === 'translate') {
      // Пользователь должен ввести перевод
      correct = userAnswer.trim().toLowerCase() === currentCard.translation.toLowerCase()
    } else if (studyMode === 'withPhoto') {
      // Пользователь должен ввести английское слово по фото
      correct = userAnswer.trim().toLowerCase() === currentCard.word.toLowerCase()
    } else if (studyMode === 'withoutPhoto') {
      // Пользователь должен ввести перевод на русском
      correct = userAnswer.trim().toLowerCase() === currentCard.translation.toLowerCase()
    } else if (studyMode === 'constructor') {
      // Проверяем собранное слово
      correct = constructorAnswer.join('').toLowerCase() === currentCard.word.toLowerCase()
    }

    setIsCorrect(correct)
    setShowAnswer(true)
    saveProgress(currentCard.id, correct)

    // Убрали автопереход - пользователь сам нажимает "Далее"
  }

  const getImageUrl = (card: Card) => {
      // Если есть сохраненное изображение, используем его
      if (card.image_url && card.image_url.trim() !== '') {
        console.log('Using saved image:', card.image_url)
        // Используем централизованную функцию из config
        return config.getFullUrl(card.image_url)
      }
      // Иначе используем Unsplash API для автогенерации
      const unsplashUrl = `https://source.unsplash.com/400x300/?${encodeURIComponent(card.word)}`
      console.log('Using Unsplash image:', unsplashUrl, 'for word:', card.word)
      return unsplashUrl
    }

  const getFallbackImageUrl = (text: string) => {
    // Локальный fallback (без внешнего интернета), чтобы картинка всегда отображалась
    const safeText = (text || '').slice(0, 40)
    const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="800" height="600">
  <defs>
    <linearGradient id="g" x1="0" y1="0" x2="1" y2="1">
      <stop offset="0%" stop-color="#EAF7FF"/>
      <stop offset="100%" stop-color="#E8E6FF"/>
    </linearGradient>
  </defs>
  <rect width="100%" height="100%" fill="url(#g)"/>
  <rect x="24" y="24" width="752" height="552" rx="24" fill="#FFFFFF" opacity="0.75"/>
  <text x="50%" y="45%" text-anchor="middle" font-family="Arial, sans-serif" font-size="44" fill="#3B3B3B">Нет фото</text>
  <text x="50%" y="55%" text-anchor="middle" font-family="Arial, sans-serif" font-size="34" fill="#6B7280">${safeText}</text>
</svg>`
    return `data:image/svg+xml;charset=utf-8,${encodeURIComponent(svg)}`
  }

  const renderCardContent = () => {
    if (!currentCard) return null

    switch (studyMode) {
      case 'view':
        // Режим просмотра - показываем все
        return (
          <div className="text-center space-y-6">
            <div className="relative">
              <img
                src={getImageUrl(currentCard)}
                alt={currentCard.word}
                className="w-full max-w-sm mx-auto rounded-lg shadow-lg"
                onError={(e) => {
                  ;(e.target as HTMLImageElement).src = getFallbackImageUrl(currentCard.word)
                }}
              />
            </div>
            <div>
              <div className="flex items-center justify-center space-x-3 mb-4">
                <h2 className="text-4xl font-bold text-text-light">{currentCard.word}</h2>
                {currentCard.audio_url && (
                  <button
                    onClick={playAudio}
                    className="text-3xl hover:scale-110 transition-transform"
                    title="Прослушать произношение"
                  >
                    🔊
                  </button>
                )}
              </div>
              {currentCard.phonetic && (
                <p className="text-xl text-gray-400 mb-2">[{currentCard.phonetic}]</p>
              )}
              <p className="text-2xl text-gray-700 font-medium mb-4">{currentCard.translation}</p>
              {currentCard.example && (
                <p className="text-lg text-gray-500 italic">"{currentCard.example}"</p>
              )}
            </div>
          </div>
        )

      case 'withPhoto':
        // Режим с фото - показываем только фото, нужно ввести английское слово
        return (
          <div className="text-center space-y-6">
            <div className="relative">
              <img
                src={getImageUrl(currentCard)}
                alt={currentCard.word}
                className="w-full max-w-sm mx-auto rounded-lg shadow-lg"
                onError={(e) => {
                  console.error('Image load error for:', currentCard.image_url)
                  ;(e.target as HTMLImageElement).src = getFallbackImageUrl(currentCard.word)
                }}
                onLoad={() => {
                  console.log('Image loaded successfully:', getImageUrl(currentCard))
                }}
              />
            </div>
            <div>
              {showAnswer ? (
                <div>
                  <div className="flex items-center justify-center space-x-3 mb-4">
                    <h2 className={`text-4xl font-bold ${isCorrect ? 'text-green-600' : 'text-red-600'}`}>
                      {currentCard.word}
                    </h2>
                    {currentCard.audio_url && (
                      <button
                        onClick={playAudio}
                        className="text-3xl hover:scale-110 transition-transform"
                      >
                        🔊
                      </button>
                    )}
                  </div>
                  {currentCard.phonetic && (
                    <p className="text-xl text-gray-400 mb-2">[{currentCard.phonetic}]</p>
                  )}
                  <p className="text-2xl text-gray-700 font-medium mb-2">{currentCard.translation}</p>
                  {currentCard.example && (
                    <p className="text-lg text-gray-500 italic">"{currentCard.example}"</p>
                  )}
                </div>
              ) : (
                <div>
                  <p className="text-lg text-gray-400 mb-4">Какое слово на английском?</p>
                  <input
                    type="text"
                    value={userAnswer}
                    onChange={(e) => setUserAnswer(e.target.value)}
                    onKeyPress={(e) => e.key === 'Enter' && handleCheckAnswer()}
                    placeholder="Введите слово на английском"
                    className="text-2xl px-4 py-3 border-2 border-gray-300 rounded-lg focus:border-link-light focus:outline-none w-full max-w-md mx-auto"
                    autoFocus
                  />
                </div>
              )}
            </div>
          </div>
        )

      case 'withoutPhoto':
        // Режим без фото - показываем слово, нужно угадать перевод
        return (
          <div className="text-center space-y-6">
            <div>
              <div className="flex items-center justify-center space-x-3 mb-4">
                <h2 className="text-5xl font-bold text-text-light">{currentCard.word}</h2>
                {currentCard.audio_url && (
                  <button
                    onClick={playAudio}
                    className="text-3xl hover:scale-110 transition-transform"
                  >
                    🔊
                  </button>
                )}
              </div>
              {currentCard.phonetic && (
                <p className="text-xl text-gray-400 mb-6">[{currentCard.phonetic}]</p>
              )}
              {showAnswer ? (
                <div>
                  <p className={`text-4xl font-bold mb-4 ${isCorrect ? 'text-green-600' : 'text-red-600'}`}>
                    {currentCard.translation}
                  </p>
                  {currentCard.example && (
                    <p className="text-lg text-gray-500 italic mt-4 border-l-2 border-link-light pl-4">"{currentCard.example}"</p>
                  )}
                </div>
              ) : (
                <div>
                  <input
                    type="text"
                    value={userAnswer}
                    onChange={(e) => setUserAnswer(e.target.value)}
                    onKeyPress={(e) => e.key === 'Enter' && userAnswer.trim() && handleCheckAnswer()}
                    placeholder="Введите перевод на русском"
                    className="text-2xl px-4 py-3 border-2 border-gray-300 rounded-lg focus:border-link-light focus:outline-none w-full max-w-md mx-auto"
                    autoFocus
                  />
                </div>
              )}
            </div>
          </div>
        )

      case 'russian':
        // Режим на русском - показываем перевод, нужно ввести английское слово
        return (
          <div className="text-center space-y-6">
            <div>
              <p className="text-3xl text-gray-700 font-medium mb-6">{currentCard.translation}</p>
              {showAnswer ? (
                <div>
                  <p className={`text-4xl font-bold ${isCorrect ? 'text-green-600' : 'text-red-600'}`}>
                    {currentCard.word}
                  </p>
                  {currentCard.phonetic && (
                    <p className="text-xl text-gray-400 mt-2">[{currentCard.phonetic}]</p>
                  )}
                  {currentCard.example && (
                    <p className="text-lg text-gray-500 italic mt-4">"{currentCard.example}"</p>
                  )}
                </div>
              ) : (
                <div>
                  <input
                    type="text"
                    value={userAnswer}
                    onChange={(e) => setUserAnswer(e.target.value)}
                    onKeyPress={(e) => e.key === 'Enter' && handleCheckAnswer()}
                    placeholder="Введите слово на английском"
                    className="text-2xl px-4 py-3 border-2 border-gray-300 rounded-lg focus:border-link-light focus:outline-none w-full max-w-md mx-auto"
                    autoFocus
                  />
                </div>
              )}
            </div>
          </div>
        )

      case 'translate':
        // Режим перевода - показываем английское слово, нужно ввести перевод
        return (
          <div className="text-center space-y-6">
            <div>
              <div className="flex items-center justify-center space-x-3 mb-4">
                <h2 className="text-5xl font-bold text-text-light">{currentCard.word}</h2>
                {currentCard.audio_url && (
                  <button
                    onClick={playAudio}
                    className="text-3xl hover:scale-110 transition-transform"
                  >
                    🔊
                  </button>
                )}
              </div>
              {currentCard.phonetic && (
                <p className="text-xl text-gray-400 mb-6">[{currentCard.phonetic}]</p>
              )}
              {showAnswer ? (
                <div>
                  <p className={`text-3xl font-medium ${isCorrect ? 'text-green-600' : 'text-red-600'}`}>
                    {currentCard.translation}
                  </p>
                  {currentCard.example && (
                    <p className="text-lg text-gray-500 italic mt-4">"{currentCard.example}"</p>
                  )}
                </div>
              ) : (
                <div>
                  <input
                    type="text"
                    value={userAnswer}
                    onChange={(e) => setUserAnswer(e.target.value)}
                    onKeyPress={(e) => e.key === 'Enter' && handleCheckAnswer()}
                    placeholder="Введите перевод"
                    className="text-2xl px-4 py-3 border-2 border-gray-300 rounded-lg focus:border-link-light focus:outline-none w-full max-w-md mx-auto"
                    autoFocus
                  />
                </div>
              )}
            </div>
          </div>
        )

      case 'constructor':
        // Режим конструктора - собрать слово из букв
        return (
          <div className="text-center space-y-6">
            <div>
              <p className="text-2xl text-gray-700 font-medium mb-6">{currentCard.translation}</p>
              {currentCard.phonetic && (
                <p className="text-lg text-gray-400 mb-4">[{currentCard.phonetic}]</p>
              )}
              
              {showAnswer ? (
                <div>
                  <p className={`text-4xl font-bold mb-4 ${isCorrect ? 'text-green-600' : 'text-red-600'}`}>
                    {currentCard.word}
                  </p>
                  {currentCard.audio_url && (
                    <button
                      onClick={playAudio}
                      className="text-3xl hover:scale-110 transition-transform mb-4"
                    >
                      🔊
                    </button>
                  )}
                  {currentCard.example && (
                    <p className="text-lg text-gray-500 italic mt-4">"{currentCard.example}"</p>
                  )}
                </div>
              ) : (
                <div className="space-y-6">
                  <p className="text-lg text-gray-500">Соберите слово из букв:</p>
                  
                  {/* Область для собранного слова */}
                  <div className="flex justify-center items-center min-h-[80px] bg-gray-100 rounded-lg p-4 flex-wrap gap-2">
                    {constructorAnswer.length === 0 ? (
                      <span className="text-gray-400">Нажмите на буквы ниже</span>
                    ) : (
                      constructorAnswer.map((letter, index) => (
                        <button
                          key={index}
                          onClick={() => {
                            const newAnswer = [...constructorAnswer]
                            newAnswer.splice(index, 1)
                            setConstructorAnswer(newAnswer)
                            setConstructorLetters([...constructorLetters, letter])
                          }}
                          className="text-3xl font-bold bg-white border-2 border-link-light text-link-light px-4 py-2 rounded-lg hover:bg-link-light hover:text-white transition-colors"
                        >
                          {letter}
                        </button>
                      ))
                    )}
                  </div>
                  
                  {/* Доступные буквы */}
                  <div className="flex justify-center items-center flex-wrap gap-2">
                    {constructorLetters.map((letter, index) => (
                      <button
                        key={index}
                        onClick={() => {
                          setConstructorAnswer([...constructorAnswer, letter])
                          const newLetters = [...constructorLetters]
                          newLetters.splice(index, 1)
                          setConstructorLetters(newLetters)
                        }}
                        className="text-3xl font-bold bg-white border-2 border-gray-300 text-gray-700 px-4 py-2 rounded-lg hover:bg-gray-100 transition-colors"
                      >
                        {letter}
                      </button>
                    ))}
                  </div>
                  
                  {/* Кнопка очистить */}
                  {constructorAnswer.length > 0 && (
                    <button
                      onClick={() => {
                        setConstructorLetters([...constructorLetters, ...constructorAnswer])
                        setConstructorAnswer([])
                      }}
                      className="text-sm text-gray-500 hover:text-gray-700 underline"
                    >
                      Очистить
                    </button>
                  )}
                </div>
              )}
            </div>
          </div>
        )

      default:
        return null
    }
  }

  if (cards.length === 0) {
    return (
      <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
        <div className="bg-card-light rounded-lg p-8 max-w-md w-full mx-4">
          <h2 className="text-2xl font-bold text-text-light mb-4">Нет карточек</h2>
          <p className="text-text-light mb-6">В этом уроке пока нет карточек для изучения.</p>
          <button
            onClick={onClose}
            className="w-full bg-link-light hover:bg-link-dark text-white px-4 py-2 rounded-lg transition-colors"
          >
            Закрыть
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-card-light rounded-lg shadow-2xl max-w-4xl w-full max-h-[90vh] overflow-y-auto">
        {/* Заголовок */}
        <div className="sticky top-0 bg-card-light border-b border-gray-200 p-4 flex items-center justify-between z-10">
          <div className="flex-1">
            <h2 className="text-2xl font-bold text-text-light">{deckTitle}</h2>
            <div className="mt-2">
              <div className="flex items-center justify-between mb-1">
                <span className="text-sm text-gray-500">Общий прогресс</span>
                <span className="text-sm font-semibold text-link-light">{Math.round(overallProgress)}%</span>
              </div>
              <div className="w-full bg-gray-200 rounded-full h-2">
                <div
                  className="bg-link-light h-2 rounded-full transition-all"
                  style={{ width: `${overallProgress}%` }}
                />
              </div>
              <p className="text-xs text-gray-400 mt-1">
                Карточка {currentIndex + 1} из {cards.length}
              </p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="text-2xl text-gray-400 hover:text-gray-600 ml-4"
          >
            ×
          </button>
        </div>

        {/* Контент карточки */}
        <div className="p-8 min-h-[400px] flex items-center justify-center">
          {renderCardContent()}
        </div>

        {/* Индикатор текущего режима */}
        <div className="border-t border-gray-200 p-4">
          <div className="flex items-center justify-center space-x-2">
            <span className="text-sm text-gray-500">Режим:</span>
            <span className="text-sm font-semibold text-link-light">
              {studyMode === 'view' && '1/5 - Изучение слов'}
              {studyMode === 'withPhoto' && '2/5 - Угадай по фото'}
              {studyMode === 'withoutPhoto' && '3/5 - Угадай перевод'}
              {studyMode === 'russian' && '4/5 - Переведи на английский'}
              {studyMode === 'constructor' && '5/5 - Собери слово'}
            </span>
          </div>
          <div className="mt-2 w-full bg-gray-200 rounded-full h-1">
            <div
              className="bg-purple-600 h-1 rounded-full transition-all"
              style={{ 
                width: `${
                  studyMode === 'view' ? 20 :
                  studyMode === 'withPhoto' ? 40 :
                  studyMode === 'withoutPhoto' ? 60 :
                  studyMode === 'russian' ? 80 :
                  100
                }%` 
              }}
            />
          </div>
        </div>

        {/* Навигация */}
        <div className="border-t border-gray-200 p-4 flex items-center justify-between">
          <button
            onClick={handlePrevious}
            disabled={currentIndex === 0}
            className="bg-gray-200 hover:bg-gray-300 disabled:opacity-50 disabled:cursor-not-allowed text-gray-700 px-6 py-2 rounded-lg transition-colors"
          >
            ← Назад
          </button>

          {/* Кнопка "Проверить" для режимов с вводом ответа */}
          {studyMode !== 'view' && !showAnswer && (studyMode === 'russian' || studyMode === 'translate' || studyMode === 'withPhoto' || studyMode === 'withoutPhoto') && (
            <button
              onClick={handleCheckAnswer}
              disabled={!userAnswer.trim()}
              className="bg-green-500 hover:bg-green-600 disabled:opacity-50 disabled:cursor-not-allowed text-white px-6 py-2 rounded-lg transition-colors"
            >
              Проверить
            </button>
          )}

          {/* Кнопка "Проверить" для режима конструктора */}
          {studyMode === 'constructor' && !showAnswer && (
            <button
              onClick={handleCheckAnswer}
              disabled={constructorAnswer.length === 0}
              className="bg-green-500 hover:bg-green-600 disabled:opacity-50 disabled:cursor-not-allowed text-white px-6 py-2 rounded-lg transition-colors"
            >
              Проверить
            </button>
          )}

          {/* Кнопка "Далее" после проверки ответа */}
          {studyMode !== 'view' && showAnswer && (
            <button
              onClick={handleNext}
              className="bg-link-light hover:bg-link-dark text-white px-6 py-2 rounded-lg transition-colors"
            >
              Далее →
            </button>
          )}

          {/* Кнопка "Далее" для режима просмотра */}
          {studyMode === 'view' && (
            <button
              onClick={handleNext}
              className="bg-link-light hover:bg-link-dark text-white px-6 py-2 rounded-lg transition-colors"
            >
              {currentIndex === cards.length - 1 ? 'Начать тренировку →' : 'Далее →'}
            </button>
          )}
        </div>
      </div>
    </div>
  )
}
