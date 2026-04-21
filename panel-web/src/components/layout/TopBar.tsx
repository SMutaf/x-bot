import { APP_NAME } from "../../app/config";

export default function TopBar() {
  return (
    <header className="topbar">
      <div className="topbar__brand">{APP_NAME}</div>

      <div className="topbar__center">
        <input
          className="topbar__search"
          type="text"
          placeholder="Ara..."
          disabled
        />
      </div>
    
    </header>
  );
}