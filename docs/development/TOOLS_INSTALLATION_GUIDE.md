# Development Tools Installation Guide

> **🛠️ Complete guide for setting up all required tools for X-Form Backend development**

## 📋 Quick Installation Checklist

```bash
# ✅ Core Tools (Required)
□ Node.js 18+
□ Go 1.21+ 
□ Python 3.8+
□ Docker Desktop
□ Git
□ Make

# 🚀 Development Tools (Recommended)
□ VS Code + Extensions
□ Postman/Insomnia
□ TablePlus/pgAdmin
□ Redis CLI
□ HTTPie

# ⚡ Performance Tools (Optional)
□ Artillery/Apache Bench
□ hey (Go load testing tool)
□ wrk (HTTP benchmarking)
```

---

## 🖥️ Core Tools Installation

### 1. **Node.js (v18+)**

#### macOS
```bash
# Option 1: Using Homebrew
brew install node@18

# Option 2: Using NVM (Recommended)
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.3/install.sh | bash
source ~/.zshrc  # or ~/.bashrc
nvm install 18
nvm use 18
nvm alias default 18

# Verify installation
node --version  # Should show v18.x.x
npm --version   # Should show 9.x.x or higher
```

#### Ubuntu/Debian
```bash
# Using NodeSource repository
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt-get install -y nodejs

# Or using NVM
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.3/install.sh | bash
source ~/.bashrc
nvm install 18
nvm use 18
```

#### Windows
```powershell
# Using Chocolatey
choco install nodejs

# Or download from: https://nodejs.org/en/download/
# Select "LTS" version
```

### 2. **Go (v1.21+)**

#### macOS
```bash
# Using Homebrew
brew install go

# Verify installation
go version  # Should show go1.21.x

# Set up Go workspace (add to ~/.zshrc or ~/.bashrc)
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
```

#### Ubuntu/Debian
```bash
# Download and install
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz

# Add to PATH (add to ~/.bashrc)
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Verify
go version
```

#### Windows
```powershell
# Using Chocolatey
choco install golang

# Or download from: https://go.dev/dl/
# Follow installer instructions
```

### 3. **Python (v3.8+)**

#### macOS
```bash
# Using Homebrew
brew install python@3.11

# Verify installation
python3 --version  # Should show 3.8+ 
pip3 --version

# Create alias (add to ~/.zshrc)
echo 'alias python=python3' >> ~/.zshrc
echo 'alias pip=pip3' >> ~/.zshrc
```

#### Ubuntu/Debian
```bash
# Update and install
sudo apt update
sudo apt install python3 python3-pip python3-venv

# Verify
python3 --version
pip3 --version
```

#### Windows
```powershell
# Using Chocolatey
choco install python

# Or download from: https://www.python.org/downloads/
# Make sure to check "Add to PATH" during installation
```

### 4. **Docker Desktop**

#### macOS
```bash
# Option 1: Using Homebrew
brew install --cask docker

# Option 2: Download from Docker website
# https://docs.docker.com/desktop/install/mac-install/

# Start Docker Desktop application
open /Applications/Docker.app

# Verify installation
docker --version
docker-compose --version
```

#### Ubuntu/Debian
```bash
# Add Docker repository
sudo apt-get update
sudo apt-get install ca-certificates curl gnupg
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg

echo \
  "deb [arch="$(dpkg --print-architecture)" signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  "$(. /etc/os-release && echo "$VERSION_CODENAME")" stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# Install Docker
sudo apt-get update
sudo apt-get install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

# Add user to docker group
sudo usermod -aG docker $USER
newgrp docker

# Verify
docker --version
docker compose version
```

#### Windows
```powershell
# Download from: https://docs.docker.com/desktop/install/windows-install/
# Follow installer instructions
# Requires WSL2 for better performance
```

### 5. **Git**

#### macOS
```bash
# Usually pre-installed, but to get latest version:
brew install git

# Configure Git
git config --global user.name "Your Name"
git config --global user.email "your.email@example.com"
```

#### Ubuntu/Debian
```bash
sudo apt install git

# Configure Git
git config --global user.name "Your Name"
git config --global user.email "your.email@example.com"
```

#### Windows
```powershell
# Using Chocolatey
choco install git

# Or download from: https://git-scm.com/download/win
```

### 6. **Make**

#### macOS
```bash
# Usually pre-installed with Xcode Command Line Tools
xcode-select --install

# Or using Homebrew
brew install make

# Verify
make --version
```

#### Ubuntu/Debian
```bash
sudo apt install build-essential
make --version
```

#### Windows
```powershell
# Using Chocolatey
choco install make

# Or use Git Bash / WSL for Make commands
```

---

## 🚀 Development Tools Installation

### 1. **VS Code + Extensions**

#### Install VS Code
```bash
# macOS
brew install --cask visual-studio-code

# Ubuntu/Debian
wget -qO- https://packages.microsoft.com/keys/microsoft.asc | gpg --dearmor > packages.microsoft.gpg
sudo install -o root -g root -m 644 packages.microsoft.gpg /etc/apt/trusted.gpg.d/
sudo sh -c 'echo "deb [arch=amd64,arm64,armhf signed-by=/etc/apt/trusted.gpg.d/packages.microsoft.gpg] https://packages.microsoft.com/repos/code stable main" > /etc/apt/sources.list.d/vscode.list'
sudo apt update
sudo apt install code

# Windows: Download from https://code.visualstudio.com/
```

#### Install Essential Extensions
```bash
# Install all recommended extensions
code --install-extension ms-vscode.vscode-typescript-next
code --install-extension golang.go
code --install-extension ms-python.python
code --install-extension ms-vscode.docker
code --install-extension humao.rest-client
code --install-extension ms-vscode.vscode-json
code --install-extension bradlc.vscode-tailwindcss
code --install-extension esbenp.prettier-vscode
code --install-extension ms-vscode.vscode-eslint
code --install-extension hashicorp.terraform
code --install-extension ms-kubernetes-tools.vscode-kubernetes-tools
```

### 2. **API Testing Tools**

#### Postman
```bash
# macOS
brew install --cask postman

# Ubuntu
sudo snap install postman

# Windows: Download from https://www.postman.com/downloads/
```

#### Insomnia (Alternative)
```bash
# macOS
brew install --cask insomnia

# Ubuntu
sudo snap install insomnia

# Windows: Download from https://insomnia.rest/download
```

#### HTTPie (Command Line)
```bash
# macOS
brew install httpie

# Ubuntu/Debian
sudo apt install httpie

# Windows
pip install httpie

# Verify
http --version
```

### 3. **Database GUI Tools**

#### TablePlus (Recommended)
```bash
# macOS
brew install --cask tableplus

# Windows/Linux: Download from https://tableplus.com/
```

#### pgAdmin (PostgreSQL)
```bash
# macOS
brew install --cask pgadmin4

# Ubuntu
curl https://www.pgadmin.org/static/packages_pgadmin_org.pub | sudo apt-key add
sudo sh -c 'echo "deb https://ftp.postgresql.org/pub/pgadmin/pgadmin4/apt/$(lsb_release -cs) pgadmin4 main" > /etc/apt/sources.list.d/pgadmin4.list'
sudo apt update
sudo apt install pgadmin4
```

#### Redis Desktop Manager
```bash
# macOS
brew install --cask redis

# Or use Redis CLI
redis-cli --version
```

### 4. **Development Utilities**

#### Hot Reload Tools
```bash
# For Node.js
npm install -g nodemon

# For Go
go install github.com/cosmtrek/air@latest

# For Python
pip install watchdog
```

#### Code Quality Tools
```bash
# Node.js tools
npm install -g eslint prettier

# Go tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install golang.org/x/tools/cmd/goimports@latest

# Python tools
pip install black flake8 mypy
```

#### API Tools
```bash
# Swagger tools
npm install -g @apidevtools/swagger-cli
npm install -g swagger-codegen-cli

# OpenAPI generator
npm install -g @openapitools/openapi-generator-cli
```

---

## ⚡ Performance and Testing Tools

### 1. **Load Testing Tools**

#### hey (Go-based HTTP load testing)
```bash
go install github.com/rakyll/hey@latest

# Verify
hey -version

# Example usage
hey -n 1000 -c 50 http://localhost:8080/health
```

#### Artillery (Node.js-based)
```bash
npm install -g artillery

# Verify
artillery version

# Example usage
artillery quick --count 10 --num 3 http://localhost:8080/health
```

#### Apache Bench
```bash
# macOS
brew install apache2

# Ubuntu
sudo apt install apache2-utils

# Example usage
ab -n 1000 -c 50 http://localhost:8080/health
```

#### wrk (Modern HTTP benchmarking)
```bash
# macOS
brew install wrk

# Ubuntu
sudo apt install wrk

# Example usage
wrk -t12 -c400 -d30s http://localhost:8080/health
```

### 2. **Monitoring Tools**

#### curl (HTTP client)
```bash
# Usually pre-installed, but to get latest:
# macOS
brew install curl

# Ubuntu
sudo apt install curl
```

#### jq (JSON processor)
```bash
# macOS
brew install jq

# Ubuntu
sudo apt install jq

# Windows
choco install jq

# Example usage
curl http://localhost:8080/health | jq
```

#### watch (Command monitoring)
```bash
# macOS
brew install watch

# Ubuntu (usually pre-installed)
sudo apt install procps

# Example usage
watch -n 2 'curl -s http://localhost:8080/health | jq'
```

---

## 🔧 Service-Specific Tools

### Node.js Development
```bash
# Package managers
npm install -g yarn pnpm

# Testing frameworks
npm install -g jest @jest/cli

# Development servers
npm install -g live-server serve

# Utility tools
npm install -g npm-check-updates
npm install -g rimraf cross-env
```

### Go Development
```bash
# Code generation
go install github.com/swaggo/swag/cmd/swag@latest

# Testing tools
go install github.com/onsi/ginkgo/v2/ginkgo@latest
go install github.com/golang/mock/mockgen@latest

# Debugging
go install github.com/go-delve/delve/cmd/dlv@latest

# Database migration
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### Python Development
```bash
# Virtual environment
pip install virtualenv pipenv

# Web frameworks
pip install fastapi uvicorn

# Testing
pip install pytest pytest-cov

# Code quality
pip install black isort autoflake

# Documentation
pip install sphinx mkdocs
```

---

## 🧪 Verification Script

Create a verification script to check all installations:

```bash
#!/bin/bash
# Save as verify-installation.sh

echo "🔍 Verifying X-Form Backend Development Environment"
echo "=================================================="

# Check Node.js
if command -v node &> /dev/null; then
    echo "✅ Node.js: $(node --version)"
else
    echo "❌ Node.js: Not installed"
fi

# Check npm
if command -v npm &> /dev/null; then
    echo "✅ npm: $(npm --version)"
else
    echo "❌ npm: Not installed"
fi

# Check Go
if command -v go &> /dev/null; then
    echo "✅ Go: $(go version)"
else
    echo "❌ Go: Not installed"
fi

# Check Python
if command -v python3 &> /dev/null; then
    echo "✅ Python: $(python3 --version)"
else
    echo "❌ Python: Not installed"
fi

# Check Docker
if command -v docker &> /dev/null; then
    echo "✅ Docker: $(docker --version)"
else
    echo "❌ Docker: Not installed"
fi

# Check Docker Compose
if command -v docker-compose &> /dev/null || docker compose version &> /dev/null; then
    echo "✅ Docker Compose: Available"
else
    echo "❌ Docker Compose: Not installed"
fi

# Check Git
if command -v git &> /dev/null; then
    echo "✅ Git: $(git --version)"
else
    echo "❌ Git: Not installed"
fi

# Check Make
if command -v make &> /dev/null; then
    echo "✅ Make: $(make --version | head -n1)"
else
    echo "❌ Make: Not installed"
fi

# Check optional tools
echo ""
echo "🚀 Optional Development Tools:"

if command -v code &> /dev/null; then
    echo "✅ VS Code: Available"
else
    echo "⚠️  VS Code: Not installed (recommended)"
fi

if command -v http &> /dev/null; then
    echo "✅ HTTPie: $(http --version)"
else
    echo "⚠️  HTTPie: Not installed (recommended)"
fi

if command -v hey &> /dev/null; then
    echo "✅ hey: Available"
else
    echo "⚠️  hey: Not installed (for load testing)"
fi

if command -v jq &> /dev/null; then
    echo "✅ jq: $(jq --version)"
else
    echo "⚠️  jq: Not installed (for JSON processing)"
fi

echo ""
echo "🎯 Setup X-Form Backend:"
echo "git clone <repo-url>"
echo "cd X-Form-Backend"
echo "make setup"
echo "make dev"
```

Make it executable and run:
```bash
chmod +x verify-installation.sh
./verify-installation.sh
```

---

## 🆘 Troubleshooting

### Common Installation Issues

#### 1. **Permission Issues**
```bash
# Fix npm permissions (macOS/Linux)
sudo chown -R $(whoami) ~/.npm
sudo chown -R $(whoami) /usr/local/lib/node_modules

# Or use nvm instead of system Node.js
```

#### 2. **Path Issues**
```bash
# Add to ~/.zshrc or ~/.bashrc
export PATH=$PATH:/usr/local/go/bin
export PATH=$PATH:$HOME/go/bin
export PATH=$PATH:$HOME/.local/bin

# Reload shell
source ~/.zshrc  # or ~/.bashrc
```

#### 3. **Docker Issues**
```bash
# Start Docker service (Linux)
sudo systemctl start docker
sudo systemctl enable docker

# Add user to docker group
sudo usermod -aG docker $USER
newgrp docker
```

#### 4. **Go Module Issues**
```bash
# Enable Go modules
export GO111MODULE=on

# Clean module cache
go clean -modcache
```

### Platform-Specific Notes

#### macOS
- Use Homebrew for most installations
- Install Xcode Command Line Tools: `xcode-select --install`
- Consider using iTerm2 instead of default Terminal

#### Ubuntu/Debian
- Update package lists: `sudo apt update`
- Install build essentials: `sudo apt install build-essential`
- Consider using fish or zsh shell

#### Windows
- Use WSL2 for better Docker performance
- Install Windows Terminal for better CLI experience
- Consider using Chocolatey for package management

---

**🎉 You're all set for X-Form Backend development!**

After installing these tools, proceed to the [Local Development Guide](LOCAL_DEVELOPMENT_COMPLETE_GUIDE.md) to start developing.
