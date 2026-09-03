export interface SessionUser {
  uid: string
  email?: string
  emailVerified: boolean
}

interface CSRFResponse {
  csrfToken: string
}

interface SessionResponse {
  user: SessionUser
}

async function csrfToken(): Promise<string> {
  const response = await fetch('/api/auth/csrf', {
    credentials: 'same-origin',
    headers: { Accept: 'application/json' },
  })
  if (!response.ok) {
    throw new Error('Could not start a secure authentication request.')
  }
  const body = (await response.json()) as CSRFResponse
  return body.csrfToken
}

export async function createSession(idToken: string): Promise<SessionUser> {
  const csrf = await csrfToken()
  const response = await fetch('/api/auth/session', {
    method: 'POST',
    credentials: 'same-origin',
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ idToken, csrfToken: csrf }),
  })
  if (!response.ok) {
    throw new Error('The sign-in link is invalid or has expired.')
  }
  const body = (await response.json()) as SessionResponse
  return body.user
}

export async function currentUser(): Promise<SessionUser | null> {
  const response = await fetch('/api/auth/me', {
    credentials: 'same-origin',
    headers: { Accept: 'application/json' },
  })
  if (response.status === 401) {
    return null
  }
  if (!response.ok) {
    throw new Error('Could not load the current session.')
  }
  const body = (await response.json()) as SessionResponse
  return body.user
}

export async function deleteSession(): Promise<void> {
  const csrf = await csrfToken()
  const response = await fetch('/api/auth/session', {
    method: 'DELETE',
    credentials: 'same-origin',
    headers: {
      Accept: 'application/json',
      'X-CSRF-Token': csrf,
    },
  })
  if (!response.ok) {
    throw new Error('Could not sign out.')
  }
}
