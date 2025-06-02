#!/usr/bin/env python3
"""
Script to create a test user for Auth Service
"""

import sys
import os
from datetime import datetime

# Add auth-service directory to Python path
sys.path.append(os.path.dirname(os.path.abspath(__file__)))

try:
    from app.db.session import SessionLocal
    from app.db.models.user import User
    from app.core.security import get_password_hash
    import uuid
    
    # Create database session
    db = SessionLocal()
    
    # Check if test user already exists
    existing_user = db.query(User).filter(User.username == "testuser").first()
    if existing_user:
        print("⚠️  Test user already exists!")
        print(f"   Username: {existing_user.username}")
        print(f"   Email: {existing_user.email}")
        db.close()
        sys.exit(0)
    
    print("Creating test user...")
    
    # Create test user
    test_user = User(
        id=uuid.uuid4(),
        username="testuser",
        email="test@example.com", 
        hashed_password=get_password_hash("Test123!"),
        role="user",
        is_active=True,
        created_at=datetime.utcnow(),
        updated_at=datetime.utcnow()
    )
    
    # Add and commit
    db.add(test_user)
    db.commit()
    db.refresh(test_user)
    
    print("✅ Test user created successfully!")
    print(f"   ID: {test_user.id}")
    print(f"   Username: {test_user.username}")
    print(f"   Email: {test_user.email}")
    print(f"   Role: {test_user.role}")
    print(f"   Active: {test_user.is_active}")
    
    db.close()
    
except Exception as e:
    print(f"❌ Error creating test user: {e}")
    sys.exit(1) 