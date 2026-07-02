# Agent Rules

## Git Commits

- Always output the git commit commands at the end of the response whenever files are modified, created, or deleted.

---

# ShopFlow AI Engineering Guide

> This document defines how AI coding agents should work on this repository.
>
> Every code generation, refactoring, and implementation must follow these rules.

---

# Project Overview

Project Name:

**ShopFlow**

Goal:

Build a production-style backend showcase that demonstrates modern backend engineering concepts using Go.

This project is **NOT** intended to become a complete Amazon clone.

The objective is to implement every major backend concept at least once using clean, production-quality code.

---

# Tech Stack

Language

- Go (latest stable version)

Database

- PostgreSQL

Cache

- Redis

Authentication

- JWT

Containerization

- Docker
- Docker Compose

Architecture

- Clean Architecture
- Repository Pattern
- Service Layer

Concurrency

- Goroutines
- Channels
- Worker Pools

Background Processing

- Event-Driven Architecture
- Background Jobs

Testing

- Unit Tests
- Integration Tests

Version Control

- Git
- GitHub

AI

- AI-assisted development
- AI-generated product description (future phase)

---

# Project Philosophy

Always prioritize:

1. Simplicity
2. Readability
3. Maintainability
4. Scalability
5. Testability

Never over-engineer.

If two solutions work, prefer the simpler one.

---

# Development Workflow

Always follow this order.

Requirements

↓

Domain Model

↓

ER Diagram

↓

Database Design

↓

API Design

↓

Architecture

↓

Implementation

↓

Testing

↓

Documentation

↓

Deployment

Never skip design.

Never implement features before the architecture is approved.

---

# Clean Architecture Rules

Keep responsibilities separated.

Handlers

- HTTP request/response only
- Validation
- Call services
- No database logic

Services

- Business logic
- Transactions
- Call repositories
- No HTTP logic

Repositories

- PostgreSQL queries
- Data persistence only

Models

- Domain models
- Database models

Middleware

- JWT
- Logging
- Recovery
- Authorization

Workers

- Background processing
- Event consumers

Events

- Event publishing
- Event handling

Cache

- Redis operations only

---

# Database Rules

Always use PostgreSQL.

Never duplicate data without business justification.

Always use:

- Primary Keys
- Foreign Keys
- Unique Constraints
- Proper Indexes

Prefer normalization.

Document every relationship.

---

# API Rules

Follow REST conventions.

Use:

/api/v1/

Return JSON only.

Use consistent response structures.

Use appropriate HTTP status codes.

Never expose internal errors to clients.

---

# Authentication Rules

Use JWT.

Passwords must always be hashed.

Never store plain text passwords.

Protect private routes with middleware.

---

# Error Handling

Errors must be meaningful.

Wrap errors with context.

Never ignore returned errors.

Never panic for expected application errors.

---

# Logging

Log important events.

Avoid unnecessary logs.

Do not log secrets.

Do not log passwords.

Do not log JWT tokens.

---

# Redis Rules

Use Redis only where caching provides value.

Prefer Cache-Aside pattern.

Do not cache everything.

---

# Worker Pool Rules

Worker Pools must use:

- Goroutines
- Channels
- Context cancellation

Workers must be graceful.

Workers must be testable.

---

# Docker Rules

The application must run using Docker Compose.

Services:

- API
- PostgreSQL
- Redis

Running the project should require minimal setup.

---

# Testing Rules

Every major service should have unit tests.

Repository layer should have integration tests.

Prefer table-driven tests where appropriate.

---

# Git Rules

Every completed task must end with a commit.

Commit messages must follow Conventional Commits.

Examples:

feat:
fix:
refactor:
docs:
test:
chore:

Never mix unrelated changes into one commit.

---

# Documentation Rules

Update documentation whenever architecture or business logic changes.

Maintain:

- README.md
- docs/domain-model.md
- docs/database-design.md
- docs/api-design.md
- docs/architecture.md
- docs/deployment.md
- docs/learning-notes.md

Documentation is part of the project.

---

# AI Agent Rules

Before writing code:

- Understand the requirement.
- Respect the current architecture.
- Reuse existing code whenever possible.

Do not:

- Rewrite unrelated files.
- Change project structure without approval.
- Introduce unnecessary dependencies.
- Generate unused code.
- Leave TODO placeholders unless requested.

When implementing a feature:

1. Explain the approach.
2. Implement only the requested scope.
3. Keep code production-ready.
4. Keep functions small and readable.
5. Prefer explicit code over clever code.

---

# Coding Standards

Write readable Go code.

Prefer descriptive names.

Keep functions focused on one responsibility.

Avoid deeply nested code.

Favor composition over duplication.

Use dependency injection.

Follow idiomatic Go practices.

---

# Project Goal

The purpose of this repository is not to build the largest application.

The purpose is to demonstrate mastery of modern backend engineering concepts, including:

- Clean Architecture
- Repository Pattern
- Service Layer
- PostgreSQL
- Redis
- JWT
- Middleware
- Event-Driven Architecture
- Worker Pools
- Background Jobs
- Docker
- Testing
- AI-Assisted Development

Every implementation should be interview-ready and easy to explain.
