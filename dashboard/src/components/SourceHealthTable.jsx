import { formatDateTime, getHealthStateLabel } from "../lib/format.js";

export default function SourceHealthTable({ sources }) {
  return (
    <section className="dashboard-panel dashboard-panel--tight source-health-panel">
      <div className="dashboard-panel__header">
        <div>
          <div className="panel-title">Source Health</div>
          <p className="panel-copy">
            Kaynaklarin aktiflik ve hata durumlarini operasyon seviyesinde izler.
          </p>
        </div>
        <div className="panel-actions">
          <a className="panel-link" href="/api/dashboard/download/source-health" target="_blank" rel="noreferrer">
            Source Health JSONL
          </a>
        </div>
      </div>

      <div className="table-shell">
        <table className="ops-table ops-table--source-health">
          <colgroup>
            <col className="source-health-col source-health-col--source" />
            <col className="source-health-col source-health-col--category" />
            <col className="source-health-col source-health-col--state" />
            <col className="source-health-col source-health-col--fails" />
            <col className="source-health-col source-health-col--error" />
            <col className="source-health-col source-health-col--disabled" />
            <col className="source-health-col source-health-col--success" />
          </colgroup>
          <thead>
            <tr>
              <th>Source</th>
              <th title="Category">Cat</th>
              <th>State</th>
              <th>Fails</th>
              <th title="Last Error">Error</th>
              <th title="Disabled Until">Disabled</th>
              <th title="Last Success">Success</th>
            </tr>
          </thead>
          <tbody>
            {sources.length === 0 ? (
              <tr>
                <td colSpan="7" className="ops-table__empty">
                  Source health verisi henuz yok.
                </td>
              </tr>
            ) : (
              sources.map((source) => {
                const state = getHealthStateLabel(source);

                return (
                  <tr key={`${source.sourceName}-${source.url}`}>
                    <td>{source.sourceName}</td>
                    <td>{source.category}</td>
                    <td>
                      <span className={`ops-state ops-state--${state}`}>{state}</span>
                    </td>
                    <td>{source.consecutiveFails ?? 0}</td>
                    <td>{source.lastErrorType || "-"}</td>
                    <td>{formatDateTime(source.disabledUntil)}</td>
                    <td>{formatDateTime(source.lastSuccessAt)}</td>
                  </tr>
                );
              })
            )}
          </tbody>
        </table>
      </div>
    </section>
  );
}
