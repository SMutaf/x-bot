import os
from datetime import datetime, timezone
from dotenv import load_dotenv
from langchain_google_genai import ChatGoogleGenerativeAI
from langchain_core.prompts import PromptTemplate
from langchain_core.output_parsers import JsonOutputParser
from pydantic import BaseModel, Field
from tenacity import retry, stop_after_attempt, wait_exponential, retry_if_exception_type
from typing import Optional
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


class TelegramOutput(BaseModel):
    hook: str = Field(description="Kısa, dikkat çekici ilk satır. Türkçe. Maksimum 10 kelime.")
    summary: str = Field(description="Haberi 1-2 kısa cümlede anlaşılır şekilde özetleyen Türkçe metin.")
    importance: str = Field(description="Bu haberin neden önemli olduğunu anlatan 1 kısa cümlelik Türkçe metin.")
    source_line: str = Field(description="Kaynak satırı. Format: 'Kaynak: X'")
    sentiment: str = Field(description="positive, negative veya neutral")


class GeminiService:
    def __init__(self):
        if not os.getenv("GOOGLE_API_KEY"):
            raise ValueError("GOOGLE_API_KEY ortam değişkeni bulunamadı!")

        self.llm = ChatGoogleGenerativeAI(
            model="gemma-3-12b-it",
            temperature=0.6,
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
İyi örnek:
Hook: Yeni modelde dikkat çeken yükseltme
Summary: Apple yeni cihazında daha güçlü çip ve gelişmiş kamera sistemi sundu.
Importance: Bu güncelleme, premium telefon pazarındaki rekabeti yeniden hızlandırabilir.
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
        retry=retry_if_exception_type(Exception),
        reraise=True
    )
    def _invoke_chain(self, chain, inputs):
        return chain.invoke(inputs)

    def _clean_text(self, text: str) -> str:
        if not text:
            return ""
        text = re.sub(r"\s+", " ", text).strip()
        text = text.replace("“", '"').replace("”", '"').replace("’", "'")
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
        news_type = self._detect_news_type(title, content or "", category)
        time_context = self._calculate_time_context(published_at)
        strategy = self._get_prompt_strategy(news_type, category, time_context)

        template = """
You are an expert Telegram news editor writing for a Turkish audience.

AMAÇ:
Verilen haberi Telegram kanalı için kısa, dikkat çekici, güvenilir ve sade bir formatta yeniden yaz.

EN KRİTİK KURAL:
- İlk satır mutlaka güçlü bir HOOK olsun.
- Hook, başlığı kopyalamasın.
- Hook, "bu neden önemli?" hissini ilk anda versin.
- Hook maksimum 10 kelime olsun.
- Hook clickbait gibi çiğ olmasın.

TELİF ve GÜVENLİ YENİDEN YAZIM KURALLARI:
- Başlığı asla birebir kopyalama
- İlk paragrafı asla birebir kopyalama
- Haberi kendi cümlelerinle yeniden kur
- Aynı anlamı daha doğal ve özgün Türkçeyle ver

DİL KURALLARI:
- Çıktı tamamen Türkçe olmalı
- Çok uzun cümle kurma
- Teknik dili gerekmedikçe sadeleştir
- Kısa, net, okunabilir ol

TON:
{tone_instruction}

HOOK KURALI:
{hook_rule}

ÖZET KURALI:
{summary_rule}

ÖNEM KURALI:
{importance_rule}

EMOJI KURALI:
{emoji_rule}

FORMAT:
1. hook
2. summary
3. importance
4. source_line

SOURCE_LINE formatı mutlaka şu biçimde olsun:
Kaynak: {source}

İYİ ÖRNEK:
{example}

HABER BİLGİLERİ:
- Kategori: {category}
- Haber Türü: {news_type}
{time_context}
- Kaynak: {source}
- Orijinal Başlık: {title}
- Orijinal İçerik: {content}
- Link: {url}

EK KURALLAR:
- summary en fazla 2 kısa cümle olsun
- importance tek cümle olsun
- source_line sadece kaynak bilgisini içersin
- link source_line içine yazma
- doğrulanmamış yorum ekleme

ÇIKTI:
{format_instructions}
        """

        prompt = PromptTemplate(
            template=template,
            input_variables=[
                "tone_instruction",
                "hook_rule",
                "summary_rule",
                "importance_rule",
                "emoji_rule",
                "example",
                "category",
                "news_type",
                "time_context",
                "source",
                "title",
                "content",
                "url",
            ],
            partial_variables={
                "format_instructions": self.parser.get_format_instructions()
            },
        )

        chain = prompt | self.llm | self.parser

        result = self._invoke_chain(chain, {
            "tone_instruction": strategy["tone"],
            "hook_rule": strategy["hook_rule"],
            "summary_rule": strategy["summary_rule"],
            "importance_rule": strategy["importance_rule"],
            "emoji_rule": strategy["emoji_rule"],
            "example": strategy["example"],
            "category": category,
            "news_type": news_type,
            "time_context": time_context,
            "source": source,
            "title": title,
            "content": content or "Detay bulunmuyor.",
            "url": url,
        })

        if not result.get("source_line"):
            result["source_line"] = f"Kaynak: {source}"

        final_message = self._build_final_message(result, url)

        return {
            "message": final_message,
            "hook": result.get("hook", ""),
            "summary": result.get("summary", ""),
            "importance": result.get("importance", ""),
            "source_line": result.get("source_line", f"Kaynak: {source}"),
            "sentiment": result.get("sentiment", "neutral"),
            "news_type": news_type,
        }