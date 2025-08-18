# X-Form Backend - Google Forms-like Survey Builder

A microservices-based backend system for building and managing surveys with real-time collaboration features.

## Architecture Overview

```
┌───────────────────────────┐
│        Frontend (React)   │
│  - Calls REST APIs        │
│  - Opens WebSocket        │
└─────────────┬─────────────┘
              │
              ▼
┌───────────────────────────┐
│       API Gateway         │
│  - Routes requests        │
│  - JWT validation         │
│  - WebSocket upgrades     │
└───┬──────────┬───────────┘
    │          │
    │          ▼
    │   ┌────────────────────┐
    │   │   Auth Service     │
    │   │   (Node.js + JWT)  │
    │   │ - /auth/*          │
    │   └────────────────────┘
    │
    │
    ▼
┌──────────────────────────────────────────────┐
│                  Core Services                │
│ ┌──────────────────────┐   ┌────────────────┐ │
│ │   Form Service (Go)  │   │ Response Svc   │ │
│ │ - CRUD forms         │   │ (Node.js)      │ │
│ │ - PostgreSQL storage │   │ - Store answers│ │
│ └──────────────────────┘   └────────────────┘ │
│         │ PostgreSQL              │ Firestore │
│         ▼                         ▼           │
│ ┌─────────────┐          ┌──────────────────┐ │
│ │  PostgreSQL │          │   Firestore DB   │ │
│ │ Users/Forms │          │  Form responses  │ │
│ └─────────────┘          └──────────────────┘ │
└──────────────────────────────────────────────┘
          │
          ▼
┌──────────────────────────────────────────────┐
│              Real-Time Service (Go)           │
│ - WebSockets: /ws/forms/:id/updates           │
│ - Redis Pub/Sub for broadcasts                │
└───────────────┬──────────────────────────────┘
                │
         ┌──────▼───────┐
         │    Redis     │
         │ WS sessions  │
         └──────────────┘

          │
          ▼
┌──────────────────────────────────────────────┐
│              Analytics Service (Python)       │
│ - /analytics/:id/summary                      │
│ - /forms/:id/export (CSV)                     │
│ - Uses Pandas + BigQuery                      │
└──────────────────────────────────────────────┘

          │
          ▼
┌──────────────────────────────────────────────┐
│              File Service (Lambda)            │
│ - /upload → S3                                │
│ - /files/:id → metadata                       │
│ - Handles thumbnails, PDF parsing             │
└──────────────────┬───────────────────────────┘
                   │
              ┌────▼────┐
              │   S3    │
              │ storage │
              └─────────┘
```

## Tech Stack

- **Languages:** Go (high-performance APIs), Node.js (auth + integrations), Python (analytics)
- **Databases:** PostgreSQL (users/forms), Firestore (responses), Redis (real-time), BigQuery (analytics)
- **Cloud & Infra:** AWS (API Gateway, Lambda, S3, RDS, EventBridge)
- **Containerization:** Docker + Kubernetes

## Services

### Core Services

1. **Auth Service** (Node.js + JWT/OAuth)
   - User authentication and authorization
   - JWT token management
   - Google OAuth integration

2. **Form Service** (Go + PostgreSQL)
   - CRUD operations for forms
   - Form schema management
   - User permissions

3. **Real-Time Service** (Go + WebSockets + Redis)
   - Live collaboration features
   - Real-time form editing
   - User presence tracking

4. **Response Service** (Node.js + Firestore)
   - Response collection and storage
   - Data validation
   - Response querying

5. **Analytics Service** (Python + BigQuery)
   - Response analytics and reporting
   - Data export (CSV, Google Sheets)
   - Statistical analysis

6. **File Service** (AWS Lambda + S3)
   - File upload handling
   - Image and document processing
   - Metadata management

## Getting Started

### Prerequisites

- Go 1.21+
- Node.js 18+
- Python 3.9+
- Docker & Docker Compose
- PostgreSQL
- Redis

### Local Development

1. Clone the repository:
```bash
git clone https://github.com/Mir00r/X-Form-Backend.git
cd X-Form-Backend
```

2. Start infrastructure services:
```bash
docker-compose up -d postgres redis
```

3. Set up environment variables:
```bash
cp .env.example .env
# Edit .env with your configuration
```

4. Start services:
```bash
# Auth Service
cd services/auth-service
npm install
npm run dev

# Form Service
cd services/form-service
go mod tidy
go run cmd/server/main.go

# Real-Time Service
cd services/realtime-service
go mod tidy
go run cmd/server/main.go

# Response Service
cd services/response-service
npm install
npm run dev

# Analytics Service
cd services/analytics-service
pip install -r requirements.txt
python main.py
```

## MVP Features

✅ **Implemented:**
- User authentication (email/password, Google OAuth)
- Form CRUD operations
- Real-time collaboration
- Response collection
- Basic analytics
- File uploads

🚧 **Planned:**
- Advanced analytics dashboards
- Team collaboration features
- Third-party integrations
- AI-powered features

## API Documentation

API documentation is available at `/docs` endpoint for each service when running in development mode.

## Deployment

The system is designed to run on Kubernetes with the following components:
- API Gateway (AWS API Gateway or NGINX Ingress)
- Microservices (EKS/GKE pods)
- Databases (AWS RDS, Firestore, ElastiCache)
- File Storage (S3)

See `deployment/` directory for Kubernetes manifests and Terraform configurations.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

MIT License - see LICENSE file for details.
