from fastapi import FastAPI
from starlette.middleware.cors import CORSMiddleware
import logging

# I need to import my settings and potentially API routers later
from app.core.config import settings
from app.api.v1.api import api_router
from app.db.session import engine
from app.db.base import Base

# Setup logging
logger = logging.getLogger(__name__)

# I should create the FastAPI application instance.
app = FastAPI(
    title=settings.PROJECT_NAME,
    openapi_url=f"{settings.API_V1_STR}/openapi.json"
)

# --- Database Table Creation ---
def create_db_tables():
    logger.info("Attempting to create database tables if they don't exist...")
    try:
        Base.metadata.create_all(bind=engine)
        logger.info("Database tables checked/created successfully.")
    except Exception as e:
        logger.error(f"Error creating database tables: {e}", exc_info=True)
        # Depending on the error, you might want to exit or handle differently
        # For now, just logging the error.

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

# --- Startup Event ---
@app.on_event("startup")
async def on_startup():
    """I should run table creation on startup."""
    create_db_tables()

# --- Main Execution (for debugging, usually run with uvicorn) ---
# if __name__ == "__main__":
#     import uvicorn
#     # Add basic logging config for direct run
#     logging.basicConfig(level=logging.INFO)
#     uvicorn.run(app, host="0.0.0.0", port=8001, log_level="info") 