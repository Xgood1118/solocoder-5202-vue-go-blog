import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  withCredentials: true
})

export const getArticles = (params) => api.get('/articles', { params })
export const getArticle = (id) => api.get(`/articles/${id}`)
export const createArticle = (data) => api.post('/articles', data)
export const updateArticle = (id, data) => api.put(`/articles/${id}`, data)
export const deleteArticle = (id) => api.delete(`/articles/${id}`)
export const likeArticle = (id) => api.post(`/articles/${id}/like`)

export const getCategories = () => api.get('/categories')
export const createCategory = (data) => api.post('/categories', data)
export const updateCategory = (id, data) => api.patch(`/categories/${id}`, data)
export const deleteCategory = (id) => api.delete(`/categories/${id}`)

export const createComment = (articleId, data) => api.post(`/articles/${articleId}/comments`, data)
export const likeComment = (id) => api.post(`/comments/${id}/like`)
export const reportComment = (id, reason) => api.post(`/comments/${id}/report`, { reason })

export const getTagCloud = () => api.get('/tags')
export const getActiveAuthors = () => api.get('/authors/active')
export const getArchive = () => api.get('/archive')
export const getReports = () => api.get('/reports')

export default api
