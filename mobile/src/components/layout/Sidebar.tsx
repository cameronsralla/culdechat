import { Image, Pressable, View } from 'react-native';
import { useStyles, type Theme } from '../../theme';
import { AppText } from '../ui/AppText';
import { TabIcon } from '../ui/Icon';
import { navItemsFor, type TabId } from './nav';

const logo = require('../../../assets/logo.jpg');

const stylesFor = (t: Theme) => ({
  rail: {
    width: t.layout.sidebarWidth,
    backgroundColor: t.colors.surface,
    borderRightWidth: t.layout.border,
    borderRightColor: t.colors.lineSoft,
    paddingTop: t.space.xxl,
    paddingHorizontal: t.space.md,
    paddingBottom: t.space.xl,
    justifyContent: 'space-between' as const,
  },
  top: { gap: t.space.xl },
  brand: {
    gap: t.space.md,
    paddingHorizontal: t.space.sm,
    paddingBottom: t.space.md,
    borderBottomWidth: t.layout.border,
    borderBottomColor: t.colors.lineSoft,
  },
  brandRow: {
    flexDirection: 'row' as const,
    alignItems: 'center' as const,
    gap: t.space.md,
  },
  logoWrap: {
    width: t.layout.sidebarLogo + 10,
    height: t.layout.sidebarLogo + 10,
    borderRadius: t.radius.md,
    backgroundColor: t.colors.brandWash,
    borderWidth: t.layout.border,
    borderColor: t.colors.brandLine,
    alignItems: 'center' as const,
    justifyContent: 'center' as const,
    overflow: 'hidden' as const,
  },
  logo: { width: t.layout.sidebarLogo, height: t.layout.sidebarLogo },
  chip: {
    alignSelf: 'flex-start' as const,
    backgroundColor: t.colors.brandSoft,
    borderRadius: t.radius.pill,
    paddingHorizontal: t.space.md,
    paddingVertical: t.space.xs,
    borderWidth: t.layout.border,
    borderColor: t.colors.brandLine,
  },
  nav: { gap: t.space.xs },
  item: {
    flexDirection: 'row' as const,
    alignItems: 'center' as const,
    gap: t.space.md,
    paddingVertical: t.space.md,
    paddingHorizontal: t.space.md,
    borderRadius: t.radius.sm,
  },
  itemActive: {
    backgroundColor: t.colors.brandSoft,
  },
  activeBar: {
    position: 'absolute' as const,
    left: 0,
    top: 10,
    bottom: 10,
    width: 3,
    borderRadius: 2,
    backgroundColor: t.colors.brand,
  },
  footer: {
    paddingHorizontal: t.space.sm,
    paddingTop: t.space.lg,
    borderTopWidth: t.layout.border,
    borderTopColor: t.colors.lineSoft,
    gap: t.space.xs,
  },
});

export function Sidebar({
  current,
  placeLabel = 'Neighbors only',
  userLabel,
  isAdmin,
  onChange,
}: {
  current: TabId;
  placeLabel?: string;
  userLabel?: string;
  isAdmin?: boolean;
  onChange: (id: TabId) => void;
}) {
  const styles = useStyles(stylesFor);
  return (
    <View style={styles.rail}>
      <View style={styles.top}>
        <View style={styles.brand}>
          <View style={styles.brandRow}>
            <View style={styles.logoWrap}>
              <Image source={logo} style={styles.logo} resizeMode="contain" />
            </View>
            <View>
              <AppText variant="subtitle">Cul-de-Chat</AppText>
              <AppText variant="caption" tone="muted">
                Your town square
              </AppText>
            </View>
          </View>
          <View style={styles.chip}>
            <AppText variant="label" tone="brand">
              {placeLabel}
            </AppText>
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
                {active ? <View style={styles.activeBar} /> : null}
                <TabIcon tab={item.id} active={active} />
                <AppText variant="subtitle" tone={active ? 'brand' : 'ink'}>
                  {item.label}
                </AppText>
              </Pressable>
            );
          })}
        </View>
      </View>
      {userLabel ? (
        <View style={styles.footer}>
          <AppText variant="caption" tone="muted">
            Signed in as
          </AppText>
          <AppText variant="label" tone="ink">
            {userLabel}
          </AppText>
        </View>
      ) : null}
    </View>
  );
}
