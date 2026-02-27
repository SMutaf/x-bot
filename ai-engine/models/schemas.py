from pydantic import BaseModel, Field
from datetime import datetime
from typing import Optional

# Go Backend'den gelecek veri formatı
class TweetRequest(BaseModel):
    title: str
    content: str = ""
    url: str
    source: str
    category: str = "GENERAL"  # BREAKING / TECH / GENERAL
    published_at: Optional[datetime] = None  # Haberin yayınlanma zamanı

# Go Backend'e döneceğimiz cevap formatı
class TweetResponse(BaseModel):
    tweet: str
    reply: str
    sentiment: str = "neutral"