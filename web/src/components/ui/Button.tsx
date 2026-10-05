import { forwardRef, type ButtonHTMLAttributes } from 'react';
import { Slot } from '@radix-ui/react-slot';
import { cn } from '@/lib/cn';

type Variant = 'primary' | 'secondary' | 'ghost' | 'danger';
type Size = 'md' | 'sm';

const VARIANT: Record<Variant, string> = {
  primary: 'bg-brand text-white hover:bg-brand-dark shadow-soft',
  secondary: 'bg-surface text-ink border border-line hover:bg-surface-muted',
  ghost: 'bg-transparent text-brand hover:bg-brand-wash',
  danger: 'bg-danger-soft text-danger hover:bg-danger hover:text-white',
};

const SIZE: Record<Size, string> = {
  md: 'min-h-control px-5 text-body',
  sm: 'min-h-control-sm px-4 text-caption',
};

export type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: Variant;
  size?: Size;
  loading?: boolean;
  full?: boolean;
  /** Render as the child element (e.g. a Link) while keeping button styling. */
  asChild?: boolean;
};

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(function Button(
  { variant = 'primary', size = 'md', loading, full, asChild, className, disabled, children, ...rest },
  ref,
) {
  const Comp = asChild ? Slot : 'button';
  return (
    <Comp
      ref={ref}
      disabled={disabled || loading}
      aria-busy={loading || undefined}
      className={cn(
        'inline-flex items-center justify-center gap-2 rounded-sm font-semibold transition-colors select-none',
        'disabled:opacity-50 disabled:pointer-events-none',
        VARIANT[variant],
        SIZE[size],
        full && 'w-full',
        className,
      )}
      {...rest}
    >
      {loading ? <Spinner /> : children}
    </Comp>
  );
});

function Spinner() {
  return (
    <span
      className="size-5 rounded-pill border-2 border-current border-r-transparent animate-spin"
      role="status"
      aria-label="Loading"
    />
  );
}
