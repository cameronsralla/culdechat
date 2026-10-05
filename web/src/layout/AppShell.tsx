import { Outlet } from 'react-router';
import { useQuery } from '@tanstack/react-query';
import { settingsApi } from '@/api/endpoints';
import { Sidebar } from './Sidebar';
import { TabBar } from './TabBar';
import { useViewport } from './useViewport';

/**
 * The signed-in frame. Wide: sidebar + content. Compact: content + bottom tabs.
 * This is the only place the two navigation trees are chosen.
 */
export function AppShell() {
  const { wide } = useViewport();
  const settings = useQuery({ queryKey: ['settings', 'public'], queryFn: settingsApi.public, staleTime: 5 * 60_000 });
  const communityName = settings.data?.community_name ?? 'Cul-de-Chat';

  return (
    <div className="flex min-h-dvh bg-paper">
      {wide && <Sidebar communityName={communityName} />}
      <div className="min-w-0 flex-1">
        <Outlet />
      </div>
      {!wide && <TabBar />}
    </div>
  );
}
