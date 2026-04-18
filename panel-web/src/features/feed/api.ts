import type { FeedItem } from "./types";

export async function fetchFeedSnapshot(
  apiBaseUrl: string,
  viewId: string,
  limit = 50
): Promise<FeedItem[]> {
  const base = apiBaseUrl.replace(/\/$/, "");
  const url = `${base}/api/feed?view=${encodeURIComponent(viewId)}&limit=${limit}`;

  const response = await fetch(url);
  if (!response.ok) {
    throw new Error(`Snapshot request failed: ${response.status}`);
  }

  return (await response.json()) as FeedItem[];
}