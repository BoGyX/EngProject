// Централизованная конфигурация приложения
// Все значения берутся из переменных окружения

// Для production (Docker) используем runtime config из window.ENV
// Для development используем import.meta.env
declare global {
  interface Window {
    ENV?: {
      VITE_API_URL?: string
    }
  }
}


// Автоматически исправляем протокол: если страница открыта по HTTPS,
// API тоже должен быть HTTPS (защита от Mixed Content)
function normalizeApiUrl(url: string): string {
  if (typeof window === 'undefined') return url
  // Если страница открыта по HTTPS, а URL указан как HTTP - исправляем
  if (window.location.protocol === 'https:' && url.startsWith('http://')) {
    return url.replace('http://', 'https://')
  }
  return url
}

function getBaseUrl(apiUrl: string): string {
  if (apiUrl === '/api') {
    return ''
  }

  return apiUrl.replace(/\/api\/?$/, '')
}

const RUNTIME_API_URL = typeof window !== 'undefined' ? window.ENV?.VITE_API_URL : undefined
const RAW_API_URL = RUNTIME_API_URL || import.meta.env.VITE_API_URL || '/api'
const API_URL = normalizeApiUrl(RAW_API_URL)
const BASE_URL = getBaseUrl(API_URL)

export const config = {
  // API URLs
  apiUrl: API_URL,
  baseUrl: API_URL.replace('/api', ''), // Базовый URL без /api
  
  // Вспомогательные функции
  getFullUrl: (path: string) => {
    if (!path) return ''
    // Если URL уже полный (начинается с http), возвращаем как есть + нормализуем протокол
    if (path.startsWith('http://') || path.startsWith('https://')) {
      return normalizeApiUrl(path)
    }
    if (path.startsWith('//')) {
      return normalizeApiUrl(`https:${path}`)
    }
    // Если относительный путь, добавляем базовый URL
    if (path.startsWith('/')) {
      return `${BASE_URL}${path}`
    }
    return path
  }
}

export default config
