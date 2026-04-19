import api from './api'

export interface PersonalTranslation {
  id: number
  user_id: string
  word: string
  translation: string
  created_at: string
  updated_at: string
}

export const personalTranslationService = {
  async getAll(word?: string): Promise<PersonalTranslation[]> {
    const query = word ? `?word=${encodeURIComponent(word)}` : ''
    const response = await api.get<PersonalTranslation[]>(`/personal-translations${query}`)
    return response.data || []
  },

  async create(word: string, translation: string): Promise<PersonalTranslation> {
    const response = await api.post<PersonalTranslation>('/personal-translations', {
      word,
      translation,
    })
    return response.data
  },

  async remove(id: number): Promise<void> {
    await api.delete(`/personal-translations/${id}`)
  },
}
