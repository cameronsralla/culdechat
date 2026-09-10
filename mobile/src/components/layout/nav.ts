export type TabId = 'home' | 'boards' | 'people' | 'admin' | 'you';

type NavItem = { id: TabId; label: string; icon: string };

const residentItems: NavItem[] = [
  { id: 'home', label: 'Home', icon: '⌂' },
  { id: 'boards', label: 'Boards', icon: '▦' },
  { id: 'people', label: 'People', icon: '✳' },
  { id: 'you', label: 'You', icon: '☺' },
];

const adminItem: NavItem = { id: 'admin', label: 'Admin', icon: '★' };

export function navItemsFor(isAdmin: boolean): NavItem[] {
  if (!isAdmin) {
    return residentItems;
  }
  return [...residentItems.slice(0, 3), adminItem, residentItems[3]];
}

export function titleForTab(tab: TabId): string {
  switch (tab) {
    case 'home':
      return 'Cul-de-Chat';
    case 'boards':
      return 'Boards';
    case 'people':
      return 'People';
    case 'admin':
      return 'Admin';
    case 'you':
      return 'You';
    default: {
      const _never: never = tab;
      return _never;
    }
  }
}

export function pathForTab(tab: TabId): string {
  switch (tab) {
    case 'home':
      return '/';
    case 'boards':
      return '/boards';
    case 'people':
      return '/directory';
    case 'admin':
      return '/admin';
    case 'you':
      return '/you';
    default: {
      const _never: never = tab;
      return _never;
    }
  }
}

export function tabFromPath(pathname: string): TabId {
  if (pathname.startsWith('/boards')) {
    return 'boards';
  }
  if (pathname.startsWith('/directory')) {
    return 'people';
  }
  if (pathname.startsWith('/admin')) {
    return 'admin';
  }
  if (pathname.startsWith('/you')) {
    return 'you';
  }
  return 'home';
}
