import { Ionicons } from '@expo/vector-icons';
import { useTheme } from '../../theme';
import type { TabId } from '../layout/nav';

const tabIcons: Record<TabId, keyof typeof Ionicons.glyphMap> = {
  home: 'home-outline',
  boards: 'grid-outline',
  people: 'people-outline',
  messages: 'chatbubbles-outline',
  admin: 'shield-checkmark-outline',
  you: 'person-circle-outline',
};

const tabIconsActive: Record<TabId, keyof typeof Ionicons.glyphMap> = {
  home: 'home',
  boards: 'grid',
  people: 'people',
  messages: 'chatbubbles',
  admin: 'shield-checkmark',
  you: 'person-circle',
};

export function TabIcon({
  tab,
  active,
  size,
}: {
  tab: TabId;
  active?: boolean;
  size?: number;
}) {
  const theme = useTheme();
  const name = active ? tabIconsActive[tab] : tabIcons[tab];
  return (
    <Ionicons
      name={name}
      size={size ?? theme.layout.navIcon}
      color={active ? theme.colors.brandDark : theme.colors.muted}
    />
  );
}

export function AppIcon({
  name,
  size = 18,
  color,
}: {
  name: keyof typeof Ionicons.glyphMap;
  size?: number;
  color?: string;
}) {
  const theme = useTheme();
  return <Ionicons name={name} size={size} color={color ?? theme.colors.muted} />;
}
