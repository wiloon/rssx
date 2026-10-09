import { AxiosError, AxiosHeaders, InternalAxiosRequestConfig } from 'axios'
import { attachToken, captureRefreshedToken, handleUnauthorized } from '@/api/http'
import { getJwtToken, setJwtToken } from '@/utils/auth'

function requestConfig (): InternalAxiosRequestConfig {
  return { headers: new AxiosHeaders() } as InternalAxiosRequestConfig
}

function errorWithStatus (status: number): AxiosError {
  return new AxiosError('request failed', 'ERR_BAD_RESPONSE', undefined, undefined, {
    status,
    statusText: '',
    headers: {},
    config: requestConfig(),
    data: {}
  })
}

describe('attachToken', () => {
  beforeEach(() => localStorage.clear())

  it('adds the stored token as a bearer Authorization header', () => {
    setJwtToken('abc.def.ghi')
    const config = attachToken(requestConfig())
    expect(config.headers.get('Authorization')).toBe('Bearer abc.def.ghi')
  })

  it('leaves the request untouched when no token is stored', () => {
    const config = attachToken(requestConfig())
    expect(config.headers.has('Authorization')).toBe(false)
  })
})

describe('captureRefreshedToken', () => {
  beforeEach(() => localStorage.clear())

  it('stores a renewed session token from the response header', () => {
    setJwtToken('old.token')
    captureRefreshedToken(new AxiosHeaders({ 'x-new-token': 'new.token' }))
    expect(getJwtToken()).toBe('new.token')
  })

  it('keeps the current token when the response has no renewal', () => {
    setJwtToken('old.token')
    captureRefreshedToken(new AxiosHeaders())
    expect(getJwtToken()).toBe('old.token')
  })
})

describe('handleUnauthorized', () => {
  beforeEach(() => localStorage.clear())

  it('drops the token and calls onUnauthorized on 401', async () => {
    setJwtToken('expired')
    const onUnauthorized = vi.fn()
    const error = errorWithStatus(401)

    await expect(handleUnauthorized(onUnauthorized)(error)).rejects.toBe(error)
    expect(getJwtToken()).toBeNull()
    expect(onUnauthorized).toHaveBeenCalledOnce()
  })

  it('keeps the token and does not redirect on other errors', async () => {
    setJwtToken('still-valid')
    const onUnauthorized = vi.fn()
    const error = errorWithStatus(500)

    await expect(handleUnauthorized(onUnauthorized)(error)).rejects.toBe(error)
    expect(getJwtToken()).toBe('still-valid')
    expect(onUnauthorized).not.toHaveBeenCalled()
  })
})
