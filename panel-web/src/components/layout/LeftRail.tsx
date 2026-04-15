import { PRESET_VIEWS } from "../../features/views/presets";

type LeftRailProps = {
  activeViewId: string;
  onChangeView: (viewId: string) => void;
};

export default function LeftRail(props: LeftRailProps) {
  const { activeViewId, onChangeView } = props;

  return (
    <aside className="left-rail">
      <div className="panel-title">Hazır Akışlar</div>

      <ul className="view-list">
        {PRESET_VIEWS.map((view) => {
          const isActive = view.id === activeViewId;

          return (
            <li key={view.id}>
              <button
                type="button"
                className={`view-item ${isActive ? "view-item--active" : ""}`}
                onClick={() => onChangeView(view.id)}
              >
                <div className="view-item__label">{view.label}</div>
                <div className="view-item__description">{view.description}</div>
              </button>
            </li>
          );
        })}
      </ul>
    </aside>
  );
}