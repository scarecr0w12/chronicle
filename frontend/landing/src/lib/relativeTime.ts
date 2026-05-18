/**
 * Format a date as a human-readable relative time string.
 * e.g. "3 hours ago", "2 days ago", "just now"
 */
export function relativeTime(date: Date): string {
  const seconds = Math.floor((Date.now() - date.getTime()) / 1000);

  if (seconds < 60) return "just now";
  if (seconds < 3600) {
    const m = Math.floor(seconds / 60);
    return `${m}m ago`;
  }
  if (seconds < 86400) {
    const h = Math.floor(seconds / 3600);
    return `${h}h ago`;
  }
  const d = Math.floor(seconds / 86400);
  if (d === 1) return "yesterday";
  if (d < 30) return `${d}d ago`;
  return date.toLocaleDateString();
}
