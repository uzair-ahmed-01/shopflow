# Project Requirements

This document defines the functional and non-functional requirements for ShopFlow.

## Functional Requirements

### 1. Authentication
- User Registration (email, password, name) with hashed passwords.
- User Login (email, password) returning a JWT.
- Protect private endpoints via a JWT authentication middleware.

### 2. Category Module
- Create Category (admin/authenticated).
- List Categories.

### 3. Product Module
- Create Product (associated with a Category).
- Update Product details.
- Delete Product.
- List Products with pagination and Redis caching.

### 4. Cart Module
- Add Product to user's cart (specifying quantity).
- Remove Product from user's cart.
- View Cart contents and subtotal.

### 5. Order Module
- Place Order (convert cart items to an order).
- View Orders history.
- View specific Order Details.
- Emit `OrderCreated` event upon successful order placement.

### 6. Background Processing
- Asynchronous Worker Pool to process background jobs.
- Update product inventory on order placement.
- Send mock email/notification on order completion.

## Non-Functional Requirements
- **Performance**: Use Redis to cache the product list to reduce DB load.
- **Robustness**: Proper database transactions when updating inventory and creating orders.
- **Graceful Shutdown**: The API server and background workers must handle OS termination signals.
