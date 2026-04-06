import os
from datetime import datetime, timezone
from dotenv import load_dotenv
from langchain_google_genai import ChatGoogleGenerativeAI
from langchain_core.prompts import PromptTemplate
from langchain_core.output_parsers import JsonOutputParser
from pydantic import BaseModel, Field
from tenacity import retry, stop_after_attempt, wait_exponential, retry_if_exception_type
from typing import Optional, Literal
import re

load_dotenv()

SENSITIVE_KEYWORDS = [
    "öldü", "ölü", "hayatını kaybetti", "şehit", "vefat", "ölüm",
    "saldırı", "bomba", "füze", "savaş", "çatışma", "terör",
    "deprem", "tsunami", "afet", "felaket", "yangın",
    "taciz", "tecavüz", "istismar", "şiddet", "cinayet",
    "kaza", "yaralı", "ağır yaralı", "hayatını kaybetmek",
    "kill", "death", "dead", "attack", "bombing", "war"
]

# Ön filtre: LLM'ye ulaşmadan direkt reject

# Başlıkta bu pattern'lardan biri varsa → direkt reject, LLM'ye gönderilmez.
# Tümü lowercase ile karşılaştırılır.
PRE_FILTER_PATTERNS = [
    # Fiyat / liste / güncelleme haberleri
    "fiyat listesi",
    "fiyat güncelleme",
    "zam geldi",
    "indirim haberi",
    "kampanya başladı",

    # Roundup / explainer / rehber
    "hakkında bildiğimiz her şey",
    "bilmeniz gereken her şey",
    "bilmeniz gerekenler",
    "nasıl kullanılır",
    "nasıl yapılır",
    "rehberi",
    "başlangıç rehberi",
    "kullanım kılavuzu",

    # Liste / karşılaştırma
    "en iyi 5",
    "en iyi 10",
    "en iyi 15",
    "en iyi 20",
    "karşılaştırma:",
    "karşılaştırması",

    # İnceleme / review
    "inceleme:",
    "incelemesi",
    "test ettik",
    "kullandık",
    "deneyimledik",

    # İngilizce roundup pattern'ları (Türkçe kaynaklarda da görünüyor)
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

# Kaynak + kategori kombinasyonuna göre ek kısıtlamalar.
# Bu kaynaklar TECH kategorisinde geliyorsa virality eşiği daha yüksek tutulur
STRICT_TECH_SOURCES = {"webtekno", "chip", "donanimhaber", "shiftdelete", "technopat"}


def pre_filter(title: str, source: str, category: str) -> Optional[str]:
    """
    LLM çağrılmadan önce haberi hızlıca filtrele.
    Elenirse reject_reason döner, geçerse None döner.
    """
    text = title.lower()

    for pattern in PRE_FILTER_PATTERNS:
        if pattern in text:
            return f"pre-filter-pattern:{pattern}"

    # Webtekno / chip gibi kaynaklar TECH kategorisinde çok geniş içerik üretiyor.
    # Başlık herhangi bir aksiyon sinyali taşımıyorsa reject et.
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


# Pydantic şema 

class TelegramOutput(BaseModel):
    decision: Literal["publish", "reject"] = Field(
        description="Haber yayınlanacaksa publish, editoryel filtreden geçmezse reject"
    )
    reject_reason: Optional[str] = Field(
        default=None,
        description="Reject ise kısa sebep."
    )
    hook: Optional[str] = Field(
        default=None,
        description="Kısa, dikkat çekici ilk satır. Türkçe. Maksimum 10 kelime."
    )
    summary: Optional[str] = Field(
        default=None,
        description="Haberi 1-2 kısa cümlede anlaşılır şekilde özetleyen Türkçe metin."
    )
    importance: Optional[str] = Field(
        default=None,
        description="Bu haberin neden önemli olduğunu anlatan 1 kısa cümlelik Türkçe metin."
    )
    source_line: Optional[str] = Field(
        default=None,
        description="Kaynak satırı. Format: 'Kaynak: X'"
    )
    sentiment: Optional[str] = Field(
        default=None,
        description="positive, negative veya neutral"
    )


# Servis

class GeminiService:
    def __init__(self):
        if not os.getenv("GOOGLE_API_KEY"):
            raise ValueError("GOOGLE_API_KEY ortam değişkeni bulunamadı!")

        self.llm = ChatGoogleGenerativeAI(
            model="gemma-3-12b-it",
            temperature=0.3,          # Daha düşük 
            convert_system_message_to_human=True
        )
        self.parser = JsonOutputParser(pydantic_object=TelegramOutput)

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

    def _get_prompt_strategy(self, news_type: str, category: str, time_context: str) -> dict:
        if news_type == "TRAGEDY":
            return {
                "tone": "Ciddi, saygılı, doğrudan. Abartı yapma. Sansasyonel dil kullanma.",
                "hook_rule": "Hook kısa, sade ve ciddi olmalı. Merak tuzağı kurma. Emoji kullanma.",
                "summary_rule": "Olayı net şekilde anlat. 1-2 cümle yeterli. Abartılı sıfat kullanma.",
                "importance_rule": "Neden önemli olduğunu sakin ve saygılı şekilde belirt.",
                "emoji_rule": "Emoji kullanma.",
                "example": """
İyi örnek:
Hook: Hastaneye saldırıda can kaybı arttı
Summary: Sudan'daki bir hastaneye düzenlenen saldırıda en az 64 kişi hayatını kaybetti.
Importance: Olay, bölgedeki insani krizin daha da derinleştiğini gösteriyor.
                """
            }

        elif news_type == "BREAKING_SERIOUS":
            return {
                "tone": "Hızlı, net, güvenilir. Fazla duygusal veya aşırı kışkırtıcı olma.",
                "hook_rule": "Hook dikkat çekici olmalı ama bağıran clickbait gibi olmamalı.",
                "summary_rule": "Ne olduğunu açık şekilde yaz. Kısa, temiz ve anlaşılır tut.",
                "importance_rule": "Piyasa, güvenlik, diplomasi veya bölgesel etkiyi tek cümlede açıkla.",
                "emoji_rule": "Gerekirse en fazla 1 emoji kullanılabilir ama şart değil.",
                "example": """
İyi örnek:
Hook: Ortadoğu'da tansiyon yeniden yükseliyor
Summary: İran'a bağlı hedeflere yönelik yeni saldırılar bölgede gerilimi artırdı.
Importance: Bu gelişme enerji fiyatları ve bölgesel güvenlik açısından yeni riskler yaratabilir.
                """
            }

        elif news_type == "TECH_LAUNCH":
            return {
                "tone": "Canlı ama profesyonel. Çok oyuncaklaştırma.",
                "hook_rule": "Hook yeni özelliği veya farkı öne çıkarsın.",
                "summary_rule": "Ürünü veya yeniliği 1-2 kısa cümlede anlat.",
                "importance_rule": "Neden dikkat çekici olduğunu kısa söyle.",
                "emoji_rule": "En fazla 1 uygun emoji kullanılabilir.",
                "example": """
İyi örnek (KABUL):
Hook: OpenAI yeni modeli piyasaya sürdü
Summary: GPT-5, önceki versiyona kıyasla çok daha hızlı ve daha az maliyetle çalışıyor.
Importance: Bu güncelleme, kurumsal yapay zeka kullanımını hızlandırabilir.

Kötü örnek (RED):
Başlık: "iOS 27 Hakkında Bildiğimiz Her Şey" → REJECT (roundup/feature yazısı)
Başlık: "2026 Fiyat Listesi Güncellendi" → REJECT (rutin fiyat güncellemesi)
                """
            }

        elif news_type == "ECONOMY_NEWS":
            return {
                "tone": "Net, sade, etkisini anlatan. Aşırı teknikleşme.",
                "hook_rule": "Hook, ekonomik etkinin ne olduğunu hissettirsin.",
                "summary_rule": "Veriyi veya kararı basit Türkçeyle özetle.",
                "importance_rule": "Türkiye, piyasalar veya küresel ekonomi etkisini belirt.",
                "emoji_rule": "Emoji kullanma ya da en fazla 1 nötr emoji kullan.",
                "example": """
İyi örnek:
Hook: Enerji fiyatlarında yeni baskı oluşuyor
Summary: Petrol fiyatlarındaki yükseliş, küresel piyasalarda maliyet baskısını artırıyor.
Importance: Bu durum enflasyon beklentilerini ve risk iştahını doğrudan etkileyebilir.
                """
            }

        return {
            "tone": "Dengeli, açık ve doğal Türkçe kullan.",
            "hook_rule": "Hook ilk bakışta haberi okutmalı.",
            "summary_rule": "Haberi 1-2 kısa cümlede sade şekilde anlat.",
            "importance_rule": "Neden önemli olduğunu tek kısa cümlede belirt.",
            "emoji_rule": "Gerekmedikçe emoji kullanma.",
            "example": """
İyi örnek:
Hook: Bölgedeki diplomasi trafiği hızlandı
Summary: Taraflar gerilimin büyümemesi için yeni temaslarda bulundu.
Importance: Bu görüşmeler, önümüzdeki günlerde atılacak adımlar açısından belirleyici olabilir.
            """
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
            return "\nZAMAN BAĞLAMI: Haber çok taze. Dili güncel ve canlı tut."
        elif minutes < 30:
            return f"\nZAMAN BAĞLAMI: Haber çok yeni ({minutes} dakika). Aciliyeti koru."
        elif minutes < 120:
            return f"\nZAMAN BAĞLAMI: Haber yakın zamanda yayımlandı ({minutes // 60} saat)."
        else:
            return "\nZAMAN BAĞLAMI: Haber nispeten eski. Aciliyet yerine önem vurgusu yap."

    @retry(
        stop=stop_after_attempt(3),
        wait=wait_exponential(multiplier=1, min=3, max=10),
        # Sadece geçici hatalar için retry; ValueError gibi kalıcı hatalar hariç
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

    def _truncate_sentence(self, text: str, max_len: int) -> str:
        text = self._clean_text(text)
        if len(text) <= max_len:
            return text

        truncated = text[:max_len].rsplit(" ", 1)[0].strip()
        if not truncated.endswith((".", "!", "?")):
            truncated += "..."
        return truncated

    def _build_final_message(self, result: dict, url: str) -> str:
        hook = self._truncate_sentence(result.get("hook", ""), 80)
        summary = self._truncate_sentence(result.get("summary", ""), 240)
        importance = self._truncate_sentence(result.get("importance", ""), 160)
        source_line = self._clean_text(result.get("source_line", ""))

        parts = []
        if hook:
            parts.append(f"**{hook}**")
        if summary:
            parts.append(summary)
        if importance:
            parts.append(f"_Neden önemli:_ {importance}")
        if source_line:
            parts.append(source_line)
        if url:
            parts.append(url)

        return "\n\n".join(parts).strip()

    def generate_telegram_post(
        self,
        title: str,
        content: str,
        url: str,
        source: str,
        category: str = "GENERAL",
        published_at: Optional[datetime] = None
    ):
        #  1. Ön filtre (LLM'siz, hızlı)
        pre_reject = pre_filter(title, source, category)
        if pre_reject:
            print(f"[PRE-FILTER] Direkt reddedildi ({pre_reject}): {title}")
            return {
                "decision": "reject",
                "reject_reason": pre_reject,
                "message": "", "hook": "", "summary": "",
                "importance": "", "source_line": "", "sentiment": "",
                "news_type": self._detect_news_type(title, content or "", category),
            }

        #  2. LLM filtresi
        news_type = self._detect_news_type(title, content or "", category)
        time_context = self._calculate_time_context(published_at)
        strategy = self._get_prompt_strategy(news_type, category, time_context)

        template = """
Sen bir Telegram haber kanalının baş editörüsün. Türk okuyuculara yönelik içerik üretiyorsun.
Görevin: Her haberi sıkı bir editoryel süzgeçten geçirmek ve sadece gerçekten değerli olanları yayına almak.

━━━ EDİTORYEL KARAR ÇERÇEVEN ━━━

YAYIN KONSEPTİ — sadece bunlar geçer:
1. Türkiye'yi doğrudan etkileyen somut gelişmeler (politika, ekonomi, güvenlik)
2. Küresel düzeyde jeopolitik, ekonomik veya güvenlik krizi niteliği taşıyan haberler
3. Piyasaları, enflasyonu ya da enerji arzını doğrudan etkileyen veriler / kararlar
4. Büyük teknoloji şirketlerinden (OpenAI, Google, Apple, Meta, Microsoft vb.) gerçek bir ürün lansmanı, büyük satın alma, büyük ceza, büyük güvenlik açığı
5. Doğal afet, savaş gelişmesi, büyük kaza gibi insanlığı etkileyen olaylar
6. Sosyal medyada gerçekten gündem yaratan, milyonlarca insanın takip ettiği olaylar

KESIN OLARAK REDDEDILECEKLER — bu listede olup olmadığını kontrol et:
✗ Fiyat listesi, güncelleme, kampanya, indirim haberleri
✗ "Hakkında bildiğimiz her şey", "bilmeniz gerekenler", roundup / özet yazıları
✗ Ürün incelemesi, test, nasıl kullanılır, rehber, karşılaştırma yazıları
✗ En iyi X listesi, top 10, sıralama haberleri
✗ Rutin yazılım / firmware güncellemesi (kritik güvenlik açığı değilse)
✗ Küçük aksesuvar, aksesuar, ara ürün tanıtımları
✗ Sadece yorum / analiz / köşe yazısı niteliğindeki içerikler
✗ Magazin, celebrity, eğlence, lifestyle haberleri (küresel etki yoksa)
✗ Türkiye ile alakasız ve küresel yankı yaratmayacak sıradan yerel dış haberler
✗ Sıradan spor maç sonuçları (büyük final, tarihi rekabet değilse)

KARAR VERİRKEN KULLANACAĞIN TEST:
"Bu haberi görünce Türkiye'deki bilinçli bir haber takipçisi 'vay be' der mi, yoksa 'bu ne ki' der mi?"
→ 'Vay be' → publish düşün
→ 'Bu ne ki' → reject

ALTIN KURAL:
- Emin değilsen REJECT ver. Hata payın publish değil, reject yönünde olsun.
- "Belki ilginçtir" veya "kısmen alakalı" gerekçesiyle publish YAPMA.
- Publish kararı için net, somut, güçlü bir gerekçen olmalı.

━━━ REJECT ÖRNEKLERİ (bunların benzerleri geçmemeli) ━━━
• "iOS 27 Hakkında Bildiğimiz Her Şey" → REJECT: roundup yazısı, haber değil
• "2026 Hyundai Fiyat Listesi Güncellendi" → REJECT: rutin fiyat güncellemesi
• "iPhone 17'yi Test Ettik: İşte İzlenimlerimiz" → REJECT: ürün incelemesi
• "En İyi 10 Ücretsiz VPN Uygulaması" → REJECT: liste yazısı
• "ChatGPT Nasıl Kullanılır? Adım Adım Rehber" → REJECT: rehber içeriği
• "Yerel Belediyeden Rutin Açıklama" → REJECT: düşük etki, yerel

━━━ PUBLISH ÖRNEKLERİ (bunların benzerleri geçmeli) ━━━
• "OpenAI, Google'ı 6.5 Milyar Dolara Geçti" → PUBLISH: somut büyük gelişme
• "Fed Faizi Beklenmedik Şekilde İndirdi" → PUBLISH: küresel piyasa etkisi
• "Ukrayna, Rus Petrol Tesisini Vurdu" → PUBLISH: jeopolitik gelişme
• "Apple, Yapay Zeka Arama Motorunu Duyurdu" → PUBLISH: büyük ürün lansmanı
• "İstanbul'da 5.8 Büyüklüğünde Deprem" → PUBLISH: Türkiye doğrudan etkisi

━━━ HABER BİLGİLERİ ━━━
Kategori: {category}
Haber Türü: {news_type}
{time_context}
Kaynak: {source}
Başlık: {title}
İçerik: {content}
Link: {url}

━━━ ÇIKTI FORMATI ━━━

REJECT ise SADECE şunu döndür:
  decision: "reject"
  reject_reason: (tek kelime veya kısa ifade, örn: "routine-update", "roundup-article", "low-impact")

PUBLISH ise şunu döndür:
  decision: "publish"
  hook: (max 10 kelime, güçlü ve özgün, başlığı kopyalama)
  summary: (1-2 kısa cümle, sade Türkçe)
  importance: (1 cümle, neden önemli)
  source_line: "Kaynak: {source}"
  sentiment: (positive / negative / neutral)

TON: {tone_instruction}
HOOK KURALI: {hook_rule}
ÖZET KURALI: {summary_rule}
ÖNEM KURALI: {importance_rule}
EMOJİ KURALI: {emoji_rule}

{format_instructions}
        """

        prompt = PromptTemplate(
            template=template,
            input_variables=[
                "tone_instruction", "hook_rule", "summary_rule",
                "importance_rule", "emoji_rule",
                "category", "news_type", "time_context",
                "source", "title", "content", "url",
            ],
            partial_variables={
                "format_instructions": self.parser.get_format_instructions()
            },
        )

        chain = prompt | self.llm | self.parser

        try:
            result = self._invoke_chain(chain, {
                "tone_instruction": strategy["tone"],
                "hook_rule": strategy["hook_rule"],
                "summary_rule": strategy["summary_rule"],
                "importance_rule": strategy["importance_rule"],
                "emoji_rule": strategy["emoji_rule"],
                "category": category,
                "news_type": news_type,
                "time_context": time_context,
                "source": source,
                "title": title,
                "content": content or "Detay bulunmuyor.",
                "url": url,
            })
        except Exception as e:
            print(f"[LLM HATA] {e} → {title}")
            return {
                "decision": "reject",
                "reject_reason": "llm-exception",
                "message": "", "hook": "", "summary": "",
                "importance": "", "source_line": "", "sentiment": "",
                "news_type": news_type,
            }

        decision = self._clean_text(result.get("decision", "")).lower()

        if decision == "reject":
            reason = result.get("reject_reason", "editorial-reject")
            print(f"[LLM REJECT] ({reason}): {title}")
            return {
                "decision": "reject",
                "reject_reason": reason,
                "message": "", "hook": "", "summary": "",
                "importance": "", "source_line": "", "sentiment": "",
                "news_type": news_type,
            }

        if decision != "publish":
            print(f"[LLM GEÇERSİZ KARAR] ({decision}): {title}")
            return {
                "decision": "reject",
                "reject_reason": "invalid-decision",
                "message": "", "hook": "", "summary": "",
                "importance": "", "source_line": "", "sentiment": "",
                "news_type": news_type,
            }

        if not result.get("source_line"):
            result["source_line"] = f"Kaynak: {source}"

        final_message = self._build_final_message(result, url)

        return {
            "decision": "publish",
            "reject_reason": "",
            "message": final_message,
            "hook": result.get("hook", ""),
            "summary": result.get("summary", ""),
            "importance": result.get("importance", ""),
            "source_line": result.get("source_line", f"Kaynak: {source}"),
            "sentiment": result.get("sentiment", "neutral"),
            "news_type": news_type,
        }