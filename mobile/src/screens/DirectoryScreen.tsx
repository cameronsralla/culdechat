import { useCallback, useState } from 'react';
import { View } from 'react-native';
import { useFocusEffect } from 'expo-router';
import { listDirectory } from '../api/community';
import type { DirectoryUser } from '../api/types';
import { AppText, Avatar, Card, EmptyState, ErrorBanner, Stack } from '../components/ui';
import { Screen } from '../components/layout/Screen';
import { useStyles, type Theme } from '../theme';

const stylesFor = (t: Theme) => ({
  row: { flexDirection: 'row' as const, alignItems: 'center' as const, gap: t.space.md },
});

export function DirectoryScreen() {
  const styles = useStyles(stylesFor);
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
        <AppText variant="title">Neighbors</AppText>
        <AppText tone="muted">Residents who opted into the directory.</AppText>
        <ErrorBanner message={error} />
        {loaded && people.length === 0 ? (
          <EmptyState
            title={error ? 'Could not load the directory.' : 'Nobody has opted in yet.'}
            actionLabel={error ? 'Try again' : undefined}
            onAction={error ? () => void load() : undefined}
          />
        ) : null}
        {people.map((person) => (
          <Card key={person.id}>
            <View style={styles.row}>
              <Avatar path={person.profile_picture_url} label={person.name} />
              <Stack gap="xs">
                <AppText variant="subtitle">{person.name}</AppText>
                <AppText variant="caption" tone="muted">
                  Unit {person.unit_number}
                </AppText>
              </Stack>
            </View>
          </Card>
        ))}
      </Stack>
    </Screen>
  );
}
