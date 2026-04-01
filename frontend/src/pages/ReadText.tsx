import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import api from '../services/api'
import { dictionaryService } from '../services/dictionaryService'

interface ReadingText {
  id: number
  user_id: string
  title: string
  content: string
  created_at: string
  updated_at: string
}

interface WordTranslation {
  word: string
  translation: string
  phonetic?: string
}

export default function ReadText() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [text, setText] = useState<ReadingText | null>(null)
  const [loading, setLoading] = useState(true)
  const [wordTranslation, setWordTranslation] = useState<WordTranslation | null>(null)
  const [loadingTranslation, setLoadingTranslation] = useState(false)

  const getUserId = () => {
    const authStorage = localStorage.getItem('auth-storage')
    if (!authStorage) return null
    const parsed = JSON.parse(authStorage)
    return parsed?.state?.user?.id
  }

  useEffect(() => {
    loadText()
  }, [id])

  const loadText = async () => {
    try {
      setLoading(true)
      const userId = getUserId()
      if (!userId || !id) {
        navigate('/reader')
        return
      }

      const response = await api.get<ReadingText>(`/reading-texts/${id}?user_id=${userId}`)
      setText(response.data)
    } catch (error) {
      console.error('Error loading text:', error)
      alert('Ошибка при загрузке текста')
      navigate('/reader')
    } finally {
      setLoading(false)
    }
  }

  const handleWordClick = async (word: string) => {
    const cleanWord = word.replace(/[.,!?;:"""''()[\]{}]/g, '').toLowerCase()
    if (!cleanWord) return

    try {
      setLoadingTranslation(true)
      const result = await dictionaryService.getWordInfo(cleanWord)
      setWordTranslation({
        word: cleanWord,
        translation: result.translation || 'Перевод не найден',
        phonetic: result.phonetic
      })
    } catch (error) {
      console.error('Error translating word:', error)
      setWordTranslation({
        word: cleanWord,
        translation: 'Ошибка при получении перевода'
      })
    } finally {
      setLoadingTranslation(false)
    }
  }

  const renderInteractiveText = (content: string) => {
    // Разбиваем текст на абзацы
    const paragraphs = content.split(/\n\n+/)
    
    return paragraphs.map((paragraph, pIndex) => {
      const parts = paragraph.split(/(\s+|[.,!?;:"""''()[\]{}—–-])/g)
      
      return (
        <p key={pIndex} className="mb-4">
          {parts.map((part, index) => {
            if (/^\s+$/.test(part) || /^[.,!?;:"""''()[\]{}—–-]$/.test(part)) {
              return <span key={index}>{part}</span>
            }
            
            if (part.trim()) {
              return (
                <span
                  key={index}
                  className="cursor-pointer hover:bg-yellow-200 hover:px-1 rounded transition-colors inline-block"
                  onClick={() => handleWordClick(part)}
                >
                  {part}
                </span>
              )
            }
            
            return null
          })}
        </p>
      )
    })
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center">
        <div className="text-center">
          <div className="text-6xl mb-4">📖</div>
          <p className="text-xl text-gray-700">Загрузка текста...</p>
        </div>
      </div>
    )
  }

  if (!text) {
    return null
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100">
      {/* Шапка */}
      <div className="bg-white shadow-md border-b border-gray-200 sticky top-0 z-10">
        <div className="max-w-5xl mx-auto px-6 py-4 flex items-center justify-between">
          <button
            onClick={() => navigate('/reader')}
            className="flex items-center space-x-2 text-blue-600 hover:text-blue-800 transition-colors"
          >
            <span className="text-2xl">←</span>
            <span className="font-semibold">Назад к списку</span>
          </button>
          <h1 className="text-2xl font-bold text-gray-800">{text.title}</h1>
          <div className="w-32"></div>
        </div>
      </div>

      {/* Основной контент */}
      <div className="max-w-4xl mx-auto px-6 py-8">
        <div className="bg-white rounded-lg shadow-xl p-8 mb-6">
          <div className="mb-6 pb-4 border-b border-gray-200">
            <p className="text-sm text-gray-500">
              💡 Наведите на слово, чтобы выделить его. Нажмите, чтобы увидеть перевод.
            </p>
          </div>
          
          <div className="prose max-w-none">
            <div className="text-gray-800 leading-relaxed text-lg" style={{ lineHeight: '2.5' }}>
              {renderInteractiveText(text.content)}
            </div>
          </div>
        </div>

        {/* Панель перевода */}
        {wordTranslation && (
          <div className="bg-gradient-to-r from-green-50 to-blue-50 rounded-lg shadow-xl p-6 border-2 border-blue-200 sticky bottom-6">
            <div className="flex items-start space-x-4">
              <div className="text-4xl">📚</div>
              <div className="flex-1">
                <div className="flex items-baseline space-x-3 mb-2">
                  <h3 className="text-2xl font-bold text-gray-800">{wordTranslation.word}</h3>
                  {wordTranslation.phonetic && (
                    <span className="text-sm text-gray-500">[{wordTranslation.phonetic}]</span>
                  )}
                </div>
                <p className="text-lg text-gray-700">{wordTranslation.translation}</p>
              </div>
              <button
                onClick={() => setWordTranslation(null)}
                className="text-gray-400 hover:text-gray-600 text-2xl transition-colors"
              >
                ×
              </button>
            </div>
          </div>
        )}

        {/* Индикатор загрузки */}
        {loadingTranslation && (
          <div className="bg-white rounded-lg shadow-xl p-4 text-center sticky bottom-6">
            <p className="text-gray-600">Загрузка перевода...</p>
          </div>
        )}
      </div>
    </div>
  )
}
