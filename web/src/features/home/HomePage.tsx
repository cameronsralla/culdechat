import { useAuth } from '@/auth/AuthContext';
import { Card, EmptyState, PageHeader, Text } from '@/components/ui';
import { Screen } from '@/layout/Screen';

export function HomePage() {
  const { user } = useAuth();
  return (
    <Screen title="The square">
      <PageHeader title={`Hi, ${user?.display_name?.split(' ')[0] || 'neighbor'}`} description="What's happening around the neighborhood." />
      <Card tone="wash">
        <Text variant="subtitle">Framework check</Text>
        <Text variant="caption" tone="muted">
          Auth, layout, and the component kit are wired. Boards, calendar, and messages land here next.
        </Text>
      </Card>
      <EmptyState icon="home" title="Nothing posted yet" body="Seeded boards and the first community posts will show up here." />
    </Screen>
  );
}
