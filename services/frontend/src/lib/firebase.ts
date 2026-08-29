import { getApp, getApps, initializeApp, type FirebaseOptions } from 'firebase/app'
import {
  connectAuthEmulator,
  getAuth,
  initializeAuth,
  inMemoryPersistence,
  setPersistence,
  type Auth,
} from 'firebase/auth'

export interface FirebaseWebConfig extends FirebaseOptions {
  apiKey: string
  appId: string
  authDomain: string
  projectId: string
  useEmulator: boolean
}

interface FirebaseClient {
  auth: Auth
  config: FirebaseWebConfig
}

let clientPromise: Promise<FirebaseClient> | undefined

export function firebaseClient(): Promise<FirebaseClient> {
  if (typeof window === 'undefined') {
    return Promise.reject(new Error('Firebase Auth is only available in the browser.'))
  }
  clientPromise ??= createFirebaseClient()
  return clientPromise
}

async function createFirebaseClient(): Promise<FirebaseClient> {
  const response = await fetch('/api/auth/config', {
    credentials: 'same-origin',
    headers: { Accept: 'application/json' },
  })
  if (!response.ok) {
    throw new Error('Firebase Auth is not configured.')
  }
  const config = (await response.json()) as FirebaseWebConfig
  if (!config.apiKey || !config.appId || !config.authDomain || !config.projectId) {
    throw new Error('Firebase Auth is not configured.')
  }

  const appName = 'bloodforbuds-web'
  const app = getApps().some((candidate) => candidate.name === appName)
    ? getApp(appName)
    : initializeApp(config, appName)

  let auth: Auth
  try {
    auth = initializeAuth(app, { persistence: inMemoryPersistence })
  } catch {
    auth = getAuth(app)
    await setPersistence(auth, inMemoryPersistence)
  }

  if (config.useEmulator && auth.emulatorConfig === null) {
    connectAuthEmulator(auth, window.location.origin, { disableWarnings: true })
  }

  return { auth, config }
}
