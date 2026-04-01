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
  created_by?: string
  is_custom: boolean
}

export interface UserCourse {
  id: number
  user_id: string
  course_id: number
  progress_percentage: number
  completed_decks_count: number
  total_decks_count: number
  is_active: boolean
  attempt_number?: number
  last_opened_at?: string
}

export interface UserDeck {
  id: number
  user_id: string
  deck_id: number
  user_course_id?: number
  status: string
  learned_cards_count: number
  total_cards_count: number
  progress_percentage: number
  is_active: boolean
  completed_at?: string
  created_at?: string
  updated_at?: string
  last_opened_at?: string
}

export interface UserCard {
  id: number
  user_id: string
  card_id: number
  user_deck_id?: number
  status: string
  correct_count: number
  wrong_count: number
  mode_view: boolean
  mode_choice: boolean
  mode_with_photo: boolean
  mode_without_photo: boolean
  mode_russian: boolean
  mode_constructor: boolean
  current_mode: string
  progress_percentage: number
}

export interface TrainingCardState {
  session_card_id: number
  card_id: number
  deck_id: number
  word: string
  translation: string
  phonetic?: string
  audio_url?: string
  image_url?: string
  example?: string
  is_custom: boolean
  sequence_number: number
  current_mode: string
  progress_percentage: number
  is_completed: boolean
  options?: string[]
}

export interface TrainingSessionPayload {
  id: number
  user_id?: string
  course_id?: number
  deck_id?: number
  user_deck_id?: number
  started_at: string
  finished_at?: string
}

export interface TrainingSessionState {
  session: TrainingSessionPayload
  cards: TrainingCardState[]
  current_card?: TrainingCardState | null
  remaining_cards: number
}

export interface TrainingAnswerResponse {
  is_correct: boolean
  session: TrainingSessionState
}

export interface CreateCustomCardResult {
  card: Card
  user_deck: UserDeck
}

export const studyService = {
  async getCourses(): Promise<Course[]> {
    const response = await api.get<Course[]>('/courses')
    return response.data || []
  },

  async getCourse(id: number): Promise<Course> {
    const response = await api.get<Course>(`/courses/${id}`)
    return response.data
  },

  async getCourseBySlug(slug: string): Promise<Course> {
    const response = await api.get<Course>(`/courses/by-slug/${slug}`)
    return response.data
  },

  async getDecksByCourse(courseId: number): Promise<Deck[]> {
    const response = await api.get<Deck[]>(`/decks?course_id=${courseId}`)
    return response.data || []
  },

  async getDeck(deckId: number): Promise<Deck> {
    const response = await api.get<Deck>(`/decks/${deckId}`)
    return response.data
  },

  async getDeckBySlug(courseSlug: string, deckSlug: string): Promise<Deck> {
    const response = await api.get<Deck>(`/decks/by-slug/${courseSlug}/${deckSlug}`)
    return response.data
  },

  async getCardsByDeck(deckId: number): Promise<Card[]> {
    const response = await api.get<Card[]>(`/cards?deck_id=${deckId}`)
    return response.data || []
  },

  async getUserCards(userId: string): Promise<UserCard[]> {
    const response = await api.get<UserCard[]>(`/user-cards/user/${userId}`)
    return response.data || []
  },

  async getUserCourses(userId: string): Promise<UserCourse[]> {
    const response = await api.get<UserCourse[]>(`/user-courses/user/${userId}`)
    return response.data || []
  },

  async getUserDecks(userId: string): Promise<UserDeck[]> {
    const response = await api.get<UserDeck[]>(`/user-decks/user/${userId}`)
    return response.data || []
  },

  async getLegacyVocabulary(userId: string) {
    const response = await api.get(`/vocabulary?user_id=${userId}`)
    return response.data || []
  },

  async getActiveCourse(): Promise<UserCourse> {
    const response = await api.get<UserCourse>('/user-courses/active')
    return response.data
  },

  async getActiveDeck(): Promise<UserDeck> {
    const response = await api.get<UserDeck>('/user-decks/active')
    return response.data
  },

  async activateCourse(courseId: number): Promise<UserCourse> {
    const response = await api.post<UserCourse>(`/user-courses/${courseId}/activate`)
    return response.data
  },

  async activateDeck(deckId: number): Promise<UserDeck> {
    const response = await api.post<UserDeck>(`/user-decks/${deckId}/activate`)
    return response.data
  },

  async startTraining(deckId?: number, courseId?: number): Promise<TrainingSessionState> {
    const payload: { deck_id?: number; course_id?: number } = {}
    if (deckId) payload.deck_id = deckId
    if (courseId) payload.course_id = courseId

    const response = await api.post<TrainingSessionState>('/training-sessions/start', payload)
    return response.data
  },

  async getTrainingSession(sessionId: number): Promise<TrainingSessionState> {
    const response = await api.get<TrainingSessionState>(`/training-sessions/${sessionId}`)
    return response.data
  },

  async answerTraining(sessionId: number, cardId: number, answer = ''): Promise<TrainingAnswerResponse> {
    const response = await api.post<TrainingAnswerResponse>(`/training-sessions/${sessionId}/answer`, {
      card_id: cardId,
      answer,
    })
    return response.data
  },

  async createCustomCard(payload: {
    deck_id?: number
    word: string
    translation: string
    phonetic?: string
    audio_url?: string
    image_url?: string
    example?: string
  }): Promise<CreateCustomCardResult> {
    const response = await api.post<CreateCustomCardResult>('/cards/custom', payload)
    return response.data
  },
}
