import { useEffect, useMemo, useState } from "react";
import { createFeedStream } from "../features/stream/sseClient";
import type { StreamConnectionState, StreamEventName } from "../features/stream/types";
import { env } from "../lib/env";

type UseFeedStreamResult = {
  connection: StreamConnectionState;
  latestEventName: StreamEventName | null;
  latestRawData: string | null;
};

export function useFeedStream(viewId: string): UseFeedStreamResult {
  const [connection, setConnection] = useState<StreamConnectionState>({
    isConnected: false,
    lastEventAt: null,
    lastError: null
  });

  const [latestEventName, setLatestEventName] = useState<StreamEventName | null>(null);
  const [latestRawData, setLatestRawData] = useState<string | null>(null);

  const streamUrl = useMemo(() => {
    const base = env.apiBaseUrl.replace(/\/$/, "");
    return `${base}/api/feed/stream?view=${encodeURIComponent(viewId)}`;
  }, [viewId]);

  useEffect(() => {
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

    return () => {
      source.close();
      setConnection((prev) => ({
        ...prev,
        isConnected: false
      }));
    };
  }, [streamUrl]);

  return {
    connection,
    latestEventName,
    latestRawData
  };
}