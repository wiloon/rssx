/**
 * The shared axios instance and the HttpClient adapter the reader API client
 * expects. The backend has no /api prefix; the dev server (and nginx in prod)
 * strips it, so requests are made against /api here.
 */

import axios from 'axios'
import type { HttpClient } from '@/api/reader'

export const axiosInstance = axios.create({ baseURL: '/api' })

export const http: HttpClient = {
  get: (url, config) => axiosInstance.get(url, config as object),
  post: (url, data, config) => axiosInstance.post(url, data, config as object),
  delete: (url, config) => axiosInstance.delete(url, config as object)
}
