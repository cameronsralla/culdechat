import { useCallback, useState } from 'react';
import { useFocusEffect, useRouter } from 'expo-router';
import { listDirectory } from '../api/community';
import type { DirectoryUser } from '../api/types';
import {
  AppText,
  Avatar,
  Button,
  EmptyState,
  ErrorBanner,
  ListGroup,
  ListRow,
  PageHeader,
  Stack,
} from '../components/ui';
import { Screen } from '../components/layout/Screen';
import { useCompactLayout } from '../components/layout/useCompactLayout';

export function DirectoryScreen() {
  const router = useRouter();
  const compact = useCompactLayout();
  const [people, setPeople] = useState<DirectoryUser[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [refreshing, setRefreshing] = useState(false);
  const [loaded, setLoaded] = useState(false);

  const load = useCallback(async () => {
    try {
      setPeople(await listDirectory());
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not load directory');
    } finally {
      setLoaded(true);
    }
  }, []);

  useFocusEffect(
    useCallback(() => {
      void load();
    }, [load]),
  );

  return (
    <Screen
      scroll
      inShell
      refreshing={refreshing}
      onRefresh={() => {
        setRefreshing(true);
        void load().finally(() => setRefreshing(false));
      }}
    >
      <Stack gap="lg">
        <PageHeader
          title="People"
          subtitle="Neighbors who chose to be listed. Say hello — or message by unit if someone stays private."
          eyebrow="Your community"
          hideTitleOnCompact
          compact={compact}
        />
        <ErrorBanner message={error} />
        {loaded && people.length === 0 ? (
          <EmptyState
            title={error ? 'Could not load the directory.' : 'Directory is empty for now.'}
            subtitle={
              error ? undefined : 'When neighbors opt in, you’ll find them here. Quiet is fine until then.'
            }
            actionLabel={error ? 'Try again' : undefined}
            onAction={error ? () => void load() : undefined}
            icon="people-outline"
          />
        ) : null}
        {people.length > 0 ? (
          <ListGroup>
            {people.map((person, index) => (
              <ListRow
                key={person.id}
                last={index === people.length - 1}
                leading={<Avatar path={person.profile_picture_url} label={person.name} />}
                trailing={
                  <Button
                    label="Message"
                    variant="soft"
                    size="sm"
                    onPress={() =>
                      router.push({
                        pathname: '/messages/new',
                        params: { userId: person.id, unit: person.unit_number, name: person.name },
                      })
                    }
                  />
                }
              >
                <AppText variant="subtitle">{person.name}</AppText>
                <AppText variant="caption" tone="muted">
                  Unit {person.unit_number}
                </AppText>
              </ListRow>
            ))}
          </ListGroup>
        ) : null}
      </Stack>
    </Screen>
  );
}
