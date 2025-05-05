from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session
from typing import List

from app.db import crud # Using __init__ perhaps? No, let's be explicit.
from app.db.crud import crud_user
from app.schemas import user as user_schemas # Alias to avoid naming conflicts
from app.db.session import get_db # Dependency to get DB session

# I need to create an API router for user-related endpoints.
router = APIRouter()

@router.post("/", response_model=user_schemas.User, status_code=status.HTTP_201_CREATED)
def create_new_user(
    *, # Makes db and user_in keyword-only arguments
    db: Session = Depends(get_db),
    user_in: user_schemas.UserCreate
):
    """
    I need an endpoint to create a new user.

    - It receives user details in the request body (UserCreate schema).
    - Checks if a user with the same email or username already exists.
    - If exists, raises an HTTP 409 Conflict error.
    - Otherwise, creates the user using the CRUD function.
    - Returns the created user data (User schema).
    """
    # I should check if a user with this username already exists.
    existing_user_by_username = crud_user.get_user_by_username(db, username=user_in.username)
    if existing_user_by_username:
        raise HTTPException(
            status_code=status.HTTP_409_CONFLICT,
            detail="A user with this username already exists.",
        )

    # I should also check if a user with this email already exists.
    existing_user_by_email = crud_user.get_user_by_email(db, email=user_in.email)
    if existing_user_by_email:
        raise HTTPException(
            status_code=status.HTTP_409_CONFLICT,
            detail="A user with this email already exists.",
        )

    # If checks pass, I can create the user.
    user = crud_user.create_user(db=db, user_in=user_in)
    return user

# I might add other user endpoints here later (get users, get user by ID, update, delete)
# @router.get("/", response_model=List[user_schemas.User])
# def read_users(
#     db: Session = Depends(get_db),
#     skip: int = 0,
#     limit: int = 100,
#     # current_user: models.User = Depends(get_current_active_superuser) # Need dependency for auth
# ):
#     """Retrieve users (requires admin privileges - auth not implemented yet)."""
#     users = crud_user.get_users(db, skip=skip, limit=limit)
#     return users 