import type { HTMLAttributes } from 'react';
import { cn } from '@/lib/cn';

type Tone = 'brand' | 'neutral' | 'pin' | 'danger' | 'success';
const TONE: Record<Tone, string> = {
  brand: 'bg-brand-soft text-brand',
  neutral: 'bg-surface-muted text-muted',
  pin: 'bg-pin text-ink border border-pin-line',
  danger: 'bg-danger-soft text-danger',
  success: 'bg-success-soft text-success',
};

/** Compact status/meta pill. Not uppercase by default — quieter craft look. */
export function Badge({ tone = 'brand', className, ...rest }: HTMLAttributes<HTMLSpanElement> & { tone?: Tone }) {
  return (
    <span
      className={cn(
        'inline-flex items-center rounded-pill px-2 py-0.5 text-caption font-medium leading-none',
        TONE[tone],
        className,
      )}
      {...rest}
    />
  );
}
