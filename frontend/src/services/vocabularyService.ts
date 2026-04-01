import api from './api'

export interface VocabularyWord {
  id: number
  word: string
  translation: string
  phonetic?: string
  audio_url?: string
  example?: string
  notes?: string
  tags: string[]
  status: 'new' | 'learning' | 'learned'
  correct_count: number
  wrong_count: number
  created_at: string
  updated_at: string
}

// Расширенный тип: сразу указываем пользователя и статус,
// чтобы можно было автоматически добавлять выученные слова
export interface CreateVocabularyRequest {
  user_id: string
  word: string
  translation: string
  phonetic?: string
  audio_url?: string
  example?: string
  notes?: string
  tags?: string[]
  status?: 'new' | 'learning' | 'learned'
}

export interface UpdateVocabularyRequest {
  translation?: string
  phonetic?: string
  audio_url?: string
  example?: string
  notes?: string
  tags?: string[]
  status?: 'new' | 'learning' | 'learned'
}

export const vocabularyService = {
  async getAll(): Promise<VocabularyWord[]> {
    const response = await api.get<VocabularyWord[]>('/vocabulary')
    return response.data
  },

  async addWord(data: CreateVocabularyRequest): Promise<VocabularyWord> {
    const response = await api.post<VocabularyWord>('/vocabulary', data)
    return response.data
  },

  async updateWord(id: number, data: UpdateVocabularyRequest): Promise<VocabularyWord> {
    const response = await api.put<VocabularyWord>(`/vocabulary/${id}`, data)
    return response.data
  },

  async deleteWord(id: number): Promise<void> {
    await api.delete(`/vocabulary/${id}`)
  },
}
