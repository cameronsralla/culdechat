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
  row: { flexDirection: 'row' as const },
  column: { flex: 1, minWidth: 0 },
  body: { flex: 1 },
});

export function AppShell({
  tab,
  subtitle,
  isAdmin,
  children,
}: {
  tab: TabId;
  subtitle?: string;
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
      <View style={[styles.shell, styles.row, { paddingTop: insets.top }]}>
        <Sidebar current={tab} subtitle={subtitle} isAdmin={isAdmin} onChange={go} />
        <View style={styles.column}>
          <AppHeader title={titleForTab(tab)} subtitle={subtitle} />
          <View style={styles.body}>{children}</View>
        </View>
      </View>
    );
  }

  return (
    <View style={[styles.shell, { paddingTop: insets.top }]}>
      <AppHeader title={titleForTab(tab)} subtitle={subtitle} />
      <View style={styles.body}>{children}</View>
      <TabBar current={tab} isAdmin={isAdmin} onChange={go} />
    </View>
  );
}
