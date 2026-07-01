# Learning Notes

This document acts as a log for backend engineering takeaways from building ShopFlow.

## Clean Architecture
- Delivery layer (Handlers) should not leak any database details or models.
- Service layer contains all business logic and controls transactions.
- Repository layer abstracts database interaction.

## Concurrency and Worker Pools
- TBD

## Redis Caching
- TBD

## PostgreSQL & Event Consistency
- TBD
