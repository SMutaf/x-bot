import type { StreamEventName } from "./types";

export type StreamEventCallback = (
  eventName: StreamEventName,
  rawData: string
) => void;

export function createFeedStream(
  url: string,
  onEvent: StreamEventCallback,
  onOpen?: () => void,
  onError?: (message: string) => void
): EventSource {
  const source = new EventSource(url);

  source.onopen = () => {
    console.log("[SSE] connected:", url);
    onOpen?.();
  };

  source.onerror = () => {
    console.error("[SSE] connection error");
    onError?.("SSE connection error");
  };

  source.addEventListener("news.published", (event) => {
    onEvent("news.published", (event as MessageEvent).data);
  });

  source.addEventListener("heartbeat", (event) => {
    onEvent("heartbeat", (event as MessageEvent).data);
  });

  source.addEventListener("connected", (event) => {
    onEvent("connected", (event as MessageEvent).data);
  });

  source.addEventListener("news.updated", (event) => {
    onEvent("news.updated", (event as MessageEvent).data);
  });

  source.addEventListener("news.removed", (event) => {
    onEvent("news.removed", (event as MessageEvent).data);
  });

  return source;
}