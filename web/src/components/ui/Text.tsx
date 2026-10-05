import type { HTMLAttributes, ElementType } from 'react';
import { cn } from '@/lib/cn';

type Variant = 'display' | 'title' | 'subtitle' | 'body' | 'label' | 'caption' | 'mono';
type Tone = 'ink' | 'muted' | 'brand' | 'danger' | 'success' | 'inverse';

const VARIANT: Record<Variant, string> = {
  display: 'text-display font-bold',
  title: 'text-title font-semibold',
  subtitle: 'text-subtitle font-semibold',
  body: 'text-body font-normal',
  label: 'text-label font-medium',
  caption: 'text-caption font-normal',
  mono: 'text-caption font-mono font-medium tracking-tight',
};

const TONE: Record<Tone, string> = {
  ink: 'text-ink',
  muted: 'text-muted',
  brand: 'text-brand',
  danger: 'text-danger',
  success: 'text-success',
  inverse: 'text-white',
};

const DEFAULT_TAG: Record<Variant, ElementType> = {
  display: 'h1',
  title: 'h2',
  subtitle: 'h3',
  body: 'p',
  label: 'span',
  caption: 'span',
  mono: 'span',
};

type Props = HTMLAttributes<HTMLElement> & { variant?: Variant; tone?: Tone; as?: ElementType };

/** Typography primitive. All text in the app goes through here or a kit component. */
export function Text({ variant = 'body', tone = 'ink', as, className, ...rest }: Props) {
  const Tag = as ?? DEFAULT_TAG[variant];
  return <Tag className={cn('m-0', VARIANT[variant], TONE[tone], className)} {...rest} />;
}
