const localStorageTokenKey = 'token'

export function getJwtToken (): string | null {
  return localStorage.getItem(localStorageTokenKey)
}

export function setJwtToken (jwtToken: string): void {
  localStorage.setItem(localStorageTokenKey, jwtToken)
}

export function removeJwtToken (): void {
  localStorage.removeItem(localStorageTokenKey)
}
