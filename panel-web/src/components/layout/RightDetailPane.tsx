export default function RightDetailPane() {
  return (
    <aside className="right-detail-pane">
      <div className="panel-title">Detay</div>

      <div className="detail-card">
        <h3 className="detail-card__title">Seçili Haber</h3>
        <p className="detail-card__text">
          Stream listesinde bir
          karta tıklanınca detay burada açılacak.
        </p>
      </div>

      <div className="detail-card">
        <h3 className="detail-card__title">Cluster Bilgisi</h3>
        <p className="detail-card__text">
          Aynı olayın kaç farklı kaynaktan geldiği gösterilecek.
        </p>
      </div>

      <div className="detail-card">
        <h3 className="detail-card__title">Output Preview</h3>
        <p className="detail-card__text">
          Son ürün metni burada yer alacak.
        </p>
      </div>
    </aside>
  );
}