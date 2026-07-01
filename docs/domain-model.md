# Domain Model

This document defines the domain entities and aggregate roots for ShopFlow.

## Entities

### User
- `ID`: Unique identifier (UUID or Auto-increment)
- `Email`: Unique string
- `PasswordHash`: Hashed string
- `Name`: String
- `CreatedAt`/`UpdatedAt`: Timestamps

### Category
- `ID`: Unique identifier
- `Name`: Unique string
- `Description`: String
- `CreatedAt`/`UpdatedAt`: Timestamps

### Product
- `ID`: Unique identifier
- `CategoryID`: Foreign key to Category
- `Name`: String
- `Description`: String
- `Price`: Decimal/Numeric (stored as cents or fixed-point integer)
- `Stock`: Integer
- `CreatedAt`/`UpdatedAt`: Timestamps

### Cart
- `ID`: Unique identifier
- `UserID`: Foreign key to User
- `Items`: List of CartItem

### CartItem
- `ProductID`: Foreign key to Product
- `Quantity`: Integer

### Order
- `ID`: Unique identifier
- `UserID`: Foreign key to User
- `Status`: String (e.g., Pending, Paid, Completed, Cancelled)
- `TotalAmount`: Decimal/Numeric
- `Items`: List of OrderItem
- `CreatedAt`/`UpdatedAt`: Timestamps

### OrderItem
- `ID`: Unique identifier
- `OrderID`: Foreign key to Order
- `ProductID`: Foreign key to Product
- `Quantity`: Integer
- `PriceAtPurchase`: Decimal/Numeric
