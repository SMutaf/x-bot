from datetime import datetime
from typing import Optional, Literal
from pydantic import BaseModel, Field


class EditorialAnalysisRequest(BaseModel):
    title: str
    description: str
    category: str
    source: str
    published_at: Optional[datetime] = None

    cluster_count: int
    virality: int


class EditorialAnalysisResponse(BaseModel):
    decision: Literal["PUBLISH", "REJECT"] = Field(
        description="Editoryal karar: PUBLISH veya REJECT"
    )
    reject_reason: Optional[str] = Field(
        default="",
        description="REJECT ise kısa sebep"
    )

    hook: str = Field(
        default="",
        description="Kısa dikkat çekici ilk satır"
    )
    summary: str = Field(
        default="",
        description="Haberi 1-2 kısa cümlede özetleyen metin"
    )
    importance: str = Field(
        default="",
        description="Neden önemli olduğunu anlatan kısa cümle"
    )
    sentiment: Literal["positive", "negative", "neutral"] = Field(
        default="neutral",
        description="Haber tonu"
    )