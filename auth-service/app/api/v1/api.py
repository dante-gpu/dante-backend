from fastapi import APIRouter

from app.api.v1.endpoints import users # I need to import the users router
# Import other endpoint routers here later, e.g.:
from app.api.v1.endpoints import login # I need to import the login router

# I need to create the main router for API v1.
api_router = APIRouter()

# I should include the users router with a prefix and tags.
api_router.include_router(users.router, prefix="/users", tags=["Users"])

# Include other routers here:
api_router.include_router(login.router, prefix="/login", tags=["Login"]) # Include login router 