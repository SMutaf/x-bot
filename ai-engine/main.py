from fastapi import FastAPI
from models.schemas import TweetRequest, TweetResponse
from services.llm import GeminiService 
import uvicorn

app = FastAPI(title="XNewsBot AI Engine (Gemini Powered) 💎")

# Gemini servisini başlat
ai_service = GeminiService()

@app.get("/")
def read_root():
    return {"status": "Gemini AI Engine is Online 🚀", "model": "gemini-1.5-flash"}

@app.post("/generate-tweet", response_model=TweetResponse)
def generate_tweet_endpoint(request: TweetRequest):
    print(f"💎 Gemini Çalışıyor: {request.title}...")
    
    #  Yapay Zekaya Gönder
    result = ai_service.generate_viral_tweet(
        title=request.title,
        content=request.content,
        url=request.url,
        source=request.source
    )
    
    return TweetResponse(
        tweet=result["tweet"],
        reply=result["reply"],
        sentiment=result["sentiment"]
    )

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)