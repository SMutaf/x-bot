import { useEffect } from "react";
import type { FeedItem } from "../../features/feed/types";
import { useFeedStream } from "../../hooks/useFeedStream";

type MainStreamProps = {
  activeViewId: string;
  selectedItem: FeedItem | null;
  onSelectItem: (item: FeedItem) => void;
};

export default function MainStream(props: MainStreamProps) {
  const { activeViewId, selectedItem, onSelectItem } = props;
  const { connection, latestEventName, latestRawData, items, isInitialLoading } =
    useFeedStream(activeViewId);

  useEffect(() => {
    if (!selectedItem && items.length > 0) {
      onSelectItem(items[0]);
    }
  }, [items, selectedItem, onSelectItem]);

  return (
    <main className="main-stream">
      <div className="main-stream__header">
        <div>
          <h1 className="main-stream__title">Canlı Akış</h1>
          <p className="main-stream__subtitle">Aktif görünüm: {activeViewId}</p>
        </div>

        <div className="live-status">
          <span
            className={`live-dot ${connection.isConnected ? "live-dot--ok" : "live-dot--off"}`}
          />
          <span>{connection.isConnected ? "Bağlı" : "Bağlantı Yok"}</span>
        </div>
      </div>

      <section className="stream-debug-card">
        <div className="stream-debug-row">
          <strong>Son event:</strong> {latestEventName ?? "Henüz event yok"}
        </div>
        <div className="stream-debug-row">
          <strong>Son zaman:</strong> {connection.lastEventAt ?? "-"}
        </div>
        <div className="stream-debug-row">
          <strong>Hata:</strong> {connection.lastError ?? "-"}
        </div>
      </section>

      <section className="stream-preview">
        <div className="panel-title">Ham Event Önizleme</div>
        <pre className="code-block">{latestRawData ?? "Yeni event bekleniyor..."}</pre>
      </section>

      <section className="stream-placeholder-card">
        <div className="panel-title">Feed Listesi</div>

        {isInitialLoading ? (
          <p>İlk veri yükleniyor...</p>
        ) : items.length === 0 ? (
          <p>Bu görünüm için henüz kayıt yok.</p>
        ) : (
          <div className="feed-list">
            {items.map((item) => {
              const key = item.link || `${item.title}-${item.time}`;
              const isSelected =
                !!selectedItem &&
                (selectedItem.link === item.link ||
                  (selectedItem.title === item.title && selectedItem.time === item.time));

              return (
                <article
                  key={key}
                  className={`feed-card ${isSelected ? "feed-card--selected" : ""}`}
                  onClick={() => onSelectItem(item)}
                  role="button"
                  tabIndex={0}
                  onKeyDown={(event) => {
                    if (event.key === "Enter" || event.key === " ") {
                      event.preventDefault();
                      onSelectItem(item);
                    }
                  }}
                >
                  <div className="feed-card__meta">
                    <span>{item.category}</span>
                    <span>•</span>
                    <span>{item.source}</span>
                  </div>

                  <h3 className="feed-card__title">{item.title}</h3>

                  <div className="feed-card__stats">
                    <span>Virality: {item.virality ?? "-"}</span>
                    <span>Cluster: {item.clusterCount ?? "-"}</span>
                    <span>
                      Saat: {item.time ? new Date(item.time).toLocaleString("tr-TR") : "-"}
                    </span>
                  </div>

                  <a
                    className="feed-card__link"
                    href={item.link}
                    target="_blank"
                    rel="noreferrer"
                    onClick={(e) => e.stopPropagation()}
                  >
                    Haberi aç
                  </a>
                </article>
              );
            })}
          </div>
        )}
      </section>
    </main>
  );
}