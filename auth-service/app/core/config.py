from pydantic_settings import BaseSettings, SettingsConfigDict
from pydantic import AnyHttpUrl, EmailStr, validator
from typing import List, Optional
import os

class Settings(BaseSettings):
    # I need to define configuration variables here.
    # Pydantic Settings will automatically load them from environment variables
    # or a .env file.

    # --- Core Settings ---
    PROJECT_NAME: str = "Dante Auth Service"
    API_V1_STR: str = "/api/v1"
    SECRET_KEY: str = "c3ab8ff13720e8ad9047dd39466b3c8974e592c2fa383d4a3960714caef0c4f2"
    # I should use a secure random key generation method for production.
    # Example: openssl rand -hex 32

    # --- Security Settings (JWT) ---
    # Note: While the gateway primarily issues tokens, this service might need to validate them
    # or potentially issue specific tokens (e.g., password reset).
    ALGORITHM: str = "HS256"
    ACCESS_TOKEN_EXPIRE_MINUTES: int = 60 * 24 * 7 # 7 days

    # --- Database Settings ---
    DATABASE_URL: str = "postgresql+psycopg2://dante_user:dante_password@localhost:5432/dante_auth"
    # Example: postgresql+psycopg2://db_user:db_password@db_host:db_port/db_name

    # --- CORS Settings (if needed directly by this service) ---
    BACKEND_CORS_ORIGINS: List[AnyHttpUrl] = []

    @validator("BACKEND_CORS_ORIGINS", pre=True)
    def assemble_cors_origins(cls, v: str | List[str]) -> List[str] | str:
        # I need a validator to allow origins to be passed as a comma-separated string
        # in environment variables.
        if isinstance(v, str) and not v.startswith("["):
            return [i.strip() for i in v.split(",")]
        elif isinstance(v, (list, str)):
            return v
        raise ValueError(v)

    # --- Superuser Settings (for initial setup/admin) ---
    FIRST_SUPERUSER_EMAIL: Optional[EmailStr] = None
    FIRST_SUPERUSER_PASSWORD: Optional[str] = None

    # --- Model Configuration ---
    # I should configure Pydantic Settings to read from a .env file.
    model_config = SettingsConfigDict(env_file=".env", extra='ignore')

# I will create a single instance of the settings to be imported across the app.
settings = Settings() 