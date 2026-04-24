import { useMemo, useState } from "react";
import TopBar from "./components/TopBar.jsx";
import MainStream from "./components/MainStream.jsx";
import { views } from "./config/views.js";

export default function App() {
  const [activeViewId, setActiveViewId] = useState("all");

  const activeView = useMemo(
    () => views.find((view) => view.id === activeViewId) ?? views[0],
    [activeViewId]
  );

  return (
    <div className="app-shell">
      <MainStream activeView={activeView} onChangeView={setActiveViewId} topBar={TopBar} />
    </div>
  );
}
