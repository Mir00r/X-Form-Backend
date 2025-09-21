#!/bin/bash

# GitLab Spec Kit Tag Creation Script
# This script creates GitLab tags for spec kit releases with proper metadata

set -e

# Configuration
GITLAB_API_URL="${CI_API_V4_URL:-https://gitlab.com/api/v4}"
PROJECT_ID="${CI_PROJECT_ID}"
GITLAB_TOKEN="${GITLAB_TOKEN}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Validate required environment variables
validate_environment() {
    log_info "Validating environment variables..."
    
    if [ -z "$GITLAB_TOKEN" ]; then
        log_error "GITLAB_TOKEN is not set"
        exit 1
    fi
    
    if [ -z "$PROJECT_ID" ]; then
        log_error "CI_PROJECT_ID is not set"
        exit 1
    fi
    
    log_success "Environment validation passed"
}

# Generate version number
generate_version() {
    local version_type=${1:-patch}
    
    log_info "Generating new version number (type: $version_type)..."
    
    # Get latest tag
    local latest_tag=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
    log_info "Latest tag: $latest_tag"
    
    # Remove 'v' prefix and split version
    local version=${latest_tag#v}
    local major=$(echo $version | cut -d. -f1)
    local minor=$(echo $version | cut -d. -f2)
    local patch=$(echo $version | cut -d. -f3)
    
    # Increment version based on type
    case $version_type in
        major)
            major=$((major + 1))
            minor=0
            patch=0
            ;;
        minor)
            minor=$((minor + 1))
            patch=0
            ;;
        patch)
            patch=$((patch + 1))
            ;;
        *)
            log_error "Invalid version type: $version_type (use major, minor, or patch)"
            exit 1
            ;;
    esac
    
    echo "v$major.$minor.$patch"
}

# Generate changelog
generate_changelog() {
    local previous_tag=${1:-$(git describe --tags --abbrev=0 HEAD~1 2>/dev/null)}
    local current_commit=${2:-HEAD}
    
    log_info "Generating changelog from $previous_tag to $current_commit..."
    
    if [ -z "$previous_tag" ]; then
        echo "## 🎉 Initial Release

### Features
- Complete OpenAPI specification suite for X-Form Backend
- Automated validation and documentation generation
- GitLab CI/CD integration
- Interactive documentation portal

### Services Included
- Auth Service API
- Form Service API  
- Response Service API
- Analytics Service API
- Real-time Service API
- Collaboration Service API
- Event Bus Service API
- API Gateway Service API"
    else
        echo "## 📋 Changes

### Commits"
        git log --pretty=format:"- %s (%h) by %an" $previous_tag..$current_commit
        
        echo -e "\n\n### Statistics"
        local commit_count=$(git rev-list --count $previous_tag..$current_commit)
        local files_changed=$(git diff --name-only $previous_tag..$current_commit | wc -l)
        echo "- $commit_count commits"
        echo "- $files_changed files changed"
        
        # Check for breaking changes
        if git log --grep="BREAKING" --grep="breaking" --grep="!:" $previous_tag..$current_commit | grep -q .; then
            echo -e "\n### ⚠️ Breaking Changes"
            git log --pretty=format:"- %s (%h)" --grep="BREAKING" --grep="breaking" --grep="!:" $previous_tag..$current_commit
        fi
        
        # Check for new features
        if git log --grep="feat" --grep="feature" $previous_tag..$current_commit | grep -q .; then
            echo -e "\n### ✨ New Features"
            git log --pretty=format:"- %s (%h)" --grep="feat" --grep="feature" $previous_tag..$current_commit
        fi
        
        # Check for bug fixes
        if git log --grep="fix" --grep="bug" $previous_tag..$current_commit | grep -q .; then
            echo -e "\n### 🐛 Bug Fixes"
            git log --pretty=format:"- %s (%h)" --grep="fix" --grep="bug" $previous_tag..$current_commit
        fi
    fi
}

# Create GitLab tag
create_gitlab_tag() {
    local tag_name=$1
    local tag_message=$2
    local release_description=$3
    
    log_info "Creating GitLab tag: $tag_name"
    
    # Create the tag
    local tag_response=$(curl -s -X POST \
        -H "PRIVATE-TOKEN: $GITLAB_TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"tag_name\": \"$tag_name\",
            \"ref\": \"$CI_COMMIT_SHA\",
            \"message\": \"$tag_message\"
        }" \
        "$GITLAB_API_URL/projects/$PROJECT_ID/repository/tags")
    
    if echo "$tag_response" | grep -q "error"; then
        log_error "Failed to create tag: $tag_response"
        return 1
    fi
    
    log_success "Tag created successfully: $tag_name"
    
    # Create the release
    log_info "Creating GitLab release for tag: $tag_name"
    
    local release_response=$(curl -s -X POST \
        -H "PRIVATE-TOKEN: $GITLAB_TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"name\": \"X-Form API Spec Kit $tag_name\",
            \"tag_name\": \"$tag_name\",
            \"description\": \"$release_description\",
            \"assets\": {
                \"links\": [
                    {
                        \"name\": \"API Documentation Portal\",
                        \"url\": \"https://docs.x-form.com\",
                        \"link_type\": \"other\"
                    },
                    {
                        \"name\": \"OpenAPI Specification Bundle\",
                        \"url\": \"$CI_PROJECT_URL/-/jobs/artifacts/$tag_name/download?job=build:docs\",
                        \"link_type\": \"package\"
                    },
                    {
                        \"name\": \"Generated SDKs\",
                        \"url\": \"$CI_PROJECT_URL/-/jobs/artifacts/$tag_name/download?job=generate:docs\",
                        \"link_type\": \"package\"
                    }
                ]
            }
        }" \
        "$GITLAB_API_URL/projects/$PROJECT_ID/releases")
    
    if echo "$release_response" | grep -q "error"; then
        log_warning "Failed to create release (tag created successfully): $release_response"
        return 0
    fi
    
    log_success "Release created successfully: $tag_name"
}

# Validate specifications before tagging
validate_specs() {
    log_info "Validating OpenAPI specifications..."
    
    # Run validation
    if npm run validate; then
        log_success "Specification validation passed"
    else
        log_error "Specification validation failed"
        exit 1
    fi
    
    # Run linting
    if npm run lint; then
        log_success "Specification linting passed"
    else
        log_error "Specification linting failed"
        exit 1
    fi
}

# Main function
main() {
    local version_type=${1:-patch}
    local force_version=$2
    
    log_info "🏷️  GitLab Spec Kit Tag Creation Starting..."
    
    # Validate environment
    validate_environment
    
    # Validate specs before tagging
    validate_specs
    
    # Generate or use provided version
    local new_version
    if [ -n "$force_version" ]; then
        new_version=$force_version
        log_info "Using provided version: $new_version"
    else
        new_version=$(generate_version $version_type)
        log_success "Generated version: $new_version"
    fi
    
    # Generate changelog
    local changelog=$(generate_changelog)
    
    # Create tag message
    local tag_message="X-Form API Spec Kit $new_version

This release includes:
- Updated OpenAPI specifications for all services
- Enhanced documentation portal
- Automated validation and testing

See release notes for detailed changes."
    
    # Create GitLab tag and release
    if create_gitlab_tag "$new_version" "$tag_message" "$changelog"; then
        log_success "🎉 Successfully created GitLab tag and release: $new_version"
        
        # Output for pipeline
        echo "SPEC_KIT_VERSION=$new_version" >> gitlab_tag.env
        echo "RELEASE_CREATED=true" >> gitlab_tag.env
        
        log_info "📦 Tag information saved to gitlab_tag.env"
        log_info "🔗 View release: $CI_PROJECT_URL/-/releases/$new_version"
    else
        log_error "❌ Failed to create GitLab tag and release"
        exit 1
    fi
}

# Script usage
usage() {
    echo "Usage: $0 [version_type] [force_version]"
    echo ""
    echo "Arguments:"
    echo "  version_type   Version increment type: major, minor, patch (default: patch)"
    echo "  force_version  Force specific version (e.g., v1.2.3)"
    echo ""
    echo "Examples:"
    echo "  $0              # Create patch version"
    echo "  $0 minor        # Create minor version" 
    echo "  $0 major        # Create major version"
    echo "  $0 patch v1.2.3 # Force specific version"
    echo ""
    echo "Environment Variables:"
    echo "  GITLAB_TOKEN    GitLab access token (required)"
    echo "  CI_PROJECT_ID   GitLab project ID (required)"
    echo "  CI_API_V4_URL   GitLab API URL (optional)"
}

# Check if help is requested
if [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
    usage
    exit 0
fi

# Run main function
main "$@"
