import { formatDateTime } from "../lib/format.js";

export default function EventTable({
  title,
  subtitle,
  items,
  columns,
  downloadHref,
  panelClassName = ""
}) {
  return (
    <section className={`dashboard-panel dashboard-panel--tight event-panel ${panelClassName}`.trim()}>
      <div className="dashboard-panel__header">
        <div>
          <div className="panel-title">{title}</div>
          <p className="panel-copy">{subtitle}</p>
        </div>
        <div className="panel-actions">
          <a className="panel-link" href={downloadHref} target="_blank" rel="noreferrer">
            JSONL indir
          </a>
        </div>
      </div>

      <div className="table-shell">
        <table className="ops-table">
          <thead>
            <tr>
              {columns.map((column) => (
                <th key={column.key}>{column.label}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {items.length === 0 ? (
              <tr>
                <td colSpan={columns.length} className="ops-table__empty">
                  Kayit yok.
                </td>
              </tr>
            ) : (
              items.map((item) => (
                <tr key={item.link || `${item.title}-${item.time}`}>
                  {columns.map((column) => (
                    <td key={column.key}>{renderCell(item, column.key)}</td>
                  ))}
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </section>
  );
}

function renderCell(item, key) {
  switch (key) {
    case "time":
      return formatDateTime(item.time);
    case "title":
      if (item.link) {
        return (
          <a href={item.link} target="_blank" rel="noreferrer" className="table-link">
            {item.title}
          </a>
        );
      }
      return item.title;
    case "virality":
      return item.virality ?? "-";
    case "clusterCount":
      return item.clusterCount ?? "-";
    default:
      return item[key] || "-";
  }
}
