# Contributing to X-Form Backend

> **🤝 Thank you for your interest in contributing to X-Form Backend!**

This guide provides everything you need to know to contribute effectively to our microservices platform.

## 📋 Table of Contents

1. [🚀 Getting Started](#-getting-started)
2. [💻 Development Environment](#-development-environment)
3. [🏗️ Project Structure](#️-project-structure)
4. [🔧 Development Workflow](#-development-workflow)
5. [✅ Code Standards](#-code-standards)
6. [🧪 Testing Guidelines](#-testing-guidelines)
7. [📝 Documentation](#-documentation)
8. [🔄 Pull Request Process](#-pull-request-process)
9. [🐛 Issue Reporting](#-issue-reporting)
10. [🏆 Recognition](#-recognition)

---

## 🚀 Getting Started

### Prerequisites for Contributors

Before you begin, ensure you have the following installed:

```bash
# Required tools
- Node.js (v18+)
- Go (v1.21+) 
- Python (v3.8+)
- Docker & Docker Compose
- Git
- Make

# Development tools (recommended)
- VS Code with extensions
- Postman or similar API testing tool
- TablePlus or similar database GUI
```

### Quick Setup for Contributors

```bash
# 1. Fork the repository on GitHub
# 2. Clone your fork
git clone https://github.com/YOUR_USERNAME/X-Form-Backend.git
cd X-Form-Backend

# 3. Add upstream remote
git remote add upstream https://github.com/original-owner/X-Form-Backend.git

# 4. Setup development environment
make setup
make dev

# 5. Verify everything works
make health
make test
```

---

## 💻 Development Environment

### Environment Setup

1. **Copy and configure environment variables:**
```bash
cp configs/environments/.env.example .env
# Edit .env with your preferred editor
```

2. **Install all dependencies:**
```bash
make install-deps
```

3. **Start development environment:**
```bash
make dev  # Starts all services with hot reload
```

### Development Tools Setup

#### VS Code Extensions
```bash
# Install recommended extensions
code --install-extension ms-vscode.vscode-typescript-next
code --install-extension golang.go
code --install-extension ms-python.python
code --install-extension ms-vscode.docker
code --install-extension humao.rest-client
code --install-extension ms-vscode.vscode-json
```

#### Git Hooks Setup
```bash
# Install pre-commit hooks
npm install
# This will setup pre-commit hooks for code quality checks
```

---

## 🏗️ Project Structure

### Directory Layout
```
X-Form-Backend/
├── apps/                    # Microservices
│   ├── auth-service/       # Node.js + TypeScript
│   ├── form-service/       # Go + Gin
│   ├── response-service/   # Node.js + Express
│   ├── realtime-service/   # Go + WebSockets
│   ├── analytics-service/  # Python + FastAPI
│   └── api-gateway/        # Go + Enhanced middleware
├── packages/               # Shared libraries
├── infrastructure/         # Infrastructure as Code
├── configs/               # Configuration files
├── docs/                  # Documentation
├── tools/                 # Development tools
└── migrations/            # Database migrations
```

### Service Technologies

| Service | Technology | Port | Purpose |
|---------|------------|------|---------|
| **Auth Service** | Node.js + TypeScript | 3001 | Authentication & User Management |
| **Form Service** | Go + Gin + GORM | 8001 | Form CRUD Operations |
| **Response Service** | Node.js + Express | 3002 | Form Response Handling |
| **Realtime Service** | Go + WebSockets | 8002 | Real-time Collaboration |
| **Analytics Service** | Python + FastAPI | 5001 | Analytics & Reporting |
| **API Gateway** | Go + Enhanced Middleware | 8080 | API Gateway & Routing |

---

## 🔧 Development Workflow

### Branch Naming Convention

```bash
# Feature branches
feature/auth-service-oauth-integration
feature/form-builder-drag-drop
feature/analytics-dashboard-improvements

# Bug fixes
fix/auth-token-validation-issue
fix/form-submission-validation-error

# Hotfixes
hotfix/security-vulnerability-patch

# Documentation
docs/api-documentation-update
docs/contributing-guide-improvements
```

### Development Process

#### 1. **Create Feature Branch**
```bash
# Update main branch
git checkout main
git pull upstream main

# Create feature branch
git checkout -b feature/your-feature-name
```

#### 2. **Development**
```bash
# Start development environment
make dev

# Make your changes
# ... edit code ...

# Run tests frequently
make test
npm test  # In specific service directories
go test ./...  # For Go services
pytest  # For Python services
```

#### 3. **Quality Checks**
```bash
# Run all quality checks
make quality-check

# Individual checks
make lint          # Code linting
make test          # All tests
make security-scan # Security scanning
```

#### 4. **Commit Changes**
```bash
# Add changes
git add .

# Commit with conventional commit format
git commit -m "feat(auth): add OAuth2 Google integration"
git commit -m "fix(forms): resolve validation error for required fields"
git commit -m "docs(api): update authentication endpoint documentation"
```

#### 5. **Push and Create PR**
```bash
# Push to your fork
git push origin feature/your-feature-name

# Create Pull Request on GitHub
# Fill out the PR template completely
```

### Conventional Commit Format

We use [Conventional Commits](https://www.conventionalcommits.org/) format:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or modifying tests
- `chore`: Maintenance tasks

**Examples:**
```bash
feat(auth): add Google OAuth integration
fix(forms): resolve validation error for empty required fields
docs(api): update authentication endpoint documentation
refactor(realtime): improve WebSocket connection handling
test(response): add integration tests for form submission
chore(deps): update dependencies to latest versions
```

---

## ✅ Code Standards

### Node.js/TypeScript Services

#### Code Style
```typescript
// Use TypeScript strict mode
// Follow clean architecture principles
// Use dependency injection

// Example: Service layer
export class AuthService {
  constructor(
    private userRepository: IUserRepository,
    private jwtService: IJWTService
  ) {}

  async authenticateUser(credentials: LoginCredentials): Promise<AuthResult> {
    // Implementation
  }
}
```

#### Standards
- **ESLint + Prettier** for code formatting
- **Jest** for testing
- **Clean Architecture** pattern
- **Dependency Injection** using containers
- **Error Handling** with standardized error responses

#### Commands
```bash
cd apps/auth-service

# Code quality
npm run lint
npm run lint:fix
npm run format

# Testing
npm test
npm run test:coverage
npm run test:watch
```

### Go Services

#### Code Style
```go
// Follow Go conventions
// Use interfaces for abstraction
// Implement clean architecture

// Example: Handler with dependency injection
type FormHandler struct {
    formService service.FormService
    logger      *logrus.Logger
}

func (h *FormHandler) CreateForm(c *gin.Context) {
    // Implementation with proper error handling
}
```

#### Standards
- **gofmt** for formatting
- **golint** for linting
- **testify** for testing
- **Clean Architecture** with layers
- **Dependency Injection** patterns
- **Structured Logging** with logrus

#### Commands
```bash
cd apps/form-service

# Code quality
go fmt ./...
go vet ./...
golangci-lint run

# Testing
go test ./...
go test -race ./...
go test -cover ./...
```

### Python Services

#### Code Style
```python
# Follow PEP 8
# Use type hints
# Implement clean architecture

# Example: FastAPI service
from fastapi import FastAPI, Depends
from typing import List

class AnalyticsService:
    def __init__(self, repository: AnalyticsRepository):
        self.repository = repository

    async def get_form_analytics(self, form_id: str) -> FormAnalytics:
        # Implementation
```

#### Standards
- **Black** for code formatting
- **flake8** for linting
- **pytest** for testing
- **Type hints** for better code quality
- **Pydantic** for data validation
- **Clean Architecture** patterns

#### Commands
```bash
cd apps/analytics-service

# Code quality
black .
flake8 .
mypy .

# Testing
pytest
pytest --cov=app
```

---

## 🧪 Testing Guidelines

### Testing Strategy

We follow a comprehensive testing strategy:

1. **Unit Tests** - Test individual functions and methods
2. **Integration Tests** - Test service interactions
3. **Contract Tests** - Test API contracts between services
4. **End-to-End Tests** - Test complete user flows

### Testing Standards

#### Unit Tests
```bash
# Node.js services
npm test
npm run test:coverage  # Must maintain 80%+ coverage

# Go services
go test ./...
go test -cover ./...  # Must maintain 80%+ coverage

# Python services
pytest
pytest --cov=app  # Must maintain 80%+ coverage
```

#### Integration Tests
```bash
# Run integration tests
make test-integration

# Test specific service integration
make test-auth-integration
make test-form-integration
```

#### API Contract Tests
```bash
# Test API contracts
make test-contracts

# Generate and validate OpenAPI specs
swagger-codegen validate -i swagger.json
```

#### End-to-End Tests
```bash
# Run E2E tests
make test-e2e

# Test complete user flows
npm run test:e2e
```

### Test Coverage Requirements

- **Minimum Coverage**: 80% for all services
- **Critical Paths**: 95% coverage for authentication, payment flows
- **New Code**: 90% coverage requirement
- **Integration**: All public APIs must have integration tests

---

## 📝 Documentation

### Documentation Standards

#### Code Documentation
```typescript
/**
 * Authenticates user credentials and returns JWT tokens
 * @param credentials - User login credentials
 * @returns Promise<AuthResult> - Authentication result with tokens
 * @throws {ValidationError} When credentials are invalid
 * @throws {ServiceError} When authentication service is unavailable
 */
async authenticateUser(credentials: LoginCredentials): Promise<AuthResult>
```

#### API Documentation
- **OpenAPI/Swagger** specifications for all endpoints
- **Request/Response examples** with real data
- **Error responses** with all possible error codes
- **Authentication requirements** clearly documented

#### README Updates
When contributing to a service, update the relevant README files:
- Service-specific README in `apps/[service-name]/README.md`
- Main project README if adding new features
- Architecture documentation if changing system design

#### Documentation Commands
```bash
# Generate API documentation
make docs

# Validate documentation
swagger-codegen validate -i http://localhost:8080/swagger/swagger.json

# Serve documentation locally
make docs-serve
```

---

## 🔄 Pull Request Process

### PR Checklist

Before submitting a pull request, ensure:

- [ ] **Code Quality**
  - [ ] All tests pass (`make test`)
  - [ ] Code coverage maintained (80%+)
  - [ ] Linting passes (`make lint`)
  - [ ] No security vulnerabilities (`make security-scan`)

- [ ] **Documentation**
  - [ ] API changes documented in OpenAPI specs
  - [ ] README updated if needed
  - [ ] Code comments added for complex logic
  - [ ] Changelog entry added

- [ ] **Testing**
  - [ ] Unit tests added for new functionality
  - [ ] Integration tests updated if needed
  - [ ] Manual testing performed
  - [ ] API endpoints tested via Swagger UI

- [ ] **Git Hygiene**
  - [ ] Conventional commit messages used
  - [ ] Feature branch based on latest main
  - [ ] No merge conflicts
  - [ ] Commits squashed if appropriate

### PR Template

When creating a PR, use this template:

```markdown
## Description
Brief description of changes and motivation.

## Type of Change
- [ ] Bug fix (non-breaking change that fixes an issue)
- [ ] New feature (non-breaking change that adds functionality)
- [ ] Breaking change (fix or feature that causes existing functionality to change)
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing performed
- [ ] All tests pass

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review of code completed
- [ ] Documentation updated
- [ ] No new warnings introduced
- [ ] Related issues linked

## Screenshots (if applicable)
Add screenshots for UI changes or API documentation updates.

## Additional Notes
Any additional information for reviewers.
```

### Review Process

1. **Automated Checks**: All CI/CD checks must pass
2. **Code Review**: At least 2 reviewers required for major changes
3. **Testing**: Manual testing by reviewers if needed
4. **Documentation Review**: Technical writer review for major features
5. **Security Review**: Security team review for auth/security changes

---

## 🐛 Issue Reporting

### Bug Reports

Use this template for bug reports:

```markdown
## Bug Description
Clear and concise description of the bug.

## Environment
- OS: [e.g., macOS 12.0]
- Node.js version: [e.g., 18.17.0]
- Go version: [e.g., 1.21.0]
- Docker version: [e.g., 20.10.0]

## Steps to Reproduce
1. Go to '...'
2. Click on '...'
3. Scroll down to '...'
4. See error

## Expected Behavior
What you expected to happen.

## Actual Behavior
What actually happened.

## Screenshots/Logs
Add screenshots or error logs if applicable.

## Additional Context
Any other context about the problem.
```

### Feature Requests

Use this template for feature requests:

```markdown
## Feature Description
Clear and concise description of the feature.

## Problem Statement
What problem does this feature solve?

## Proposed Solution
Describe your proposed solution.

## Alternatives Considered
Describe alternatives you've considered.

## Additional Context
Any other context or screenshots about the feature.
```

### Issue Labels

We use these labels for issue management:

- **Type**: `bug`, `feature`, `enhancement`, `documentation`
- **Priority**: `low`, `medium`, `high`, `critical`
- **Service**: `auth-service`, `form-service`, `response-service`, etc.
- **Status**: `needs-triage`, `in-progress`, `blocked`, `ready-for-review`

---

## 🏆 Recognition

### Contributors

We recognize contributors in several ways:

1. **README Credits**: Contributors listed in project README
2. **Release Notes**: Significant contributions mentioned in releases
3. **GitHub Discussions**: Shout-outs in community discussions
4. **Contributor Badge**: Special badge for regular contributors

### Contribution Types

We value all types of contributions:

- **Code**: New features, bug fixes, performance improvements
- **Documentation**: API docs, tutorials, guides
- **Testing**: Test improvements, test coverage
- **Design**: UX/UI improvements, architecture design
- **Community**: Issue triage, helping other contributors
- **Translation**: Documentation and UI translations

### Becoming a Maintainer

Active contributors can become maintainers by:

1. **Consistent Contributions**: Regular, high-quality contributions
2. **Community Involvement**: Helping other contributors
3. **Code Review**: Participating in code reviews
4. **Issue Triage**: Helping with issue management
5. **Documentation**: Improving project documentation

---

## 📞 Getting Help

### Development Support

- **GitHub Discussions**: General questions and discussions
- **GitHub Issues**: Bug reports and feature requests
- **Stack Overflow**: Tag questions with `x-form-backend`
- **Discord/Slack**: Real-time chat (if available)

### Development Resources

- **[Local Development Guide](docs/development/LOCAL_DEVELOPMENT_COMPLETE_GUIDE.md)**: Complete setup guide
- **[Quick Reference](docs/development/DEVELOPER_QUICK_REFERENCE.md)**: Daily development commands
- **[Architecture Guide](docs/architecture/ARCHITECTURE_V2.md)**: System architecture overview
- **[API Documentation](http://localhost:8080/swagger/)**: Interactive API docs

---

## 📜 Code of Conduct

We are committed to providing a welcoming and inclusive environment. Please read our [Code of Conduct](CODE_OF_CONDUCT.md) before contributing.

## 📄 License

By contributing to X-Form Backend, you agree that your contributions will be licensed under the same license as the project.

---

**🎉 Thank you for contributing to X-Form Backend!**

Your contributions help make this project better for everyone. We appreciate your time and effort in improving our microservices platform.
