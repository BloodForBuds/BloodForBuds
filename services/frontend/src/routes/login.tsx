import {
  Alert,
  Anchor,
  Button,
  Container,
  Paper,
  Stack,
  Text,
  TextInput,
  Title,
} from '@mantine/core'
import { createFileRoute } from '@tanstack/react-router'

import { useEmailLinkSignIn } from '../hooks/useEmailLinkSignIn'

const Login = () => {
  const {
    developmentLink,
    email,
    error,
    message,
    needsEmailConfirmation,
    onSubmit,
    pending,
    setEmail,
  } = useEmailLinkSignIn()
  const buttonLabel = needsEmailConfirmation ? 'Complete sign in' : 'Email me a sign-in link'

  return (
    <main className="min-h-screen bg-slate-50 px-6 py-20">
      <Container size="xs">
        <Paper radius="lg" p="xl" shadow="sm" withBorder>
          <Stack gap="lg">
            <div>
              <Title order={1}>Sign in</Title>
              <Text c="dimmed" mt="xs">
                {needsEmailConfirmation
                  ? 'Confirm the email address that received this link.'
                  : 'We will send a passwordless sign-in link to your email.'}
              </Text>
            </div>

            {message && <Alert color="green">{message}</Alert>}
            {developmentLink && (
              <Alert color="blue" title="Development Auth Emulator">
                <Anchor href={developmentLink}>Open the local sign-in link</Anchor>
              </Alert>
            )}
            {error && <Alert color="red">{error}</Alert>}

            <form onSubmit={onSubmit}>
              <Stack>
                <TextInput
                  autoComplete="email"
                  label="Email address"
                  name="email"
                  onChange={(event) => setEmail(event.currentTarget.value)}
                  required
                  type="email"
                  value={email}
                />
                <Button loading={pending} type="submit">
                  {buttonLabel}
                </Button>
              </Stack>
            </form>

            <Anchor href="/">Back to BloodForBuds</Anchor>
          </Stack>
        </Paper>
      </Container>
    </main>
  )
}

export const Route = createFileRoute('/login')({ component: Login })
