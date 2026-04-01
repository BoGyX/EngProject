import api from './api'

export interface RegisterData {
  email: string
  password: string
  name?: string
}

export interface LoginData {
  email: string
  password: string
}

export interface MoodleLoginData {
  username: string
  password: string
}

export interface AuthResponse {
  access_token: string
  refresh_token: string
  user: {
    id: string
    email: string
    name?: string
    role: string
  }
}

export const authService = {
  async register(data: RegisterData): Promise<AuthResponse> {
    const response = await api.post<AuthResponse>('/auth/register', data)
    return response.data
  },

  async login(data: LoginData): Promise<AuthResponse> {
    const response = await api.post<AuthResponse>('/auth/login', data)
    return response.data
  },

  async loginMoodle(data: MoodleLoginData): Promise<AuthResponse> {
    const response = await api.post<AuthResponse>('/auth/login/moodle', data)
    return response.data
  },

  async registerMoodle(data: MoodleLoginData): Promise<AuthResponse> {
    const response = await api.post<AuthResponse>('/auth/register/moodle', data)
    return response.data
  },

  async refreshToken(refreshToken: string): Promise<{ access_token: string }> {
    const response = await api.post<{ access_token: string }>('/auth/refresh', {
      refresh_token: refreshToken,
    })
    return response.data
  },

  async logout(refreshToken?: string): Promise<void> {
    if (refreshToken) {
      await api.post('/auth/logout', { refresh_token: refreshToken })
    }
  },

  async getMe(): Promise<any> {
    const response = await api.get('/me')
    return response.data
  },
}
