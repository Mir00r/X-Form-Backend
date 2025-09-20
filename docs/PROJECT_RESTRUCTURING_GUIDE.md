# Project Restructuring Implementation Guide

## 🎯 Overview

This document provides step-by-step instructions for implementing the new project structure for X-Form Backend. The restructuring follows industry best practices for microservices and monorepo architecture.

## 📋 Migration Checklist

### Phase 1: Documentation and Configuration Migration

#### ✅ Completed
- [x] Created new directory structure
- [x] Moved documentation to organized structure
- [x] Created comprehensive development guides
- [x] Set up new Makefile with automation
- [x] Created essential scripts (setup.sh, health-check.sh)
- [x] Set up proper .editorconfig and .nvmrc
- [x] Created .dockerignore with comprehensive patterns

#### 🔄 Next Steps (Manual Migration Required)

**1. Move Documentation Files**
```bash
# Architecture documentation
mv ARCHITECTURE_V2.md docs/architecture/overview.md  # Already done
mv ARCHITECTURE_ANALYSIS.md docs/architecture/analysis.md
mv ARCHITECTURE_ADAPTATION_GUIDE.md docs/architecture/adaptation-guide.md

# Development documentation  
mv CONTRIBUTING.md docs/development/contributing.md  # Already done
mv IMPLEMENTATION_GUIDE.md docs/development/implementation-guide.md
mv TUTORIAL.md docs/development/tutorial.md

# Operations documentation
mv OBSERVABILITY_COMPLETE.md docs/operations/observability.md
mv OBSERVABILITY_IMPLEMENTATION_GUIDE.md docs/operations/observability-guide.md

# Deployment documentation
mv CI_CD_INFRASTRUCTURE_IMPLEMENTATION_GUIDE.md docs/deployment/ci-cd-guide.md
mv ENHANCED_ARCHITECTURE_IMPLEMENTATION.md docs/deployment/enhanced-architecture.md
```

**2. Move Configuration Files**
```bash
# Environment configurations
mv .env.example configs/environments/.env.example
mv alert_rules.yml infrastructure/monitoring/alerting/alert_rules.yml
mv prometheus.yml infrastructure/monitoring/prometheus/prometheus.yml
mv otel-collector-config.yaml infrastructure/monitoring/otel-collector-config.yaml

# Docker configurations  
mv docker-compose.yml infrastructure/docker/environments/docker-compose.legacy.yml
mv docker-compose-traefik.yml infrastructure/docker/environments/docker-compose.prod.yml
mv docker-compose-v2.yml infrastructure/docker/environments/docker-compose.v2.yml

# Traefik configurations
mv infrastructure/traefik/* configs/traefik/
```

### Phase 2: Services Migration

**3. Migrate Services to New Structure**

**Auth Service (Node.js)**
```bash
# Create new structure
mkdir -p apps/auth-service/src/{application,domain,infrastructure,interfaces}
mkdir -p apps/auth-service/{tests,docs}

# Move existing code
mv services/auth-service/src/* apps/auth-service/src/
mv services/auth-service/test/* apps/auth-service/tests/
mv services/auth-service/package.json apps/auth-service/
mv services/auth-service/tsconfig.json apps/auth-service/
mv services/auth-service/Dockerfile apps/auth-service/
mv services/auth-service/*.md apps/auth-service/docs/
```

**Form Service (Go)**
```bash
# Create new structure  
mkdir -p apps/form-service/{cmd,internal,pkg}
mkdir -p apps/form-service/internal/{application,domain,infrastructure,interfaces}
mkdir -p apps/form-service/{tests,docs}

# Move existing code
mv services/form-service/cmd/* apps/form-service/cmd/
mv services/form-service/internal/* apps/form-service/internal/
mv services/form-service/go.mod apps/form-service/
mv services/form-service/go.sum apps/form-service/
mv services/form-service/Dockerfile apps/form-service/
mv services/form-service/*.md apps/form-service/docs/
```

**Repeat for other services:**
- response-service → apps/response-service
- realtime-service → apps/realtime-service  
- analytics-service → apps/analytics-service
- collaboration-service → apps/collaboration-service
- file-upload-service → apps/file-service

### Phase 3: Infrastructure Migration

**4. Organize Infrastructure**
```bash
# Kubernetes configurations
mv deployment/k8s/* infrastructure/kubernetes/base/

# Helm charts
mv infrastructure/helm/* infrastructure/helm/

# Terraform
mv infrastructure/terraform/* infrastructure/terraform/

# Monitoring
mv monitoring/* infrastructure/monitoring/
```

### Phase 4: Shared Packages Creation

**5. Extract Shared Code**
```bash
# Create shared packages
mkdir -p packages/{shared-types,shared-utils,shared-middleware,shared-config}

# Move common TypeScript types
# (Extract from services and move to packages/shared-types)

# Move common utilities  
# (Extract from services and move to packages/shared-utils)

# Move reusable middleware
# (Extract from services and move to packages/shared-middleware)
```

### Phase 5: Database and Scripts Migration

**6. Organize Database Files**
```bash
# Move database files
mv scripts/init-db.sql migrations/postgres/001_initial_schema.sql
mv scripts/migrate-to-traefik.sh tools/scripts/migrate-to-traefik.sh

# Move other scripts
mv scripts/* tools/scripts/
```

### Phase 6: Testing Infrastructure

**7. Set Up Testing**
```bash
# Create test structure
mkdir -p tests/{integration,e2e,performance,fixtures}

# Move existing tests
# (Extract integration tests from services to tests/integration)
# (Move any E2E tests to tests/e2e)
```

## 🔧 Configuration Updates Required

After moving files, update the following configurations:

### 1. Update Docker Compose Files
```yaml
# Update service paths in docker-compose files
services:
  auth-service:
    build: 
      context: .
      dockerfile: apps/auth-service/Dockerfile
    # ... rest of config
```

### 2. Update Import Paths
```typescript
// In TypeScript services, update imports to use shared packages
import { UserType } from '@x-form/shared-types';
import { validateEmail } from '@x-form/shared-utils';
```

### 3. Update Go Module Paths
```go
// In Go services, update import paths
import (
    "github.com/your-org/x-form-backend/packages/shared-go/types"
)
```

### 4. Update CI/CD Pipelines
```yaml
# Update .github/workflows/*.yml to reflect new paths
- name: Test Auth Service
  run: |
    cd apps/auth-service
    npm test
```

## 🚀 Execution Plan

### Recommended Order:
1. **Documentation Migration** (30 minutes)
2. **Configuration Migration** (45 minutes)  
3. **Services Migration** (2-3 hours)
4. **Infrastructure Migration** (1 hour)
5. **Shared Packages Creation** (2-4 hours)
6. **Testing and Validation** (1-2 hours)

### Before Starting:
```bash
# Create backup
git branch backup-before-restructure
git add -A
git commit -m "Backup before restructuring"

# Create new feature branch
git checkout -b feature/project-restructure
```

### After Completion:
```bash
# Test the new structure
make setup
make dev
make health
make test

# Commit changes
git add -A
git commit -m "Restructure project following industry best practices"
```

## 🔍 Validation Steps

1. **Build Verification**
   ```bash
   make build
   ```

2. **Service Health**
   ```bash
   make health
   ```

3. **Test Execution**
   ```bash
   make test
   ```

4. **Docker Compose**
   ```bash
   make start
   ```

5. **Documentation Review**
   - Verify all links work
   - Check that guides are accessible
   - Validate examples

## 📚 Post-Migration Tasks

1. **Update README.md** - Reflect new structure
2. **Update team documentation** - Inform team of changes
3. **Update IDE configurations** - Adjust workspace settings
4. **Update deployment scripts** - Reflect new paths
5. **Training** - Conduct team walkthrough

## 🎯 Benefits After Migration

- ✅ **Consistent Structure**: All services follow same pattern
- ✅ **Improved Developer Experience**: Clear organization and automation
- ✅ **Better Documentation**: Centralized and organized
- ✅ **Enhanced Maintainability**: Industry-standard patterns
- ✅ **Easier Onboarding**: Comprehensive guides and setup
- ✅ **Scalable Architecture**: Ready for new services and features

## 🆘 Rollback Plan

If issues arise during migration:

```bash
# Return to backup branch
git checkout backup-before-restructure

# Or reset specific files
git checkout HEAD~1 -- path/to/file
```

The migration is designed to be incremental, so you can pause and resume at any phase.
