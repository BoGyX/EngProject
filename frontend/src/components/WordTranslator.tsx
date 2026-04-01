import { useState, useEffect, useRef } from 'react'
import { dictionaryService } from '../services/dictionaryService'

interface TooltipPosition {
  x: number
  y: number
}

export default function WordTranslator() {
  const [selectedWord, setSelectedWord] = useState('')
  const [translation, setTranslation] = useState<any>(null)
  const [position, setPosition] = useState<TooltipPosition | null>(null)
  const [loading, setLoading] = useState(false)
  const timeoutRef = useRef<number | null>(null)

  useEffect(() => {
    const handleMouseUp = async (e: MouseEvent) => {
      // Получаем выделенный текст
      const selection = window.getSelection()
      const text = selection?.toString().trim()

      if (text && text.length > 0 && text.length < 50) {
        // Проверяем, что это одно слово (без пробелов)
        const words = text.split(/\s+/)
        if (words.length === 1) {
          const word = words[0].toLowerCase().replace(/[^a-z]/gi, '')
          
          if (word.length >= 2) {
            setSelectedWord(word)
            setPosition({ x: e.clientX, y: e.clientY })
            
            // Загружаем перевод
            setLoading(true)
            try {
              const result = await dictionaryService.getWordInfo(word)
              setTranslation(result)
            } catch (error) {
              console.error('Translation error:', error)
              setTranslation(null)
            } finally {
              setLoading(false)
            }
          }
        }
      } else {
        // Скрываем tooltip если ничего не выделено
        setPosition(null)
        setTranslation(null)
      }
    }


    const handleClick = (e: MouseEvent) => {
      // Закрываем tooltip при клике вне его
      const target = e.target as HTMLElement
      if (!target.closest('.word-translator-tooltip')) {
        setPosition(null)
        setTranslation(null)
      }
    }

    document.addEventListener('mouseup', handleMouseUp)
    document.addEventListener('click', handleClick)

    return () => {
      document.removeEventListener('mouseup', handleMouseUp)
      document.removeEventListener('click', handleClick)
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current)
      }
    }
  }, [])

  if (!position || !selectedWord) return null

  return (
    <div
      className="word-translator-tooltip fixed z-50 bg-white rounded-lg shadow-2xl border-2 border-blue-500 p-4 max-w-sm"
      style={{
        left: `${position.x + 10}px`,
        top: `${position.y + 10}px`,
      }}
    >
      {loading ? (
        <div className="flex items-center space-x-2">
          <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-500"></div>
          <span className="text-sm text-gray-600">Перевод...</span>
        </div>
      ) : translation ? (
        <div className="space-y-2">
          <div className="flex items-center justify-between border-b pb-2">
            <h3 className="text-lg font-bold text-gray-800">{translation.word}</h3>
            {translation.audio_url && (
              <button
                onClick={() => {
                  const audio = new Audio(translation.audio_url.startsWith('//') ? `https:${translation.audio_url}` : translation.audio_url)
                  audio.play()
                }}
                className="text-blue-500 hover:text-blue-700 text-xl"
                title="Прослушать"
              >
                🔊
              </button>
            )}
          </div>
          
          {translation.phonetic && (
            <p className="text-sm text-gray-500">[{translation.phonetic}]</p>
          )}
          
          {translation.definition && (
            <p className="text-base text-gray-700">
              <span className="font-semibold">Перевод:</span> {translation.definition}
            </p>
          )}
          
          {translation.example && (
            <p className="text-sm text-gray-600 italic border-l-2 border-blue-300 pl-2">
              "{translation.example}"
            </p>
          )}
          
          <button
            onClick={() => {
              setPosition(null)
              setTranslation(null)
            }}
            className="text-xs text-gray-400 hover:text-gray-600 mt-2"
          >
            Закрыть
          </button>
        </div>
      ) : (
        <div className="text-sm text-gray-600">
          <p className="font-semibold mb-1">{selectedWord}</p>
          <p className="text-xs text-red-500">Перевод не найден</p>
        </div>
      )}
    </div>
  )
}
