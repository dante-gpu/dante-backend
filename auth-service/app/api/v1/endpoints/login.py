from fastapi import APIRouter, Depends, HTTPException, status
from fastapi.security import OAuth2PasswordRequestForm
from sqlalchemy.orm import Session
from typing import Any

from app.db.session import get_db
from app.db.crud import crud_user
from app.schemas import user as user_schemas # Alias for clarity

# I need to create a router for login related endpoints.
router = APIRouter()

@router.post("/login", response_model=user_schemas.User)
def login_for_user_details(
    # I should use OAuth2PasswordRequestForm for standard form data input
    form_data: OAuth2PasswordRequestForm = Depends(),
    db: Session = Depends(get_db)
) -> Any:
    """
    I need an endpoint to handle user login and return user details upon success.
    It uses the OAuth2 standard for receiving username and password in form data.

    - Receives username and password via form data.
    - Authenticates the user using the CRUD helper.
    - If authentication fails (user not found, inactive, wrong password), raises HTTP 401.
    - If successful, returns the authenticated user's details (User schema).
    """
    # I should use the authenticate_user CRUD function.
    user = crud_user.authenticate_user(
        db, username=form_data.username, password=form_data.password
    )
    if not user:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Incorrect username or password",
            headers={"WWW-Authenticate": "Bearer"}, # Standard header for 401
        )
    # Although authenticate_user checks is_active, double-checking here is safe.
    if not user.is_active:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST, # Or 401? 400 seems appropriate for inactive.
            detail="Inactive user"
        )

    # If authentication is successful, I return the user object.
    # The API Gateway will use this info (user_id, username, role) to generate the JWT.
    return user

# Note: This endpoint intentionally does *not* return a JWT token.
# Token generation is the responsibility of the API Gateway based on the
# successful authentication response (containing user details) from this service. 