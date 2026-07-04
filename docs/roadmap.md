# Project Roadmap

This document outlines the milestones and timeline of ShopFlow.

## Phase 1: Project Initialization & Design

- [x] Initial design rules and agent guidelines
- [x] Initialize project structure and Go module
- [x] Write detailed requirements and domain model
- [x] Define ER diagram and database design
- [x] Design REST APIs (endpoints, requests, responses)

## Phase 2: Core Modules & Architecture

- [x] Authentication Module (JWT, Password hashing, Middleware)
- [x] Category Module
- [x] Product Module (CRUD, pointer-based partial updates)
- [x] Cart Module
- [x] Order Module (Transactional placements, stock verification, cart cleanup)

## Phase 3: Background Processing & Concurrency

- [x] Worker pool execution for background order status transitions (PENDING -> PROCESSING)
- [x] Graceful worker shutdowns and cancellation context

## Phase 4: Containerization & Deployment

- [ ] Dockerfiles for API service
- [ ] Docker Compose setup (API, Postgres, Redis)

## Phase 5: Verification & Documentation

- [ ] Unit testing (Services)
- [ ] Integration testing (Repositories)
- [x] Fill learning-notes.md
- [x] Create Go concurrency study guide (go-concurrency-guide.md)

## Phase 6: Future Enhancements & Postponed Modules

- [ ] **Redis Integration**: Cache product details and lists to reduce PostgreSQL query load.
- [ ] **Event-Driven Architecture**: Publish events like `OrderCreated` to a message broker (RabbitMQ/Redis Pub-Sub).
- [ ] **Asynchronous Micro-services**: Handle mock email dispatch, push notifications, and analytics in background subscribers.
- [ ] **Distributed Task Queue**: Upgrade in-memory worker pool to **Asynq** or **River** for multi-server reliability.
- [ ] **Dockerization**: Complete containerization deployment using Docker Compose.
