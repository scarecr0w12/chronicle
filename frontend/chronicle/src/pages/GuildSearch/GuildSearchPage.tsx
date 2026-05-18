import { Users } from "lucide-react";
import { GuildSearchContent } from "./GuildSearchContent";

export function GuildSearchPage() {
  return (
    <div className="max-w-5xl mx-auto p-4 md:p-8">
      <div className="mb-6">
        <h1 className="text-2xl font-bold flex items-center gap-2">
          <Users className="h-6 w-6 text-primary" />
          Guilds
        </h1>
        <p className="text-sm text-muted-foreground mt-1">
          Browse guilds seen in uploaded combat logs.
        </p>
      </div>
      <GuildSearchContent />
    </div>
  );
}
