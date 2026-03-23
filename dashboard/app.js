const API_BASE = "http://localhost:8081/api/dashboard";

async function fetchJSON(path) {
  const res = await fetch(`${API_BASE}${path}`);
  if (!res.ok) {
    throw new Error(`HTTP ${res.status}`);
  }
  return res.json();
}

function downloadFile(path) {
  window.open(`${API_BASE}${path}`, "_blank");
}

function renderSummary(summary) {
  const container = document.getElementById("summaryCards");
  const items = [
    ["Published", summary.publishedCount],
    ["Rejected", summary.rejectedCount],
    ["Healthy Sources", summary.healthySources],
    ["Disabled Sources", summary.disabledSources],
    ["Tracked Sources", summary.trackedSourceSize],
  ];

  container.innerHTML = items
    .map(([label, value]) => `
      <div class="card">
        <div class="label">${label}</div>
        <div class="value">${value}</div>
      </div>
    `)
    .join("");
}

function renderSources(rows) {
  const tbody = document.querySelector("#sourcesTable tbody");
  tbody.innerHTML = rows.map(r => `
    <tr>
      <td>${r.sourceName}</td>
      <td>${r.category}</td>
      <td class="state-${r.state}">${r.state}</td>
      <td>${r.consecutiveFails}</td>
      <td>${r.lastErrorType || "-"}</td>
      <td>${r.disabledUntil || "-"}</td>
      <td>${r.lastSuccessAt || "-"}</td>
    </tr>
  `).join("");
}

function renderPublished(rows) {
  const tbody = document.querySelector("#publishedTable tbody");
  const ordered = [...rows].reverse().slice(0, 100);

  tbody.innerHTML = ordered.map(r => `
    <tr>
      <td>${formatTime(r.time)}</td>
      <td>${r.category}</td>
      <td>${r.source}</td>
      <td>${r.virality}</td>
      <td>${r.clusterCount}</td>
      <td class="title-cell">
        <a href="${r.link}" target="_blank" rel="noopener noreferrer">${r.title}</a>
      </td>
    </tr>
  `).join("");
}

function renderRejected(rows) {
  const tbody = document.querySelector("#rejectedTable tbody");
  const ordered = [...rows].reverse().slice(0, 100);

  tbody.innerHTML = ordered.map(r => `
    <tr>
      <td>${formatTime(r.time)}</td>
      <td>${r.category}</td>
      <td>${r.source}</td>
      <td>${r.reason}</td>
      <td class="title-cell">${r.title}</td>
    </tr>
  `).join("");
}

function formatTime(value) {
  if (!value) return "-";
  try {
    return new Date(value).toLocaleString("tr-TR");
  } catch {
    return value;
  }
}

async function refresh() {
  try {
    const [summary, sources, published, rejected] = await Promise.all([
      fetchJSON("/summary"),
      fetchJSON("/sources"),
      fetchJSON("/published"),
      fetchJSON("/rejected"),
    ]);

    renderSummary(summary);
    renderSources(sources);
    renderPublished(published);
    renderRejected(rejected);
  } catch (err) {
    console.error("Dashboard refresh error:", err);
  }
}

refresh();
setInterval(refresh, 5000);