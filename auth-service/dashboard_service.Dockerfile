# Dashboard Service Dockerfile for Dante GPU Rental Platform
FROM python:3.11-slim

# Set working directory
WORKDIR /app

# Set environment variables
ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1
ENV PYTHONPATH=/app

# Install system dependencies from auth-service Dockerfile (libpq-dev for psycopg2)
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        build-essential \
        curl \
        libpq-dev \
    && rm -rf /var/lib/apt/lists/*

# Copy requirements from auth-service and install Python dependencies
# Assuming dashboard_service shares requirements with auth_service
COPY requirements.txt .
RUN pip install --no-cache-dir --upgrade pip \
    && pip install --no-cache-dir -r requirements.txt

# Copy only the necessary files for the dashboard service
# This includes the service script itself, and any shared modules like core, db, schemas
COPY dashboard_service.py .

# Create non-root user (similar to auth-service Dockerfile)
RUN adduser --disabled-password --gecos '' appuser \
    && chown -R appuser:appuser /app
USER appuser

# Expose port for dashboard service
EXPOSE 8091

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8091/health || exit 1

# Run the dashboard service application
# Make sure dashboard_service.py uses uvicorn to run or is a FastAPI app object
CMD ["uvicorn", "dashboard_service:app", "--host", "0.0.0.0", "--port", "8091"] 