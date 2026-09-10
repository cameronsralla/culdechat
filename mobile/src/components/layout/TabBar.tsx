import { Pressable, View } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { useStyles, useTheme, type Theme } from '../../theme';
import { AppText } from '../ui/AppText';
import { navItemsFor, type TabId } from './nav';

export type { TabId };

const stylesFor = (t: Theme) => ({
  bar: {
    flexDirection: 'row' as const,
    borderTopWidth: t.layout.border,
    borderTopColor: t.colors.line,
    backgroundColor: t.colors.surface,
    paddingTop: t.space.sm,
  },
  item: { flex: 1, alignItems: 'center' as const, gap: t.space.xxs },
  icon: { fontSize: t.layout.navIcon, lineHeight: t.layout.navIcon + 4 },
  idle: { color: t.colors.muted },
  active: { color: t.colors.brandDark, fontFamily: t.type.subtitle.fontFamily },
  activeLabel: { fontFamily: t.type.label.fontFamily },
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
            <AppText style={[styles.icon, active ? styles.active : styles.idle]}>{tab.icon}</AppText>
            <AppText variant="caption" tone={active ? 'brand' : 'muted'} style={active ? styles.activeLabel : undefined}>
              {tab.label}
            </AppText>
          </Pressable>
        );
      })}
    </View>
  );
}
