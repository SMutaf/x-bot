import os
import re
from datetime import datetime, timezone
from typing import Optional, Literal

from dotenv import load_dotenv
from langchain_google_genai import ChatGoogleGenerativeAI
from langchain_core.prompts import PromptTemplate
from langchain_core.output_parsers import JsonOutputParser
from pydantic import BaseModel, Field
from tenacity import retry, stop_after_attempt, wait_exponential, retry_if_exception_type

load_dotenv()

SENSITIVE_KEYWORDS = [
    "öldü", "ölü", "hayatını kaybetti", "şehit", "vefat", "ölüm",
    "saldırı", "bomba", "füze", "savaş", "çatışma", "terör",
    "deprem", "tsunami", "afet", "felaket", "yangın",
    "taciz", "tecavüz", "istismar", "şiddet", "cinayet",
    "kaza", "yaralı", "ağır yaralı", "hayatını kaybetmek",
    "kill", "death", "dead", "attack", "bombing", "war"
]

PRE_FILTER_PATTERNS = [
    "fiyat listesi",
    "fiyat güncelleme",
    "zam geldi",
    "indirim haberi",
    "kampanya başladı",
    "hakkında bildiğimiz her şey",
    "bilmeniz gereken her şey",
    "bilmeniz gerekenler",
    "nasıl kullanılır",
    "nasıl yapılır",
    "rehberi",
    "başlangıç rehberi",
    "kullanım kılavuzu",
    "en iyi 5",
    "en iyi 10",
    "en iyi 15",
    "en iyi 20",
    "karşılaştırma:",
    "karşılaştırması",
    "inceleme:",
    "incelemesi",
    "test ettik",
    "kullandık",
    "deneyimledik",
    "everything we know",
    "all you need to know",
    "how to use",
    "hands-on",
    "review:",
    "explainer:",
    "analysis:",
    " explained",
    "best of",
    "top 10",
    "top 5",
]

STRICT_TECH_SOURCES = {"webtekno", "chip", "donanimhaber", "shiftdelete", "technopat"}


def pre_filter(title: str, source: str, category: str) -> Optional[str]:
    text = title.lower()

    for pattern in PRE_FILTER_PATTERNS:
        if pattern in text:
            return f"pre-filter-pattern:{pattern}"

    if category == "TECH" and source.lower() in STRICT_TECH_SOURCES:
        action_signals = [
            "duyurdu", "tanıttı", "açıkladı", "yasakladı", "kapattı",
            "satın aldı", "birleşti", "çöktü", "hacklendi", "sızdırıldı",
            "erişime kapandı", "iflas", "ceza", "rekor", "devralma",
            "launched", "acquired", "banned", "hacked", "fined",
        ]
        if not any(sig in text for sig in action_signals):
            return "pre-filter-no-action-signal"

    return None


class EditorialAnalysisOutput(BaseModel):
    decision: Literal["PUBLISH", "REJECT"] = Field(
        description="Haber yayınlanacaksa PUBLISH, editoryel filtreden geçmezse REJECT"
    )
    reject_reason: Optional[str] = Field(
        default="",
        description="REJECT ise kısa sebep"
    )
    hook: Optional[str] = Field(
        default="",
        description="Kısa, dikkat çekici ilk satır. Türkçe. Maksimum 10 kelime."
    )
    summary: Optional[str] = Field(
        default="",
        description="Haberi 1-2 kısa cümlede anlaşılır şekilde özetleyen Türkçe metin."
    )
    importance: Optional[str] = Field(
        default="",
        description="Bu haberin neden önemli olduğunu anlatan 1 kısa cümlelik Türkçe metin."
    )
    sentiment: Literal["positive", "negative", "neutral"] = Field(
        default="neutral",
        description="positive, negative veya neutral"
    )


class GeminiService:
    def __init__(self):
        if not os.getenv("GOOGLE_API_KEY"):
            raise ValueError("GOOGLE_API_KEY ortam değişkeni bulunamadı!")

        self.llm = ChatGoogleGenerativeAI(
            model="gemma-3-12b-it",
            temperature=0.3,
            convert_system_message_to_human=True
        )
        self.parser = JsonOutputParser(pydantic_object=EditorialAnalysisOutput)

    def _detect_news_type(self, title: str, content: str, category: str) -> str:
        text = (title + " " + content).lower()

        for keyword in SENSITIVE_KEYWORDS:
            if keyword in text:
                return "TRAGEDY"

        if category == "BREAKING":
            return "BREAKING_SERIOUS"

        if category == "TECH":
            if any(word in text for word in ["tanıttı", "duyurdu", "çıktı", "launch", "announce", "reveal"]):
                return "TECH_LAUNCH"

        if category == "ECONOMY":
            return "ECONOMY_NEWS"

        return "GENERAL_NEWS"

    def _get_prompt_strategy(self, news_type: str) -> dict:
        if news_type == "TRAGEDY":
            return {
                "tone": "Ciddi, saygılı, doğrudan. Abartı yapma. Sansasyonel dil kullanma.",
                "hook_rule": "Hook kısa, sade ve ciddi olmalı. Merak tuzağı kurma. Emoji kullanma.",
                "summary_rule": "Olayı net şekilde anlat. 1-2 cümle yeterli. Abartılı sıfat kullanma.",
                "importance_rule": "Neden önemli olduğunu sakin ve saygılı şekilde belirt.",
            }

        if news_type == "BREAKING_SERIOUS":
            return {
                "tone": "Hızlı, net, güvenilir. Fazla duygusal veya aşırı kışkırtıcı olma.",
                "hook_rule": "Hook dikkat çekici olmalı ama bağıran clickbait gibi olmamalı.",
                "summary_rule": "Ne olduğunu açık şekilde yaz. Kısa, temiz ve anlaşılır tut.",
                "importance_rule": "Piyasa, güvenlik, diplomasi veya bölgesel etkiyi tek cümlede açıkla.",
            }

        if news_type == "TECH_LAUNCH":
            return {
                "tone": "Canlı ama profesyonel. Çok oyuncaklaştırma.",
                "hook_rule": "Hook yeni özelliği veya farkı öne çıkarsın.",
                "summary_rule": "Ürünü veya yeniliği 1-2 kısa cümlede anlat.",
                "importance_rule": "Neden dikkat çekici olduğunu kısa söyle.",
            }

        if news_type == "ECONOMY_NEWS":
            return {
                "tone": "Net, sade, etkisini anlatan. Aşırı teknikleşme.",
                "hook_rule": "Hook, ekonomik etkinin ne olduğunu hissettirsin.",
                "summary_rule": "Veriyi veya kararı basit Türkçeyle özetle.",
                "importance_rule": "Türkiye, piyasalar veya küresel ekonomi etkisini belirt.",
            }

        return {
            "tone": "Dengeli, açık ve doğal Türkçe kullan.",
            "hook_rule": "Hook ilk bakışta haberi okutmalı.",
            "summary_rule": "Haberi 1-2 kısa cümlede sade şekilde anlat.",
            "importance_rule": "Neden önemli olduğunu tek kısa cümlede belirt.",
        }

    def _calculate_time_context(self, published_at: Optional[datetime]) -> str:
        if not published_at:
            return ""

        now = datetime.now(timezone.utc)
        if published_at.tzinfo is None:
            published_at = published_at.replace(tzinfo=timezone.utc)

        diff = now - published_at
        minutes = int(diff.total_seconds() / 60)

        if minutes < 5:
            return "Haber çok taze. Dili güncel ve canlı tut."
        if minutes < 30:
            return f"Haber çok yeni ({minutes} dakika). Aciliyeti koru."
        if minutes < 120:
            return f"Haber yakın zamanda yayımlandı ({minutes // 60} saat)."
        return "Haber nispeten eski. Aciliyet yerine önem vurgusu yap."

    @retry(
        stop=stop_after_attempt(3),
        wait=wait_exponential(multiplier=1, min=3, max=10),
        retry=retry_if_exception_type((TimeoutError, ConnectionError, OSError)),
        reraise=True
    )
    def _invoke_chain(self, chain, inputs):
        return chain.invoke(inputs)

    def _clean_text(self, text: str) -> str:
        if not text:
            return ""
        text = re.sub(r"\s+", " ", text).strip()
        text = text.replace("\u201c", '"').replace("\u201d", '"').replace("\u2019", "'")
        return text

    def _normalize_sentiment(self, sentiment: str) -> str:
        sentiment = self._clean_text(sentiment).lower()
        if sentiment not in {"positive", "negative", "neutral"}:
            return "neutral"
        return sentiment

    def analyze_editorial(
        self,
        title: str,
        content: str,
        source: str,
        category: str = "GENERAL",
        published_at: Optional[datetime] = None,
        cluster_count: int = 1,
        virality: int = 0,
    ):
        pre_reject = pre_filter(title, source, category)
        if pre_reject:
            print(f"[PRE-FILTER] Direkt reddedildi ({pre_reject}): {title}")
            return {
                "decision": "REJECT",
                "reject_reason": pre_reject,
                "hook": "",
                "summary": "",
                "importance": "",
                "sentiment": "neutral",
            }

        news_type = self._detect_news_type(title, content or "", category)
        strategy = self._get_prompt_strategy(news_type)
        time_context = self._calculate_time_context(published_at)

        template = """
Sen Türkiye odaklı bir haber sisteminin baş editörüsün.
Görevin bir haberi yayınlayıp yayınlamamaya karar vermek ve yayınlanacaksa kısa editoryel alanları üretmek.

YAYINLANABİLECEK HABERLER:
- Türkiye'yi doğrudan etkileyen somut gelişmeler
- Küresel jeopolitik, ekonomik, enerji veya güvenlik etkisi olan haberler
- Büyük teknoloji şirketlerinden önemli lansman, büyük güvenlik açığı, büyük satın alma, büyük ceza
- Piyasaları veya risk algısını etkileyebilecek güçlü gelişmeler
- Büyük kriz, savaş, afet, diplomatik kırılma, kritik ekonomik veri

TÜRKİYE BAĞLANTI KURALLARI:
Aşağıdaki durumlarda haberin Türkiye ile "dolaylı bağlantısı" olduğunu kabul et ve PUBLISH yönünde değerlendir:
- ABD, AB veya Çin ekonomik/ticaret kararları → TL kuru, BIST, ihracat/ithalat doğrudan etkilenir
- Fed veya ECB faiz kararları → gelişen piyasa (EM) sermaye akışı üzerinden Türkiye'yi etkiler
- Orta Doğu güvenlik gelişmeleri (Suriye, İran, İsrail, Körfez) → enerji fiyatları, turizm, sınır güvenliği
- NATO veya savunma gelişmeleri → Türkiye NATO üyesidir, her karar doğrudan bağlayıcıdır
- Körfez / Hürmüz Boğazı / LNG enerji haberleri → Türkiye enerjisinin büyük bölümünü ithal eder
- Küresel gıda veya hammadde fiyat şokları → Türkiye'nin cari açığını doğrudan etkiler
- Rusya-Ukrayna gelişmeleri → enerji, tahıl, turizm ve döviz üzerinden kritik etki
Bu durumlarda "dolaylı etki var" diyerek PUBLISH kararı verebilirsin; kesin kanıt arama.

KESİN RED:
- rehber, roundup, "bilmeniz gerekenler", "everything we know"
- inceleme, karşılaştırma, liste, test yazıları
- rutin fiyat/list price/update içerikleri
- düşük etkili yerel dış haberler
- sadece yorum / analiz yazıları
- düşük etkili magazin / spor / lifestyle içerikleri

ÖNEMLİ:
- Emin değilsen REJECT ver.
- PUBLISH ancak net etki varsa ver.
- Hook başlığı kopyalamasın.
- Çıktıyı sadece JSON formatında ver.

Kategori: {category}
Kaynak: {source}
Başlık: {title}
İçerik: {content}
NewsType: {news_type}
PublishedAtContext: {time_context}
ClusterCount: {cluster_count}
Virality: {virality}

TON: {tone}
HOOK KURALI: {hook_rule}
ÖZET KURALI: {summary_rule}
ÖNEM KURALI: {importance_rule}

REJECT ise:
- decision = "REJECT"
- reject_reason = kısa sebep
- hook, summary, importance boş string olabilir

PUBLISH ise:
- decision = "PUBLISH"
- hook = kısa dikkat çekici ilk satır
- summary = 1-2 kısa cümle
- importance = neden önemli
- sentiment = positive / negative / neutral

{format_instructions}
"""

        prompt = PromptTemplate(
            template=template,
            input_variables=[
                "category", "source", "title", "content", "news_type",
                "time_context", "cluster_count", "virality",
                "tone", "hook_rule", "summary_rule", "importance_rule",
            ],
            partial_variables={
                "format_instructions": self.parser.get_format_instructions()
            },
        )

        chain = prompt | self.llm | self.parser

        try:
            result = self._invoke_chain(chain, {
                "category": category,
                "source": source,
                "title": title,
                "content": content or "Detay bulunmuyor.",
                "news_type": news_type,
                "time_context": time_context,
                "cluster_count": cluster_count,
                "virality": virality,
                "tone": strategy["tone"],
                "hook_rule": strategy["hook_rule"],
                "summary_rule": strategy["summary_rule"],
                "importance_rule": strategy["importance_rule"],
            })
        except Exception as e:
            print(f"[LLM HATA] {e} → {title}")
            return {
                "decision": "REJECT",
                "reject_reason": "llm-exception",
                "hook": "",
                "summary": "",
                "importance": "",
                "sentiment": "neutral",
            }

        decision = self._clean_text(result.get("decision", "")).upper()

        if decision == "REJECT":
            reason = self._clean_text(result.get("reject_reason", "editorial-reject"))
            print(f"[LLM REJECT] ({reason}): {title}")
            return {
                "decision": "REJECT",
                "reject_reason": reason or "editorial-reject",
                "hook": "",
                "summary": "",
                "importance": "",
                "sentiment": "neutral",
            }

        if decision != "PUBLISH":
            print(f"[LLM GEÇERSİZ KARAR] ({decision}): {title}")
            return {
                "decision": "REJECT",
                "reject_reason": "invalid-decision",
                "hook": "",
                "summary": "",
                "importance": "",
                "sentiment": "neutral",
            }

        return {
            "decision": "PUBLISH",
            "reject_reason": "",
            "hook": self._clean_text(result.get("hook", "")),
            "summary": self._clean_text(result.get("summary", "")),
            "importance": self._clean_text(result.get("importance", "")),
            "sentiment": self._normalize_sentiment(result.get("sentiment", "neutral")),
        }