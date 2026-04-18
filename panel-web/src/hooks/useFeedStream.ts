import { useEffect, useMemo, useState } from "react";
import { createFeedStream } from "../features/stream/sseClient";
import type { StreamConnectionState, StreamEventName } from "../features/stream/types";
import type { FeedItem } from "../features/feed/types";
import { fetchFeedSnapshot } from "../features/feed/api";
import { env } from "../lib/env";

type UseFeedStreamResult = {
  connection: StreamConnectionState;
  latestEventName: StreamEventName | null;
  latestRawData: string | null;
  items: FeedItem[];
  isInitialLoading: boolean;
};

function mergeFeedItems(current: FeedItem[], incoming: FeedItem[]): FeedItem[] {
  const map = new Map<string, FeedItem>();

  for (const item of current) {
    map.set(item.link || `${item.title}-${item.time}`, item);
  }

  for (const item of incoming) {
    map.set(item.link || `${item.title}-${item.time}`, item);
  }

  return Array.from(map.values()).sort((a, b) => {
    const aTime = new Date(a.time).getTime();
    const bTime = new Date(b.time).getTime();
    return bTime - aTime;
  });
}

export function useFeedStream(viewId: string): UseFeedStreamResult {
  const [connection, setConnection] = useState<StreamConnectionState>({
    isConnected: false,
    lastEventAt: null,
    lastError: null
  });

  const [latestEventName, setLatestEventName] = useState<StreamEventName | null>(null);
  const [latestRawData, setLatestRawData] = useState<string | null>(null);
  const [items, setItems] = useState<FeedItem[]>([]);
  const [isInitialLoading, setIsInitialLoading] = useState<boolean>(true);

  const streamUrl = useMemo(() => {
    const base = env.apiBaseUrl.replace(/\/$/, "");
    return `${base}/api/feed/stream?view=${encodeURIComponent(viewId)}`;
  }, [viewId]);

  useEffect(() => {
    setItems([]);
    setLatestEventName(null);
    setLatestRawData(null);
    setIsInitialLoading(true);

    let isCancelled = false;

    const source = createFeedStream(
      streamUrl,
      (eventName, rawData) => {
        console.log("[SSE EVENT]", eventName, rawData);

        setLatestEventName(eventName);
        setLatestRawData(rawData);
        setConnection((prev) => ({
          ...prev,
          lastEventAt: new Date().toISOString(),
          lastError: null
        }));

        if (eventName === "news.published") {
          try {
            const item = JSON.parse(rawData) as FeedItem;

            if (!isCancelled) {
              setItems((prev) => mergeFeedItems(prev, [item]));
            }
          } catch (error) {
            console.error("[SSE PARSE ERROR]", error);
          }
        }
      },
      () => {
        setConnection({
          isConnected: true,
          lastEventAt: new Date().toISOString(),
          lastError: null
        });
      },
      (message) => {
        setConnection((prev) => ({
          ...prev,
          isConnected: false,
          lastError: message
        }));
      }
    );

    async function loadSnapshot() {
      try {
        const snapshot = await fetchFeedSnapshot(env.apiBaseUrl, viewId, 50);

        if (!isCancelled) {
          setItems((prev) => mergeFeedItems(prev, snapshot));
        }
      } catch (error) {
        console.error("[SNAPSHOT ERROR]", error);
      } finally {
        if (!isCancelled) {
          setIsInitialLoading(false);
        }
      }
    }

    loadSnapshot();

    return () => {
      isCancelled = true;
      source.close();
      setConnection((prev) => ({
        ...prev,
        isConnected: false
      }));
    };
  }, [streamUrl, viewId]);

  return {
    connection,
    latestEventName,
    latestRawData,
    items,
    isInitialLoading
  };
}