import api from './api'

export const uploadService = {
  async uploadImage(file: File): Promise<{ url: string; filename: string }> {
    const formData = new FormData()
    formData.append('file', file)
    
    // Получаем токен для авторизации
    let token = localStorage.getItem('access_token')
    if (!token) {
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
    
    const response = await api.post<{ url: string; filename: string; message: string }>(
      '/upload/image',
      formData,
      {
        headers: {
          // НЕ устанавливаем Content-Type вручную - axios сам добавит boundary
          ...(token && { Authorization: `Bearer ${token}` }),
        },
      }
    )
    return response.data
  },

  async uploadAudio(file: File): Promise<{ url: string; filename: string }> {
    const formData = new FormData()
    formData.append('file', file)
    
    // Получаем токен для авторизации
    let token = localStorage.getItem('access_token')
    if (!token) {
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
    
    const response = await api.post<{ url: string; filename: string; message: string }>(
      '/upload/audio',
      formData,
      {
        headers: {
          // НЕ устанавливаем Content-Type вручную - axios сам добавит boundary
          ...(token && { Authorization: `Bearer ${token}` }),
        },
      }
    )
    return response.data
  },

  async deleteFile(type: 'image' | 'audio', filename: string): Promise<void> {
    await api.delete(`/upload/delete?type=${type}&filename=${filename}`)
  },
}
