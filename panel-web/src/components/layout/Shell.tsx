import { useState } from "react";
import { env } from "../../lib/env";
import type { FeedItem } from "../../features/feed/types";
import TopBar from "./TopBar";
import LeftRail from "./LeftRail";
import MainStream from "./MainStream";
import RightDetailPane from "./RightDetailPane";

export default function Shell() {
  const [activeViewId, setActiveViewId] = useState<string>(env.defaultView);
  const [selectedItem, setSelectedItem] = useState<FeedItem | null>(null);

  return (
    <div className="app-shell">
      <TopBar />
      <div className="app-body">
        <LeftRail activeViewId={activeViewId} onChangeView={setActiveViewId} />
        <MainStream
          activeViewId={activeViewId}
          selectedItem={selectedItem}
          onSelectItem={setSelectedItem}
        />
        <RightDetailPane selectedItem={selectedItem} />
      </div>
    </div>
  );
}