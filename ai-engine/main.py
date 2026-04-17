import uvicorn
from fastapi import FastAPI
from fastapi.exceptions import RequestValidationError

from models.schemas import EditorialAnalysisRequest, EditorialAnalysisResponse
from services.llm import GeminiService
from core.exceptions import global_exception_handler, validation_exception_handler

app = FastAPI(title="Editorial AI Engine")

app.add_exception_handler(Exception, global_exception_handler)
app.add_exception_handler(RequestValidationError, validation_exception_handler)

ai_service = GeminiService()


@app.get("/")
def read_root():
    return {"status": "AI Engine Running", "mode": "editorial"}


@app.post("/analyze", response_model=EditorialAnalysisResponse)
async def analyze(req: EditorialAnalysisRequest):
    print(f"[AI] Analyzing: {req.title}...")

    result = ai_service.analyze_editorial(
        title=req.title,
        content=req.description,
        source=req.source,
        category=req.category,
        published_at=req.published_at,
        cluster_count=req.cluster_count,
        virality=req.virality,
    )

    return EditorialAnalysisResponse(
        decision=result.get("decision", "REJECT"),
        reject_reason=result.get("reject_reason", ""),
        hook=result.get("hook", ""),
        summary=result.get("summary", ""),
        importance=result.get("importance", ""),
        sentiment=result.get("sentiment", "neutral"),
    )


if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)