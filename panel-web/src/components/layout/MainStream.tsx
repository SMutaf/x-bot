import { useFeedStream } from "../../hooks/useFeedStream";

type MainStreamProps = {
  activeViewId: string;
};

export default function MainStream(props: MainStreamProps) {
  const { activeViewId } = props;
  const { connection, latestEventName, latestRawData } = useFeedStream(activeViewId);

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
        <p>
          <code>news.published</code> eventlerini kart
          olarak render edilicek.
        </p>
      </section>
    </main>
  );
}