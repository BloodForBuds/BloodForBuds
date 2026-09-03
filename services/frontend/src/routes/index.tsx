import { Alert, Badge, Button, Container, Group, Paper, Stack, Text, Title } from '@mantine/core'
import { createFileRoute } from '@tanstack/react-router'
import { useEffect, useState } from 'react'

import { currentUser, deleteSession, type SessionUser } from '../lib/api'

const Home = () => {
  const [count, setCount] = useState(0)
  const [user, setUser] = useState<SessionUser | null>()
  const [error, setError] = useState<string>()

  useEffect(() => {
    void currentUser()
      .then(setUser)
      .catch((cause: unknown) => {
        setError(cause instanceof Error ? cause.message : 'Could not load the current session.')
        setUser(null)
      })
  }, [])

  async function signOut() {
    setError(undefined)
    try {
      await deleteSession()
      setUser(null)
    } catch (cause: unknown) {
      setError(cause instanceof Error ? cause.message : 'Could not sign out.')
    }
  }

  return (
    <main className="min-h-screen bg-slate-50 px-6 py-20">
      <Container size="sm">
        <Paper radius="lg" p="xl" shadow="sm" withBorder>
          <Stack gap="lg">
            <Badge variant="light" w="fit-content">
              TanStack Start
            </Badge>
            <div>
              <Title order={1}>BloodForBuds</Title>
              <Text c="dimmed" mt="xs">
                React, TypeScript, Tailwind CSS, and Mantine UI are ready.
              </Text>
            </div>
            <Group>
              <Button onClick={() => setCount((value) => value + 1)}>
                Count: {count}
              </Button>
              <Text size="sm" c="dimmed">
                Edit src/routes/index.tsx to get started.
              </Text>
            </Group>
            {error && <Alert color="red">{error}</Alert>}
            {user === undefined ? (
              <Text c="dimmed" size="sm">Checking your session…</Text>
            ) : user ? (
              <Group justify="space-between">
                <Text size="sm">Signed in as {user.email ?? user.uid}</Text>
                <Button onClick={() => void signOut()} variant="light">Sign out</Button>
              </Group>
            ) : (
              <Button component="a" href="/login" variant="light">Sign in by email</Button>
            )}
          </Stack>
        </Paper>
      </Container>
    </main>
  )
}

export const Route = createFileRoute('/')({ component: Home })
