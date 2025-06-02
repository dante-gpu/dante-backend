#!/usr/bin/env python3
"""Initialize database with tables and default users."""

from app.db.session import engine, SessionLocal
from app.db.base import Base
from app.db.models.user import User
from app.db.crud.crud_user import create_user
from app.schemas.user import UserCreate

def init_database():
    # Create all tables
    Base.metadata.create_all(bind=engine)
    print("Database tables created!")
    
    # Create default users
    db = SessionLocal()
    try:
        # Create admin user
        admin_data = UserCreate(
            username='admin',
            email='admin@dantegpu.com',
            password='admin123',
            role='admin'
        )
        try:
            admin_user = create_user(db, user_in=admin_data)
            print(f"Admin user created: {admin_user.username}")
        except Exception as e:
            print(f"Admin user may already exist: {e}")
        
        # Create demo user
        demo_data = UserCreate(
            username='demo',
            email='demo@dantegpu.com', 
            password='demo123456',
            role='user'
        )
        try:
            demo_user = create_user(db, user_in=demo_data)
            print(f"Demo user created: {demo_user.username}")
        except Exception as e:
            print(f"Demo user may already exist: {e}")
            
    finally:
        db.close()
    
    print("Database initialization complete!")

if __name__ == "__main__":
    init_database() 