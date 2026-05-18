import { useRef, useEffect, useState } from "react";
import { ModelViewer, createCdnResolver } from "classic-wow-model-viewer";
import type { PlayerOutfit } from "@/api/typesGenerated";
import { resolveEquipment } from "./modelEquipment";

const CDN_BASE = "https://models.chronicleclassic.com";

/** Map ArmoryPlayer.race → model-viewer race slug */
function mapRace(race: string): string {
  const map: Record<string, string> = {
    Human: "human",
    Orc: "orc",
    Dwarf: "dwarf",
    NightElf: "nightelf",
    Scourge: "scourge",
    Tauren: "tauren",
    Gnome: "gnome",
    Troll: "troll",
    BloodElf: "bloodelf",
    Draenei: "draenei",
    HighElf: "highelf",
  };
  return map[race] ?? race.toLowerCase();
}

interface CharacterModelViewerProps {
  race: string;
  gender: string;
  gear: PlayerOutfit;
  /** Called with outfit indices that failed to load display data. */
  onSlotErrors?: (failedSlots: Set<number>) => void;
}

export function CharacterModelViewer({ race, gender, gear, onSlotErrors }: CharacterModelViewerProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const viewerRef = useRef<ModelViewer | null>(null);
  const [status, setStatus] = useState<"loading" | "ready" | "error">("loading");

  // Resolve equipment first, then load character + equip in one go.
  // This avoids the camera reset that happens when equip() reloads the model.
  useEffect(() => {
    let cancelled = false;

    (async () => {
      try {
        // Fetch all item display data before touching the viewer
        const { equipment, pathToSlot, failedSlots } = await resolveEquipment(gear);
        if (cancelled) return;

        // Intercept CDN fetches to detect 404s on model/texture assets
        const cdnErrors = new Set<number>();
        const origFetch = window.fetch;
        window.fetch = (input: RequestInfo | URL, init?: RequestInit) => {
          const url = typeof input === "string" ? input : input instanceof Request ? input.url : String(input);
          if (url.startsWith(CDN_BASE)) {
            return origFetch.call(window, input, init).then((resp) => {
              if (!resp.ok) {
                // Map the failed CDN URL back to an outfit slot via path prefixes
                const rel = url.slice(CDN_BASE.length);
                for (const [prefix, idx] of pathToSlot) {
                  if (rel.startsWith(prefix)) {
                    cdnErrors.add(idx);
                    break;
                  }
                }
              }
              return resp;
            });
          }
          return origFetch.call(window, input, init);
        };

        try {
          // Now create or reuse the viewer and load everything at once
          if (!containerRef.current) return;

          if (viewerRef.current) {
            viewerRef.current.dispose();
          }
          const viewer = new ModelViewer({
            container: containerRef.current,
            assets: createCdnResolver(CDN_BASE),
            backgroundColor: 0x1a1a1a,
          });
          viewerRef.current = viewer;

          const slug = mapRace(race);
          const g = gender.toLowerCase() as "male" | "female";
          await viewer.loadCharacter(slug, g);
          if (cancelled) return;

          await viewer.equip(equipment);
          if (cancelled) return;
        } finally {
          window.fetch = origFetch;
        }

        // Merge API failures + CDN failures
        const allFailed = new Set([...failedSlots, ...cdnErrors]);
        onSlotErrors?.(allFailed);

        setStatus("ready");
      } catch {
        if (!cancelled) {
          setStatus("error");
        }
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [race, gender, gear, onSlotErrors]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      viewerRef.current?.dispose();
      viewerRef.current = null;
    };
  }, []);

  if (status === "error") {
    return <p className="text-xs text-zinc-600 italic">Model unavailable</p>;
  }

  return (
    <div className="w-full h-full relative">
      {status === "loading" && (
        <div className="absolute inset-0 flex items-center justify-center">
          <p className="text-xs text-zinc-500">Loading model…</p>
        </div>
      )}
      <div ref={containerRef} className={`w-full h-full ${status === "loading" ? "invisible" : ""}`} />
    </div>
  );
}
