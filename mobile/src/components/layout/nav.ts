export type TabId = 'home' | 'boards' | 'people' | 'messages' | 'admin' | 'you';

type NavItem = { id: TabId; label: string };

const residentItems: NavItem[] = [
  { id: 'home', label: 'Square' },
  { id: 'boards', label: 'Boards' },
  { id: 'people', label: 'People' },
  { id: 'messages', label: 'Messages' },
  { id: 'you', label: 'You' },
];

const adminItem: NavItem = { id: 'admin', label: 'Admin' };

export function navItemsFor(isAdmin: boolean): NavItem[] {
  if (!isAdmin) {
    return residentItems;
  }
  return [...residentItems.slice(0, 4), adminItem, residentItems[4]];
}

export function titleForTab(tab: TabId): string {
  switch (tab) {
    case 'home':
      return 'The square';
    case 'boards':
      return 'Boards';
    case 'people':
      return 'People';
    case 'messages':
      return 'Messages';
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
    case 'messages':
      return '/messages';
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
  if (pathname.startsWith('/messages')) {
    return 'messages';
  }
  if (pathname.startsWith('/admin')) {
    return 'admin';
  }
  if (pathname.startsWith('/you')) {
    return 'you';
  }
  return 'home';
}
