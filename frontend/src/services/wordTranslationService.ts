import api from './api'

export interface UserWordTranslation {
  id: number
  user_id: string
  word: string
  translation: string
  created_at: string
  updated_at: string
}

export const wordTranslationService = {
  async upsert(word: string, translation: string): Promise<UserWordTranslation> {
    const response = await api.post<UserWordTranslation>('/word-translations', { word, translation })
    return response.data
  },

  async get(word: string): Promise<UserWordTranslation | null> {
    const response = await api.get<{ translation: null } | UserWordTranslation>(`/word-translations?word=${encodeURIComponent(word)}`)
    const data = response.data as any
    if (!data || data.translation === null || data.translation === undefined) return null
    return data as UserWordTranslation
  },

  async delete(word: string): Promise<void> {
    await api.delete(`/word-translations?word=${encodeURIComponent(word)}`)
  },
}
