# API Design

This document details the REST API endpoints, method types, request payloads, and response payloads.

All endpoints are prefixed with `/api/v1`.

## Auth Endpoints
- `POST /api/v1/auth/register` - Create new user account.
- `POST /api/v1/auth/login` - Authenticate user and return JWT.

## Category Endpoints
- `POST /api/v1/categories` - Create a new product category (Requires Auth).
- `GET /api/v1/categories` - Get list of categories.

## Product Endpoints
- `POST /api/v1/products` - Create a new product (Requires Auth).
- `PUT /api/v1/products/:id` - Update product details (Requires Auth).
- `DELETE /api/v1/products/:id` - Delete product (Requires Auth).
- `GET /api/v1/products` - List products (cached in Redis, supports pagination).

## Cart Endpoints
- `GET /api/v1/cart` - View user's cart (Requires Auth).
- `POST /api/v1/cart/items` - Add product to cart / update quantity (Requires Auth).
- `DELETE /api/v1/cart/items/:productId` - Remove product from cart (Requires Auth).

## Order Endpoints
- `POST /api/v1/orders` - Place an order from cart items (Requires Auth).
- `GET /api/v1/orders` - View order history (Requires Auth).
- `GET /api/v1/orders/:id` - View order details (Requires Auth).

## Standard Response Format

### Success
```json
{
  "success": true,
  "data": {}
}
```

### Error
```json
{
  "success": false,
  "error": {
    "message": "Human readable error message",
    "code": "ERROR_CODE"
  }
}
```
