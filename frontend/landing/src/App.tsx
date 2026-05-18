import { ServerGrid } from "./components/ServerGrid";
import { Footer } from "./components/Footer";
import { SERVERS } from "./data/servers";

export function App() {
  return (
    <div className="flex min-h-dvh flex-col">
      <main className="flex-1">
        <ServerGrid servers={SERVERS} />
      </main>
      <Footer />
    </div>
  );
}
