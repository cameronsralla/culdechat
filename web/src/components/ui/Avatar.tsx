import * as RadixAvatar from '@radix-ui/react-avatar';
import { cn } from '@/lib/cn';

type Size = 'sm' | 'md' | 'lg';
const SIZE: Record<Size, string> = { sm: 'size-10 text-caption', md: 'size-12 text-body', lg: 'size-22 text-title' };

export function Avatar({ name, src, size = 'md', className }: { name: string; src?: string; size?: Size; className?: string }) {
  return (
    <RadixAvatar.Root
      className={cn('inline-flex shrink-0 items-center justify-center overflow-hidden rounded-pill bg-surface-muted text-muted font-semibold select-none', SIZE[size], className)}
    >
      {src && <RadixAvatar.Image src={src} alt={name} className="size-full object-cover" />}
      <RadixAvatar.Fallback delayMs={src ? 300 : 0}>{initials(name)}</RadixAvatar.Fallback>
    </RadixAvatar.Root>
  );
}

export function initials(name: string): string {
  const parts = name.trim().split(/\s+/).filter(Boolean);
  if (parts.length === 0) return '?';
  if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase();
  return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
}
