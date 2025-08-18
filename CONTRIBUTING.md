# Contributing to X-Form Backend

Thank you for your interest in contributing to X-Form Backend! This document provides guidelines and information for contributors.

## 🚀 Getting Started

### Prerequisites

- **Node.js** 18+ (for auth and response services)
- **Go** 1.21+ (for form and real-time services)
- **Python** 3.9+ (for analytics service)
- **Docker** and **Docker Compose**
- **PostgreSQL** 15+
- **Redis** 7+

### Development Setup

1. **Clone the repository**:
```bash
git clone https://github.com/Mir00r/X-Form-Backend.git
cd X-Form-Backend
```

2. **Run setup script**:
```bash
make setup
```

3. **Configure environment**:
```bash
cp .env.example .env
# Edit .env with your configuration
```

4. **Start development environment**:
```bash
make start
```

## 🏗️ Project Structure

```
X-Form-Backend/
├── services/
│   ├── auth-service/          # Node.js - Authentication
│   ├── form-service/          # Go - Form management
│   ├── response-service/      # Node.js - Response collection
│   ├── realtime-service/      # Go - WebSocket/real-time
│   └── analytics-service/     # Python - Analytics & reporting
├── deployment/
│   ├── k8s/                   # Kubernetes manifests
│   └── nginx.conf             # API Gateway config
├── scripts/
│   ├── setup.sh               # Development setup
│   └── init-db.sql           # Database initialization
└── docker-compose.yml         # Local development
```

## 🔧 Development Workflow

### Working on Services

Each service can be developed independently:

```bash
# Auth Service (Node.js)
cd services/auth-service
npm run dev

# Form Service (Go)
cd services/form-service
go run cmd/server/main.go

# Response Service (Node.js)
cd services/response-service
npm run dev

# Analytics Service (Python)
cd services/analytics-service
python main.py
```

### Testing

```bash
# Run all tests
make test

# Test specific service
make test-auth
make test-form
make test-response

# Run linting
make lint
```

### Code Standards

#### **Node.js Services**
- Use **ESLint** with Airbnb configuration
- **Prettier** for code formatting
- **Jest** for testing
- Follow **RESTful API** conventions
- Use **async/await** over promises

#### **Go Services**
- Follow **Go conventions** (gofmt, golint)
- Use **testify** for testing
- Follow **clean architecture** patterns
- Use **GORM** for database operations
- Use **Gin** for HTTP routing

#### **Python Services**
- Follow **PEP 8** style guide
- Use **Black** for code formatting
- Use **pytest** for testing
- Use **FastAPI** for HTTP APIs
- Use **Pydantic** for data validation

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
type(scope): description

feat(auth): add Google OAuth integration
fix(form): resolve form validation issue
docs(api): update authentication endpoints
test(response): add integration tests
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `test`: Tests
- `refactor`: Code refactoring
- `style`: Code style changes
- `chore`: Maintenance tasks

## 🚦 Pull Request Process

1. **Create a feature branch**:
```bash
git checkout -b feature/your-feature-name
```

2. **Make your changes**:
   - Write clean, documented code
   - Add tests for new functionality
   - Update documentation if needed

3. **Test your changes**:
```bash
make test
make lint
```

4. **Commit your changes**:
```bash
git add .
git commit -m "feat(service): add new feature"
```

5. **Push and create PR**:
```bash
git push origin feature/your-feature-name
```

6. **Create Pull Request** with:
   - Clear description of changes
   - Link to related issues
   - Screenshots/examples if applicable

## 🐛 Bug Reports

When reporting bugs, please include:

1. **Environment details**:
   - OS and version
   - Node.js/Go/Python versions
   - Service version

2. **Steps to reproduce**:
   - Clear, numbered steps
   - Expected vs actual behavior
   - Error messages/logs

3. **Additional context**:
   - Screenshots if applicable
   - Configuration details
   - Related issues

## 💡 Feature Requests

For new features:

1. **Check existing issues** first
2. **Describe the problem** you're solving
3. **Propose a solution** with examples
4. **Consider alternatives** and their trade-offs
5. **Think about impact** on existing features

## 📋 Issue Labels

- `bug` - Something isn't working
- `enhancement` - New feature or request
- `documentation` - Improvements or additions to docs
- `good first issue` - Good for newcomers
- `help wanted` - Extra attention is needed
- `priority:high` - High priority issue
- `service:auth` - Related to auth service
- `service:form` - Related to form service
- `service:response` - Related to response service
- `service:analytics` - Related to analytics service

## 🔐 Security

For security vulnerabilities:

1. **DO NOT** create public issues
2. **Email** security concerns to: security@xform.dev
3. **Include** detailed description and steps to reproduce
4. **Wait** for acknowledgment before public disclosure

## 📚 Documentation

Help improve documentation:

- **API docs** - Update OpenAPI/Swagger specs
- **README** - Keep setup instructions current
- **Code comments** - Document complex logic
- **Architecture** - Update diagrams and explanations

## 🎯 Areas for Contribution

### High Priority
- Complete Form Service handlers
- Implement Real-time Service
- Add comprehensive testing
- Improve error handling

### Medium Priority
- OAuth integrations
- File upload service
- Analytics improvements
- Performance optimizations

### Good First Issues
- Documentation improvements
- Code formatting/linting
- Basic test additions
- Configuration improvements

## 💬 Community

- **Discussions**: Use GitHub Discussions for questions
- **Issues**: Use GitHub Issues for bugs and features
- **Discord**: [Join our Discord server](https://discord.gg/xform)
- **Email**: team@xform.dev

## 📄 License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

**Thank you for contributing to X-Form Backend!** 🙏
