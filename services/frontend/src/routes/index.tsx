import { Badge, Button, Container, Group, Paper, Stack, Text, Title } from '@mantine/core'
import { createFileRoute } from '@tanstack/react-router'
import { useState } from 'react'

export const Route = createFileRoute('/')({ component: Home })

function Home() {
  const [count, setCount] = useState(0)

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
          </Stack>
        </Paper>
      </Container>
    </main>
  )
}
