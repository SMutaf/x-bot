import uvicorn
from fastapi import FastAPI
from fastapi.exceptions import RequestValidationError
from models.schemas import TweetRequest, TweetResponse
from services.llm import GeminiService
from core.exceptions import global_exception_handler, validation_exception_handler

app = FastAPI(title="Telegram News Bot AI Engine")

app.add_exception_handler(Exception, global_exception_handler)
app.add_exception_handler(RequestValidationError, validation_exception_handler)

ai_service = GeminiService()


@app.get("/")
def read_root():
    return {"status": "Telegram AI Engine is Online", "model": "gemma-3-12b-it"}


@app.post("/generate-message", response_model=TweetResponse)
async def generate_tweet_endpoint(request: TweetRequest):
    print(f"Telegram içerik üretimi: {request.title}...")

    result = ai_service.generate_telegram_post(
        title=request.title,
        content=request.content,
        url=request.url,
        source=request.source,
        category=request.category,
        published_at=request.published_at,
    )

    return TweetResponse(
        message=result["message"],
        hook=result["hook"],
        summary=result["summary"],
        importance=result["importance"],
        source_line=result["source_line"],
        sentiment=result["sentiment"],
        news_type=result["news_type"],
    )


if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)