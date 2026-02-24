from pydantic import BaseModel

# Go Backend'den gelecek veri formatı
class TweetRequest(BaseModel):
    title: str
    content: str = "" # Özet veya içerik (boş olabilir)
    url: str
    source: str

# Go Backend'e döneceğimiz cevap formatı
class TweetResponse(BaseModel):
    tweet: str
    reply: str  # Link
    sentiment: str = "neutral" # Opsiyonel: Haberin duygusu