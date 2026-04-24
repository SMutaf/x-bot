import { formatDateTime } from "../lib/format.js";

export default function TopBar({ status }) {
  const pythonState = status?.pythonState || (status?.pythonConnected ? "online" : "offline");
  const pythonValue = pythonState === "busy"
    ? `Busy${status?.pythonInFlight ? ` (${status.pythonInFlight})` : ""}`
    : pythonState === "online"
      ? "Online"
      : "Offline";
  const pythonTone = pythonState === "busy" ? "warn" : status?.pythonConnected ? "ok" : "off";

  const items = [
    { label: "Published", value: status?.publishedCount ?? "-" },
    { label: "Rejected", value: status?.rejectedCount ?? "-" },
    { label: "Healthy RSS", value: status?.healthyRssSources ?? "-" },
    { label: "Unhealthy RSS", value: status?.unhealthyRssSources ?? "-" },
    { label: "Redis", value: status?.redisConnected ? "Online" : "Offline", tone: status?.redisConnected ? "ok" : "off" },
    { label: "Python", value: pythonValue, tone: pythonTone }
  ];

  return (
    <header className="topbar topbar--ops">
      <div className="topbar__main">
        <div className="topbar__brand">
          <div className="topbar__logo">x</div>
          <div>
            <div className="topbar__title">x-bot Dashboard</div>
            <div className="topbar__caption">Operations Console</div>
          </div>
        </div>
        <div className="topbar__center">
          <div className="topbar__hint">
            Son publish: {formatDateTime(status?.lastPublishedAt)} • Son reject: {formatDateTime(status?.lastRejectedAt)}
          </div>
        </div>
      </div>
      <div className="status-strip">
        {items.map((item) => (
          <article key={item.label} className={`status-tile ${item.tone ? `status-tile--${item.tone}` : ""}`}>
            <div className="status-tile__label">{item.label}</div>
            <div className="status-tile__value">{item.value}</div>
          </article>
        ))}
      </div>
    </header>
  );
}
