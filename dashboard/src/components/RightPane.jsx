import { formatDateTime, getHealthStateLabel } from "../lib/format.js";

export default function RightPane({ selectedItem }) {
  return (
    <aside className="right-detail-pane">
      <section className="detail-card">
        <div className="panel-title">Selected Story</div>
        {selectedItem ? (
          <>
            <h2 className="detail-card__title">{selectedItem.title}</h2>
            <p className="detail-card__text">{selectedItem.summary || selectedItem.description || "Detay yok."}</p>
            <DetailRow label="Hook" value={selectedItem.hook || "-"} />
            <DetailRow label="Importance" value={selectedItem.importance || "-"} />
            <DetailRow label="Source" value={selectedItem.source || "-"} />
            <DetailRow label="Category" value={selectedItem.category || "-"} />
            <DetailRow label="Virality" value={String(selectedItem.virality ?? "-")} />
            <DetailRow label="Cluster" value={String(selectedItem.clusterCount ?? "-")} />
            <DetailRow label="Time" value={formatDateTime(selectedItem.time)} />
            <a className="detail-card__link" href={selectedItem.link} target="_blank" rel="noreferrer">
              Habere git
            </a>
          </>
        ) : (
          <p className="detail-card__text">Listeden bir haber secince detay burada acilir.</p>
        )}
      </section>

      <section className="detail-card">
        <div className="panel-title">Health Legend</div>
        <LegendItem state={getHealthStateLabel({ consecutiveFails: 0, disabledUntil: "" })} />
        <LegendItem state={getHealthStateLabel({ consecutiveFails: 2, disabledUntil: "" })} />
        <LegendItem state={getHealthStateLabel({ consecutiveFails: 2, disabledUntil: new Date(Date.now() + 60000).toISOString() })} />
      </section>
    </aside>
  );
}

function DetailRow({ label, value }) {
  return (
    <>
      <div className="detail-card__label">{label}</div>
      <p className="detail-card__value">{value}</p>
    </>
  );
}

function LegendItem({ state }) {
  return (
    <div className="legend-item">
      <span className={`ops-state ops-state--${state}`}>{state}</span>
    </div>
  );
}
