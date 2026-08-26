from fastapi import FastAPI

from app.api.payment_callbacks import router as payment_callback_router
from app.errors import APIError, api_error_handler

app = FastAPI(title="Payment Callback Migration Assignment", version="0.1.0")
app.add_exception_handler(APIError, api_error_handler)
app.include_router(payment_callback_router)


@app.get("/health")
def health() -> dict[str, str]:
    return {"status": "ok"}
