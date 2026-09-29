/**
 * The shared axios instance and the HttpClient adapter the reader API client
 * expects. The backend has no /api prefix; the dev server (and nginx in prod)
 * strips it, so requests are made against /api here.
 */

import axios, { AxiosError, InternalAxiosRequestConfig } from 'axios'
import type { HttpClient } from '@/api/reader'
import { getJwtToken, removeJwtToken } from '@/utils/auth'

export const axiosInstance = axios.create({ baseURL: '/api' })

/** Adds the stored JWT as a bearer token to every outgoing request. */
export function attachToken (config: InternalAxiosRequestConfig): InternalAxiosRequestConfig {
  const token = getJwtToken()
  if (token) {
    config.headers.set('Authorization', `Bearer ${token}`)
  }
  return config
}

/**
 * Builds the response error handler: on 401 the stored token is dropped and
 * `onUnauthorized` runs (redirect to login); the error is always re-thrown.
 */
export function handleUnauthorized (onUnauthorized: () => void) {
  return (error: AxiosError): Promise<never> => {
    if (error.response?.status === 401) {
      removeJwtToken()
      onUnauthorized()
    }
    return Promise.reject(error)
  }
}

axiosInstance.interceptors.request.use(attachToken)
axiosInstance.interceptors.response.use(
  (response) => response,
  // Imported lazily: the router pulls in views that import this module.
  handleUnauthorized(() => {
    void import('@/router').then(({ default: router }) => {
      if (router.currentRoute.value.name !== 'Login') {
        void router.push({ name: 'Login' })
      }
    })
  })
)

export const http: HttpClient = {
  get: (url, config) => axiosInstance.get(url, config as object),
  post: (url, data, config) => axiosInstance.post(url, data, config as object),
  put: (url, data, config) => axiosInstance.put(url, data, config as object),
  delete: (url, config) => axiosInstance.delete(url, config as object)
}
