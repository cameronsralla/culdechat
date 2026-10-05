import type { HTMLAttributes, ElementType } from 'react';
import { cn } from '@/lib/cn';

type Variant = 'display' | 'title' | 'subtitle' | 'body' | 'label' | 'caption';
type Tone = 'ink' | 'muted' | 'brand' | 'danger' | 'inverse';

const VARIANT: Record<Variant, string> = {
  display: 'text-display font-extrabold',
  title: 'text-title font-bold',
  subtitle: 'text-subtitle font-semibold',
  body: 'text-body',
  label: 'text-label font-semibold uppercase',
  caption: 'text-caption',
};

const TONE: Record<Tone, string> = {
  ink: 'text-ink',
  muted: 'text-muted',
  brand: 'text-brand',
  danger: 'text-danger',
  inverse: 'text-white',
};

const DEFAULT_TAG: Record<Variant, ElementType> = {
  display: 'h1',
  title: 'h2',
  subtitle: 'h3',
  body: 'p',
  label: 'span',
  caption: 'span',
};

type Props = HTMLAttributes<HTMLElement> & { variant?: Variant; tone?: Tone; as?: ElementType };

/** Typography primitive. All text in the app goes through here or a kit component. */
export function Text({ variant = 'body', tone = 'ink', as, className, ...rest }: Props) {
  const Tag = as ?? DEFAULT_TAG[variant];
  return <Tag className={cn('m-0', VARIANT[variant], TONE[tone], className)} {...rest} />;
}
