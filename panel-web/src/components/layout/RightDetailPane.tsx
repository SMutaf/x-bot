import type { FeedItem } from "../../features/feed/types";

type RightDetailPaneProps = {
  selectedItem: FeedItem | null;
};

export default function RightDetailPane(props: RightDetailPaneProps) {
  const { selectedItem } = props;

  if (!selectedItem) {
    return (
      <aside className="right-detail-pane">
        <div className="panel-title">Detay</div>

        <div className="detail-card">
          <h3 className="detail-card__title">Secili Haber</h3>
          <p className="detail-card__text">
            Stream listesinden bir karta tiklaninca detay burada acilacak.
          </p>
        </div>
      </aside>
    );
  }

  const displayTitle = selectedItem.hook || selectedItem.title;
  const displaySummary = selectedItem.summary;
  const rawDescription = selectedItem.description;
  const translatedDescription = selectedItem.descriptionTr;
  const displayImportance = selectedItem.importance;

  return (
    <aside className="right-detail-pane">
      <div className="panel-title">Detay</div>

      <div className="detail-card">
        <h3 className="detail-card__title">Secili Haber</h3>
        <p className="detail-card__text detail-card__label">Baslik</p>
        <p className="detail-card__value">{displayTitle}</p>

        <p className="detail-card__text detail-card__label">Kaynak</p>
        <p className="detail-card__value">{selectedItem.source}</p>

        <p className="detail-card__text detail-card__label">Kategori</p>
        <p className="detail-card__value">{selectedItem.category}</p>

        <p className="detail-card__text detail-card__label">Zaman</p>
        <p className="detail-card__value">
          {selectedItem.time ? new Date(selectedItem.time).toLocaleString("tr-TR") : "-"}
        </p>
      </div>

      <div className="detail-card">
        <h3 className="detail-card__title">Cluster Bilgisi</h3>
        <p className="detail-card__text detail-card__label">Cluster Count</p>
        <p className="detail-card__value">{selectedItem.clusterCount ?? "-"}</p>

        <p className="detail-card__text detail-card__label">Virality</p>
        <p className="detail-card__value">{selectedItem.virality ?? "-"}</p>
      </div>

      <div className="detail-card">
        <h3 className="detail-card__title">Output Preview</h3>
        <a
          className="feed-card__link"
          href={selectedItem.link}
          target="_blank"
          rel="noreferrer"
        >
          Orijinal haberi ac
        </a>
      </div>

      <div className="detail-card">
        <h3 className="detail-card__title">Detay</h3>
        <p>{translatedDescription || rawDescription || "-"}</p>
      </div>

      <div className="detail-card">
        <h3 className="detail-card__title">Önem</h3>
        <p>{displayImportance || "-"}</p>
      </div>
    </aside>
  );
}
