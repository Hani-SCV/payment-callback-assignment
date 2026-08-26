from fastapi import Request
from fastapi.responses import JSONResponse


class APIError(Exception):
    def __init__(self, *, status_code: int, code: str, message: str) -> None:
        self.status_code = status_code
        self.code = code
        self.message = message
        super().__init__(message)


class AssignmentIncompleteError(APIError):
    def __init__(self) -> None:
        super().__init__(
            status_code=501,
            code="ASSIGNMENT_INCOMPLETE",
            message="Payment callbacks are not implemented yet.",
        )


async def api_error_handler(_: Request, exc: APIError) -> JSONResponse:
    return JSONResponse(
        status_code=exc.status_code,
        content={"error": {"code": exc.code, "message": exc.message}},
    )
