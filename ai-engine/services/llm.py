import os
from dotenv import load_dotenv
from langchain_google_genai import ChatGoogleGenerativeAI
from langchain_core.prompts import PromptTemplate 
from langchain_core.output_parsers import JsonOutputParser
from pydantic import BaseModel, Field

# 1. Ortam değişkenlerini yükle
load_dotenv()

# 2. Çıktı formatını belirle 
class TweetOutput(BaseModel):
    tweet: str = Field(description="Viral, engaging tweet content in Turkish without links.")
    reply: str = Field(description="Reply content containing the source link and hashtags.")
    sentiment: str = Field(description="The sentiment of the news: positive, negative, or neutral")

class GeminiService:
    def __init__(self):
        self.llm = ChatGoogleGenerativeAI(
            model="gemma-3-12b-it", 
            temperature=0.7,
            convert_system_message_to_human=True
        )
        self.parser = JsonOutputParser(pydantic_object=TweetOutput)

    def generate_viral_tweet(self, title: str, content: str, url: str, source: str):
        template = """
        You are an algorithm-aware Social Media Strategist for a tech news account.

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
            input_variables=["source", "title", "content", "url"],
            partial_variables={
                "format_instructions": self.parser.get_format_instructions()
            }
        )

        chain = prompt | self.llm | self.parser

        try:
            response = chain.invoke({
                "source": source,
                "title": title,
                "content": content or "Detaylar linkte.",
                "url": url
            })
            return response

        except Exception as e:
            print(f"Gemini Hatası: {e}")
            return {
                "tweet": f" {title}\n\nTeknoloji dünyasındaki bu gelişme hakkında ne düşünüyorsunuz?",
                "reply": f"Kaynak: {source}\nDetaylar: {url}",
                "sentiment": "neutral"
            }