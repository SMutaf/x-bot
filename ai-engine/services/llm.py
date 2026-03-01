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

# Hassas kelimeler - ciddi muamele gerektiren konular
SENSITIVE_KEYWORDS = [
    "öldü", "ölü", "hayatını kaybetti", "şehit", "vefat", "ölüm",
    "saldırı", "bomba", "füze", "savaş", "çatışma", "terör",
    "deprem", "tsunami", "afet", "felaket", "yangın",
    "taciz", "tecavüz", "istismar", "şiddet", "cinayet",
    "kaza", "yaralı", "ağır yaralı", "hayatını kaybetmek",
    "kill", "death", "dead", "attack", "bombing", "war"
]

class TweetOutput(BaseModel):
    tweet: str = Field(description="Viral, engaging tweet content in Turkish without links.")
    reply: str = Field(description="Reply content containing the source link and hashtags.")
    sentiment: str = Field(description="The sentiment of the news: positive, negative, or neutral")

class GeminiService:
    def __init__(self):
        if not os.getenv("GOOGLE_API_KEY"):
            raise ValueError("GOOGLE_API_KEY ortam değişkeni bulunamadı!")

        self.llm = ChatGoogleGenerativeAI(
            model="gemma-3-12b-it",
            temperature=0.7,
            convert_system_message_to_human=True
        )
        self.parser = JsonOutputParser(pydantic_object=TweetOutput)

    def _detect_news_type(self, title: str, content: str, category: str) -> str:
        """
        Haberin türünü tespit et:
        - TRAGEDY: Ölüm, savaş, terör, kaza, afet
        - BREAKING_SERIOUS: Ciddi politik/ekonomik gelişme
        - TECH_LAUNCH: Ürün lansman, yeni teknoloji
        - GENERAL_NEWS: Normal haber
        """
        text = (title + " " + content).lower()
        
        # Hassas kelime kontrolü
        for keyword in SENSITIVE_KEYWORDS:
            if keyword in text:
                return "TRAGEDY"
        
        # Breaking + ciddi konular
        if category == "BREAKING":
            if any(word in text for word in ["başkan", "cumhurbaşkanı", "minister", "president", "hükümet"]):
                return "BREAKING_SERIOUS"
            return "BREAKING_SERIOUS"
        
        # Teknoloji lansmanları
        if category == "TECH":
            if any(word in text for word in ["tanıttı", "duyurdu", "çıktı", "launch", "announce", "reveal"]):
                return "TECH_LAUNCH"
        
        return "GENERAL_NEWS"

    def _get_prompt_strategy(self, news_type: str, category: str, time_context: str) -> dict:
        """
        Haber türüne göre prompt stratejisi belirle
        """
        
        if news_type == "TRAGEDY":
            return {
                "tone": "EXTREMELY SERIOUS. NO emojis. NO questions. NO engagement tricks.",
                "structure": "State the facts clearly and respectfully. Period. No commentary.",
                "emoji_rule": "ZERO emojis. Not even 🔴. Use text only: SON DAKİKA or BREAKING.",
                "question_rule": "NO questions. NO 'Ne düşünüyorsunuz?'. Just facts.",
                "example": """
Good: "SON DAKİKA | İran'da okul saldırısında 15 öğrenci hayatını kaybetti. (Al Jazeera)"
Bad: "İran'da trajedi 😢 Okul saldırısı... Ne düşünüyorsunuz?"
                """
            }
        
        elif news_type == "BREAKING_SERIOUS":
            return {
                "tone": "Urgent but neutral. Maximum 1 red dot emoji (🔴). No playful tone.",
                "structure": "Lead with action. Add context. No questions unless genuinely important.",
                "emoji_rule": "Only 🔴 for SON DAKİKA. No other emojis.",
                "question_rule": "Avoid questions. If used, make it rhetorical and serious.",
                "example": """
Good: "🔴 SON DAKİKA | Merkez Bankası faiz oranını %45'e yükseltti."
Bad: "Merkez Bankası faiz artırdı! 🚀 Bu karar ekonomiyi nasıl etkiler?"
                """
            }
        
        elif news_type == "TECH_LAUNCH":
            return {
                "tone": "Excited but informative. Max 2 tech-related emojis.",
                "structure": "Highlight innovation → Key specs → Optional question.",
                "emoji_rule": "Tech emojis OK: 🚀 💻 📱 ⚡ (max 2)",
                "question_rule": "Questions OK for tech: 'Almayı düşünür müsünüz?', 'Hangisi daha iyi?'",
                "example": """
Good: "iPhone 17 tanıtıldı! 🚀 6.1 inç OLED, A20 çip, 1TB depolama. Almayı düşünür müsünüz?"
Bad: "iPhone 17 geldi işte! 😍🔥💯 Sizce bu telefon efsane mi olacak?"
                """
            }
        
        else:  # GENERAL_NEWS
            return {
                "tone": "Balanced. Informative. Slightly engaging.",
                "structure": "Lead → Context → Light question if appropriate.",
                "emoji_rule": "Max 1-2 contextual emojis.",
                "question_rule": "Questions OK but not mandatory.",
                "example": """
Good: "Türkiye'de elektrikli araç satışları %40 arttı. Altyapı yeterli mi?"
                """
            }

    def _calculate_time_context(self, published_at: Optional[datetime]) -> str:
        """
        Haberin ne kadar yeni olduğunu hesaplar ve prompt'a eklenecek bağlamı döner.
        """
        if not published_at:
            return ""
        
        # Timezone-aware karşılaştırma
        now = datetime.now(timezone.utc)
        if published_at.tzinfo is None:
            published_at = published_at.replace(tzinfo=timezone.utc)
        
        diff = now - published_at
        minutes = int(diff.total_seconds() / 60)
        
        if minutes < 5:
            return "\n🔥 ULTRA-FRESH NEWS (< 5 minutes old): Use present tense, emphasize urgency."
        elif minutes < 30:
            return f"\n⚡ FRESH NEWS ({minutes} minutes old): Maintain urgency, recent past tense."
        elif minutes < 120:
            return f"\n📰 RECENT NEWS ({minutes // 60} hours old): Balance timeliness with context."
        else:
            return "\n📚 Older news: Focus on evergreen value."

    @retry(
        stop=stop_after_attempt(3),
        wait=wait_exponential(multiplier=1, min=4, max=10),
        retry=retry_if_exception_type(Exception),
        reraise=True
    )
    def _invoke_chain(self, chain, inputs):
        return chain.invoke(inputs)

    def generate_viral_tweet(
        self, 
        title: str, 
        content: str, 
        url: str, 
        source: str, 
        category: str = "GENERAL",
        published_at: Optional[datetime] = None 
    ):
        # 1. Haber türünü tespit et
        news_type = self._detect_news_type(title, content, category)
        
        # 2. Zaman bağlamını hesapla
        time_context = self._calculate_time_context(published_at)
        
        # 3. Prompt stratejisini al
        strategy = self._get_prompt_strategy(news_type, category, time_context)

        # 4. YENİ TELİF KORUMALI DİNAMİK PROMPT
        template = """
        You are a CONTEXT-AWARE Social Media Strategist for a professional news account.

         CRITICAL COPYRIGHT COMPLIANCE:
        - NEVER copy the headline verbatim
        - NEVER copy the first paragraph
        - NEVER use exact quotes longer than 5 words
        - ALWAYS rewrite in your own words
        - ALWAYS add original analysis/context
        - This is TRANSFORMATIVE USE, not reproduction

        NEWS TYPE: {news_type}
        {time_context}

        TONE & STYLE REQUIREMENTS:
        {tone_instruction}

        STRUCTURE:
        {structure_instruction}

        EMOJI RULES:
        {emoji_rule}

        QUESTION RULES:
        {question_rule}

        COPYRIGHT-SAFE REWRITING RULES:
        1. Read the headline → Understand the core fact
        2. REWRITE completely in different words
        3. Add your own angle/context
        4. Make it conversational, not journalistic
        5. If the news is "X announced Y", write "Y geldi! X duyurdu." (different structure)

        EXAMPLES OF TRANSFORMATIVE REWRITING:

        ❌ BAD (Too similar to source):
        Source: "Apple announces iPhone 17 with 200MP camera"
        Tweet: "Apple iPhone 17'yi 200MP kamerayla duyurdu"
        Problem: Nearly identical, just translated

        ✅ GOOD (Transformative):
        Source: "Apple announces iPhone 17 with 200MP camera"
        Tweet: "200 megapiksel kamera geliyor! iPhone 17 tanıtıldı."
        Why: Different structure, adds excitement, original phrasing

        ❌ BAD (Too similar):
        Source: "Merkez Bankası faiz oranını %45'e yükseltti"
        Tweet: "TCMB faiz oranını yüzde 45'e yükseltti"
        Problem: Just rephrased minimally

        ✅ GOOD (Transformative):
        Source: "Merkez Bankası faiz oranını %45'e yükseltti"
        Tweet: "Faizde sert adım! TCMB %45 kararını açıkladı."
        Why: Adds interpretation, different angle

        CRITICAL LANGUAGE REQUIREMENT:
        - Output MUST be 100% in Turkish (except proper nouns)
        - Use natural, native Turkish phrasing
        - Do NOT translate word-by-word

        UNIVERSAL HARD RULES:
        1. Main tweet under 280 characters
        2. NEVER include link in main tweet
        3. Be factually accurate but USE YOUR OWN WORDS
        4. Match the tone to the content severity

        SOURCE NEWS DATA:
        - Source: {source}
        - Original Headline: {title}
        - Original Content: {content}
        - URL: {url}

        IMPORTANT: The reply field MUST include:
        - Source attribution: "{source} haberi:"
        - The link: {url}
        - Relevant hashtags

        OUTPUT FORMAT:
        {format_instructions}
        """

        prompt = PromptTemplate(
            template=template,
            input_variables=[
                "news_type",
                "time_context", 
                "tone_instruction",
                "structure_instruction",
                "emoji_rule",
                "question_rule",
                "example",
                "source", 
                "title", 
                "content", 
                "url"
            ],
            partial_variables={"format_instructions": self.parser.get_format_instructions()}
        )

        chain = prompt | self.llm | self.parser

        result = self._invoke_chain(chain, {
            "news_type": news_type,
            "time_context": time_context,
            "tone_instruction": strategy["tone"],
            "structure_instruction": strategy["structure"],
            "emoji_rule": strategy["emoji_rule"],
            "question_rule": strategy["question_rule"],
            "example": strategy["example"],
            "source": source,
            "title": title,
            "content": content or "Detaylar linkte.",
            "url": url,
        })
        
        # Post-processing: Kaynak belirtme kontrolü
        if source.lower() not in result["reply"].lower():
            result["reply"] = f"{source} haberi: {result['reply']}"
        
        return result