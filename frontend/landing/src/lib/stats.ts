import type { ServerStats } from "../types";

const CACHE_TTL_MS = 5 * 60 * 1000; // 5 minutes
const FETCH_TIMEOUT_MS = 5_000;

interface CacheEntry {
  stats: ServerStats;
  fetchedAt: number;
}

function cacheKey(chronicleUrl: string): string {
  return `chronicle-stats:${chronicleUrl}`;
}

function readCache(chronicleUrl: string): ServerStats | null {
  try {
    const raw = sessionStorage.getItem(cacheKey(chronicleUrl));
    if (!raw) return null;
    const entry: CacheEntry = JSON.parse(raw);
    if (Date.now() - entry.fetchedAt > CACHE_TTL_MS) {
      sessionStorage.removeItem(cacheKey(chronicleUrl));
      return null;
    }
    return {
      ...entry.stats,
      lastActivityAt: entry.stats.lastActivityAt
        ? new Date(entry.stats.lastActivityAt)
        : null,
    };
  } catch {
    return null;
  }
}

function writeCache(chronicleUrl: string, stats: ServerStats): void {
  try {
    const entry: CacheEntry = { stats, fetchedAt: Date.now() };
    sessionStorage.setItem(cacheKey(chronicleUrl), JSON.stringify(entry));
  } catch {
    // sessionStorage full or unavailable — ignore
  }
}

/**
 * Fetch recent raid stats from a Chronicle deployment.
 * Returns null on any failure (timeout, network, CORS, bad response).
 * Results are cached in sessionStorage for 5 minutes.
 */
export async function fetchServerStats(
  chronicleUrl: string,
): Promise<ServerStats | null> {
  const cached = readCache(chronicleUrl);
  if (cached) return cached;

  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), FETCH_TIMEOUT_MS);

  try {
    const res = await fetch(
      `${chronicleUrl}/api/v1/raidlogs/recent`,
      { signal: controller.signal },
    );
    if (!res.ok) return null;

    const data = await res.json();

    // The response has an `instances` array with `started_at` timestamps.
    // We take the first 5 most recent and compute stats from those.
    const instances: Array<{ started_at: string }> =
      Array.isArray(data?.instances) ? data.instances : [];

    const now = Date.now();
    const sevenDaysAgo = now - 7 * 24 * 60 * 60 * 1000;

    let recentLogs = 0;
    let lastActivityAt: Date | null = null;

    // Only look at up to 5 most recent (they should already be sorted desc)
    const slice = instances.slice(0, 5);
    for (const inst of slice) {
      const ts = new Date(inst.started_at);
      if (ts.getTime() >= sevenDaysAgo) {
        recentLogs++;
      }
      if (!lastActivityAt || ts > lastActivityAt) {
        lastActivityAt = ts;
      }
    }

    const stats: ServerStats = { recentLogs, lastActivityAt };
    writeCache(chronicleUrl, stats);
    return stats;
  } catch {
    return null;
  } finally {
    clearTimeout(timeout);
  }
}
