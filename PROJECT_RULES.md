# PROJECT_RULES.md

# ShopFlow Project Rules

> This document defines the project scope, engineering decisions, development strategy, and long-term roadmap.

---

# Project Vision

ShopFlow is **not** an Amazon clone.

It is a **Production-Style Backend Showcase**.

The goal is to demonstrate modern backend engineering concepts using a single, well-designed project.

Every major backend concept should be implemented at least once in a clean and understandable way.

---

# Primary Objective

The purpose of this project is to showcase the following backend engineering skills:

* Go
* PostgreSQL
* Redis
* JWT Authentication
* Docker
* Docker Compose
* Clean Architecture
* Repository Pattern
* Service Layer
* Middleware
* Worker Pools
* Goroutines
* Channels
* Event-Driven Architecture
* Background Jobs
* Unit Testing
* Integration Testing
* Git & GitHub
* AI-Assisted Development

---

# Project Scope

This project intentionally implements only one complete example of each major backend concept.

The goal is learning and demonstrating engineering skills, not recreating a commercial e-commerce platform.

---

# Core Modules

## Authentication

Features

* Register
* Login
* JWT Authentication
* Protected Routes

---

## Product Module

Features

* Create Product
* Update Product
* Delete Product
* List Products

---

## Category Module

Features

* Create Category
* List Categories

---

## Cart Module

Features

* Add Product
* Remove Product
* View Cart

---

## Order Module

Features

* Place Order
* View Orders
* View Order Details

---

## Event System

Implement one production-style event.

Example

OrderCreated

This event should trigger background processing.

---

## Worker Pool

Implement one worker pool that processes events asynchronously.

Examples

* Inventory Update
* Notification
* Email (Mock)

---

## Redis

Implement one real caching use case.

Recommended:

Product List Cache

---

## Docker

Containerize

* API
* PostgreSQL
* Redis

Application should run using Docker Compose.

---

## Testing

Implement

* Unit Tests
* Integration Tests

Focus on quality over quantity.

---

## AI Feature

Implement one practical AI feature.

Recommended

Generate Product Description

or

Generate Product Summary

The AI feature must solve a real business problem.

---

# Out of Scope

The following features are intentionally excluded.

Do NOT implement unless the project scope changes.

* Payment Gateway
* Wishlist
* Reviews & Ratings
* Coupons
* Shipping Integration
* Multi Address Support
* Returns
* Refunds
* Recommendation Engine
* Analytics Dashboard UI
* Multi Vendor Marketplace
* Real Email Service
* SMS Service
* Third-Party Logistics
* Mobile Application

---

# Development Workflow

Every feature follows the same lifecycle.

Requirements

↓

Design

↓

Implementation

↓

Testing

↓

Documentation

↓

Git Commit

↓

Git Push

No shortcuts.

---

# Git Workflow

Every completed feature must have:

* Working code
* Documentation
* Meaningful commit message

Examples

docs: complete database design

feat: implement jwt authentication

feat: implement product module

feat: add redis caching

test: add authentication service tests

---

# Documentation Rules

Documentation is mandatory.

Maintain

README.md

docs/

* roadmap.md
* requirements.md
* domain-model.md
* database-design.md
* api-design.md
* architecture.md
* deployment.md
* learning-notes.md

Every completed phase should update the relevant document.

---

# Code Review Checklist

Before every commit, verify:

* Project builds successfully.
* No unused code.
* Error handling is implemented.
* Documentation is updated.
* Functions remain small and readable.
* Business logic stays inside the service layer.
* Database logic stays inside the repository layer.

---

# Engineering Principles

Always prefer

* Readability
* Simplicity
* Maintainability
* Scalability
* Testability

Avoid unnecessary abstraction.

Keep architecture clean.

Implement only what is needed today.

Design so that future features can be added without major rewrites.

---

# Learning Goal

By the end of this project, I should be able to confidently explain:

* Why this architecture was chosen
* Database relationships
* Repository Pattern
* Service Layer
* Dependency Injection
* JWT Authentication
* PostgreSQL Design
* Transactions
* Redis Caching
* Event-Driven Architecture
* Worker Pools
* Goroutines
* Docker
* Testing Strategy
* AI Integration

without relying on notes.

---

# Definition of Done

A feature is considered complete only when:

* Requirements are implemented.
* Code compiles successfully.
* Tests pass.
* Documentation is updated.
* Git commit is created.
* Code is understandable and interview-ready.

---

# Success Criteria

This repository should demonstrate the mindset of a backend engineer rather than the number of implemented features.

The final project should be something that can be confidently presented during technical interviews and used as a long-term portfolio project.
