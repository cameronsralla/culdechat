import type { HTMLAttributes } from 'react';
import { cn } from '@/lib/cn';

type Tone = 'brand' | 'neutral' | 'pin' | 'danger';
const TONE: Record<Tone, string> = {
  brand: 'bg-brand-soft text-brand-dark',
  neutral: 'bg-ink-faint text-muted',
  pin: 'bg-pin text-ink',
  danger: 'bg-danger-soft text-danger',
};

export function Badge({ tone = 'brand', className, ...rest }: HTMLAttributes<HTMLSpanElement> & { tone?: Tone }) {
  return (
    <span
      className={cn('inline-flex items-center rounded-pill px-2.5 py-0.5 text-label font-semibold uppercase', TONE[tone], className)}
      {...rest}
    />
  );
}
