from fastapi import FastAPI
from models.schemas import TweetRequest, TweetResponse
import uvicorn

app = FastAPI(title="XNewsBot AI Engine")

@app.get("/")
def read_root():
    return {"status": "AI Engine is Online", "model": "GPT-4o (Waiting...)"}

# Go servisinin çağıracağı endpoint
@app.post("/generate-tweet", response_model=TweetResponse)
def generate_tweet_endpoint(request: TweetRequest):
    print(f"Haber Geldi: {request.title} ({request.source})")
    
    # --- BURAYA SONRA GERÇEK AI GELECEK ---
    # Şimdilik sistemin çalıştığını görmek için sahte cevap dönüyoruz.
    mock_tweet = f"Bu haber çok konuşulur!{request.title} hakkında detaylar şaşırtıcı. Siz ne düşünüyorsunuz?"
    mock_reply = f"Kaynağı incelemek isteyenler için: {request.url}"
    
    return TweetResponse(
        tweet=mock_tweet,
        reply=mock_reply,
        sentiment="positive"
    )

if __name__ == "__main__":
    # 8000 portunda çalıştır
    uvicorn.run(app, host="0.0.0.0", port=8000)