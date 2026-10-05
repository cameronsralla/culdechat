import { useSyncExternalStore } from 'react';

/**
 * The single compact/wide switch for layout components. Mirrors the `md`
 * breakpoint in tokens.css (800px). Use Tailwind `md:` utilities for styling
 * differences; use this hook only when the component *tree* differs
 * (e.g. Sidebar vs TabBar).
 */
const QUERY = '(min-width: 50rem)';

function subscribe(cb: () => void) {
  const mq = window.matchMedia(QUERY);
  mq.addEventListener('change', cb);
  return () => mq.removeEventListener('change', cb);
}

export function useViewport(): { compact: boolean; wide: boolean } {
  const wide = useSyncExternalStore(
    subscribe,
    () => window.matchMedia(QUERY).matches,
    () => true,
  );
  return { compact: !wide, wide };
}
