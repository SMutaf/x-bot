export async function fetchJson(path) {
  const response = await fetch(path);
  if (!response.ok) {
    throw new Error(`HTTP ${response.status}`);
  }

  return response.json();
}

export function buildDashboardPath(path) {
  return `/api/dashboard${path}`;
}

export function buildFeedStreamPath(viewId) {
  const query = viewId ? `?view=${encodeURIComponent(viewId)}` : "";
  return `/api/feed/stream${query}`;
}
