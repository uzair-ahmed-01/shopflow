# Project Roadmap

This document outlines the milestones and timeline of ShopFlow.

## Phase 1: Project Initialization & Design

- [x] Initial design rules and agent guidelines
- [x] Initialize project structure and Go module
- [x] Write detailed requirements and domain model
- [x] Define ER diagram and database design
- [x] Design REST APIs (endpoints, requests, responses)

## Phase 2: Core Modules & Architecture

- [ ] Authentication Module (JWT, Password hashing, Middleware)
- [ ] Category Module
- [ ] Product Module & Caching (Redis integration)
- [ ] Cart Module
- [ ] Order Module & Events

## Phase 3: Background Processing & Concurrency

- [ ] Worker pool execution for events (Inventory, notifications, mock email)
- [ ] Graceful worker shutdowns and cancellation context

## Phase 4: Containerization & Deployment

- [ ] Dockerfiles for API service
- [ ] Docker Compose setup (API, Postgres, Redis)

## Phase 5: Verification & Documentation

- [ ] Unit testing (Services)
- [ ] Integration testing (Repositories)
- [ ] Fill learning-notes.md
