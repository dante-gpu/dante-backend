import uuid
from sqlalchemy import Column, String, ForeignKey, DateTime, Table, func, UUID as pgUUID
from sqlalchemy.orm import relationship
from app.db.base_class import Base

# Association table for User and Role (many-to-many)
user_roles_table = Table('user_roles', Base.metadata,
    Column('user_id', pgUUID(as_uuid=True), ForeignKey('users.id', ondelete='CASCADE'), primary_key=True),
    Column('role_id', pgUUID(as_uuid=True), ForeignKey('roles.id', ondelete='CASCADE'), primary_key=True)
)

# Association table for Role and Permission (many-to-many)
role_permissions_table = Table('role_permissions', Base.metadata,
    Column('role_id', pgUUID(as_uuid=True), ForeignKey('roles.id', ondelete='CASCADE'), primary_key=True),
    Column('permission_id', pgUUID(as_uuid=True), ForeignKey('permissions.id', ondelete='CASCADE'), primary_key=True)
)

class Role(Base):
    __tablename__ = "roles"
    id = Column(pgUUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    name = Column(String(50), unique=True, index=True, nullable=False) # e.g., ROLE_USER, ROLE_ADMIN, ROLE_PROVIDER
    description = Column(String(255), nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), onupdate=func.now(), server_default=func.now())

    permissions = relationship("Permission", secondary=role_permissions_table, back_populates="roles", lazy="selectin")
    users = relationship("User", secondary=user_roles_table, back_populates="roles", lazy="selectin")

    def __repr__(self):
        return f"<Role(id={self.id}, name='{self.name}')>"

class Permission(Base):
    __tablename__ = "permissions"
    id = Column(pgUUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    name = Column(String(100), unique=True, index=True, nullable=False) # e.g., users:create, users:read_all, providers:manage_own
    description = Column(String(255), nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now())

    roles = relationship("Role", secondary=role_permissions_table, back_populates="permissions", lazy="selectin")

    def __repr__(self):
        return f"<Permission(id={self.id}, name='{self.name}')>" 