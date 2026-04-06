from datetime import datetime
from typing import Optional, Literal
from pydantic import BaseModel, Field


class TweetRequest(BaseModel):
    title: str
    content: str
    url: str
    source: str
    category: str
    published_at: Optional[datetime] = None


class TweetResponse(BaseModel):
    decision: Literal["publish", "reject"] = Field(
        description="LLM editoryal kararı: publish veya reject"
    )
    reject_reason: Optional[str] = Field(
        default="",
        description="Reject ise kısa sebep"
    )

    message: str = Field(
        default="",
        description="Telegram için hazırlanmış nihai mesaj"
    )
    hook: str = Field(
        default="",
        description="İlk dikkat çekici satır"
    )
    summary: str = Field(
        default="",
        description="Kısa haber özeti"
    )
    importance: str = Field(
        default="",
        description="Neden önemli açıklaması"
    )
    source_line: str = Field(
        default="",
        description="Kaynak satırı"
    )
    sentiment: str = Field(
        default="",
        description="positive, negative veya neutral"
    )
    news_type: str = Field(
        default="",
        description="TRAGEDY, BREAKING_SERIOUS, TECH_LAUNCH, ECONOMY_NEWS veya GENERAL_NEWS"
    )