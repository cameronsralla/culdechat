import { Image, Pressable, View } from 'react-native';
import { useStyles, type Theme } from '../../theme';
import { AppText } from '../ui/AppText';
import { navItemsFor, type TabId } from './nav';

const logo = require('../../../assets/logo.jpg');

const stylesFor = (t: Theme) => ({
  rail: {
    width: t.layout.sidebarWidth,
    backgroundColor: t.colors.surface,
    borderRightWidth: t.layout.border,
    borderRightColor: t.colors.line,
    paddingTop: t.space.xl,
    paddingHorizontal: t.space.md,
  },
  brand: {
    flexDirection: 'row' as const,
    alignItems: 'center' as const,
    gap: t.space.md,
    paddingHorizontal: t.space.sm,
    marginBottom: t.space.xl,
  },
  logo: { width: t.layout.sidebarLogo, height: t.layout.sidebarLogo },
  nav: { gap: t.space.xs },
  item: {
    flexDirection: 'row' as const,
    alignItems: 'center' as const,
    gap: t.space.md,
    paddingVertical: t.space.md,
    paddingHorizontal: t.space.md,
    borderRadius: t.radius.sm,
  },
  itemActive: { backgroundColor: t.colors.brandSoft },
  icon: { fontSize: t.layout.navIcon, lineHeight: t.layout.navIcon + 4, width: t.layout.navIcon + 4, textAlign: 'center' as const },
  iconIdle: { color: t.colors.muted },
  iconActive: { color: t.colors.brandDark, fontFamily: t.type.subtitle.fontFamily },
});

export function Sidebar({
  current,
  subtitle,
  isAdmin,
  onChange,
}: {
  current: TabId;
  subtitle?: string;
  isAdmin?: boolean;
  onChange: (id: TabId) => void;
}) {
  const styles = useStyles(stylesFor);
  return (
    <View style={styles.rail}>
      <View style={styles.brand}>
        <Image source={logo} style={styles.logo} resizeMode="contain" />
        <View>
          <AppText variant="subtitle">Cul-de-Chat</AppText>
          {subtitle ? (
            <AppText variant="caption" tone="muted">
              {subtitle}
            </AppText>
          ) : null}
        </View>
      </View>
      <View style={styles.nav}>
        {navItemsFor(Boolean(isAdmin)).map((item) => {
          const active = item.id === current;
          return (
            <Pressable
              key={item.id}
              accessibilityRole="button"
              accessibilityState={{ selected: active }}
              onPress={() => onChange(item.id)}
              style={[styles.item, active ? styles.itemActive : null]}
            >
              <AppText style={[styles.icon, active ? styles.iconActive : styles.iconIdle]}>{item.icon}</AppText>
              <AppText variant="subtitle" tone={active ? 'brand' : 'ink'}>
                {item.label}
              </AppText>
            </Pressable>
          );
        })}
      </View>
    </View>
  );
}
