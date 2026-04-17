import uvicorn
from fastapi import FastAPI
from fastapi.exceptions import RequestValidationError

from models.schemas import EditorialAnalysisRequest, EditorialAnalysisResponse
from core.exceptions import global_exception_handler, validation_exception_handler

app = FastAPI(title="Editorial AI Engine")

app.add_exception_handler(Exception, global_exception_handler)
app.add_exception_handler(RequestValidationError, validation_exception_handler)


@app.get("/")
def read_root():
    return {"status": "AI Engine Running", "mode": "editorial"}


@app.post("/analyze", response_model=EditorialAnalysisResponse)
def analyze(req: EditorialAnalysisRequest):

    print(f"[AI] Analyzing: {req.title}")

    # ŞİMDİLİK MOCK
    return EditorialAnalysisResponse(
        decision="PUBLISH",
        reject_reason="",

        summary="Test summary",
        importance="MEDIUM",
        sentiment="NEUTRAL",

        hook="Test hook"
    )


if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)