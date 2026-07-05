# Domain Model

This document defines the domain entities, aggregates, properties, business rules, and constraints for ShopFlow.

## Domain Entities

### User
Represents an identity registered on the platform.
*   **Properties**:
    *   `ID` (integer): Auto-increment primary key.
    *   `Name` (string): User's full name. Must be 2-100 characters.
    *   `Email` (string): Unique email address. Must match RFC 5322 email format.
    *   `PasswordHash` (string): Secure bcrypt hash. Original password is never stored.
    *   `Role` (string): Permission tier (`customer` or `admin`). Default: `customer`.
    *   `CreatedAt` (time.Time): Timestamp of registration.
    *   `UpdatedAt` (time.Time): Timestamp of last profile update.
*   **Business Rules**:
    *   Emails are case-insensitive and converted to lowercase before saving.
    *   A newly created user is automatically assigned an empty Cart.
    *   Roles regulate administrative API endpoint clearance (RBAC authorization).

### Category
Represents a group under which products are cataloged.
*   **Properties**:
    *   `ID` (integer): Auto-increment primary key.
    *   `Name` (string): Unique category name. Must be 2-50 characters.
    *   `Description` (string): Description of products in this category.
    *   `CreatedAt` (time.Time): Timestamp of creation.
    *   `UpdatedAt` (time.Time): Timestamp of last metadata update.
*   **Business Rules**:
    *   A Category cannot be deleted if there are products assigned to it (referenced integrity).

### Product
Represents an item available for purchase.
*   **Properties**:
    *   `ID` (integer): Auto-increment primary key.
    *   `CategoryID` (integer): References associated category. Must be valid.
    *   `Name` (string): Product title. Must be 2-100 characters.
    *   `Description` (string): Detailed specifications.
    *   `Price` (integer): Price represented in cents (e.g., $10.99 is stored as `1099`). Must be > 0.
    *   `Stock` (integer): Number of units physically in stock. Must be >= 0.
    *   `CreatedAt` (time.Time): Timestamp of creation.
    *   `UpdatedAt` (time.Time): Timestamp of last catalog update.
*   **Business Rules**:
    *   Products must belong to exactly one Category.
    *   Stock updates must be atomic. Stock cannot go below zero.

### Cart
Represents the current shopping cart state for a user.
*   **Properties**:
    *   `ID` (integer): Auto-increment primary key.
    *   `UserID` (integer): Unique key referencing the User.
    *   `Items` ([]CartItem): Collection of items in the cart.
*   **Business Rules**:
    *   One cart per User.
    *   Cart subtotal is calculated as the sum of `Price * Quantity` for all items.

### CartItem
An item quantity in a Cart.
*   **Properties**:
    *   `CartID` (integer): References parent Cart.
    *   `ProductID` (integer): References selected Product.
    *   `Quantity` (integer): Must be >= 1.
*   **Business Rules**:
    *   Adding an item already in the cart increases `Quantity` of the existing record.
    *   Quantity must not exceed available Product stock at the time of modification.

### Order
Represents a finalized purchase invoice.
*   **Properties**:
    *   `ID` (integer): Auto-increment primary key.
    *   `UserID` (integer): References User who placed the order.
    *   `Status` (OrderStatus): State machine (values: `PENDING`, `PAID`, `CANCELLED`).
    *   `TotalAmount` (integer): Cumulative purchase price of all items in cents.
    *   `Items` ([]OrderItem): Items snapshot.
    *   `CreatedAt` (time.Time): Timestamp of placement.
    *   `UpdatedAt` (time.Time): Timestamp of status transitions.
*   **OrderStatus Transitions**:
    ```
    [PENDING] ----(Payment Success)---> [PAID]
       |
       +-------(Payment Failure)-----> [CANCELLED]
    ```

### OrderItem
An immutable historical snapshot of a purchased product inside an Order.
*   **Properties**:
    *   `ID` (integer): Auto-increment primary key.
    *   `OrderID` (integer): References parent Order.
    *   `ProductID` (integer): References original Product.
    *   `Quantity` (integer): Must be >= 1.
    *   `PriceAtPurchase` (integer): Cost in cents at the moment checkout occurred.
*   **Business Rules**:
    *   `PriceAtPurchase` is fixed and does not change even if the current product price changes in catalog database.

### RefreshToken
Represents a token used to maintain user authentication sessions.
*   **Properties**:
    *   `ID` (integer): Auto-increment primary key.
    *   `UserID` (integer): References the User who owns the session.
    *   `Token` (string): Cryptographically secure random 32-byte hex token.
    *   `ExpiresAt` (time.Time): Absolute time boundary when token becomes invalid.
    *   `CreatedAt` (time.Time): Timestamp of token generation.
    *   `RevokedAt` (*time.Time): Optional timestamp indicating manual logout or session revocation.
*   **Business Rules**:
    *   A RefreshToken is invalid if `ExpiresAt` is in the past, or if `RevokedAt` is populated.
    *   Used refresh tokens are rotated during token refresh requests.
