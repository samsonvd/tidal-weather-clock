import { createFileRoute, Navigate } from '@tanstack/react-router'
import { localDateStr } from '@/lib/format'

export const Route = createFileRoute('/')({
  component: () => <Navigate to="/day/$date" params={{ date: localDateStr() }} />,
})
