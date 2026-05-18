import type { ServerEntry } from "../types";

/**
 * Static registry of Chronicle-enabled WoW private servers.
 *
 * To add a server, append an entry here. Logos go in public/servers/<id>/logo.png.
 * Banners (optional) go in public/servers/<id>/banner.webp.
 */
export const SERVERS: ServerEntry[] = [
  {
    id: "turtle",
    name: "Turtle WoW",
    tagline: "Vanilla+ with custom content",
    description:
      "Vanilla 1.12-based server with extensive custom quests, zones, dungeons, raids, races, and class changes. Focuses on expanding the original Azeroth while preserving a Classic-style experience.",
    logo: "servers/turtle/logo.png",
    banner: "servers/turtle/banner.webp",
    accentColor: "#4ade80",
    expansion: "vanilla",
    client: "1.12.1",
    logging: "client",
    engine: "unknown",
    chronicleUrl: "https://turtle.chronicleclassic.com",
    homepageUrl: "https://turtlecraft.gg",
    status: ["closed", "custom-content"],
  },
  {
    id: "oldmanwarcraft",
    name: "Old Man Warcraft",
    tagline: "WotLK with PlayerBots",
    description:
      "Laid-back WotLK 3.3.5a community server built around AzerothCore and PlayerBots. Designed to support small-group progression with bot-assisted dungeon and raid play.",
    logo: "servers/oldmanwarcraft/logo.png",
    banner: "servers/oldmanwarcraft/banner.webp",
    accentColor: "#d97706",
    expansion: "wotlk",
    client: "3.3.5a",
    logging: "server",
    engine: "azerothcore",
    chronicleUrl: "https://logs.oldmanwarcraft.com",
    homepageUrl: "https://oldmanwarcraft.com",
  },
  {
    id: "kronos",
    name: "Kronos",
    tagline: "Authentic Vanilla project",
    description:
      "Long-running Vanilla project under TwinStar focused on an authentic 1.12-style experience. Emphasizes high-quality scripting across raids, dungeons, and non-raid content.",
    logo: "servers/kronos/logo.png",
    banner: "servers/kronos/banner.jpg",
    accentColor: "#fbbf24",
    expansion: "vanilla",
    client: "1.12.1",
    logging: "client",
    engine: "unknown",
    chronicleUrl: "https://kronos.chronicleclassic.com",
    homepageUrl: "https://www.kronos-wow.com",
  },
  {
    id: "vanillaplus",
    name: "Vanilla+",
    tagline: "Rebalanced Vanilla PvP",
    description:
      "Vanilla+ PvP server with rebalanced classes, new challenges, custom bosses, battlegrounds, and community-driven content. Designed around new builds without leaving the Vanilla framework.",
    logo: "servers/vanillaplus/logo.png",
    banner: "servers/vanillaplus/banner.png",
    accentColor: "#a78bfa",
    expansion: "vanilla",
    client: "1.12.1",
    logging: "client",
    engine: "unknown",
    chronicleUrl: "https://vanillaplus.chronicleclassic.com",
    homepageUrl: "https://vanillaplus.org",
  },
  {
    id: "octowow",
    name: "Octo WoW",
    tagline: "Vanilla+ with custom content",
    description:
      "A Vanilla 1.12-based server with custom quests, dungeons, and content additions that expand on the Classic experience.",
    logo: "servers/octowow/logo.webp",
    accentColor: "#38bdf8",
    expansion: "vanilla",
    client: "1.12.1",
    logging: "client",
    engine: "unknown",
    chronicleUrl: "https://octo.chronicleclassic.com",
    homepageUrl: "https://octowow.st/",
    status: ["custom-content"],
  },
];
