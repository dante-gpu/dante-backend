from sqlalchemy import Column, String, Boolean, DateTime, func, UUID as pgUUID
from sqlalchemy.orm import relationship
import uuid

from app.db.base_class import Base # I need to import the Declarative Base

class User(Base):
    # I should define the table name.
    __tablename__ = "users"

    # I need to define the columns for the users table.
    id = Column(pgUUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    email = Column(String, unique=True, index=True, nullable=False)
    username = Column(String, unique=True, index=True, nullable=False)
    hashed_password = Column(String, nullable=False)
    role = Column(String, nullable=False, default="user", index=True)
    is_active = Column(Boolean(), default=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), onupdate=func.now(), server_default=func.now())

    # I could add relationships here later if needed, e.g., to roles or profiles.
    # items = relationship("Item", back_populates="owner")

    def __repr__(self):
        # I should add a representation for easier debugging.
        return f"<User(id={self.id}, email='{self.email}', username='{self.username}', role='{self.role}')>" 