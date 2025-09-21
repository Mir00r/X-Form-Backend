# GitLab Spec Kit Integration - Complete Setup Guide

This document provides comprehensive instructions for setting up GitHub Spec Kit with GitLab CI/CD integration for the X-Form Backend project.

## 🎯 Overview

The GitLab integration provides:
- **Automated Validation**: OpenAPI spec validation on every commit
- **Automated Documentation**: Generate and deploy API documentation
- **Version Management**: Automatic tagging and release creation
- **Package Publishing**: Publish spec kit to GitLab Package Registry
- **Security Scanning**: Validate specifications for security issues
- **Breaking Change Detection**: Identify API breaking changes

## 🔧 Setup Instructions

### 1. GitLab CI/CD Variables

Configure the following variables in your GitLab project:

#### Required Variables
```bash
# GitLab Access Token (Project Settings > Access Tokens)
GITLAB_TOKEN=glpat-xxxxxxxxxxxxxxxxxxxx

# NPM Token for package publishing
NPM_TOKEN=npm_xxxxxxxxxxxxxxxxxxxx

# Deployment Tokens
DEV_DEPLOY_TOKEN=dev_xxxxxxxxxxxxxxxxxxxx
STAGING_DEPLOY_TOKEN=staging_xxxxxxxxxxxxxxxxxxxx
PROD_DEPLOY_TOKEN=prod_xxxxxxxxxxxxxxxxxxxx

# Webhook URLs for deployments
DEV_DEPLOY_WEBHOOK=https://api-dev.x-form.com/deploy/webhook
STAGING_DEPLOY_WEBHOOK=https://api-staging.x-form.com/deploy/webhook
PROD_DEPLOY_WEBHOOK=https://api.x-form.com/deploy/webhook
```

#### Optional Variables
```bash
# Slack Integration
SLACK_WEBHOOK_URL=https://hooks.slack.com/services/xxx/xxx/xxx

# Jira Integration
JIRA_SERVER_URL=https://company.atlassian.net
JIRA_TOKEN=xxxxxxxxxxxxxxxxxxxx

# Additional Configuration
SPECTRAL_VERSION=6.11.0
NODE_VERSION=18.x
```

### 2. GitLab Runner Configuration

Ensure your GitLab runners have Docker support and the following tags:
- `docker` - For containerized jobs
- `kubernetes` - For Kubernetes deployments (if applicable)

### 3. Project Structure

The GitLab integration expects this structure:
```
├── .gitlab-ci.yml                    # Main CI/CD pipeline
├── .gitlab/
│   ├── merge_request_templates/       # MR templates
│   └── issue_templates/               # Issue templates
├── specs/
│   ├── .gitlab-spec-kit.yml          # Spec kit configuration
│   ├── openapi.yaml                  # Main OpenAPI spec
│   ├── services/                     # Service specifications
│   ├── docs/                         # Documentation portal
│   └── dist/                         # Generated files
└── scripts/
    └── gitlab/
        └── create-spec-kit-tag.sh     # Tag creation script
```

## 🚀 Pipeline Stages

### Stage 1: Validate
- **validate:openapi**: Validates all OpenAPI specifications using Spectral
- **security:openapi**: Runs security checks on API specifications

### Stage 2: Build
- **build:docs**: Builds the documentation portal
- **generate:docs**: Generates comprehensive API documentation and SDKs

### Stage 3: Test
- **test:portal**: Tests the documentation portal functionality
- **test:integration**: Runs integration tests against specifications

### Stage 4: Deploy
- **deploy:dev**: Deploys to development environment (automatic on `develop` branch)
- **deploy:staging**: Deploys to staging environment (manual on `main` branch)
- **deploy:production**: Deploys to production environment (manual on version tags)

### Stage 5: Publish
- **publish:spec-kit**: Publishes package to GitLab Package Registry
- **release:notes**: Generates release notes and creates GitLab releases

## 🏷️ Tagging Strategy

### Automatic Tagging
The pipeline automatically creates tags for:
- **Patch versions**: Bug fixes and documentation updates
- **Minor versions**: New features and backwards-compatible changes
- **Major versions**: Breaking changes and major releases

### Manual Tagging
Use the provided script for manual tag creation:
```bash
# Create patch version
./scripts/gitlab/create-spec-kit-tag.sh patch

# Create minor version
./scripts/gitlab/create-spec-kit-tag.sh minor

# Create major version
./scripts/gitlab/create-spec-kit-tag.sh major

# Force specific version
./scripts/gitlab/create-spec-kit-tag.sh patch v2.1.0
```

## 📦 Package Publishing

### GitLab Package Registry
Packages are automatically published to:
```
https://gitlab.com/your-org/x-form-backend/-/packages
```

### NPM Installation
Install the published spec kit:
```bash
# Install from GitLab Package Registry
npm install @x-form/api-spec-kit

# Install specific version
npm install @x-form/api-spec-kit@1.2.3
```

## 🔍 Validation Rules

### Spectral Rules
The pipeline uses custom Spectral rules defined in `.spectral.yml`:
- API design best practices
- Security validation
- Documentation completeness
- Naming conventions

### Breaking Change Detection
Automatically detects breaking changes by comparing:
- Endpoint removals
- Required field additions
- Response schema changes
- Authentication changes

## 🚨 Security Scanning

### API Security Checks
- No API keys in URLs
- Proper authentication schemes
- Input validation requirements
- Rate limiting configurations

### Dependency Security
- NPM audit for known vulnerabilities
- Docker image security scanning
- License compliance checking

## 📊 Monitoring & Analytics

### Pipeline Metrics
Track the following in GitLab Analytics:
- Pipeline success/failure rates
- Deployment frequency
- Time to deploy
- Change failure rate

### API Documentation Metrics
Monitor documentation portal:
- Page views and usage
- API endpoint popularity
- Error rates
- Performance metrics

## 🎯 Best Practices

### Branch Protection
Configure branch protection rules:
- Require MR approval for `main` branch
- Require pipeline success
- Require up-to-date branches
- Restrict force pushes

### Code Review Process
1. Create feature branch from `develop`
2. Make API specification changes
3. Run local validation: `npm run validate`
4. Create merge request using template
5. Address review feedback
6. Merge after approval and pipeline success

### Release Process
1. Merge features to `develop` branch
2. Test in development environment
3. Create release MR to `main` branch
4. Deploy to staging for final testing
5. Create version tag for production release
6. Deploy to production with manual approval

## 🔧 Troubleshooting

### Common Issues

#### Pipeline Validation Failures
```bash
# Check specification locally
npm run validate
npm run lint

# Fix validation errors
npm run lint:fix
```

#### Authentication Errors
```bash
# Verify GitLab token permissions
curl -H "PRIVATE-TOKEN: $GITLAB_TOKEN" \
  "https://gitlab.com/api/v4/user"
```

#### Deployment Failures
```bash
# Check deployment webhook URLs
curl -X POST \
  -H "Authorization: Bearer $DEV_DEPLOY_TOKEN" \
  "$DEV_DEPLOY_WEBHOOK/health"
```

### Debug Commands
```bash
# Test Spectral validation
npx spectral lint specs/openapi.yaml --verbose

# Test documentation generation
npm run build-docs

# Test package publishing (dry run)
npm publish --dry-run
```

## 📞 Support

### Getting Help
- **GitLab Issues**: Create issues using provided templates
- **Team Chat**: #api-team Slack channel
- **Documentation**: https://docs.x-form.com/spec-kit
- **Wiki**: https://gitlab.com/your-org/x-form-backend/-/wikis

### Escalation
For critical issues:
1. Create high-priority GitLab issue
2. Tag @api-team and @platform-team
3. Notify #api-alerts Slack channel
4. Contact DevOps team for infrastructure issues

## 🔄 Continuous Improvement

### Feedback Collection
- Developer experience surveys
- Pipeline performance analysis
- Documentation usage analytics
- API consumer feedback

### Automation Opportunities
- AI-powered API documentation
- Automated SDK generation
- Performance regression testing
- Automated security patching

---

## 📋 Checklist for New Team Members

- [ ] GitLab project access granted
- [ ] Local development environment setup
- [ ] Spec kit documentation reviewed
- [ ] Pipeline execution permissions configured
- [ ] Slack channels joined
- [ ] First MR created and merged
- [ ] Release process practiced
- [ ] Emergency procedures understood

---

**Document Version**: 1.0.0  
**Last Updated**: ${new Date().toISOString().split('T')[0]}  
**Maintained By**: X-Form API Team
