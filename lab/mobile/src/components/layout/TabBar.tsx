import { Pressable, View } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { useStyles, useTheme, type Theme } from '../../theme';
import { AppText } from '../ui/AppText';
import { TabIcon } from '../ui/Icon';
import { navItemsFor, type TabId } from './nav';

export type { TabId };

const stylesFor = (t: Theme) => ({
  bar: {
    flexDirection: 'row' as const,
    borderTopWidth: t.layout.border,
    borderTopColor: t.colors.lineSoft,
    backgroundColor: t.colors.surface,
    paddingTop: t.space.sm,
  },
  item: { flex: 1, alignItems: 'center' as const, gap: 1 },
  activeDot: {
    width: 5,
    height: 5,
    borderRadius: 3,
    backgroundColor: t.colors.brand,
    marginTop: 2,
  },
});

export function TabBar({
  current,
  isAdmin,
  onChange,
}: {
  current: TabId;
  isAdmin?: boolean;
  onChange: (id: TabId) => void;
}) {
  const insets = useSafeAreaInsets();
  const theme = useTheme();
  const styles = useStyles(stylesFor);
  return (
    <View style={[styles.bar, { paddingBottom: Math.max(insets.bottom, theme.space.sm) }]}>
      {navItemsFor(Boolean(isAdmin)).map((tab) => {
        const active = tab.id === current;
        return (
          <Pressable
            key={tab.id}
            accessibilityRole="tab"
            accessibilityState={{ selected: active }}
            onPress={() => onChange(tab.id)}
            style={styles.item}
          >
            <TabIcon tab={tab.id} active={active} />
            <AppText variant="caption" tone={active ? 'brand' : 'muted'}>
              {tab.label}
            </AppText>
            {active ? <View style={styles.activeDot} /> : <View style={{ height: 6 }} />}
          </Pressable>
        );
      })}
    </View>
  );
}
