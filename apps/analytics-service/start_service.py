#!/usr/bin/env python3
"""
Analytics Service - Final Working Startup Script
"""
import sys
import os
from pathlib import Path

# Add app directory to Python path
current_dir = Path(__file__).parent
app_dir = current_dir / "app"
sys.path.insert(0, str(app_dir))

# Load environment variables
try:
    from dotenv import load_dotenv
    load_dotenv()
except ImportError:
    print("⚠️  python-dotenv not available, using environment variables as-is")

def main():
    try:
        # Import the app
        from app.main import app
        print("✅ Analytics Service modules loaded successfully!")
        
        # Start server
        import uvicorn
        
        print("\n🚀 Analytics Service with Comprehensive Swagger Documentation")
        print("=" * 70)
        print("📚 Swagger UI (Interactive):  http://localhost:8085/docs")
        print("📖 ReDoc Documentation:      http://localhost:8085/redoc")
        print("🏥 Health Check:             http://localhost:8085/health")
        print("🔗 API Root:                 http://localhost:8085/")
        print("")
        print("🎯 API Endpoints:")
        print("  📊 Analytics API:         /api/analytics/")
        print("  📋 Reports & Export:      /api/reports/")
        print("  🤖 AI Insights:           /api/insights/")
        print("  🔴 Live Streaming:        /api/streaming/")
        print("")
        print("🔐 Authentication: Use Bearer token 'dev-token' for testing")
        print("🌟 All endpoints documented with comprehensive Swagger schemas")
        print("=" * 70)
        print("")
        
        # Run the server
        uvicorn.run(
            app,
            host="0.0.0.0",
            port=8085,
            reload=False,
            log_level="info"
        )
        
    except Exception as e:
        print(f"❌ Error starting service: {e}")
        print("\n🔄 Attempting to start with basic configuration...")
        
        # Fallback to simple app
        from fastapi import FastAPI
        from fastapi.middleware.cors import CORSMiddleware
        
        app = FastAPI(
            title="Analytics Service",
            description="Analytics Service API with Swagger Documentation",
            version="1.0.0",
            docs_url="/docs",
            redoc_url="/redoc"
        )
        
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
                "message": "Analytics Service is running in fallback mode",
                "status": "success",
                "documentation": "/docs"
            }
        
        @app.get("/health")
        async def health():
            return {"status": "healthy", "service": "analytics-service"}
        
        import uvicorn
        print("🚀 Starting Analytics Service in fallback mode...")
        uvicorn.run(app, host="0.0.0.0", port=8085)

if __name__ == "__main__":
    main()
