export type StreamEventName =
  | "news.published"
  | "heartbeat"
  | "connected"
  | "news.updated"
  | "news.removed";

export interface StreamConnectionState {
  isConnected: boolean;
  lastEventAt: string | null;
  lastError: string | null;
}