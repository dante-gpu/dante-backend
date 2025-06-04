import uuid
from sqlalchemy import Column, String, ForeignKey, DateTime, func, Text, Boolean, Index, UUID as pgUUID
from sqlalchemy.orm import relationship
from app.db.base_class import Base
import secrets

class UserApiKey(Base):
    __tablename__ = "user_api_keys"
    id = Column(pgUUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    user_id = Column(pgUUID(as_uuid=True), ForeignKey('users.id', ondelete='CASCADE'), nullable=False, index=True)
    key_name = Column(String(100), nullable=False)
    prefix = Column(String(8), unique=True, nullable=False)
    hashed_key = Column(String(255), nullable=False, unique=True)
    scopes = Column(Text, nullable=True)
    expires_at = Column(DateTime(timezone=True), nullable=True)
    last_used_at = Column(DateTime(timezone=True), nullable=True)
    is_active = Column(Boolean, default=True, nullable=False)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), onupdate=func.now(), server_default=func.now())

    user = relationship("User", back_populates="api_keys")

    __table_args__ = (Index('ix_user_api_keys_user_id_key_name', 'user_id', 'key_name', unique=True),)
    
    def __repr__(self):
        return f"<UserApiKey(user_id={self.user_id}, name='{self.key_name}', prefix='{self.prefix}')>"

    @staticmethod
    def generate_key_components(prefix_val="dsk"):
        visible_part = secrets.token_urlsafe(18)
        secret_part = secrets.token_urlsafe(24)
        full_key = f"{prefix_val}_{visible_part}_{secret_part}"
        key_prefix_to_store = f"{prefix_val}_{visible_part[:4]}"
        return full_key, key_prefix_to_store


class LoginHistory(Base):
    __tablename__ = "login_history"
    id = Column(pgUUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    user_id = Column(pgUUID(as_uuid=True), ForeignKey('users.id', ondelete='CASCADE'), nullable=False, index=True)
    login_timestamp = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    ip_address = Column(String(45), nullable=True)
    user_agent = Column(String(512), nullable=True)
    login_successful = Column(Boolean, nullable=False)
    failure_reason = Column(String(255), nullable=True)

    user = relationship("User")

    def __repr__(self):
        return f"<LoginHistory(user_id={self.user_id}, timestamp='{self.login_timestamp}', success={self.login_successful})>" 