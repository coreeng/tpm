import type { PreviewPayload } from "./types";

export async function fetchPreview(): Promise<PreviewPayload> {
  const response = await fetch("/api/preview", { cache: "no-store" });
  if (!response.ok) {
    const text = await response.text();
    throw new Error(text.trim() || `Preview request failed with ${response.status}`);
  }
  return (await response.json()) as PreviewPayload;
}
