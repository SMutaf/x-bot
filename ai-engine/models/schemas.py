from pydantic import BaseModel

# Go Backend'den gelecek veri formatı
class TweetRequest(BaseModel):
    title: str
    content: str = ""
    url: str
    source: str
    category: str = "GENERAL"  # BREAKING / TECH / GENERAL

# Go Backend'e döneceğimiz cevap formatı
class TweetResponse(BaseModel):
    tweet: str
    reply: str
    sentiment: str = "neutral"