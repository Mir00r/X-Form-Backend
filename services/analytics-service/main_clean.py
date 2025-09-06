"""
Analytics Service - Main Application Entry Point
Enhanced with comprehensive Swagger documentation following industry best practices
"""
import sys
import os
from pathlib import Path

# Add the app directory to the Python path
current_dir = Path(__file__).parent
app_dir = current_dir / "app"
sys.path.insert(0, str(app_dir))

try:
    from app.main import app
    print("✅ Analytics Service modules loaded successfully!")
except ImportError as e:
    print(f"❌ Import error: {e}")
    print("📝 Creating simplified app for testing...")
    
    # Fallback to simple app if modules fail
    from fastapi import FastAPI
    from fastapi.middleware.cors import CORSMiddleware
    
    app = FastAPI(
        title="Analytics Service",
        description="Analytics Service API with Swagger Documentation",
        version="1.0.0",
        docs_url="/docs",
        redoc_url="/redoc"
    )
    
    # CORS middleware
    app.add_middleware(
        CORSMiddleware,
        allow_origins=["*"],
        allow_credentials=True,
        allow_methods=["*"],
        allow_headers=["*"],
    )
    
    @app.get("/")
    async def root():
        return {
            "message": "Analytics Service is running", 
            "status": "success",
            "documentation": "/docs"
        }
    
    @app.get("/health")
    async def health_check():
        return {
            "status": "healthy",
            "service": "analytics-service",
            "version": "1.0.0"
        }

if __name__ == "__main__":
    import uvicorn
    
    print("🚀 Starting Analytics Service...")
    print("📚 Swagger Documentation: http://localhost:8084/docs")
    print("📖 ReDoc Documentation: http://localhost:8084/redoc")
    print("🏥 Health Check: http://localhost:8084/health")
    
    uvicorn.run(
        app,
        host="0.0.0.0",
        port=8084,
        reload=False
    )
