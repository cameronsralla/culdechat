export function relativeTime(iso?: string | null): string {
  if (!iso) {
    return '';
  }
  const then = new Date(iso);
  if (Number.isNaN(then.getTime())) {
    return '';
  }
  const delta = Date.now() - then.getTime();
  const minute = 60 * 1000;
  const hour = 60 * minute;
  const day = 24 * hour;
  if (delta < minute) {
    return 'just now';
  }
  if (delta < hour) {
    const n = Math.floor(delta / minute);
    return `${n}m ago`;
  }
  if (delta < day) {
    const n = Math.floor(delta / hour);
    return `${n}h ago`;
  }
  if (delta < 7 * day) {
    const n = Math.floor(delta / day);
    return `${n}d ago`;
  }
  return then.toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' });
}
