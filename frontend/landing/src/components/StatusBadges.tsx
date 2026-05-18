import type { StatusTag } from "../types";

const TAG_CONFIG: Record<StatusTag, { label: string; color: string }> = {
  closed: { label: "Closed", color: "bg-red-500/20 text-red-400 border-red-500/40" },
  beta: { label: "Beta", color: "bg-yellow-500/15 text-yellow-400 border-yellow-500/30" },
  new: { label: "New", color: "bg-green-500/15 text-green-400 border-green-500/30" },
  hardcore: { label: "Hardcore", color: "bg-red-500/15 text-red-400 border-red-500/30" },
  fresh: { label: "Fresh", color: "bg-cyan-500/15 text-cyan-400 border-cyan-500/30" },
  progression: { label: "Progression", color: "bg-purple-500/15 text-purple-400 border-purple-500/30" },
  "custom-content": { label: "Custom Content", color: "bg-emerald-500/15 text-emerald-400 border-emerald-500/30" },
};

export function StatusBadges({ tags }: { tags?: StatusTag[] }) {
  if (!tags?.length) return null;

  return (
    <div className="flex flex-wrap gap-1.5">
      {tags.map((tag) => {
        const config = TAG_CONFIG[tag];
        return (
          <span
            key={tag}
            className={`inline-flex items-center rounded-full border px-2 py-0.5 text-xs font-medium ${config.color}`}
          >
            {config.label}
          </span>
        );
      })}
    </div>
  );
}
