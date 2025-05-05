from pydantic import BaseModel, EmailStr, Field
from typing import Optional
from datetime import datetime
import uuid

# --- Base Schemas ---
# I need a base schema with fields common to reading and creation/update.
class UserBase(BaseModel):
    # Using Optional fields that might not be present in all operations
    email: Optional[EmailStr] = None
    username: Optional[str] = None
    role: Optional[str] = Field(default="user") # Default role is 'user'
    is_active: Optional[bool] = True

# --- Schemas for Creation ---
# I need a schema for creating a new user, inheriting from Base.
# Password is required during creation.
class UserCreate(UserBase):
    email: EmailStr = Field(..., example="user@example.com")
    username: str = Field(..., min_length=3, max_length=50, example="newuser")
    password: str = Field(..., min_length=8, example="aSecurePassword123")
    role: str = Field(default="user", example="user") # Explicitly require role on create? Or keep default?

# --- Schemas for Update ---
# Schema for updating an existing user. All fields are optional.
class UserUpdate(UserBase):
    password: Optional[str] = Field(default=None, min_length=8, example="aNewSecurePassword123") # Optional new password

# --- Schemas for Reading ---
# I need a base schema for reading user data, inheriting from UserBase.
# It should definitely *not* include the password.
class UserInDBBase(UserBase):
    id: uuid.UUID # Use UUID for primary key
    created_at: datetime
    updated_at: datetime

    # I should configure Pydantic to work with ORM models.
    model_config = {
        "from_attributes": True # Renamed from orm_mode in Pydantic v2
    }

# The final schema for returning user data via the API.
class User(UserInDBBase):
    # No extra fields needed currently, inheriting all from UserInDBBase.
    pass

# Schema for representing a user stored in the database, including the hashed password.
# This should *never* be returned by the API directly.
class UserInDB(UserInDBBase):
    hashed_password: str 