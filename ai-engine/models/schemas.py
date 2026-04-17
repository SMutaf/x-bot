from datetime import datetime
from typing import Optional, Literal
from pydantic import BaseModel


class EditorialAnalysisRequest(BaseModel):
    title: str
    description: str
    category: str
    source: str
    published_at: datetime

    cluster_count: int
    virality: int


class EditorialAnalysisResponse(BaseModel):
    decision: Literal["PUBLISH", "REJECT"]
    reject_reason: Optional[str] = ""

    summary: str
    importance: str
    sentiment: str

    hook: str