import { views } from "../config/views.js";

export default function LeftRail({ activeViewId, onChangeView }) {
  return (
    <aside className="left-rail">
      <div className="panel-title">Views</div>
      <div className="view-list">
        {views.map((view) => (
          <button
            key={view.id}
            type="button"
            className={`view-item ${activeViewId === view.id ? "view-item--active" : ""}`}
            onClick={() => onChangeView(view.id)}
          >
            <div className="view-item__label">{view.label}</div>
            <div className="view-item__description">{view.description}</div>
          </button>
        ))}
      </div>
    </aside>
  );
}
