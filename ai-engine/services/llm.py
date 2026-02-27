import os
from datetime import datetime, timezone
from dotenv import load_dotenv
from langchain_google_genai import ChatGoogleGenerativeAI
from langchain_core.prompts import PromptTemplate
from langchain_core.output_parsers import JsonOutputParser
from pydantic import BaseModel, Field
from tenacity import retry, stop_after_attempt, wait_exponential, retry_if_exception_type
from typing import Optional

load_dotenv()

# Kategori bazlı ton talimatları
CATEGORY_INSTRUCTIONS = {
    "BREAKING": (
        "This is BREAKING NEWS. "
        "Be extremely concise and punchy. "
        "Impact statement must feel urgent. "
        "Question must demand immediate opinion."
    ),
    "TECH": (
        "This is a technology deep-dive. "
        "Impact statement should highlight technical significance. "
        "Question should invite expert discussion."
    ),
    "GENERAL": (
        "This is a general tech news item. "
        "Balance informativeness with engagement. "
        "Question should be broadly relatable."
    ),
}

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
            return "\n🔥 ULTRA-FRESH NEWS (< 5 minutes old): Use present tense, emphasize 'breaking right now'."
        elif minutes < 30:
            return f"\n⚡ FRESH NEWS ({minutes} minutes old): Maintain urgency, use recent past tense."
        elif minutes < 120:
            return f"\n📰 RECENT NEWS ({minutes // 60} hours old): Balance timeliness with context."
        else:
            return "\n📚 Older news: Focus on evergreen value, not urgency."

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
        # Kategori talimatını al
        category_instruction = CATEGORY_INSTRUCTIONS.get(category, CATEGORY_INSTRUCTIONS["GENERAL"])
        
        # Zaman bağlamını hesapla
        time_context = self._calculate_time_context(published_at)

        template = """
        You are an algorithm-aware Social Media Strategist for a tech news account.

        CATEGORY CONTEXT:
        {category_instruction}
        {time_context}

        CRITICAL LANGUAGE REQUIREMENT:
        - The output MUST be 100% in Turkish.
        - Do NOT use English words unless proper nouns.
        - Output only Turkish text.

        Your objective is to create a high-engagement Twitter (X) post 
        while maintaining honesty and credibility.

        OPTIMIZATION STRATEGY:
        - Encourage meaningful replies
        - Increase dwell time
        - Improve retweet probability
        - Avoid clickbait or misleading tone

        HARD RULES:
        1. Main tweet under 280 characters.
        2. NEVER include link in main tweet.
        3. Max 2 emojis.
        4. Be factually aligned.

        STRUCTURE:
        - Impact statement
        - Why it matters
        - Open-ended question

        NEWS DATA:
        - Source: {source}
        - Title: {title}
        - Content Snippet: {content}
        - URL: {url}

        OUTPUT FORMAT:
        {format_instructions}
        """

        prompt = PromptTemplate(
            template=template,
            input_variables=["source", "title", "content", "url", "category_instruction", "time_context"],
            partial_variables={"format_instructions": self.parser.get_format_instructions()}
        )

        chain = prompt | self.llm | self.parser

        return self._invoke_chain(chain, {
            "source": source,
            "title": title,
            "content": content or "Detaylar linkte.",
            "url": url,
            "category_instruction": category_instruction,
            "time_context": time_context, 
        })