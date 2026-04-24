import { useEffect, useState } from "react";
import { env } from "../../lib/env";
import type { FeedItem } from "../../features/feed/types";
import TopBar from "./TopBar";
import LeftRail from "./LeftRail";
import MainStream from "./MainStream";
import RightDetailPane from "./RightDetailPane";

export default function Shell() {
  const [activeViewId, setActiveViewId] = useState<string>(env.defaultView);
  const [selectedItem, setSelectedItem] = useState<FeedItem | null>(null);
  const [searchQuery, setSearchQuery] = useState<string>("");

  useEffect(() => {
    setSelectedItem(null);
    setSearchQuery("");
  }, [activeViewId]);

  return (
    <div className="app-shell">
      <TopBar searchQuery={searchQuery} onSearchChange={setSearchQuery} />
      <div className="app-body">
        <LeftRail activeViewId={activeViewId} onChangeView={setActiveViewId} />
        <MainStream
          activeViewId={activeViewId}
          searchQuery={searchQuery}
          selectedItem={selectedItem}
          onSelectItem={setSelectedItem}
        />
        <RightDetailPane selectedItem={selectedItem} />
      </div>
    </div>
  );
}
