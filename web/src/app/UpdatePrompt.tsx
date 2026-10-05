import { useRegisterSW } from 'virtual:pwa-register/react';
import { Button, Text } from '@/components/ui';

/** PWA: when a new build is deployed, offer a one-tap reload. */
export function UpdatePrompt() {
  const {
    needRefresh: [needRefresh, setNeedRefresh],
    updateServiceWorker,
  } = useRegisterSW();

  if (!needRefresh) return null;
  return (
    <div className="fixed inset-x-4 top-4 z-50 mx-auto flex max-w-md items-center gap-3 rounded-md border border-line bg-raised px-4 py-3 shadow-raised">
      <Text variant="caption" className="flex-1">
        A new version is available.
      </Text>
      <Button size="sm" onClick={() => void updateServiceWorker(true)}>
        Update
      </Button>
      <Button size="sm" variant="ghost" onClick={() => setNeedRefresh(false)}>
        Later
      </Button>
    </div>
  );
}
