const localStorageTokenKey = 'token'

export function getJwtToken () {
  return localStorage.getItem(localStorageTokenKey)
}

export function setJwtToken (jwtToken) {
  return localStorage.setItem(localStorageTokenKey, jwtToken)
}

export function removeJwtToken () {
  return localStorage.removeItem(localStorageTokenKey)
}
