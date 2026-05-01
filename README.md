# x-bot — Turkey-Focused News Monitoring & Delivery System

An AI-assisted news system that monitors developments affecting Turkey and high-impact global events in real time, then delivers approved stories to Telegram and a live web panel.

---

## Screenshots

### panel-web — Live Feed
![panel-web article detail](/screenshots/panel-web.png)

### Operations Dashboard
![dashboard article detail](/screenshots/dashboard.png)

---

## Purpose

x-bot is built to answer three practical questions:

- What happened in the world that affects Turkey?
- Is there a critical global development worth paying attention to?
- Is this story actually important, or just noise?

The system monitors 30+ global and local news sources, filters out low-signal content, scores events by importance, and runs shortlisted stories through a Gemini-powered editorial layer before publishing them.

---

## How It Works

1. **RSS ingestion (Go)**
   The backend polls global and local RSS feeds at configurable intervals.

2. **Fast filtering (Go)**
   Duplicate, low-value, and weakly relevant content is removed using Redis-backed deduplication and rule-based filtering.

3. **Processing (Go)**
   Remaining items are clustered, scored, and checked against category-specific publication policies.

4. **Editorial analysis (Python / FastAPI / Gemini)**
   The AI layer makes the final PUBLISH / REJECT decision and generates:
   - a short hook
   - a Turkish summary
   - a Turkish detail translation
   - sentiment
   - importance note

5. **Delivery**
   Approved items are:
   - sent to the Telegram delivery flow
   - streamed to the live panel-web
   - exposed through the internal operations dashboard

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────┐
│                    RSS SCRAPER (Go)                     │
│   BBC · NYT · Guardian · Al Jazeera · Bloomberg · AA    │
└────────────────────────┬────────────────────────────────┘
                         │ Raw Articles
                         ▼
┌─────────────────────────────────────────────────────────┐
│                  INGESTION PIPELINE (Go)                │
│  Dedup (Redis)  →  Keyword Filter  →  Relevance Filter  │
└────────────────────────┬────────────────────────────────┘
                         │ Filtered Articles
                         ▼
┌─────────────────────────────────────────────────────────┐
│              PROCESSING PIPELINE (Go)                   │
│  Event Clustering  →  Scoring  →  Policy Check          │
└────────────────────────┬────────────────────────────────┘
                         │ High-Signal Candidates
                         ▼
┌─────────────────────────────────────────────────────────┐
│            AI ENGINE (Python · FastAPI · Gemini)        │
│   Editorial Decision → Summary → Detail TR → Sentiment  │
└────────────────────────┬────────────────────────────────┘
                         │ PUBLISH / REJECT
                         ▼
┌─────────────────────────────────────────────────────────┐
│                   DELIVERY LAYER (Go)                   │
└──────────────────┬──────────────────┬───────────────────┘
                   │                  │
       ┌───────────▼──────┐  ┌────────▼────────────────┐
       │  TELEGRAM FLOW   │  │  PANEL-WEB (Live Feed)  │
       └──────────────────┘  └─────────────────────────┘
```

---

## Components

### 1. `backend/` — Core Orchestration (Go)

The Go backend manages the full pipeline end to end.

| Package | Responsibility |
|---|---|
| `ingestion/scraper` | Polls RSS sources at configured intervals |
| `ingestion/dedup` | Removes duplicate articles via Redis |
| `ingestion/filter` | Applies Turkey relevance and low-noise filtering |
| `ingestion/sourcehealth` | Tracks source health, failures, and cooldown state |
| `processing/cluster` | Groups similar articles into the same event |
| `processing/scoring` | Scores stories by recency, keywords, clustering, and burst |
| `processing/policy` | Applies category-specific publication rules |
| `delivery/telegram` | Sends approved items into the Telegram flow |
| `api/dashboard` | REST endpoints used by the dashboard |
| `api/stream` | SSE stream used by panel-web |

### 2. `ai-engine/` — AI Editorial Engine (Python · FastAPI)

Gemini-powered editorial layer.

It is responsible for:

- final PUBLISH / REJECT decision
- short hook generation
- Turkish summary generation
- Turkish detail translation of the source description
- sentiment labeling
- short importance explanation

### 3. `panel-web/` — Live User Panel (React · TypeScript · Vite)

This is the primary end-user surface for approved stories.

Current features include:

- live feed via SSE
- preset high-signal views:
  - Turkey Critical
  - Global Impact
  - Economy
  - Tech Watch
- keyword search
- article detail pane
- category badges and Turkey-focus highlighting

### 4. `dashboard/` — Operations Panel (React · Vite)

Internal monitoring panel for operating the pipeline.

Current features include:

- published and rejected news tables
- source health table
- Redis / Python service health
- live feed connection state
- top-level counts for published, rejected, and source status
- JSONL exports for operational data

---

## Filtering & Scoring Logic

### Turkey Focus

The system prioritizes stories that:

- **directly involve Turkey**
  - Turkey, Türkiye, Ankara, Istanbul, TCMB, TRY, BIST, Erdoğan
- **indirectly affect Turkey**
  - regional conflict, energy shocks, NATO developments, global macro moves, sanctions, major market events

### Category Scope

| Category | Scope |
|---|---|
| `BREAKING` | War, attacks, earthquakes, diplomatic crises, sanctions, critical incidents |
| `ECONOMY` | Central bank decisions, FX moves, commodities, macro and market shocks |
| `TECH` | Major launches, outages, breaches, bans, acquisitions |
| `GENERAL` | Political and diplomatic developments relevant to Turkey |

### Automatically Rejected

Examples of content that is filtered out early:

- guides and explainers
- reviews and comparisons
- "top 10" / listicle-style content
- weak low-impact stories
- duplicate coverage of the same event

### Scoring Signals

Stories are scored using a combination of signals:

| Signal | Meaning |
|---|---|
| Recency Score | How recent the article is |
| Keyword Score | Presence of high-impact keywords |
| Cluster Score | How many sources confirm the same event |
| Burst Score | Whether the topic is spiking unusually fast |
| Turkey Relevance | Whether the event directly or indirectly matters to Turkey |

---

## News Sources

The system monitors a mix of global and Turkish/local feeds.

Examples include:

| Category | Sources |
|---|---|
| `BREAKING` | BBC World, NYT World, The Guardian, Al Jazeera, Politico EU, Sky News, DW |
| `ECONOMY` | Bloomberg, Financial Times, CNBC, MarketWatch, Bloomberg HT, BBC Business |
| `TECH` | TechCrunch, The Verge, Ars Technica, Webtekno |
| `GENERAL` | AA, TRT Haber, NTV, Habertürk, Cumhuriyet, BBC Türkçe |

---

## Tech Stack

| Layer | Technology |
|---|---|
| Backend | Go |
| AI Engine | Python, FastAPI, LangChain, Google Gemini |
| Frontend | React, Vite, TypeScript |
| Cache / Dedup | Redis |
| Streaming | SSE |
| Delivery | Telegram Bot API |

---

## Project Structure

```
x-bot/
├── backend/
│   ├── cmd/main.go
│   ├── config/
│   └── internal/
│       ├── api/
│       ├── delivery/
│       ├── domain/models/
│       ├── ingestion/
│       ├── infra/
│       └── processing/
│
├── ai-engine/
│   ├── main.py
│   ├── services/llm.py
│   └── models/schemas.py
│
├── panel-web/
└── dashboard/
```

---

## Output Surfaces

### Telegram

Approved stories are formatted and sent into the Telegram delivery flow.

### panel-web

Approved stories appear in the live feed with:

- hook
- summary
- category badge
- virality score
- source
- publication time
- detail pane with translated description

### dashboard

Operators can inspect:

- source health
- accepted vs rejected flow
- service health
- recent published and rejected records

---

## Design Principles

**Low noise, high signal**
Sending irrelevant news is worse than sending none.

**Turkey first**
The system is optimized for Turkey-relevant monitoring rather than generic world news.

**Multi-source awareness**
Event clustering helps separate isolated noise from confirmed developments.

**Single pipeline, multiple surfaces**
The same approved flow powers Telegram, panel-web, and the ops dashboard.

**AI as editor, rules as guardrails**