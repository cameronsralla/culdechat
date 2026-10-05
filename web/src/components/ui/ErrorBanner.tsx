import { Icon } from './Icon';
import { Text } from './Text';

export function ErrorBanner({ message }: { message?: string | null }) {
  if (!message) return null;
  return (
    <div role="alert" className="flex items-start gap-2 rounded-sm border border-danger/30 bg-danger-soft px-3 py-2.5">
      <Icon name="alert" size={18} className="mt-0.5 shrink-0 text-danger" />
      <Text variant="caption" tone="danger">
        {message}
      </Text>
    </div>
  );
}

/** Pull a human message off any thrown value. */
export function errorMessage(err: unknown, fallback = 'Something went wrong.'): string {
  if (err instanceof Error && err.message) return err.message;
  return fallback;
}
