"""
Simple FastAPI test for Analytics Service
"""
import sys
import os

# Add the app directory to the Python path
sys.path.append(os.path.join(os.path.dirname(__file__), 'app'))

try:
    from fastapi import FastAPI
    from fastapi.responses import JSONResponse
    
    app = FastAPI(
        title="Analytics Service",
        description="Analytics Service API",
        version="1.0.0",
        docs_url="/docs",
        redoc_url="/redoc"
    )
    
    @app.get("/")
    async def root():
        return {"message": "Analytics Service is running", "status": "success"}
    
    @app.get("/health")
    async def health_check():
        return {
            "status": "healthy",
            "service": "analytics-service",
            "version": "1.0.0"
        }
    
    print("FastAPI app created successfully!")
    
except ImportError as e:
    print(f"Import error: {e}")
    print("Please install dependencies: pip install fastapi uvicorn")
    sys.exit(1)

if __name__ == "__main__":
    import uvicorn
    print("Starting Analytics Service...")
    uvicorn.run(app, host="0.0.0.0", port=8084)
