const SERVER_NAME = import.meta.env.VITE_SERVER_NAME ?? "turtle";

/** Features that may differ per server. */
export interface ServerCapabilities {
  armory: boolean;
  /** Which faction Blood Elf belongs to on this server. */
  bloodElfFaction: "Horde" | "Alliance";
}

const CAPABILITIES: Record<string, ServerCapabilities> = {
  turtle: { armory: true, bloodElfFaction: "Alliance" },
  octowow: { armory: true, bloodElfFaction: "Alliance" },
};

const DEFAULT_CAPABILITIES: ServerCapabilities = {
  armory: true,
  bloodElfFaction: "Horde",
};

/** Capabilities for the current server. */
export const serverCapabilities: ServerCapabilities =
  CAPABILITIES[SERVER_NAME] ?? DEFAULT_CAPABILITIES;
