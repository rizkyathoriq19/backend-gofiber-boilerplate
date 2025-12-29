# Contributing to Go Fiber Boilerplate

Thank you for your interest in contributing! This document provides guidelines for contributing to this project.

## 📋 Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Pull Request Process](#pull-request-process)
- [Coding Standards](#coding-standards)

## Code of Conduct

Please be respectful and constructive in all interactions. We welcome contributors of all experience levels.

## Getting Started

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/backend-gofiber-boilerplate.git
   cd backend-gofiber-boilerplate
   ```
3. Add the upstream remote:
   ```bash
   git remote add upstream https://github.com/ORIGINAL_OWNER/backend-gofiber-boilerplate.git
   ```

## Development Setup

### Prerequisites

- Go 1.24+
- PostgreSQL 17+
- Redis 7+
- Docker & Docker Compose (optional)

### Local Setup

```bash
# Copy environment file
cp .env.example .env

# Install dependencies
go mod download

# Run database migrations
make migrate-up

# Start development server with hot reload
air
```

### Docker Setup

```bash
# Development with hot reload
docker-compose -f docker-compose.dev.yml up

# Production build
docker-compose up --build
```

## Making Changes

1. Create a feature branch:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make your changes following our [coding standards](#coding-standards)

3. Run tests:
   ```bash
   make test
   ```

4. Run linter:
   ```bash
   golangci-lint run
   ```

5. Commit your changes:
   ```bash
   git commit -m "feat: add your feature description"
   ```

### Commit Message Format

We follow [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation changes
- `style:` - Code style changes (formatting, etc.)
- `refactor:` - Code refactoring
- `test:` - Adding or updating tests
- `chore:` - Maintenance tasks

## Pull Request Process

1. Update documentation if needed
2. Ensure all tests pass
3. Update the README.md if you've added new features
4. Create a Pull Request with a clear title and description
5. Wait for review and address any feedback

## Coding Standards

### Go Guidelines

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Keep functions small and focused
- Write meaningful comments for exported functions
- Handle errors explicitly

### Project Structure

```
internal/
├── config/          # Configuration
├── database/        # Database connections
├── middleware/      # HTTP middleware
├── module/          # Feature modules (Clean Architecture)
│   └── {module}/
│       ├── controller.go   # HTTP handlers
│       ├── service.go      # Business logic
│       ├── repository.go   # Data access
│       ├── entity.go       # Database entities
│       ├── domain.go       # Interfaces
│       ├── request.go      # Request DTOs
│       └── response.go     # Response DTOs
└── pkg/             # Shared packages
```

### Testing

- Write unit tests for services
- Write integration tests for handlers
- Use table-driven tests when appropriate
- Aim for meaningful coverage, not 100%

## Questions?

Feel free to open an issue for any questions or concerns.
