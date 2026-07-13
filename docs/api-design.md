# API Design Specification

This document details the REST API endpoints, request payloads, response schemas, and HTTP status codes for ShopFlow.

All endpoints are prefixed with `/api/v1`.

## Interactive Swagger API Documentation
An interactive Swagger UI is available to explore and test endpoints in real time:
- **URL**: `http://localhost:8080/swagger/index.html`
- **Re-generation**: Run `make swagger-gen` after modifying handler comments to regenerate the documentation specs.

---

## Standard JSON Response Formats

### Success Response Envelope
Returned for successful operations. Returns 200 OK, 201 Created, or 204 No Content.
```json
{
  "success": true,
  "data": {}
}
```

### Error Response Envelope
Returned when validation fails, permissions are lacking, or resources are missing. Status codes range from 400 to 500.
```json
{
  "success": false,
  "error": {
    "message": "Detailed explanation of what failed.",
    "code": "BAD_REQUEST"
  }
}
```

---

## Auth Endpoints

### 1. Register User
*   **Path**: `POST /api/v1/auth/register`
*   **Authentication**: None
*   **Request Body**:
    ```json
    {
      "name": "John Doe",
      "email": "john.doe@example.com",
      "password": "securepassword123"
    }
    ```
*   **Success Response**: `201 Created`
    ```json
    {
      "success": true,
      "data": {
        "user_id": 1,
        "name": "John Doe",
        "email": "john.doe@example.com",
        "created_at": "2026-07-01T14:00:00Z"
      }
    }
    ```

### 2. Login User
*   **Path**: `POST /api/v1/auth/login`
*   **Authentication**: None
*   **Request Body**:
    ```json
    {
      "email": "john.doe@example.com",
      "password": "securepassword123"
    }
    ```
*   **Success Response**: `200 OK`
    ```json
    {
      "success": true,
      "data": {
        "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "refresh_token": "8dfa27a8106f0e4b85c3b123...",
        "expires_in_seconds": 900
      }
    }
    ```

### 3. Refresh Access Token
*   **Path**: `POST /api/v1/auth/refresh`
*   **Authentication**: None (Requires valid active refresh token in payload)
*   **Request Body**:
    ```json
    {
      "refresh_token": "8dfa27a8106f0e4b85c3b123..."
    }
    ```
*   **Success Response**: `200 OK`
    ```json
    {
      "success": true,
      "data": {
        "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "refresh_token": "92ea81c2f901ab8872bca532...",
        "expires_in_seconds": 900
      }
    }
    ```

### 4. Logout User
*   **Path**: `POST /api/v1/auth/logout`
*   **Authentication**: None (Requires active refresh token in payload to invalidate session)
*   **Request Body**:
    ```json
    {
      "refresh_token": "8dfa27a8106f0e4b85c3b123..."
    }
    ```
*   **Success Response**: `200 OK`
    ```json
    {
      "success": true,
      "data": {
        "message": "successfully logged out"
      }
    }
    ```

---

## Category Endpoints

### 1. Create Category
*   **Path**: `POST /api/v1/categories`
*   **Headers**: `Authorization: Bearer <token>` (Requires **Admin** role)
*   **Request Body**:
    ```json
    {
      "name": "Electronics",
      "description": "Smartphones, Laptops, Accessories"
    }
    ```
*   **Success Response**: `201 Created`
    ```json
    {
      "success": true,
      "data": {
        "id": 1,
        "name": "Electronics",
        "description": "Smartphones, Laptops, Accessories"
      }
    }
    ```

### 2. List Categories
*   **Path**: `GET /api/v1/categories`
*   **Authentication**: None
*   **Success Response**: `200 OK`
    ```json
    {
      "success": true,
      "data": [
        {
          "id": 1,
          "name": "Electronics",
          "description": "Smartphones, Laptops, Accessories"
        }
      ]
    }
    ```

---

## Product Endpoints

### 1. Create Product
*   **Path**: `POST /api/v1/products`
*   **Headers**: `Authorization: Bearer <token>` (Requires **Admin** role)
*   **Request Body**:
    ```json
    {
      "category_id": 1,
      "name": "iPhone 15 Pro",
      "description": "128GB Black Titanium",
      "price": 99900,
      "stock": 50
    }
    ```
*   **Success Response**: `201 Created`
    ```json
    {
      "success": true,
      "data": {
        "id": 101,
        "category_id": 1,
        "name": "iPhone 15 Pro",
        "description": "128GB Black Titanium",
        "price": 99900,
        "stock": 50,
        "created_at": "2026-07-01T14:10:00Z"
      }
    }
    ```

### 2. Update Product
*   **Path**: `PUT /api/v1/products/:id`
*   **Headers**: `Authorization: Bearer <token>` (Requires **Admin** role)
*   **Request Body** (Partial updates supported):
    ```json
    {
      "price": 94900,
      "stock": 45
    }
    ```
*   **Success Response**: `200 OK`
    ```json
    {
      "success": true,
      "data": {
        "id": 101,
        "category_id": 1,
        "name": "iPhone 15 Pro",
        "description": "128GB Black Titanium",
        "price": 94900,
        "stock": 45,
        "updated_at": "2026-07-01T14:15:00Z"
      }
    }
    ```

### 3. Delete Product
*   **Path**: `DELETE /api/v1/products/:id`
*   **Headers**: `Authorization: Bearer <token>` (Requires **Admin** role)
*   **Success Response**: `204 No Content`

### 4. List Products (Cached)
*   **Path**: `GET /api/v1/products?page=1&limit=10`
*   **Authentication**: None
*   **Success Response**: `200 OK`
    ```json
    {
      "success": true,
      "data": {
        "products": [
          {
            "id": 101,
            "category_id": 1,
            "name": "iPhone 15 Pro",
            "price": 94900,
            "stock": 45
          }
        ],
        "pagination": {
          "current_page": 1,
          "limit": 10,
          "total_items": 1
        }
      }
    }
    ```

---

## Cart Endpoints

### 1. View Cart
*   **Path**: `GET /api/v1/cart`
*   **Headers**: `Authorization: Bearer <token>`
*   **Success Response**: `200 OK`
    ```json
    {
      "success": true,
      "data": {
        "cart_id": 12,
        "items": [
          {
            "product_id": 101,
            "product_name": "iPhone 15 Pro",
            "price": 94900,
            "quantity": 2
          }
        ],
        "total": 189800
      }
    }
    ```

### 2. Add / Update Cart Item
*   **Path**: `POST /api/v1/cart/items`
*   **Headers**: `Authorization: Bearer <token>`
*   **Request Body**:
    ```json
    {
      "product_id": 101,
      "quantity": 2
    }
    ```
*   **Success Response**: `200 OK`
    ```json
    {
      "success": true,
      "data": {
        "product_id": 101,
        "quantity": 2
      }
    }
    ```

### 3. Remove Cart Item
*   **Path**: `DELETE /api/v1/cart/items/:productId`
*   **Headers**: `Authorization: Bearer <token>`
*   **Success Response**: `204 No Content`

---

## Order Endpoints

### 1. Place Order
*   **Path**: `POST /api/v1/orders`
*   **Headers**: `Authorization: Bearer <token>`
*   **Request Body**: None (checkout process uses active cart items)
*   **Success Response**: `201 Created`
    ```json
    {
      "success": true,
      "data": {
        "order_id": 501,
        "total_amount": 189800,
        "status": "PENDING",
        "created_at": "2026-07-01T14:20:00Z"
      }
    }
    ```

### 2. View Orders History
*   **Path**: `GET /api/v1/orders`
*   **Headers**: `Authorization: Bearer <token>`
*   **Success Response**: `200 OK`
    ```json
    {
      "success": true,
      "data": [
        {
          "order_id": 501,
          "total_amount": 189800,
          "status": "PENDING",
          "created_at": "2026-07-01T14:20:00Z"
        }
      ]
    }
    ```

### 3. View Order Details
*   **Path**: `GET /api/v1/orders/:id`
*   **Headers**: `Authorization: Bearer <token>`
*   **Success Response**: `200 OK`
    ```json
    {
      "success": true,
      "data": {
        "order_id": 501,
        "status": "PENDING",
        "total_amount": 189800,
        "items": [
          {
            "product_id": 101,
            "product_name": "iPhone 15 Pro",
            "quantity": 2,
            "price_at_purchase": 94900
          }
        ],
        "created_at": "2026-07-01T14:20:00Z"
      }
    }
    ```
