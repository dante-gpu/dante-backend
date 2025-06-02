#!/usr/bin/env python3
"""
Script to create database tables for Auth Service
"""

import sys
import os

# Add auth-service directory to Python path
sys.path.append(os.path.dirname(os.path.abspath(__file__)))

try:
    from app.db.session import engine
    from app.db.models import user
    from app.db.base_class import Base
    
    print("Creating database tables for Auth Service...")
    
    # Create all tables
    Base.metadata.create_all(bind=engine)
    
    print("✅ Database tables created successfully!")
    print("📊 Tables created:")
    print("   - users")
    
except Exception as e:
    print(f"❌ Error creating tables: {e}")
    sys.exit(1) 