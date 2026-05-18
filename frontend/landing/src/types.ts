export type Expansion = "vanilla" | "tbc" | "wotlk";
export type Client = "1.12.1" | "2.4.3" | "3.3.5a";
export type Logging = "server" | "client";
export type Engine =
  | "azerothcore"
  | "trinitycore"
  | "mangos"
  | "custom"
  | "unknown";

export type StatusTag =
  | "closed"
  | "beta"
  | "new"
  | "hardcore"
  | "fresh"
  | "progression"
  | "custom-content";

export interface ServerEntry {
  /** URL slug and Chronicle subdomain prefix. */
  id: string;
  name: string;
  /** One-line tagline shown below the name. */
  tagline: string;
  /** 1–3 sentence description. */
  description: string;

  // Visual
  logo: string;
  banner?: string;
  /** CSS color for optional accent glow on card hover. */
  accentColor?: string;

  // Attributes
  expansion: Expansion;
  client: Client;
  logging: Logging;
  engine: Engine;

  // Links
  chronicleUrl: string;
  homepageUrl?: string;
  discordUrl?: string;

  // Tags
  status?: StatusTag[];

  /** Reserved for future sponsorship tier; ignored for now. */
  sponsored?: boolean;
}

/** Stats fetched live from each Chronicle deployment. */
export interface ServerStats {
  recentLogs: number;
  lastActivityAt: Date | null;
}
