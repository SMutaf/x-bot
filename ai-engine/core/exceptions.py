from fastapi import Request, status
from fastapi.responses import JSONResponse
from fastapi.exceptions import RequestValidationError
import logging

# Loglama ayarı (Hataları konsola basmak için)
logging.basicConfig(level=logging.ERROR)
logger = logging.getLogger(__name__)

async def global_exception_handler(request: Request, exc: Exception):
    """
    Tüm beklenmeyen hataları (500 Internal Server Error) yakalar.
    """
    logger.error(f"Global Exception: {str(exc)} - URL: {request.url}")
    
    return JSONResponse(
        status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
        content={
            "success": False,
            "error": "Internal Server Error",
            "message": f"Beklenmedik bir hata oluştu: {str(exc)}" # Prod ortamında bu detay gizlenmeli
        },
    )

async def validation_exception_handler(request: Request, exc: RequestValidationError):
    """
    Pydantic validasyon hatalarını (422 Unprocessable Entity) yakalar ve formatlar.
    """
    error_messages = []
    for error in exc.errors():
        field = " -> ".join(str(loc) for loc in error["loc"])
        message = error["msg"]
        error_messages.append(f"{field}: {message}")

    return JSONResponse(
        status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
        content={
            "success": False,
            "error": "Validation Error",
            "message": "Gönderilen veride hatalar var.",
            "details": error_messages
        },
    )