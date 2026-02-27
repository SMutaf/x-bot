import uvicorn
from fastapi import FastAPI
from fastapi.exceptions import RequestValidationError
from models.schemas import TweetRequest, TweetResponse
from services.llm import GeminiService
from core.exceptions import global_exception_handler, validation_exception_handler 

app = FastAPI(title="XNewsBot AI Engine (Gemini Powered)")

# --- GLOBAL HATA YÖNETİMİ (MIDDLEWARE) ---
# Tüm beklenmedik hataları (500) yakalar
app.add_exception_handler(Exception, global_exception_handler)

# Veri formatı hatalarını (422) yakalar ve düzenler
app.add_exception_handler(RequestValidationError, validation_exception_handler)
# ----------------------------------------

# Gemini servisini başlat
ai_service = GeminiService()

@app.get("/")
def read_root():
    return {"status": "Gemini AI Engine is Online", "model": "gemini-1.5-flash"}

@app.post("/generate-tweet", response_model=TweetResponse)
async def generate_tweet_endpoint(request: TweetRequest):
    print(f"Gemini Çalışıyor: {request.title}...")
    
    result = ai_service.generate_viral_tweet(
        title=request.title,
        content=request.content,
        url=request.url,
        source=request.source,
        category=request.category,
    )
    
    return TweetResponse(
        tweet=result["tweet"],
        reply=result["reply"],
        sentiment=result["sentiment"]
    )

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)