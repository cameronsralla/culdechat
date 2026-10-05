import { type ReactNode } from 'react';
import { View } from 'react-native';
import { type Href, useRouter } from 'expo-router';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { useStyles, type Theme } from '../../theme';
import { AppHeader } from './AppHeader';
import { Sidebar } from './Sidebar';
import { TabBar } from './TabBar';
import { pathForTab, titleForTab, type TabId } from './nav';
import { useCompactLayout } from './useCompactLayout';

export { pathForTab, tabFromPath, titleForTab } from './nav';
export type { TabId } from './nav';

const stylesFor = (t: Theme) => ({
  shell: { flex: 1, backgroundColor: t.colors.paper },
  wash: {
    position: 'absolute' as const,
    top: 0,
    left: 0,
    right: 0,
    height: 220,
    backgroundColor: t.colors.brandWash,
    opacity: 0.45,
  },
  row: { flexDirection: 'row' as const, flex: 1 },
  column: { flex: 1, minWidth: 0, backgroundColor: 'transparent' },
  body: { flex: 1 },
});

export function AppShell({
  tab,
  userLabel,
  isAdmin,
  children,
}: {
  tab: TabId;
  userLabel?: string;
  isAdmin?: boolean;
  children: ReactNode;
}) {
  const router = useRouter();
  const insets = useSafeAreaInsets();
  const compact = useCompactLayout();
  const styles = useStyles(stylesFor);

  const go = (id: TabId) => {
    router.replace(pathForTab(id) as Href);
  };

  if (!compact) {
    return (
      <View style={[styles.shell, { paddingTop: insets.top }]}>
        <View style={styles.wash} />
        <View style={styles.row}>
          <Sidebar current={tab} userLabel={userLabel} isAdmin={isAdmin} onChange={go} />
          <View style={styles.column}>
            <View style={styles.body}>{children}</View>
          </View>
        </View>
      </View>
    );
  }

  return (
    <View style={[styles.shell, { paddingTop: insets.top }]}>
      <View style={styles.wash} />
      <AppHeader title={titleForTab(tab)} />
      <View style={styles.body}>{children}</View>
      <TabBar current={tab} isAdmin={isAdmin} onChange={go} />
    </View>
  );
}
