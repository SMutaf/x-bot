import { useDashboardData } from "../hooks/useDashboardData.jsx";
import { formatDateTime } from "../lib/format.js";
import TopBar from "./TopBar.jsx";
import SourceHealthTable from "./SourceHealthTable.jsx";
import EventTable from "./EventTable.jsx";
import { views } from "../config/views.js";

export default function MainStream({ activeView, onChangeView }) {
  const { status, summary, sources, published, rejected, connection } = useDashboardData(activeView);

  return (
    <main className="ops-page">
      <TopBar status={status} />

      <div className="ops-layout">
        <section className="dashboard-panel dashboard-panel--controls ops-layout__controls">
          <div className="dashboard-panel__header">
            <div>
              <div className="panel-title">Control Surface</div>
            </div>
          </div>

          <div className="control-grid">
            <div className="control-block">
              <div className="control-block__label">Active View</div>
              <div className="view-tabs">
                {views.map((view) => (
                  <button
                    key={view.id}
                    type="button"
                    className={`view-tab ${activeView.id === view.id ? "view-tab--active" : ""}`}
                    onClick={() => onChangeView(view.id)}
                  >
                    {view.label}
                  </button>
                ))}
              </div>
            </div>

            <div className="control-block">
              <div className="control-block__label">Live Feed</div>
              <div className="control-kpis">
                <span className="control-pill">State: {connection.isConnected ? "Canli" : "Kopuk"}</span>
                <span className="control-pill">Son event: {formatDateTime(connection.lastEventAt)}</span>
                <span className="control-pill">Son snapshot: {formatDateTime(connection.lastSnapshotAt)}</span>
              </div>
            </div>

            {summary ? (
              <div className="control-block">
                <div className="control-block__label">RSS Overview</div>
                <div className="control-kpis">
                  <span className="control-pill">Tracked: {summary.trackedSourceSize}</span>
                  <span className="control-pill">Healthy: {summary.healthySources}</span>
                  <span className="control-pill">Degraded: {summary.degradedSources}</span>
                  <span className="control-pill">Disabled: {summary.disabledSources}</span>
                </div>
              </div>
            ) : null}
          </div>
        </section>

        <section className="cockpit-grid">
          <SourceHealthTable sources={sources} />

          <EventTable
            title="Published News"
            subtitle="Yayina cikmis haberlerin operasyon listesi."
            items={published}
            downloadHref="/api/dashboard/download/published"
            panelClassName="dashboard-panel--tight"
            columns={[
              { key: "time", label: "Time" },
              { key: "category", label: "Category" },
              { key: "source", label: "Source" },
              { key: "virality", label: "Virality" },
              { key: "clusterCount", label: "Cluster" },
              { key: "title", label: "Title" }
            ]}
          />

          <EventTable
            title="Rejected News"
            subtitle="Filtrelenen veya reddedilen haberlerin operasyon listesi."
            items={rejected}
            downloadHref="/api/dashboard/download/rejected"
            panelClassName="dashboard-panel--tight"
            columns={[
              { key: "time", label: "Time" },
              { key: "category", label: "Category" },
              { key: "source", label: "Source" },
              { key: "reason", label: "Reason" },
              { key: "title", label: "Title" }
            ]}
          />
        </section>
      </div>
    </main>
  );
}
