import type { IconName } from '@/components/ui/Icon';

export type NavItem = { id: string; label: string; to: string; icon: IconName; adminOnly?: boolean; end?: boolean };

/** Primary navigation. Order matters: Sidebar and TabBar render the same list. */
export const NAV: NavItem[] = [
  { id: 'home', label: 'Square', to: '/', icon: 'home', end: true },
  { id: 'people', label: 'People', to: '/directory', icon: 'people' },
  { id: 'admin', label: 'Admin', to: '/admin', icon: 'shield', adminOnly: true },
  { id: 'you', label: 'You', to: '/you', icon: 'user' },
];

export function navFor(isAdmin: boolean): NavItem[] {
  return NAV.filter((n) => !n.adminOnly || isAdmin);
}
