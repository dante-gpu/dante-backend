import asyncio
import uuid
from datetime import datetime, timedelta
from typing import Optional, List

import uvicorn
from fastapi import FastAPI, HTTPException, Depends, status, Security
from fastapi.middleware.cors import CORSMiddleware
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from sqlalchemy.orm import Session
from passlib.context import CryptContext
from jose import JWTError, jwt
import secrets

from app.core.config import Settings
from app.db.session import get_db
from app.db.models.user import User
from app.db.crud.crud_user import (
    create_user, 
    get_user_by_email, 
    get_user_by_username, 
    authenticate_user,
    get_user,
    get_users
)
from app.schemas.user import UserCreate, User as UserSchema, UserUpdate
from app.core.security import verify_password, get_password_hash

# Initialize settings
settings = Settings()

# Initialize FastAPI app
app = FastAPI(
    title="DanteGPU Authentication Service",
    description="Professional authentication service for DanteGPU platform with JWT tokens and user management",
    version="1.0.0",
    docs_url="/docs",
    redoc_url="/redoc"
)

# CORS configuration
app.add_middleware(
    CORSMiddleware,
    allow_origins=[
        "http://localhost:3000", 
        "http://localhost:3001",  # Frontend Next.js
        "http://localhost:8080",  # API Gateway
        "http://localhost:9999",  # Test server
        "http://127.0.0.1:3000",
        "http://127.0.0.1:3001",
        "http://127.0.0.1:8080",
        "http://127.0.0.1:9999",
    ],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Password hashing
pwd_context = CryptContext(schemes=["bcrypt"], deprecated="auto")

# JWT Security
security = HTTPBearer()

# Token models
from pydantic import BaseModel, EmailStr
from typing import Optional

class Token(BaseModel):
    access_token: str
    refresh_token: str
    token_type: str
    expires_at: datetime
    user_id: str
    username: str
    email: str
    role: str

class TokenData(BaseModel):
    username: Optional[str] = None
    user_id: Optional[str] = None
    email: Optional[str] = None
    role: Optional[str] = None

class LoginRequest(BaseModel):
    username: str
    password: str

class RegisterRequest(BaseModel):
    username: str
    email: EmailStr
    password: str
    role: Optional[str] = "user"

class AuthResponse(BaseModel):
    token: str
    refresh_token: str
    expires_at: datetime
    user_id: str
    username: str
    email: str
    role: str
    permissions: List[str] = []

class UserProfile(BaseModel):
    id: str
    username: str
    email: str
    role: str
    is_active: bool
    created_at: datetime
    updated_at: datetime

class PasswordChangeRequest(BaseModel):
    current_password: str
    new_password: str

class PasswordResetRequest(BaseModel):
    email: EmailStr

class PasswordResetConfirm(BaseModel):
    token: str
    new_password: str

# JWT Functions
def create_access_token(data: dict, expires_delta: Optional[timedelta] = None):
    to_encode = data.copy()
    if expires_delta:
        expire = datetime.utcnow() + expires_delta
    else:
        expire = datetime.utcnow() + timedelta(hours=24)
    
    to_encode.update({"exp": expire, "type": "access"})
    encoded_jwt = jwt.encode(to_encode, settings.SECRET_KEY, algorithm=settings.ALGORITHM)
    return encoded_jwt, expire

def create_refresh_token(data: dict):
    to_encode = data.copy()
    expire = datetime.utcnow() + timedelta(days=30)
    to_encode.update({"exp": expire, "type": "refresh"})
    encoded_jwt = jwt.encode(to_encode, settings.SECRET_KEY, algorithm=settings.ALGORITHM)
    return encoded_jwt, expire

def verify_token(token: str) -> TokenData:
    try:
        payload = jwt.decode(token, settings.SECRET_KEY, algorithms=[settings.ALGORITHM])
        username: str = payload.get("sub")
        user_id: str = payload.get("user_id")
        email: str = payload.get("email")
        role: str = payload.get("role")
        
        if username is None:
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="Invalid authentication credentials",
                headers={"WWW-Authenticate": "Bearer"},
            )
        
        token_data = TokenData(
            username=username,
            user_id=user_id,
            email=email,
            role=role
        )
        return token_data
    except JWTError:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Invalid authentication credentials",
            headers={"WWW-Authenticate": "Bearer"},
        )

async def get_current_user(
    credentials: HTTPAuthorizationCredentials = Security(security),
    db: Session = Depends(get_db)
) -> User:
    token = credentials.credentials
    token_data = verify_token(token)
    
    user = get_user_by_username(db, username=token_data.username)
    if user is None:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="User not found",
            headers={"WWW-Authenticate": "Bearer"},
        )
    
    if not user.is_active:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="Inactive user"
        )
    
    return user

async def get_current_active_user(current_user: User = Depends(get_current_user)) -> User:
    if not current_user.is_active:
        raise HTTPException(status_code=400, detail="Inactive user")
    return current_user

# Authentication Routes
@app.post("/api/v1/auth/register", response_model=AuthResponse)
async def register(user_data: RegisterRequest, db: Session = Depends(get_db)):
    """Register a new user account"""
    
    # Check if user already exists
    existing_user_email = get_user_by_email(db, email=user_data.email)
    if existing_user_email:
        raise HTTPException(
            status_code=400,
            detail="Email already registered"
        )
    
    existing_user_username = get_user_by_username(db, username=user_data.username)
    if existing_user_username:
        raise HTTPException(
            status_code=400,
            detail="Username already taken"
        )
    
    # Validate password strength
    if len(user_data.password) < 8:
        raise HTTPException(
            status_code=400,
            detail="Password must be at least 8 characters long"
        )
    
    # Create user
    user_create = UserCreate(
        username=user_data.username,
        email=user_data.email,
        password=user_data.password,
        role=user_data.role
    )
    
    try:
        user = create_user(db, user_in=user_create)
    except Exception as e:
        raise HTTPException(
            status_code=500,
            detail=f"Failed to create user: {str(e)}"
        )
    
    # Generate tokens
    access_token, access_expires = create_access_token(
        data={
            "sub": user.username,
            "user_id": str(user.id),
            "email": user.email,
            "role": user.role
        }
    )
    
    refresh_token, _ = create_refresh_token(
        data={
            "sub": user.username,
            "user_id": str(user.id)
        }
    )
    
    return AuthResponse(
        token=access_token,
        refresh_token=refresh_token,
        expires_at=access_expires,
        user_id=str(user.id),
        username=user.username,
        email=user.email,
        role=user.role,
        permissions=["user:read", "wallet:read", "job:submit"] if user.role == "user" else ["admin:all"]
    )

@app.post("/api/v1/auth/login", response_model=AuthResponse)
async def login(login_data: LoginRequest, db: Session = Depends(get_db)):
    """Authenticate user and return access tokens"""
    
    # Authenticate user
    user = authenticate_user(db, login_data.username, login_data.password)
    if not user:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Invalid username or password",
            headers={"WWW-Authenticate": "Bearer"},
        )
    
    if not user.is_active:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="User account is disabled"
        )
    
    # Generate tokens
    access_token, access_expires = create_access_token(
        data={
            "sub": user.username,
            "user_id": str(user.id),
            "email": user.email,
            "role": user.role
        }
    )
    
    refresh_token, _ = create_refresh_token(
        data={
            "sub": user.username,
            "user_id": str(user.id)
        }
    )
    
    return AuthResponse(
        token=access_token,
        refresh_token=refresh_token,
        expires_at=access_expires,
        user_id=str(user.id),
        username=user.username,
        email=user.email,
        role=user.role,
        permissions=["user:read", "wallet:read", "job:submit"] if user.role == "user" else ["admin:all"]
    )

@app.post("/api/v1/auth/refresh", response_model=AuthResponse)
async def refresh_token(refresh_token: str, db: Session = Depends(get_db)):
    """Refresh access token using refresh token"""
    
    try:
        payload = jwt.decode(refresh_token, settings.SECRET_KEY, algorithms=[settings.ALGORITHM])
        username: str = payload.get("sub")
        token_type: str = payload.get("type")
        
        if username is None or token_type != "refresh":
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="Invalid refresh token"
            )
        
        user = get_user_by_username(db, username=username)
        if user is None or not user.is_active:
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="User not found or inactive"
            )
        
        # Generate new access token
        access_token, access_expires = create_access_token(
            data={
                "sub": user.username,
                "user_id": str(user.id),
                "email": user.email,
                "role": user.role
            }
        )
        
        return AuthResponse(
            token=access_token,
            refresh_token=refresh_token,  # Keep the same refresh token
            expires_at=access_expires,
            user_id=str(user.id),
            username=user.username,
            email=user.email,
            role=user.role,
            permissions=["user:read", "wallet:read", "job:submit"] if user.role == "user" else ["admin:all"]
        )
        
    except JWTError:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Invalid refresh token"
        )

@app.get("/api/v1/auth/profile", response_model=UserProfile)
async def get_user_profile(current_user: User = Depends(get_current_active_user)):
    """Get current user profile information"""
    
    return UserProfile(
        id=str(current_user.id),
        username=current_user.username,
        email=current_user.email,
        role=current_user.role,
        is_active=current_user.is_active,
        created_at=current_user.created_at,
        updated_at=current_user.updated_at
    )

@app.put("/api/v1/auth/profile", response_model=UserProfile)
async def update_user_profile(
    user_update: UserUpdate,
    current_user: User = Depends(get_current_active_user),
    db: Session = Depends(get_db)
):
    """Update current user profile"""
    
    # Check if email is being changed and if it's already taken
    if user_update.email and user_update.email != current_user.email:
        existing_user = get_user_by_email(db, email=user_update.email)
        if existing_user and existing_user.id != current_user.id:
            raise HTTPException(
                status_code=400,
                detail="Email already registered"
            )
    
    # Check if username is being changed and if it's already taken
    if user_update.username and user_update.username != current_user.username:
        existing_user = get_user_by_username(db, username=user_update.username)
        if existing_user and existing_user.id != current_user.id:
            raise HTTPException(
                status_code=400,
                detail="Username already taken"
            )
    
    try:
        from app.db.crud.crud_user import update_user
        updated_user = update_user(db, db_user=current_user, user_in=user_update)
        
        return UserProfile(
            id=str(updated_user.id),
            username=updated_user.username,
            email=updated_user.email,
            role=updated_user.role,
            is_active=updated_user.is_active,
            created_at=updated_user.created_at,
            updated_at=updated_user.updated_at
        )
    except Exception as e:
        raise HTTPException(
            status_code=500,
            detail=f"Failed to update profile: {str(e)}"
        )

@app.post("/api/v1/auth/change-password")
async def change_password(
    password_data: PasswordChangeRequest,
    current_user: User = Depends(get_current_active_user),
    db: Session = Depends(get_db)
):
    """Change user password"""
    
    # Verify current password
    if not verify_password(password_data.current_password, current_user.hashed_password):
        raise HTTPException(
            status_code=400,
            detail="Current password is incorrect"
        )
    
    # Validate new password
    if len(password_data.new_password) < 8:
        raise HTTPException(
            status_code=400,
            detail="New password must be at least 8 characters long"
        )
    
    try:
        # Update password
        current_user.hashed_password = get_password_hash(password_data.new_password)
        current_user.updated_at = datetime.utcnow()
        db.commit()
        db.refresh(current_user)
        
        return {"message": "Password changed successfully"}
    except Exception as e:
        db.rollback()
        raise HTTPException(
            status_code=500,
            detail=f"Failed to change password: {str(e)}"
        )

@app.post("/api/v1/auth/logout")
async def logout(current_user: User = Depends(get_current_active_user)):
    """Logout user (client-side token removal)"""
    # In a real-world scenario, you might want to maintain a blacklist of tokens
    # For now, we'll just return a success message
    return {"message": "Successfully logged out"}

# Admin Routes
@app.get("/api/v1/auth/users", response_model=List[UserProfile])
async def list_users(
    skip: int = 0,
    limit: int = 100,
    current_user: User = Depends(get_current_active_user),
    db: Session = Depends(get_db)
):
    """List all users (admin only)"""
    
    if current_user.role != "admin":
        raise HTTPException(
            status_code=403,
            detail="Insufficient permissions"
        )
    
    users = get_users(db, skip=skip, limit=limit)
    return [
        UserProfile(
            id=str(user.id),
            username=user.username,
            email=user.email,
            role=user.role,
            is_active=user.is_active,
            created_at=user.created_at,
            updated_at=user.updated_at
        )
        for user in users
    ]

# Health check
@app.get("/health")
async def health_check():
    """Health check endpoint"""
    return {
        "status": "healthy",
        "service": "DanteGPU Authentication Service",
        "timestamp": datetime.utcnow().isoformat(),
        "version": "1.0.0"
    }

@app.get("/")
async def root():
    """Root endpoint"""
    return {
        "message": "DanteGPU Authentication Service",
        "version": "1.0.0",
        "docs": "/docs",
        "health": "/health"
    }

# Database initialization
@app.on_event("startup")
async def startup_event():
    """Initialize database and create default admin user"""
    try:
        from app.db.session import engine
        from app.db.base_class import Base
        
        # Create tables
        Base.metadata.create_all(bind=engine)
        
        # Create default admin user if it doesn't exist
        db = next(get_db())
        try:
            admin_user = get_user_by_username(db, username="admin")
            if not admin_user:
                admin_create = UserCreate(
                    username="admin",
                    email="admin@dantegpu.com",
                    password="admin123",
                    role="admin"
                )
                create_user(db, user_in=admin_create)
                print("✅ Default admin user created: admin / admin123")
            
            # Create demo user if it doesn't exist
            demo_user = get_user_by_username(db, username="demo")
            if not demo_user:
                demo_create = UserCreate(
                    username="demo",
                    email="demo@dantegpu.com",
                    password="demo12345",
                    role="user"
                )
                create_user(db, user_in=demo_create)
                print("✅ Demo user created: demo / demo12345")
                
        finally:
            db.close()
            
        print("🚀 DanteGPU Authentication Service started successfully!")
        
    except Exception as e:
        print(f"❌ Failed to initialize database: {e}")

if __name__ == "__main__":
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=8090,
        reload=True,
        log_level="info"
    ) 