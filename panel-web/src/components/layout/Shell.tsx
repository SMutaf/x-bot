import { useState } from "react";
import { env } from "../../lib/env";
import TopBar from "./TopBar";
import LeftRail from "./LeftRail";
import MainStream from "./MainStream";
import RightDetailPane from "./RightDetailPane";

export default function Shell() {
  const [activeViewId, setActiveViewId] = useState<string>(env.defaultView);

  return (
    <div className="app-shell">
      <TopBar />
      <div className="app-body">
        <LeftRail activeViewId={activeViewId} onChangeView={setActiveViewId} />
        <MainStream activeViewId={activeViewId} />
        <RightDetailPane />
      </div>
    </div>
  );
}