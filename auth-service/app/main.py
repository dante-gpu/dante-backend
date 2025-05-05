from fastapi import FastAPI
from starlette.middleware.cors import CORSMiddleware

# I need to import my settings and potentially API routers later
from app.core.config import settings
from app.api.v1.api import api_router

# I should create the FastAPI application instance.
app = FastAPI(
    title=settings.PROJECT_NAME,
    openapi_url=f"{settings.API_V1_STR}/openapi.json"
)

# --- Middleware Setup ---

# I need to set up CORS middleware if origins are defined in the settings.
if settings.BACKEND_CORS_ORIGINS:
    app.add_middleware(
        CORSMiddleware,
        allow_origins=[str(origin) for origin in settings.BACKEND_CORS_ORIGINS],
        allow_credentials=True,
        allow_methods=["GET", "POST", "PUT", "DELETE", "OPTIONS"], # Allow common methods
        allow_headers=["*"], # Allow all headers for simplicity, can be restricted
    )

# --- Router Setup ---

# I will include the main API router here later.
app.include_router(api_router, prefix=settings.API_V1_STR)

# --- Basic Root Endpoint ---

@app.get("/", tags=["Root"])
async def read_root():
    """I should provide a simple root endpoint indicating the service is running."""
    return {"message": f"Welcome to {settings.PROJECT_NAME}!", "docs": "/docs"}

# --- Database Connection (Example - to be refined) ---
# I might add startup/shutdown events to manage DB connections.
# @app.on_event("startup")
# async def startup_db_client():
#     # Connect to DB
#     pass
#
# @app.on_event("shutdown")
# async def shutdown_db_client():
#     # Disconnect DB
#     pass

# --- Main Execution (for debugging, usually run with uvicorn) ---
# if __name__ == "__main__":
#     import uvicorn
#     uvicorn.run(app, host="0.0.0.0", port=8001, log_level="info") 