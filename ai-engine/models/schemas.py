from datetime import datetime
from typing import Optional
from pydantic import BaseModel, Field


class TweetRequest(BaseModel):
    title: str
    content: str
    url: str
    source: str
    category: str
    published_at: Optional[datetime] = None


class TweetResponse(BaseModel):
    message: str = Field(description="Telegram için hazırlanmış nihai mesaj")
    hook: str = Field(description="İlk dikkat çekici satır")
    summary: str = Field(description="Kısa haber özeti")
    importance: str = Field(description="Neden önemli açıklaması")
    source_line: str = Field(description="Kaynak satırı")
    sentiment: str = Field(description="positive, negative veya neutral")
    news_type: str = Field(description="TRAGEDY, BREAKING_SERIOUS, TECH_LAUNCH, ECONOMY_NEWS veya GENERAL_NEWS")