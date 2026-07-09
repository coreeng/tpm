import { AlertTriangle, RefreshCw } from "lucide-react";
import { StrictMode, useCallback, useEffect, useState } from "react";
import { createRoot } from "react-dom/client";
import { fetchPreview } from "./api";
import "./index.css";
import { LabPreview } from "./lab/LabPreview";
import { ModulePreview } from "./module/ModulePreview";
import type { PreviewPayload } from "./types";

function App() {
  const [data, setData] = useState<PreviewPayload | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const load = useCallback(async ({ background = false }: { background?: boolean } = {}) => {
    if (!background) {
      setLoading(true);
    }
    setError(null);
    try {
      setData(await fetchPreview());
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      if (!background) {
        setLoading(false);
      }
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  useEffect(() => {
    const events = new EventSource("/api/events");
    events.addEventListener("preview-changed", () => {
      void load({ background: true });
    });
    events.addEventListener("preview-error", (event) => {
      try {
        const payload = JSON.parse(event instanceof MessageEvent ? event.data : "{}") as { error?: string };
        if (payload.error) {
          setError(payload.error);
        }
      } catch {
        setError("Preview watcher reported an unreadable error");
      }
    });
    return () => {
      events.close();
    };
  }, [load]);

  useEffect(() => {
    if (data?.kind !== "lab") {
      return;
    }
    const interval = window.setInterval(() => {
      void load({ background: true });
    }, 5000);
    return () => {
      window.clearInterval(interval);
    };
  }, [data?.kind, load]);

  if (loading && !data) {
    return (
      <main className="grid min-h-screen place-items-center bg-slate-50 px-6">
        <div className="rounded-lg border border-slate-200 bg-white px-5 py-4 text-base text-slate-700 shadow-sm sm:text-sm">
          Loading preview...
        </div>
      </main>
    );
  }

  if (error && !data) {
    return <ErrorState error={error} onRetry={() => void load()} />;
  }

  return (
    <>
      {error ? <RefreshError error={error} onRetry={() => void load()} /> : null}
      {data?.kind === "module" ? <ModulePreview data={data} /> : null}
      {data?.kind === "lab" ? <LabPreview data={data} /> : null}
    </>
  );
}

function ErrorState({ error, onRetry }: { error: string; onRetry: () => void }) {
  return (
    <main className="grid min-h-screen place-items-center bg-slate-50 px-6">
      <div className="max-w-xl rounded-lg border border-red-200 bg-white p-6 shadow-sm">
        <div className="flex items-center gap-2 text-lg font-semibold text-red-800">
          <AlertTriangle className="size-5" />
          Preview failed
        </div>
        <pre className="mt-3 rounded-md bg-red-50 p-3 text-sm leading-6 text-red-900">{error}</pre>
        <button
          className="mt-4 inline-flex items-center gap-2 rounded-md bg-red-700 px-3 py-2 text-sm font-semibold text-white hover:bg-red-800"
          onClick={onRetry}
          type="button"
        >
          <RefreshCw className="size-4" />
          Retry
        </button>
      </div>
    </main>
  );
}

function RefreshError({ error, onRetry }: { error: string; onRetry: () => void }) {
  return (
    <div className="fixed bottom-4 right-4 z-50 max-w-md rounded-lg border border-red-200 bg-white p-4 text-sm shadow-xl">
      <div className="flex items-start gap-3">
        <AlertTriangle className="mt-0.5 size-4 text-red-700" />
        <div className="min-w-0">
          <div className="font-semibold text-red-800">Refresh failed</div>
          <div className="mt-1 line-clamp-3 text-red-900">{error}</div>
          <button className="mt-2 text-sm font-semibold text-red-700 hover:text-red-800" onClick={onRetry} type="button">
            Retry
          </button>
        </div>
      </div>
    </div>
  );
}

const root = document.getElementById("root");
if (!root) {
  throw new Error("Root element not found");
}

createRoot(root).render(
  <StrictMode>
    <App />
  </StrictMode>,
);
