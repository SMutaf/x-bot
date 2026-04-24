import { useEffect, useMemo, useState } from "react";
import { buildDashboardPath, buildFeedStreamPath, fetchJson } from "../lib/api.js";

const PUBLISHED_LIMIT = 500;
const REJECTED_LIMIT = 300;

export function useDashboardData(activeView) {
  const [status, setStatus] = useState(null);
  const [summary, setSummary] = useState(null);
  const [sources, setSources] = useState([]);
  const [published, setPublished] = useState([]);
  const [rejected, setRejected] = useState([]);
  const [connection, setConnection] = useState({
    isConnected: false,
    lastEventAt: null,
    lastSnapshotAt: null,
    lastError: null
  });

  const feedView = activeView.feedView;

  useEffect(() => {
    let cancelled = false;

    async function loadSnapshot() {
      try {
        const publishedPath = buildDashboardPath(
          `/published?limit=${PUBLISHED_LIMIT}${feedView ? `&view=${encodeURIComponent(feedView)}` : ""}`
        );

        const [nextStatus, nextSummary, nextSources, nextPublished, nextRejected] = await Promise.all([
          fetchJson(buildDashboardPath("/status")),
          fetchJson(buildDashboardPath("/summary")),
          fetchJson(buildDashboardPath("/sources")),
          fetchJson(publishedPath),
          fetchJson(buildDashboardPath(`/rejected?limit=${REJECTED_LIMIT}`))
        ]);

        if (cancelled) {
          return;
        }

        setStatus(nextStatus);
        setSummary(nextSummary);
        setSources(nextSources);
        setPublished(nextPublished);
        setRejected(nextRejected);
        setConnection((prev) => ({
          ...prev,
          lastSnapshotAt: new Date().toISOString(),
          lastError: null
        }));
      } catch (error) {
        if (cancelled) {
          return;
        }

        setConnection((prev) => ({
          ...prev,
          lastError: error instanceof Error ? error.message : String(error)
        }));
      }
    }

    setPublished([]);
    loadSnapshot();

    return () => {
      cancelled = true;
    };
  }, [feedView]);

  useEffect(() => {
    const source = new EventSource(buildFeedStreamPath(feedView));

    source.onopen = () => {
      setConnection((prev) => ({
        ...prev,
        isConnected: true,
        lastError: null
      }));
    };

    source.onerror = () => {
      setConnection((prev) => ({
        ...prev,
        isConnected: false,
        lastError: "Canli akis gecici olarak kesildi"
      }));
    };

    source.addEventListener("news.published", (event) => {
      try {
        const item = JSON.parse(event.data);
        setPublished((current) => mergePublished(current, item, PUBLISHED_LIMIT));
        setConnection((prev) => ({
          ...prev,
          isConnected: true,
          lastEventAt: new Date().toISOString(),
          lastError: null
        }));
      } catch {
        setConnection((prev) => ({
          ...prev,
          lastError: "Canli akis verisi parse edilemedi"
        }));
      }
    });

    return () => source.close();
  }, [feedView]);

  useEffect(() => {
    const summaryTimer = window.setInterval(async () => {
      try {
        const [nextStatus, nextSummary] = await Promise.all([
          fetchJson(buildDashboardPath("/status")),
          fetchJson(buildDashboardPath("/summary"))
        ]);
        setStatus(nextStatus);
        setSummary(nextSummary);
        setConnection((prev) => ({
          ...prev,
          lastSnapshotAt: new Date().toISOString()
        }));
      } catch (error) {
        setConnection((prev) => ({
          ...prev,
          lastError: error instanceof Error ? error.message : String(error)
        }));
      }
    }, 10000);

    const sourcesTimer = window.setInterval(async () => {
      try {
        const nextSources = await fetchJson(buildDashboardPath("/sources"));
        setSources(nextSources);
      } catch (error) {
        setConnection((prev) => ({
          ...prev,
          lastError: error instanceof Error ? error.message : String(error)
        }));
      }
    }, 15000);

    const rejectedTimer = window.setInterval(async () => {
      try {
        const nextRejected = await fetchJson(buildDashboardPath(`/rejected?limit=${REJECTED_LIMIT}`));
        setRejected(nextRejected);
      } catch (error) {
        setConnection((prev) => ({
          ...prev,
          lastError: error instanceof Error ? error.message : String(error)
        }));
      }
    }, 20000);

    return () => {
      window.clearInterval(summaryTimer);
      window.clearInterval(sourcesTimer);
      window.clearInterval(rejectedTimer);
    };
  }, []);

  const latestItem = useMemo(() => published[0] ?? null, [published]);

  return {
    status,
    summary,
    sources,
    published,
    rejected,
    latestItem,
    connection
  };
}

function mergePublished(current, incoming, limit) {
  const map = new Map();

  for (const item of [incoming, ...current]) {
    const key = item.link || `${item.title}-${item.time}`;
    map.set(key, item);
  }

  return Array.from(map.values())
    .sort((a, b) => new Date(b.time).getTime() - new Date(a.time).getTime())
    .slice(0, limit);
}
