import { useRef, useState } from 'react'
import { config } from '../config'
import { uploadService } from '../services/uploadService'

interface FileUploadProps {
  type: 'image' | 'audio'
  currentUrl?: string
  onUrlChange: (url: string) => void
  label: string
  placeholder: string
}

const IMAGE_TYPES = ['image/jpeg', 'image/jpg', 'image/png', 'image/gif', 'image/webp']
const AUDIO_TYPES = ['audio/mpeg', 'audio/mp3', 'audio/wav', 'audio/ogg', 'audio/m4a']

export default function FileUpload({ type, currentUrl, onUrlChange, label, placeholder }: FileUploadProps) {
  const [uploading, setUploading] = useState(false)
  const [uploadMode, setUploadMode] = useState<'url' | 'file'>('url')
  const fileInputRef = useRef<HTMLInputElement>(null)
  const previewUrl = currentUrl ? config.getFullUrl(currentUrl) : ''

  const validateFile = (file: File) => {
    if (type === 'image') {
      if (!IMAGE_TYPES.includes(file.type)) {
        throw new Error('Недопустимый формат изображения. Разрешены JPG, PNG, GIF и WEBP.')
      }
      if (file.size > 5 * 1024 * 1024) {
        throw new Error('Изображение слишком большое. Максимальный размер 5 МБ.')
      }
      return
    }

    const hasAllowedAudioType =
      AUDIO_TYPES.includes(file.type) ||
      file.name.toLowerCase().endsWith('.mp3') ||
      file.name.toLowerCase().endsWith('.m4a')

    if (!hasAllowedAudioType) {
      throw new Error('Недопустимый формат аудио. Разрешены MP3, WAV, OGG и M4A.')
    }
    if (file.size > 10 * 1024 * 1024) {
      throw new Error('Аудиофайл слишком большой. Максимальный размер 10 МБ.')
    }
  }

  const handleFileSelect = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    if (!file) {
      return
    }

    try {
      validateFile(file)
      setUploading(true)

      const result = type === 'image' ? await uploadService.uploadImage(file) : await uploadService.uploadAudio(file)

      // Store the relative uploads path so data survives domain or protocol changes.
      onUrlChange(result.url)
    } catch (error: any) {
      console.error('Error uploading file:', error)
    } finally {
      setUploading(false)
      if (fileInputRef.current) {
        fileInputRef.current.value = ''
      }
    }
  }

  return (
    <div>
      <label className="mb-2 block text-sm font-medium text-text-light">{label}</label>

      <div className="mb-2 flex space-x-2">
        <button
          type="button"
          onClick={() => setUploadMode('url')}
          className={`rounded-lg px-3 py-1 text-sm transition-colors ${
            uploadMode === 'url' ? 'bg-link-light text-white' : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
          }`}
        >
          URL
        </button>
        <button
          type="button"
          onClick={() => setUploadMode('file')}
          className={`rounded-lg px-3 py-1 text-sm transition-colors ${
            uploadMode === 'file' ? 'bg-link-light text-white' : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
          }`}
        >
          Загрузить файл
        </button>
      </div>

      {uploadMode === 'url' ? (
        <input
          type="url"
          placeholder={placeholder}
          value={currentUrl || ''}
          onChange={(event) => onUrlChange(event.target.value)}
          className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm transition focus:border-link-light focus:outline-none focus:ring-2 focus:ring-link-light"
        />
      ) : (
        <div>
          <input
            ref={fileInputRef}
            type="file"
            accept={
              type === 'image'
                ? 'image/jpeg,image/jpg,image/png,image/gif,image/webp'
                : 'audio/mpeg,audio/mp3,audio/wav,audio/ogg,audio/m4a'
            }
            onChange={handleFileSelect}
            disabled={uploading}
            className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm transition file:mr-4 file:rounded-lg file:border-0 file:bg-link-light file:px-4 file:py-2 file:text-sm file:font-semibold file:text-white hover:file:bg-link-dark focus:border-link-light focus:outline-none focus:ring-2 focus:ring-link-light disabled:cursor-not-allowed disabled:opacity-50"
          />
          {uploading && <p className="mt-2 text-sm text-blue-600">Загрузка файла...</p>}
          <p className="mt-1 text-xs text-gray-500">
            {type === 'image'
              ? 'Форматы: JPG, PNG, GIF, WEBP. Максимальный размер: 5 МБ.'
              : 'Форматы: MP3, WAV, OGG, M4A. Максимальный размер: 10 МБ.'}
          </p>
        </div>
      )}

      {type === 'image' && currentUrl && (
        <div className="mt-2">
          <img
            src={previewUrl}
            alt="Preview"
            className="h-32 w-full max-w-xs rounded-lg border border-gray-300 object-cover"
            onError={(event) => {
              ;(event.target as HTMLImageElement).style.display = 'none'
            }}
          />
        </div>
      )}

      {type === 'audio' && currentUrl && (
        <button
          type="button"
          onClick={() => {
            const audio = new Audio(previewUrl)
            audio.play().catch((error) => {
              console.error('Error playing audio:', error)
            })
          }}
          className="mt-2 rounded-lg bg-green-600 px-4 py-2 text-sm text-white transition-colors hover:bg-green-700"
        >
          Прослушать
        </button>
      )}
    </div>
  )
}
