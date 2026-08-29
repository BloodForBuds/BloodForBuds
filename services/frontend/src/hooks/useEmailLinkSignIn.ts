import {
  isSignInWithEmailLink,
  sendSignInLinkToEmail,
  signInWithEmailLink,
  signOut,
} from 'firebase/auth'
import { useCallback, useEffect, useRef, useState } from 'react'
import type { FormEvent } from 'react'

import { createSession } from '../lib/api'
import { firebaseClient } from '../lib/firebase'

const emailStorageKey = 'bloodforbuds.emailForSignIn'

export function useEmailLinkSignIn() {
  const [email, setEmail] = useState('')
  const [needsEmailConfirmation, setNeedsEmailConfirmation] = useState(false)
  const [pending, setPending] = useState(false)
  const [message, setMessage] = useState<string>()
  const [error, setError] = useState<string>()
  const [developmentLink, setDevelopmentLink] = useState<string>()
  const completing = useRef(false)

  const finishSignIn = useCallback(async (signInEmail: string) => {
    if (completing.current) {
      return
    }
    completing.current = true
    setPending(true)
    setError(undefined)
    try {
      const { auth } = await firebaseClient()
      const credential = await signInWithEmailLink(auth, signInEmail, window.location.href)
      try {
        await createSession(await credential.user.getIdToken())
      } finally {
        await signOut(auth)
      }
      window.sessionStorage.removeItem(emailStorageKey)
      window.location.replace('/')
    } catch (cause: unknown) {
      completing.current = false
      setError(errorMessage(cause))
      setPending(false)
    }
  }, [])

  useEffect(() => {
    let active = true
    void firebaseClient()
      .then(({ auth }) => {
        if (!active || !isSignInWithEmailLink(auth, window.location.href)) {
          return
        }
        const storedEmail = window.sessionStorage.getItem(emailStorageKey)
        if (storedEmail) {
          setEmail(storedEmail)
          void finishSignIn(storedEmail)
        } else {
          setNeedsEmailConfirmation(true)
        }
      })
      .catch((cause: unknown) => {
        if (active) {
          setError(errorMessage(cause))
        }
      })
    return () => {
      active = false
    }
  }, [finishSignIn])

  const requestLink = useCallback(async () => {
    setPending(true)
    setError(undefined)
    setMessage(undefined)
    setDevelopmentLink(undefined)
    try {
      const { auth, config } = await firebaseClient()
      const normalizedEmail = email.trim().toLowerCase()
      await sendSignInLinkToEmail(auth, normalizedEmail, {
        handleCodeInApp: true,
        url: `${window.location.origin}/login`,
      })
      window.sessionStorage.setItem(emailStorageKey, normalizedEmail)
      setEmail(normalizedEmail)
      setMessage('Check your email for a secure sign-in link.')
      if (config.useEmulator) {
        setDevelopmentLink(await emulatorEmailLink(config.projectId, normalizedEmail))
      }
    } catch (cause: unknown) {
      setError(errorMessage(cause))
    } finally {
      setPending(false)
    }
  }, [email])

  const onSubmit = useCallback(
    async (event: FormEvent<HTMLFormElement>) => {
      event.preventDefault()
      if (needsEmailConfirmation) {
        await finishSignIn(email.trim().toLowerCase())
        return
      }
      await requestLink()
    },
    [email, finishSignIn, needsEmailConfirmation, requestLink],
  )

  return {
    developmentLink,
    email,
    error,
    message,
    needsEmailConfirmation,
    onSubmit,
    pending,
    setEmail,
  }
}

function errorMessage(cause: unknown): string {
  return cause instanceof Error ? cause.message : 'Authentication failed. Please try again.'
}

interface EmulatorOOBCode {
  email: string
  oobLink: string
  requestType: string
}

async function emulatorEmailLink(projectID: string, email: string): Promise<string> {
  const response = await fetch(`/emulator/v1/projects/${encodeURIComponent(projectID)}/oobCodes`)
  if (!response.ok) {
    throw new Error('The Auth Emulator did not return a sign-in link.')
  }
  const body = (await response.json()) as { oobCodes?: EmulatorOOBCode[] }
  const code = [...(body.oobCodes ?? [])]
    .reverse()
    .find((candidate) => candidate.email === email && candidate.requestType === 'EMAIL_SIGNIN')
  if (!code) {
    throw new Error('The Auth Emulator did not return a sign-in link.')
  }

  const link = new URL(code.oobLink)
  link.protocol = window.location.protocol
  link.host = window.location.host
  return link.toString()
}
