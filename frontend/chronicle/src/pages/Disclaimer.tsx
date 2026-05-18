export function Disclaimer() {
  return (
    <div className="container mx-auto px-4 py-8 max-w-3xl">
      <h1 className="text-3xl font-bold mb-6">Disclaimer</h1>
      <div className="space-y-4 text-muted-foreground">
        <p>
          Chronicle is an independent, community-built project created by players.
        </p>
        <p>
          We are not affiliated with, endorsed by, or associated with Blizzard
          Entertainment. World of Warcraft® and related trademarks
          are the property of Blizzard Entertainment.
        </p>
        <p>
          Chronicle analyzes player-generated combat logs and does not modify
          gameplay or interact with game servers.
        </p>
        <p>
          Donations support development and hosting costs only and do not grant
          special access or in-game advantages.
        </p>
      </div>
    </div>
  );
}
