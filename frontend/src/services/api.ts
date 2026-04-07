import axios from 'axios'
import { config } from '../config'

const OPTIONAL_AUTH_ENDPOINTS = [
  /\/user-courses\/active$/,
  /\/user-decks\/active$/,
  /\/user-courses\/\d+\/activate$/,
  /\/user-decks\/\d+\/activate$/,
]

function shouldRedirectOnUnauthorized(error: any): boolean {
  const requestUrl = String(error?.config?.url || '')
  const authHeader = error?.config?.headers?.Authorization || error?.config?.headers?.authorization
  const hadAuthHeader = Boolean(authHeader)

  if (!hadAuthHeader) {
    return false
  }

  return !OPTIONAL_AUTH_ENDPOINTS.some((pattern) => pattern.test(requestUrl))
}

console.log('🔧 API_URL configured:', config.apiUrl)

export const api = axios.create({
  baseURL: config.apiUrl,
  headers: {
    'Content-Type': 'application/json',
  },
  transformRequest: [
    (data, headers) => {
      // Если это FormData, не трогаем его и удаляем Content-Type (axios сам установит правильный)
      if (data instanceof FormData) {
        delete headers['Content-Type']
        return data
      }
      
      // Удаляем undefined значения перед отправкой
      if (data && typeof data === 'object') {
        const cleaned = JSON.parse(JSON.stringify(data, (_key, value) => {
          return value === undefined ? null : value
        }))
        console.log('📡 Axios sending JSON:', JSON.stringify(cleaned, null, 2))
        return JSON.stringify(cleaned)
      }
      return data
    },
  ],
})

// Interceptor для добавления токена
api.interceptors.request.use(
  (config) => {
    // Пытаемся получить токен из localStorage (для совместимости)
    let token = localStorage.getItem('access_token')
    if (!token) {
      // Пытаемся получить из zustand store
      try {
        const authStorage = localStorage.getItem('auth-storage')
        if (authStorage) {
          const parsed = JSON.parse(authStorage)
          token = parsed?.state?.accessToken
        }
      } catch (e) {
        // Игнорируем ошибки
      }
    }
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Interceptor для обработки ошибок
api.interceptors.response.use(
  (response) => response,
  async (error) => {
    if (error.response?.status === 401 && shouldRedirectOnUnauthorized(error)) {
      localStorage.removeItem('access_token')
      localStorage.removeItem('refresh_token')
      localStorage.removeItem('auth-storage')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export default api
