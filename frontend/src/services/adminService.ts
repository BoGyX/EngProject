import api from './api'

export interface Course {
  id: number
  slug: string
  title: string
  description?: string
  image_url?: string
  is_published: boolean
  created_at: string
}

export interface Deck {
  id: number
  course_id: number
  slug: string
  title: string
  description?: string
  position: number
}

export interface Card {
  id: number
  deck_id: number
  word: string
  translation: string
  phonetic?: string
  audio_url?: string
  image_url?: string
  example?: string
}

export interface CreateCourseRequest {
  title: string
  slug?: string
  description?: string
  image_url?: string
  is_published?: boolean
}

export interface CreateDeckRequest {
  course_id: number
  title: string
  slug?: string
  description?: string
  position?: number
}

export interface CreateCardRequest {
  deck_id: number
  word: string
  translation: string
  phonetic?: string
  audio_url?: string
  image_url?: string
  example?: string
  is_custom?: boolean
}

export const adminService = {
  // Courses
  async getAllCourses(): Promise<Course[]> {
    const response = await api.get<Course[]>('/courses')
    return response.data
  },

  async getCourse(id: number): Promise<Course> {
    const response = await api.get<Course>(`/courses/${id}`)
    return response.data
  },

  async createCourse(data: CreateCourseRequest): Promise<Course> {
    const response = await api.post<Course>('/admin/courses', data)
    return response.data
  },

  async updateCourse(id: number, data: Partial<CreateCourseRequest>): Promise<Course> {
    const response = await api.put<Course>(`/admin/courses/${id}`, data)
    return response.data
  },

  async deleteCourse(id: number): Promise<void> {
    await api.delete(`/admin/courses/${id}`)
  },

  async publishCourse(id: number): Promise<Course> {
    const response = await api.post<Course>(`/admin/courses/${id}/publish`)
    return response.data
  },

  // Decks
  async getDecksByCourse(courseId: number): Promise<Deck[]> {
    const response = await api.get<Deck[]>(`/decks?course_id=${courseId}`)
    return response.data
  },

  async getDeck(id: number): Promise<Deck> {
    const response = await api.get<Deck>(`/decks/${id}`)
    return response.data
  },

  async createDeck(data: CreateDeckRequest): Promise<Deck> {
    const response = await api.post<Deck>('/admin/decks', data)
    return response.data
  },

  async updateDeck(id: number, data: Partial<CreateDeckRequest>): Promise<Deck> {
    const response = await api.put<Deck>(`/admin/decks/${id}`, data)
    return response.data
  },

  async deleteDeck(id: number): Promise<void> {
    await api.delete(`/admin/decks/${id}`)
  },

  // Cards
  async getCardsByDeck(deckId: number): Promise<Card[]> {
    const response = await api.get<Card[]>(`/cards?deck_id=${deckId}`)
    return response.data
  },

  async getCard(id: number): Promise<Card> {
    const response = await api.get<Card>(`/cards/${id}`)
    return response.data
  },

  async createCard(data: CreateCardRequest): Promise<Card> {
    const response = await api.post<Card>('/admin/cards', data)
    return response.data
  },

  async updateCard(id: number, data: Partial<CreateCardRequest>): Promise<Card> {
    const response = await api.put<Card>(`/admin/cards/${id}`, data)
    return response.data
  },

  async deleteCard(id: number): Promise<void> {
    await api.delete(`/admin/cards/${id}`)
  },
}
