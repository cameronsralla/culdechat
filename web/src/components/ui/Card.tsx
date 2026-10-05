import type { HTMLAttributes } from 'react';
import { cn } from '@/lib/cn';

type Props = HTMLAttributes<HTMLDivElement> & { tone?: 'surface' | 'wash' | 'pin'; padded?: boolean };

const TONE = {
  surface: 'bg-surface border-line-soft shadow-card',
  wash: 'bg-brand-wash border-brand-line',
  pin: 'bg-pin border-pin-line',
};

export function Card({ tone = 'surface', padded = true, className, ...rest }: Props) {
  return <div className={cn('rounded-md border', TONE[tone], padded && 'p-4 md:p-5', className)} {...rest} />;
}
