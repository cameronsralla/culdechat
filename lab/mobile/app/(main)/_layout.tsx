import { Redirect, Slot, usePathname } from 'expo-router';
import { useAuth } from '../../src/auth/AuthContext';
import { AppShell, tabFromPath } from '../../src/components/layout/AppShell';
import { LoadingScreen } from '../../src/components/layout/LoadingScreen';

export default function MainLayout() {
  const { ready, user } = useAuth();
  const pathname = usePathname();

  if (!ready) {
    return <LoadingScreen />;
  }
  if (!user) {
    return <Redirect href="/login" />;
  }

  const tab = tabFromPath(pathname);
  const userLabel = user.name?.trim()
    ? `${user.name} · Unit ${user.unit_number}`
    : `Unit ${user.unit_number}`;

  return (
    <AppShell tab={tab} userLabel={userLabel} isAdmin={user.is_admin}>
      <Slot />
    </AppShell>
  );
}
