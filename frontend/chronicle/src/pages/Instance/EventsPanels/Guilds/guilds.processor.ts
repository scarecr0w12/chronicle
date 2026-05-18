import type { PanelProcessor, CombatantInfoProcessorEvent, ProcessorContext } from "../processorTypes";

export interface GuildPlayerInfo {
  guid: string;
  name: string;
  heroClass: string;
}

export interface GuildEntry {
  name: string | null;
  players: Map<string, GuildPlayerInfo>;
}

export interface GuildsResult {
  /** guildName (or "" for no guild) → guild entry */
  guilds: Map<string, GuildEntry>;
}

export const guildsProcessor: PanelProcessor<GuildsResult, CombatantInfoProcessorEvent> = {
  id: "guilds",
  streams: ["combatant_info"],
  createState: () => ({ guilds: new Map() }),
  processEvent: (state, event, encounterID: string, _firstTimestamp: Date, _streamType: string, context: ProcessorContext) => {
    if (!context.selectedEncounterIds.has(encounterID)) return;

    const key = event.guildName ?? "";
    let guild = state.guilds.get(key);
    if (!guild) {
      guild = { name: event.guildName, players: new Map() };
      state.guilds.set(key, guild);
    }
    guild.players.set(event.guid, {
      guid: event.guid,
      name: event.name,
      heroClass: event.heroClass,
    });
  },
};
