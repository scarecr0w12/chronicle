import { useEffect, useState } from "react";
import type { ServerStats } from "../types";
import { fetchServerStats } from "../lib/stats";
import { relativeTime } from "../lib/relativeTime";

export function LiveStats({ chronicleUrl }: { chronicleUrl: string }) {
  const [stats, setStats] = useState<ServerStats | null>(null);

  useEffect(() => {
    let cancelled = false;
    fetchServerStats(chronicleUrl).then((s) => {
      if (!cancelled) setStats(s);
    });
    return () => { cancelled = true; };
  }, [chronicleUrl]);

  if (!stats || (stats.recentLogs === 0 && !stats.lastActivityAt)) {
    return null;
  }

  return (
    <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
      <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="shrink-0 opacity-60">
        <path d="M18 20V10" /><path d="M12 20V4" /><path d="M6 20v-6" />
      </svg>
      <span>
        {stats.recentLogs > 0 && (
          <>{stats.recentLogs} log{stats.recentLogs !== 1 ? "s" : ""} this week</>
        )}
        {stats.recentLogs > 0 && stats.lastActivityAt && " · "}
        {stats.lastActivityAt && (
          <>last activity {relativeTime(stats.lastActivityAt)}</>
        )}
      </span>
    </div>
  );
}
