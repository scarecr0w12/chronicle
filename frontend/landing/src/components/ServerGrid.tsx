import { useEffect, useMemo, useState } from "react";
import type { Expansion, Client, Logging, ServerEntry, StatusTag } from "../types";
import { ServerCard } from "./ServerCard";
import { DiscordIcon } from "./DiscordIcon";

const DISCORD_URL = "https://discord.gg/gz97ABFVAj";

function GetInTouchModal({ onClose }: { onClose: () => void }) {
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", onKey);
    document.body.style.overflow = "hidden";
    return () => {
      window.removeEventListener("keydown", onKey);
      document.body.style.overflow = "";
    };
  }, [onClose]);

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 p-4"
      onClick={onClose}
    >
      <div
        className="relative w-full max-w-md rounded-lg border border-border bg-card p-6 shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <button
          onClick={onClose}
          className="absolute right-3 top-3 rounded p-1 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
          aria-label="Close"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <line x1="18" y1="6" x2="6" y2="18" />
            <line x1="6" y1="6" x2="18" y2="18" />
          </svg>
        </button>

        <h2 className="text-xl font-semibold text-foreground">
          Get in touch via Discord
        </h2>
        <p className="mt-3 text-sm text-muted-foreground">
          We'd love to help bring Chronicle to you. Reach out on our Discord
          and we'll get you set up.
        </p>
        <p className="mt-3 text-sm text-muted-foreground">
          Chronicle is open source and{" "}
          <a
            href="https://github.com/Emyrk/chronicle/blob/main/DEPLOYING.md"
            target="_blank"
            rel="noreferrer noopener"
            className="text-primary hover:underline"
          >
            self-hosting is fully supported
          </a>
          {" "}— run it on your own infrastructure if you prefer.
        </p>

        <a
          href={DISCORD_URL}
          target="_blank"
          rel="noreferrer noopener"
          className="mt-5 inline-flex items-center gap-2 rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90"
        >
          <DiscordIcon className="h-4 w-4" />
          Join the Chronicle Discord
        </a>
      </div>
    </div>
  );
}

/** Preserve the order from the registry (sponsored entries float to top). */
function sortServers(servers: ServerEntry[]): ServerEntry[] {
  return [...servers].sort((a, b) => {
    if (a.sponsored && !b.sponsored) return -1;
    if (!a.sponsored && b.sponsored) return 1;
    return 0; // stable: keep registry order
  });
}

// --- Filter types ---

type FilterKey = "expansion" | "client" | "logging" | "status";

interface FilterOption {
  key: FilterKey;
  value: string;
  label: string;
}

/** Derive available filter options from the actual server list. */
function deriveFilters(servers: ServerEntry[]): FilterOption[] {
  const expansionLabels: Record<Expansion, string> = { vanilla: "Vanilla", tbc: "TBC", wotlk: "WotLK" };
  const clientLabels: Record<Client, string> = { "1.12.1": "1.12.1", "2.4.3": "2.4.3", "3.3.5a": "3.3.5a" };
  const loggingLabels: Record<Logging, string> = { server: "Server-side log", client: "Client-side log" };
  const statusLabels: Record<StatusTag, string> = {
    closed: "Closed", beta: "Beta", new: "New", hardcore: "Hardcore", fresh: "Fresh",
    progression: "Progression", "custom-content": "Custom Content",
  };

  const seen = new Set<string>();
  const filters: FilterOption[] = [];

  const add = (key: FilterKey, value: string, label: string) => {
    const id = `${key}:${value}`;
    if (!seen.has(id)) {
      seen.add(id);
      filters.push({ key, value, label });
    }
  };

  for (const s of servers) {
    add("expansion", s.expansion, expansionLabels[s.expansion]);
    add("client", s.client, clientLabels[s.client]);
    add("logging", s.logging, loggingLabels[s.logging]);
    for (const tag of s.status ?? []) {
      add("status", tag, statusLabels[tag]);
    }
  }

  return filters;
}

function matchesFilters(server: ServerEntry, active: Set<string>): boolean {
  if (active.size === 0) return true;

  // Group active filters by key — within a key it's OR, across keys it's AND
  const byKey = new Map<FilterKey, string[]>();
  for (const id of active) {
    const [key, value] = id.split(":") as [FilterKey, string];
    const arr = byKey.get(key) ?? [];
    arr.push(value);
    byKey.set(key, arr);
  }

  for (const [key, values] of byKey) {
    if (key === "status") {
      if (!values.some((v) => server.status?.includes(v as StatusTag))) return false;
    } else {
      if (!values.includes(server[key])) return false;
    }
  }
  return true;
}

// --- Filter pill ---

function FilterPill({
  label,
  active,
  onClick,
}: {
  label: string;
  active: boolean;
  onClick: () => void;
}) {
  return (
    <button
      onClick={onClick}
      className={
        active
          ? "rounded-full border border-primary/50 bg-primary/15 px-3 py-1 text-xs font-medium text-primary transition-colors"
          : "rounded-full border border-border px-3 py-1 text-xs text-muted-foreground transition-colors hover:border-foreground/20 hover:text-foreground"
      }
    >
      {label}
    </button>
  );
}

// --- Grid ---

export function ServerGrid({ servers }: { servers: ServerEntry[] }) {
  const [activeFilters, setActiveFilters] = useState<Set<string>>(new Set());
  const [modalOpen, setModalOpen] = useState(false);

  const filterOptions = useMemo(() => deriveFilters(servers), [servers]);

  const toggle = (key: FilterKey, value: string) => {
    setActiveFilters((prev) => {
      const next = new Set(prev);
      const id = `${key}:${value}`;
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  };

  const sorted = useMemo(() => sortServers(servers), [servers]);

  const matches = useMemo(() => {
    if (activeFilters.size === 0) return null; // no filtering active
    const set = new Set<string>();
    for (const s of servers) {
      if (matchesFilters(s, activeFilters)) set.add(s.id);
    }
    return set;
  }, [servers, activeFilters]);

  return (
    <section className="relative mx-auto w-full max-w-6xl px-4 pt-8 pb-12 sm:px-6 sm:pt-12 lg:px-8">
      {/* Subtle radial gradient background */}
      <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_top,_var(--primary-darker)_0%,_transparent_60%)] opacity-40 pointer-events-none" />

      {/* Hero header */}
      <div className="relative mb-6 text-center">
        <img
          src="chronicle-logo.svg"
          alt="Chronicle"
          className="mx-auto mb-4 h-28 sm:h-28"
          onError={(e) => {
            (e.target as HTMLImageElement).style.display = "none";
          }}
        />

        <h1 className="text-2xl font-bold tracking-tight sm:text-3xl lg:text-4xl">
          Combat log analysis for{" "}
          <span className="text-primary">Classic WoW</span>
        </h1>

        <p className="mx-auto mt-2 max-w-2xl text-base text-muted-foreground sm:text-lg">
          Chronicle transforms raid logs into clear, actionable insights.
          Select a server below to explore.
        </p>

        <div className="mt-3 flex items-center justify-center gap-4 text-sm text-muted-foreground">
          <a
            href="https://github.com/Emyrk/chronicle"
            target="_blank"
            rel="noreferrer noopener"
            className="inline-flex items-center gap-1.5 transition-colors hover:text-foreground"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor" className="shrink-0">
              <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z" />
            </svg>
            GitHub
          </a>
          <span className="text-border">·</span>
          <button
            type="button"
            onClick={() => setModalOpen(true)}
            className="transition-colors hover:text-foreground cursor-pointer"
          >
            Run Chronicle for your server →
          </button>
        </div>
      </div>

      {/* Separator */}
      <div className="relative mx-auto mb-6 h-px w-full max-w-xs bg-border/60" />

      {/* Filter bar */}
      <div className="relative mb-6 flex flex-wrap items-center justify-center gap-2">
        {filterOptions.map((f) => {
          const id = `${f.key}:${f.value}`;
          return (
            <FilterPill
              key={id}
              label={f.label}
              active={activeFilters.has(id)}
              onClick={() => toggle(f.key, f.value)}
            />
          );
        })}
        {activeFilters.size > 0 && (
          <button
            onClick={() => setActiveFilters(new Set())}
            className="ml-1 rounded-full px-2.5 py-1 text-xs text-muted-foreground transition-colors hover:text-foreground"
          >
            Clear
          </button>
        )}
      </div>

      {/* Grid — non-matching cards are greyed out instead of hidden */}
      <div className="relative grid auto-rows-[1fr] gap-6" style={{ gridTemplateColumns: "repeat(auto-fill, minmax(320px, 1fr))" }}>
        {sorted.map((server) => {
          const dimmed = matches !== null && !matches.has(server.id);
          return (
            <div
              key={server.id}
              className={`flex ${dimmed ? "opacity-30 grayscale pointer-events-none transition-all duration-200" : "transition-all duration-200"}`}
            >
              <ServerCard server={server} />
            </div>
          );
        })}
      </div>

      {modalOpen && <GetInTouchModal onClose={() => setModalOpen(false)} />}
    </section>
  );
}
