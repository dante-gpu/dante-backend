# GPU Monitor Service Dockerfile for Dante GPU Rental Platform
FROM python:3.11-slim

# Set working directory
WORKDIR /app

# Set environment variables
ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1
ENV PYTHONPATH=/app

# Install system dependencies (similar to auth-service/dashboard_service)
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        build-essential \
        curl \
        libpq-dev \
    && rm -rf /var/lib/apt/lists/*

# Copy requirements from auth-service and install Python dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir --upgrade pip \
    && pip install --no-cache-dir -r requirements.txt

# Copy necessary files for the gpu_monitor service
COPY gpu_monitor_service.py .
# COPY app/ /app/app/ # Assuming shared modules are in 'app' - Removed as it was causing issues

# Create non-root user
RUN adduser --disabled-password --gecos '' appuser \
    && chown -R appuser:appuser /app
USER appuser

# Expose port for gpu_monitor service
EXPOSE 8092

# Health check (adjust if your service has a different health endpoint)
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8092/health || exit 1

# Run the gpu_monitor service application
CMD ["uvicorn", "gpu_monitor_service:app", "--host", "0.0.0.0", "--port", "8092"] 