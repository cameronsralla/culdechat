import { useQuery } from '@tanstack/react-query';
import { usersApi } from '@/api/endpoints';
import { Avatar, EmptyState, ErrorBanner, ListGroup, ListRow, PageHeader, errorMessage } from '@/components/ui';
import { Screen } from '@/layout/Screen';

export function DirectoryPage() {
  const q = useQuery({ queryKey: ['directory'], queryFn: usersApi.directory });
  return (
    <Screen title="People">
      <PageHeader title="People" description="Residents who've chosen to be listed." />
      {q.isError && <ErrorBanner message={errorMessage(q.error)} />}
      {q.data && q.data.length === 0 && <EmptyState icon="people" title="No one listed yet" body="Residents appear here once they opt in from their profile." />}
      {q.data && q.data.length > 0 && (
        <ListGroup>
          {q.data.map((p) => (
            <ListRow key={p.id} leading={<Avatar name={p.display_name || p.email} size="sm" />} title={p.display_name || p.email} subtitle={`Unit ${p.unit_number}`} />
          ))}
        </ListGroup>
      )}
    </Screen>
  );
}
