import { useEffect } from "react";
import type { FeedItem } from "../../features/feed/types";
import { useFeedStream } from "../../hooks/useFeedStream";

type MainStreamProps = {
  activeViewId: string;
  searchQuery: string;
  selectedItem: FeedItem | null;
  onSelectItem: (item: FeedItem) => void;
};

export default function MainStream(props: MainStreamProps) {
  const { activeViewId, searchQuery, selectedItem, onSelectItem } = props;
  const { connection, latestEventName, latestRawData, items, isInitialLoading } =
    useFeedStream(activeViewId);
  const normalizedQuery = normalizeSearchText(searchQuery);
  const filteredItems = normalizedQuery
    ? items.filter((item) => matchesSearch(item, normalizedQuery))
    : items;

  useEffect(() => {
    if (filteredItems.length === 0) {
      return;
    }

    if (!selectedItem) {
      onSelectItem(filteredItems[0]);
      return;
    }

    const selectedStillVisible = filteredItems.some(
      (item) =>
        selectedItem.link === item.link ||
        (selectedItem.title === item.title && selectedItem.time === item.time)
    );

    if (!selectedStillVisible) {
      onSelectItem(filteredItems[0]);
    }
  }, [filteredItems, selectedItem, onSelectItem]);

  return (
    <main className="main-stream">
      <div className="main-stream__header">
        <div>
          <h1 className="main-stream__title">Canli Akis</h1>
          <p className="main-stream__subtitle">Aktif gorunum: {activeViewId}</p>
        </div>

        <div className="live-status">
          <span
            className={`live-dot ${connection.isConnected ? "live-dot--ok" : "live-dot--off"}`}
          />
          <span>{connection.isConnected ? "Bagli" : "Baglanti Yok"}</span>
        </div>
      </div>

      <section className="stream-debug-card">
        <div className="stream-debug-row">
          <strong>Son event:</strong> {latestEventName ?? "Henuz event yok"}
        </div>
        <div className="stream-debug-row">
          <strong>Son zaman:</strong> {connection.lastEventAt ?? "-"}
        </div>
        <div className="stream-debug-row">
          <strong>Hata:</strong> {connection.lastError ?? "-"}
        </div>
      </section>

      <section className="stream-preview">
        <div className="panel-title">Ham Event Onizleme</div>
        <pre className="code-block">{latestRawData ?? "Yeni event bekleniyor..."}</pre>
      </section>

      <section className="stream-placeholder-card">
        <div className="panel-title">Feed Listesi</div>

        {isInitialLoading ? (
          <div className="feed-list">
            {Array.from({ length: 4 }).map((_, index) => (
              <article key={index} className="feed-card feed-card--skeleton" aria-hidden="true">
                <div className="feed-card__meta">
                  <span className="skeleton skeleton--badge" />
                  <span className="skeleton skeleton--meta" />
                </div>
                <div className="skeleton skeleton--title" />
                <div className="skeleton skeleton--line" />
                <div className="skeleton skeleton--line skeleton--line-short" />
                <div className="feed-card__stats">
                  <span className="skeleton skeleton--stat" />
                  <span className="skeleton skeleton--stat" />
                  <span className="skeleton skeleton--stat skeleton--stat-wide" />
                </div>
                <div className="virality-block">
                  <div className="virality-row">
                    <span className="skeleton skeleton--virality-label" />
                    <span className="skeleton skeleton--virality-value" />
                  </div>
                  <div className="skeleton skeleton--virality-bar" />
                </div>
                <span className="skeleton skeleton--link" />
              </article>
            ))}
          </div>
        ) : items.length === 0 ? (
          <p>Bu gorunum icin henuz kayit yok.</p>
        ) : filteredItems.length === 0 ? (
          <p>Bu arama icin eslesen kayit yok.</p>
        ) : (
          <div className="feed-list">
            {filteredItems.map((item) => {
              const key = item.link || `${item.title}-${item.time}`;
              const isSelected =
                !!selectedItem &&
                (selectedItem.link === item.link ||
                  (selectedItem.title === item.title && selectedItem.time === item.time));
              const cardTitle = item.hook || item.title;
              const cardDescription = item.summary || item.description || "";
              const categoryTone = getCategoryTone(item.category);
              const hasTurkeyImpact = getTurkeyImpact(item);

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
                    <span className={`feed-card__badge feed-card__badge--${categoryTone}`}>
                      {item.category}
                    </span>
                    {hasTurkeyImpact ? (
                      <span className="feed-card__badge feed-card__badge--turkey">TR Focus</span>
                    ) : null}
                    <span>&bull;</span>
                    <span>{item.source}</span>
                  </div>

                  <h3 className="feed-card__title">{cardTitle}</h3>

                  {cardDescription ? <p>{cardDescription}</p> : null}

                  <div className="feed-card__stats">
                    <span>Cluster: {item.clusterCount ?? "-"}</span>
                    <span>
                      Saat: {item.time ? new Date(item.time).toLocaleString("tr-TR") : "-"}
                    </span>
                  </div>

                  <div className="virality-block">
                    <div className="virality-row">
                      <span className="virality-label">Virality</span>
                      <span className="virality-value">{formatVirality(item.virality)}</span>
                    </div>
                    <div className="virality-bar" aria-hidden="true">
                      <div
                        className="virality-bar__fill"
                        style={{ width: `${getViralityScore(item.virality)}%` }}
                      />
                    </div>
                  </div>

                  <a
                    className="feed-card__link"
                    href={item.link}
                    target="_blank"
                    rel="noreferrer"
                    onClick={(e) => e.stopPropagation()}
                  >
                    Haberi ac
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

function getCategoryTone(category: string) {
  switch (category?.toUpperCase()) {
    case "BREAKING":
      return "breaking";
    case "TECH":
      return "tech";
    case "ECONOMY":
      return "economy";
    case "GENERAL":
      return "general";
    default:
      return "general";
  }
}

function getTurkeyImpact(item: FeedItem) {
  const text = normalizeSearchText(
    `${item.title} ${item.description ?? ""} ${item.summary ?? ""} ${item.importance ?? ""} ${item.source}`
  );

  const signals = [
    "turkiye",
    "turkey",
    "ankara",
    "istanbul",
    "izmir",
    "tbmm",
    "tcmb",
    "bist",
    "borsa istanbul",
    "lira",
    "try",
    "erdogan",
    "cumhurbaskani"
  ];

  return signals.some((signal) => text.includes(signal));
}

function getViralityScore(value?: number) {
  if (typeof value !== "number" || Number.isNaN(value)) {
    return 0;
  }

  return Math.max(0, Math.min(100, value));
}

function formatVirality(value?: number) {
  const score = getViralityScore(value);
  return `${score}/100`;
}

function matchesSearch(item: FeedItem, query: string) {
  const haystack = normalizeSearchText(
    [
      item.title,
      item.hook,
      item.summary,
      item.description,
      item.source,
      item.category,
      item.importance,
      item.sentiment
    ]
      .filter(Boolean)
      .join(" ")
  );

  if (haystack.includes(query)) {
    return true;
  }

  const queryTokens = tokenizeSearch(query);
  if (queryTokens.length === 0) {
    return true;
  }

  const textTokens = tokenizeSearch(haystack);
  if (textTokens.length === 0) {
    return false;
  }

  return queryTokens.every((queryToken) =>
    textTokens.some((textToken) => isLooseMatch(queryToken, textToken))
  );
}

function normalizeSearchText(value: string) {
  return value
    .toLocaleLowerCase("tr-TR")
    .normalize("NFD")
    .replace(/[\u0300-\u036f]/g, "")
    .replace(/['’`]/g, "")
    .replace(/[^a-z0-9\s]/g, " ")
    .replace(/\s+/g, " ")
    .trim();
}

function tokenizeSearch(value: string) {
  return value
    .split(" ")
    .map((token) => token.trim())
    .filter((token) => token.length >= 2);
}

function isLooseMatch(queryToken: string, textToken: string) {
  if (textToken.includes(queryToken) || queryToken.includes(textToken)) {
    return true;
  }

  if (textToken.startsWith(queryToken) || queryToken.startsWith(textToken)) {
    return true;
  }

  if (Math.abs(queryToken.length - textToken.length) > 2) {
    return false;
  }

  const distance = levenshtein(queryToken, textToken);
  if (queryToken.length <= 4) {
    return distance <= 1;
  }

  return distance <= 2;
}

function levenshtein(a: string, b: string) {
  const rows = a.length + 1;
  const cols = b.length + 1;
  const dp = Array.from({ length: rows }, () => Array(cols).fill(0));

  for (let i = 0; i < rows; i += 1) {
    dp[i][0] = i;
  }

  for (let j = 0; j < cols; j += 1) {
    dp[0][j] = j;
  }

  for (let i = 1; i < rows; i += 1) {
    for (let j = 1; j < cols; j += 1) {
      const cost = a[i - 1] === b[j - 1] ? 0 : 1;
      dp[i][j] = Math.min(
        dp[i - 1][j] + 1,
        dp[i][j - 1] + 1,
        dp[i - 1][j - 1] + cost
      );
    }
  }

  return dp[rows - 1][cols - 1];
}
