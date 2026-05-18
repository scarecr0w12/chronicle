import type { ServerEntry } from "../types";

const EXPANSION_LABELS: Record<string, string> = {
  vanilla: "Vanilla",
  tbc: "TBC",
  wotlk: "WotLK",
};

const ENGINE_LABELS: Record<string, string> = {
  azerothcore: "AzerothCore",
  trinitycore: "TrinityCore",
  mangos: "MaNGOS",
  custom: "Custom",
};

function Badge({ children, highlight }: { children: React.ReactNode; highlight?: boolean }) {
  return (
    <span className={
      highlight
        ? "inline-flex items-center rounded-md border border-green-500/40 bg-green-500/10 px-2 py-0.5 text-xs font-medium text-green-400"
        : "inline-flex items-center rounded-md border border-border px-2 py-0.5 text-xs text-muted-foreground"
    }>
      {children}
    </span>
  );
}

export function AttributeBadges({ server }: { server: ServerEntry }) {
  return (
    <div className="flex flex-wrap gap-1.5">
      <Badge>{EXPANSION_LABELS[server.expansion] ?? server.expansion}</Badge>
      <Badge>{server.client}</Badge>
      <Badge highlight={server.logging === "server"}>{server.logging === "server" ? "✦ Server-side log" : "Client-side log"}</Badge>
      {server.engine !== "unknown" && (
        <Badge>{ENGINE_LABELS[server.engine] ?? server.engine}</Badge>
      )}
    </div>
  );
}
