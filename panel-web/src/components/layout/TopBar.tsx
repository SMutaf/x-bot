import { APP_NAME } from "../../app/config";

type TopBarProps = {
  searchQuery: string;
  onSearchChange: (value: string) => void;
};

export default function TopBar(props: TopBarProps) {
  const { searchQuery, onSearchChange } = props;

  return (
    <header className="topbar">
      <div className="topbar__brand">{APP_NAME}</div>

      <div className="topbar__center">
        <input
          className="topbar__search"
          type="text"
          placeholder="Ara..."
          value={searchQuery}
          onChange={(event) => onSearchChange(event.target.value)}
        />
      </div>
    </header>
  );
}
