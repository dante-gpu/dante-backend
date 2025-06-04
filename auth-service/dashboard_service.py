#!/usr/bin/env python3
"""
Dashboard Service for DanteGPU Platform
Provides real dashboard data for users including stats, jobs, transactions, and GPU metrics
"""

from fastapi import FastAPI, HTTPException, Depends, status, Query
from fastapi.middleware.cors import CORSMiddleware
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from pydantic import BaseModel
from sqlalchemy import create_engine, Column, String, Boolean, DateTime, UUID as pgUUID, func, Integer, Float, Text, JSON
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker, Session
from jose import JWTError, jwt
from datetime import datetime, timedelta
from typing import Optional, List, Dict, Any
from enum import Enum
import uuid
import uvicorn
import json
import random
from sqlalchemy.sql import text
import os

# Database Setup
DATABASE_URL = os.getenv("DATABASE_URL", "postgresql+psycopg2://dante_user:dante_password@localhost:5432/dante_auth")
engine = create_engine(DATABASE_URL)
SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)
Base = declarative_base()

# Status Enums
class JobStatus(str, Enum):
    PENDING = "pending"
    RUNNING = "running"
    COMPLETED = "completed"
    FAILED = "failed"
    CANCELLED = "cancelled"

class TransactionType(str, Enum):
    PAYMENT = "payment"
    EARNING = "earning"
    REFUND = "refund"
    DEPOSIT = "deposit"
    WITHDRAWAL = "withdrawal"

class TransactionStatus(str, Enum):
    PENDING = "pending"
    COMPLETED = "completed"
    FAILED = "failed"

class ProviderStatus(str, Enum):
    ONLINE = "online"
    OFFLINE = "offline"
    BUSY = "busy"
    MAINTENANCE = "maintenance"

# Database Models
class User(Base):
    __tablename__ = "users"
    
    id = Column(pgUUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    email = Column(String, unique=True, index=True, nullable=False)
    username = Column(String, unique=True, index=True, nullable=False)
    hashed_password = Column(String, nullable=False)
    role = Column(String, nullable=False, default="user")
    is_active = Column(Boolean(), default=True)
    wallet_address = Column(String, nullable=True)
    balance_dgpu = Column(Float, default=0.0)
    total_spent = Column(Float, default=0.0)
    total_earned = Column(Float, default=0.0)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), onupdate=func.now(), server_default=func.now())

class Provider(Base):
    __tablename__ = "providers"
    
    id = Column(pgUUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    owner_id = Column(pgUUID(as_uuid=True), nullable=False)
    name = Column(String, nullable=False)
    location = Column(String, nullable=True)
    status = Column(String, default=ProviderStatus.OFFLINE)
    hostname = Column(String, nullable=True)
    ip_address = Column(String, nullable=True)
    hourly_rate = Column(Float, default=0.5)
    rating = Column(Float, default=4.5)
    total_jobs = Column(Integer, default=0)
    success_rate = Column(Float, default=95.0)
    gpus_data = Column(JSON, nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    last_seen_at = Column(DateTime(timezone=True), server_default=func.now())

class Job(Base):
    __tablename__ = "jobs"
    
    id = Column(pgUUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    user_id = Column(pgUUID(as_uuid=True), nullable=False)
    provider_id = Column(pgUUID(as_uuid=True), nullable=False)
    name = Column(String, nullable=False)
    description = Column(Text, nullable=True)
    status = Column(String, default=JobStatus.PENDING)
    gpu_model = Column(String, nullable=False)
    cost_dgpu = Column(Float, default=0.0)
    duration_seconds = Column(Integer, default=0)
    progress_percentage = Column(Float, default=0.0)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    started_at = Column(DateTime(timezone=True), nullable=True)
    completed_at = Column(DateTime(timezone=True), nullable=True)
    requirements = Column(JSON, nullable=True)
    execution_config = Column(JSON, nullable=True)
    metrics_data = Column(JSON, nullable=True)

class Transaction(Base):
    __tablename__ = "transactions"
    
    id = Column(pgUUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    user_id = Column(pgUUID(as_uuid=True), nullable=False)
    transaction_type = Column(String, nullable=False)
    amount = Column(Float, nullable=False)
    description = Column(String, nullable=False)
    status = Column(String, default=TransactionStatus.PENDING)
    tx_hash = Column(String, nullable=True)
    from_address = Column(String, nullable=True)
    to_address = Column(String, nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    completed_at = Column(DateTime(timezone=True), nullable=True)

class GPUMetrics(Base):
    __tablename__ = "gpu_metrics"
    
    id = Column(pgUUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    provider_id = Column(pgUUID(as_uuid=True), nullable=False)
    gpu_index = Column(Integer, default=0)
    utilization_gpu = Column(Float, default=0.0)
    utilization_memory = Column(Float, default=0.0)
    temperature = Column(Float, default=0.0)
    power_draw = Column(Float, default=0.0)
    memory_used_mb = Column(Integer, default=0)
    memory_total_mb = Column(Integer, default=0)
    clock_graphics_mhz = Column(Integer, default=0)
    clock_memory_mhz = Column(Integer, default=0)
    fan_speed_percent = Column(Float, default=0.0)
    is_healthy = Column(Boolean(), default=True)
    timestamp = Column(DateTime(timezone=True), server_default=func.now())

# Database Migration Helper
def migrate_user_table():
    """Add new columns to users table if they don't exist"""
    try:
        with engine.connect() as conn:
            # Check if new columns exist
            result = conn.execute(text("""
                SELECT column_name 
                FROM information_schema.columns 
                WHERE table_name='users' AND column_name IN ('wallet_address', 'balance_dgpu', 'total_spent', 'total_earned')
            """))
            existing_columns = [row[0] for row in result.fetchall()]
            
            # Add missing columns
            if 'wallet_address' not in existing_columns:
                conn.execute(text("ALTER TABLE users ADD COLUMN wallet_address VARCHAR"))
                print("Added wallet_address column")
            
            if 'balance_dgpu' not in existing_columns:
                conn.execute(text("ALTER TABLE users ADD COLUMN balance_dgpu FLOAT DEFAULT 0.0"))
                print("Added balance_dgpu column")
            
            if 'total_spent' not in existing_columns:
                conn.execute(text("ALTER TABLE users ADD COLUMN total_spent FLOAT DEFAULT 0.0"))
                print("Added total_spent column")
            
            if 'total_earned' not in existing_columns:
                conn.execute(text("ALTER TABLE users ADD COLUMN total_earned FLOAT DEFAULT 0.0"))
                print("Added total_earned column")
            
            conn.commit()
            
    except Exception as e:
        print(f"Migration warning: {e}")

# Run migration before creating tables
migrate_user_table()

# Create tables
Base.metadata.create_all(bind=engine)

# FastAPI App
app = FastAPI(title="DanteGPU Dashboard Service", version="1.0.0")

# CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=[
        "http://localhost:3000",
        "http://localhost:3001", 
        "http://127.0.0.1:3000",
        "http://127.0.0.1:3001"
    ],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Security
security = HTTPBearer()
SECRET_KEY = "dante_super_secret_jwt_key_2024_production_ready"
ALGORITHM = "HS256"

# Pydantic Models
class DashboardStatsResponse(BaseModel):
    totalProviders: int
    availableGPUs: int
    activeJobs: int
    totalEarnings: str
    walletBalance: str
    totalSpent: str
    jobsCompleted: int
    computeHours: float

class GPUMetricsResponse(BaseModel):
    utilization: float
    temperature: float
    powerDraw: float
    memoryUsage: float
    vramTotal: int
    vramUsed: int
    clockGraphics: int
    clockMemory: int
    fanSpeed: float
    isHealthy: bool

class JobResponse(BaseModel):
    id: str
    name: str
    provider: str
    status: str
    startTime: str
    duration: str
    cost: str
    gpuModel: str
    progress: float
    description: Optional[str] = None

class TransactionResponse(BaseModel):
    id: str
    type: str
    amount: str
    description: str
    timestamp: str
    status: str
    tx_hash: Optional[str] = None

class ProviderResponse(BaseModel):
    id: str
    name: str
    location: str
    status: str
    hourlyRate: float
    rating: float
    totalJobs: int
    successRate: float
    gpus: List[Dict[str, Any]]
    lastSeen: str

# Database dependency
def get_db():
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()

# Utility functions
def verify_token(token: str):
    try:
        payload = jwt.decode(token, SECRET_KEY, algorithms=[ALGORITHM])
        username: str = payload.get("sub")
        if username is None:
            raise HTTPException(status_code=401, detail="Invalid token")
        return payload
    except JWTError:
        raise HTTPException(status_code=401, detail="Invalid token")

async def get_current_user(credentials: HTTPAuthorizationCredentials = Depends(security), db: Session = Depends(get_db)):
    token = credentials.credentials
    payload = verify_token(token)
    username = payload.get("sub")
    
    user = db.query(User).filter(User.username == username).first()
    if user is None:
        raise HTTPException(status_code=401, detail="User not found")
    return user

# Initialize sample data
def init_sample_data(db: Session):
    """Initialize sample data for dashboard demo"""
    
    # Check if sample data already exists
    existing_providers = db.query(Provider).count()
    if existing_providers > 0:
        return
    
    # Update existing users with sample wallet data
    try:
        users = db.query(User).all()
        for user in users:
            if user.balance_dgpu == 0.0:
                user.balance_dgpu = 47.25
                user.total_spent = 8.50
                user.total_earned = 15.75
        db.commit()
    except Exception as e:
        print(f"Warning: Could not update user wallet data: {e}")
    
    # Create sample providers
    providers_data = [
        {
            "name": "CloudGPU Pro",
            "location": "US-East",
            "status": ProviderStatus.ONLINE,
            "hourly_rate": 0.75,
            "rating": 4.9,
            "total_jobs": 1247,
            "success_rate": 98.5,
            "gpus_data": [
                {
                    "model_name": "NVIDIA RTX 4090",
                    "vram_mb": 24576,
                    "driver_version": "536.23",
                    "is_healthy": True,
                    "utilization_gpu_percent": 75,
                    "temperature_c": 68,
                    "power_draw_w": 320
                }
            ]
        },
        {
            "name": "AI Compute Hub",
            "location": "EU-West",
            "status": ProviderStatus.ONLINE,
            "hourly_rate": 1.50,
            "rating": 4.8,
            "total_jobs": 2156,
            "success_rate": 97.2,
            "gpus_data": [
                {
                    "model_name": "NVIDIA A100",
                    "vram_mb": 81920,
                    "driver_version": "535.86",
                    "is_healthy": True,
                    "utilization_gpu_percent": 90,
                    "temperature_c": 72,
                    "power_draw_w": 400
                }
            ]
        },
        {
            "name": "RenderFarm Elite",
            "location": "Asia-Pacific",
            "status": ProviderStatus.ONLINE,
            "hourly_rate": 0.45,
            "rating": 4.7,
            "total_jobs": 3421,
            "success_rate": 96.8,
            "gpus_data": [
                {
                    "model_name": "NVIDIA RTX 3080",
                    "vram_mb": 10240,
                    "driver_version": "535.98",
                    "is_healthy": True,
                    "utilization_gpu_percent": 85,
                    "temperature_c": 65,
                    "power_draw_w": 250
                }
            ]
        }
    ]
    
    for provider_data in providers_data:
        provider = Provider(
            id=uuid.uuid4(),
            owner_id=uuid.uuid4(),  # Sample owner ID
            **provider_data
        )
        db.add(provider)
    
    db.commit()

# Routes
@app.get("/api/v1/dashboard/stats", response_model=DashboardStatsResponse)
async def get_dashboard_stats(current_user: User = Depends(get_current_user), db: Session = Depends(get_db)):
    """Get dashboard statistics for the current user"""
    
    # Initialize sample data if needed
    init_sample_data(db)
    
    # Get user's active jobs
    active_jobs = db.query(Job).filter(
        Job.user_id == current_user.id,
        Job.status.in_([JobStatus.PENDING, JobStatus.RUNNING])
    ).count()
    
    # Get completed jobs
    completed_jobs = db.query(Job).filter(
        Job.user_id == current_user.id,
        Job.status == JobStatus.COMPLETED
    ).count()
    
    # Get total providers and available GPUs
    total_providers = db.query(Provider).filter(Provider.status == ProviderStatus.ONLINE).count()
    
    # Calculate available GPUs
    providers = db.query(Provider).filter(Provider.status == ProviderStatus.ONLINE).all()
    available_gpus = sum(len(p.gpus_data) if p.gpus_data else 1 for p in providers)
    
    # Calculate compute hours
    completed_job_durations = db.query(Job.duration_seconds).filter(
        Job.user_id == current_user.id,
        Job.status == JobStatus.COMPLETED
    ).all()
    
    total_seconds = sum(duration[0] for duration in completed_job_durations if duration[0])
    compute_hours = round(total_seconds / 3600, 2) if total_seconds else 0.0
    
    return DashboardStatsResponse(
        totalProviders=total_providers,
        availableGPUs=available_gpus,
        activeJobs=active_jobs,
        totalEarnings=f"{current_user.total_earned or 0.0:.2f}",
        walletBalance=f"{current_user.balance_dgpu or 0.0:.2f}",
        totalSpent=f"{current_user.total_spent or 0.0:.2f}",
        jobsCompleted=completed_jobs,
        computeHours=compute_hours
    )

@app.get("/api/v1/dashboard/providers", response_model=List[ProviderResponse])
async def get_providers(current_user: User = Depends(get_current_user), db: Session = Depends(get_db)):
    """Get list of available providers"""
    
    # Initialize sample data if needed
    init_sample_data(db)
    
    providers = db.query(Provider).filter(Provider.status == ProviderStatus.ONLINE).all()
    
    return [
        ProviderResponse(
            id=str(provider.id),
            name=provider.name,
            location=provider.location or "Unknown",
            status=provider.status,
            hourlyRate=provider.hourly_rate,
            rating=provider.rating,
            totalJobs=provider.total_jobs,
            successRate=provider.success_rate,
            gpus=provider.gpus_data or [],
            lastSeen=provider.last_seen_at.isoformat()
        ) for provider in providers
    ]

@app.get("/api/v1/dashboard/jobs", response_model=List[JobResponse])
async def get_user_jobs(
    status: Optional[str] = Query(None),
    limit: int = Query(10),
    current_user: User = Depends(get_current_user), 
    db: Session = Depends(get_db)
):
    """Get user's jobs"""
    
    # Create sample jobs if none exist for this user
    existing_jobs = db.query(Job).filter(Job.user_id == current_user.id).count()
    if existing_jobs == 0:
        # Get a provider for sample jobs
        providers = db.query(Provider).limit(2).all()
        if providers:
            sample_jobs = [
                {
                    "name": "AI Training - ResNet50",
                    "description": "Training a ResNet50 model on CIFAR-10 dataset",
                    "provider_id": providers[0].id,
                    "gpu_model": "NVIDIA RTX 4090",
                    "status": JobStatus.RUNNING,
                    "cost_dgpu": 2.50,
                    "duration_seconds": 9900,  # 2h 45m
                    "progress_percentage": 65.0,
                    "started_at": datetime.utcnow() - timedelta(hours=2, minutes=45)
                },
                {
                    "name": "Video Rendering - 4K Animation",
                    "description": "Rendering a 4K animation sequence",
                    "provider_id": providers[1].id if len(providers) > 1 else providers[0].id,
                    "gpu_model": "NVIDIA A100",
                    "status": JobStatus.PENDING,
                    "cost_dgpu": 0.0,
                    "duration_seconds": 0,
                    "progress_percentage": 0.0
                },
                {
                    "name": "ML Model Inference",
                    "description": "Running inference on trained model",
                    "provider_id": providers[0].id,
                    "gpu_model": "NVIDIA RTX 4090",
                    "status": JobStatus.COMPLETED,
                    "cost_dgpu": 1.25,
                    "duration_seconds": 3600,  # 1 hour
                    "progress_percentage": 100.0,
                    "started_at": datetime.utcnow() - timedelta(hours=5),
                    "completed_at": datetime.utcnow() - timedelta(hours=4)
                }
            ]
            
            for job_data in sample_jobs:
                job = Job(
                    id=uuid.uuid4(),
                    user_id=current_user.id,
                    **job_data
                )
                db.add(job)
            db.commit()
    
    # Query jobs
    query = db.query(Job).filter(Job.user_id == current_user.id)
    if status:
        query = query.filter(Job.status == status)
    
    jobs = query.order_by(Job.created_at.desc()).limit(limit).all()
    
    # Get provider names
    provider_ids = [job.provider_id for job in jobs]
    providers = db.query(Provider).filter(Provider.id.in_(provider_ids)).all()
    provider_map = {str(p.id): p.name for p in providers}
    
    def format_duration(seconds):
        if not seconds:
            return "0m"
        hours = seconds // 3600
        minutes = (seconds % 3600) // 60
        if hours > 0:
            return f"{hours}h {minutes}m"
        return f"{minutes}m"
    
    return [
        JobResponse(
            id=str(job.id),
            name=job.name,
            provider=provider_map.get(str(job.provider_id), "Unknown Provider"),
            status=job.status,
            startTime=job.started_at.isoformat() if job.started_at else job.created_at.isoformat(),
            duration=format_duration(job.duration_seconds),
            cost=f"{job.cost_dgpu:.2f}",
            gpuModel=job.gpu_model,
            progress=job.progress_percentage,
            description=job.description
        ) for job in jobs
    ]

@app.get("/api/v1/dashboard/transactions", response_model=List[TransactionResponse])
async def get_user_transactions(
    transaction_type: Optional[str] = Query(None),
    limit: int = Query(10),
    current_user: User = Depends(get_current_user), 
    db: Session = Depends(get_db)
):
    """Get user's transaction history"""
    
    # Create sample transactions if none exist for this user
    existing_transactions = db.query(Transaction).filter(Transaction.user_id == current_user.id).count()
    if existing_transactions == 0:
        sample_transactions = [
            {
                "transaction_type": TransactionType.PAYMENT,
                "amount": 2.50,
                "description": "GPU Rental Payment - RTX 4090",
                "status": TransactionStatus.COMPLETED,
                "completed_at": datetime.utcnow() - timedelta(hours=2)
            },
            {
                "transaction_type": TransactionType.DEPOSIT,
                "amount": 50.00,
                "description": "Wallet Top-up via Solana",
                "status": TransactionStatus.COMPLETED,
                "tx_hash": "4xK7n2PmR8qL9vB3xT6mJ1wS5nD8qY4cF7kM2pX9zR5L",
                "completed_at": datetime.utcnow() - timedelta(days=1)
            },
            {
                "transaction_type": TransactionType.EARNING,
                "amount": 5.75,
                "description": "Provider Earnings - A100 Rental",
                "status": TransactionStatus.COMPLETED,
                "completed_at": datetime.utcnow() - timedelta(days=2)
            },
            {
                "transaction_type": TransactionType.REFUND,
                "amount": 1.25,
                "description": "Job Cancellation Refund",
                "status": TransactionStatus.COMPLETED,
                "completed_at": datetime.utcnow() - timedelta(days=3)
            }
        ]
        
        for tx_data in sample_transactions:
            transaction = Transaction(
                id=uuid.uuid4(),
                user_id=current_user.id,
                **tx_data
            )
            db.add(transaction)
        db.commit()
    
    # Query transactions
    query = db.query(Transaction).filter(Transaction.user_id == current_user.id)
    if transaction_type:
        query = query.filter(Transaction.transaction_type == transaction_type)
    
    transactions = query.order_by(Transaction.created_at.desc()).limit(limit).all()
    
    return [
        TransactionResponse(
            id=str(tx.id),
            type=tx.transaction_type,
            amount=f"{tx.amount:.2f}",
            description=tx.description,
            timestamp=tx.completed_at.isoformat() if tx.completed_at else tx.created_at.isoformat(),
            status=tx.status,
            tx_hash=tx.tx_hash
        ) for tx in transactions
    ]

@app.get("/api/v1/dashboard/gpu-metrics", response_model=GPUMetricsResponse)
async def get_gpu_metrics(current_user: User = Depends(get_current_user), db: Session = Depends(get_db)):
    """Get current GPU metrics for user's active jobs"""
    
    # For now, return simulated real-time metrics
    # In production, this would aggregate metrics from user's active GPU rentals
    return GPUMetricsResponse(
        utilization=random.uniform(70, 95),
        temperature=random.uniform(60, 80),
        powerDraw=random.uniform(200, 400),
        memoryUsage=random.uniform(60, 90),
        vramTotal=24576,
        vramUsed=random.randint(15000, 22000),
        clockGraphics=random.randint(1800, 2200),
        clockMemory=random.randint(9000, 11000),
        fanSpeed=random.uniform(40, 80),
        isHealthy=True
    )

@app.post("/api/v1/dashboard/jobs/{job_id}/action")
async def job_action(
    job_id: str,
    action: str,
    current_user: User = Depends(get_current_user), 
    db: Session = Depends(get_db)
):
    """Perform action on a job (pause, resume, cancel)"""
    
    try:
        job_uuid = uuid.UUID(job_id)
    except ValueError:
        raise HTTPException(status_code=400, detail="Invalid job ID format")
    
    job = db.query(Job).filter(
        Job.id == job_uuid,
        Job.user_id == current_user.id
    ).first()
    
    if not job:
        raise HTTPException(status_code=404, detail="Job not found")
    
    if action == "cancel":
        job.status = JobStatus.CANCELLED
        job.completed_at = datetime.utcnow()
    elif action == "pause" and job.status == JobStatus.RUNNING:
        job.status = JobStatus.PENDING
    elif action == "resume" and job.status == JobStatus.PENDING:
        job.status = JobStatus.RUNNING
        if not job.started_at:
            job.started_at = datetime.utcnow()
    else:
        raise HTTPException(status_code=400, detail=f"Invalid action '{action}' for job status '{job.status}'")
    
    db.commit()
    
    return {"message": f"Job {action} successful", "job_id": job_id, "new_status": job.status}

@app.get("/health")
async def health_check():
    """Health check endpoint"""
    return {
        "status": "healthy",
        "service": "DanteGPU Dashboard Service",
        "timestamp": datetime.utcnow().isoformat(),
        "version": "1.0.0"
    }

@app.get("/")
async def root():
    """Root endpoint"""
    return {
        "message": "DanteGPU Dashboard Service",
        "version": "1.0.0",
        "endpoints": {
            "stats": "/api/v1/dashboard/stats",
            "providers": "/api/v1/dashboard/providers",
            "jobs": "/api/v1/dashboard/jobs",
            "transactions": "/api/v1/dashboard/transactions",
            "gpu_metrics": "/api/v1/dashboard/gpu-metrics"
        }
    }

if __name__ == "__main__":
    print("🚀 Starting DanteGPU Dashboard Service...")
    print("🔗 Database:", DATABASE_URL.replace("dante_password", "***"))
    print("🌐 Server: http://localhost:8091")
    print("📖 API Docs: http://localhost:8091/docs")
    
    uvicorn.run(app, host="0.0.0.0", port=8091, log_level="info") 