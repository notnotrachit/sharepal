# Splitwise Alternative API

A complete expense-sharing REST API built with Go, MongoDB, and Redis. This API provides all the functionality needed to build a Splitwise-like application for splitting expenses among friends and groups.

## Features

### 🔐 Authentication & Users
- User registration and login with JWT tokens
- Token refresh mechanism
- Secure password hashing

### 👥 Friend Management
- Send/accept/reject friend requests
- View friends list
- Block users
- View sent and received friend requests

### 🏘️ Group Management
- Create and manage expense groups
- Add/remove group members
- Group permissions (only creator can remove members)
- Soft delete groups

### 💰 Expense Tracking
- Create expenses with multiple split types:
  - **Equal Split**: Divide equally among participants
  - **Exact Amount**: Specify exact amounts for each person
  - **Percentage Split**: Split by percentage
- Track who paid and who owes
- Categories for expenses
- Add notes and receipts
- Update and delete expenses

### ⚖️ Balance & Settlement Management
- Real-time balance calculation
- Smart debt simplification algorithm
- Settlement tracking and completion
- Group and individual balance views

## API Endpoints

### Authentication
```
POST /v1/auth/register     # Register new user
POST /v1/auth/login        # User login
POST /v1/auth/refresh      # Refresh tokens
```

### Friends
```
GET    /v1/friends                     # Get friends list
POST   /v1/friends/request             # Send friend request
POST   /v1/friends/request/:id/respond # Accept/reject request
GET    /v1/friends/requests/received   # Get received requests
GET    /v1/friends/requests/sent       # Get sent requests
DELETE /v1/friends/:friendId           # Remove friend
POST   /v1/friends/block/:userId       # Block user
```

### Groups
```
POST   /v1/groups                 # Create group
GET    /v1/groups                 # Get user's groups
GET    /v1/groups/:id             # Get group details
DELETE /v1/groups/:id             # Delete group
POST   /v1/groups/:id/members     # Add member
GET    /v1/groups/:id/members     # Get members
DELETE /v1/groups/:id/members/:id # Remove member
GET    /v1/groups/:id/expenses    # Get group expenses
GET    /v1/groups/:id/balances    # Get balances
GET    /v1/groups/:id/simplify    # Get simplified debts
GET    /v1/groups/:id/settlements # Get settlements
```

### Expenses
```
POST   /v1/expenses     # Create expense
GET    /v1/expenses     # Get user expenses
GET    /v1/expenses/:id # Get expense details
PUT    /v1/expenses/:id # Update expense
DELETE /v1/expenses/:id # Delete expense
```

### Settlements
```
GET  /v1/settlements             # Get user settlements
POST /v1/settlements/:id/complete # Mark settlement complete
```

## Quick Start

### Prerequisites
- Go 1.20+
- MongoDB
- Redis (optional)

### Installation

1. **Clone the repository:**
```bash
git clone <your-repo-url>
cd go-rest-api-starter
```

2. **Set up environment:**
```bash
cp .env.example .env
# Edit .env with your MongoDB URI and other settings
```

3. **Install dependencies:**
```bash
go mod download
```

4. **Build and run:**
```bash
go build -o splitwise-api .
./splitwise-api
```

The API will be available at `http://localhost:8080`

### Docker Setup
```bash
docker-compose up -d
```

## Usage Examples

### 1. Register and Login
```bash
# Register
curl -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "password123"
  }'

# Login
curl -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "password123"
  }'
```

### 2. Create a Group
```bash
curl -X POST http://localhost:8080/v1/groups \
  -H "Content-Type: application/json" \
  -H "Bearer-Token: <your-jwt-token>" \
  -d '{
    "name": "Weekend Trip",
    "description": "Our weekend getaway expenses",
    "currency": "USD"
  }'
```

### 3. Add an Expense
```bash
curl -X POST http://localhost:8080/v1/expenses \
  -H "Content-Type: application/json" \
  -H "Bearer-Token: <your-jwt-token>" \
  -d '{
    "group_id": "<group-id>",
    "description": "Dinner at restaurant",
    "amount": 120.00,
    "currency": "USD",
    "split_type": "equal",
    "splits": [
      {"user_id": "<user1-id>", "amount": 0},
      {"user_id": "<user2-id>", "amount": 0},
      {"user_id": "<user3-id>", "amount": 0}
    ],
    "category": "Food"
  }'
```

### 4. Check Balances
```bash
curl -X GET http://localhost:8080/v1/groups/<group-id>/balances \
  -H "Bearer-Token: <your-jwt-token>"
```

### 5. Simplify Debts
```bash
curl -X GET http://localhost:8080/v1/groups/<group-id>/simplify \
  -H "Bearer-Token: <your-jwt-token>"
```

## Expense Split Types

### Equal Split
Everyone pays the same amount:
```json
{
  "split_type": "equal",
  "splits": [
    {"user_id": "user1", "amount": 0},
    {"user_id": "user2", "amount": 0}
  ]
}
```

### Exact Amount Split
Specify exact amounts:
```json
{
  "split_type": "exact",
  "splits": [
    {"user_id": "user1", "amount": 30.00},
    {"user_id": "user2", "amount": 20.00}
  ]
}
```

### Percentage Split
Split by percentage (must total 100%):
```json
{
  "split_type": "percentage",
  "splits": [
    {"user_id": "user1", "amount": 60},
    {"user_id": "user2", "amount": 40}
  ]
}
```

## Balance Calculation

The API automatically calculates balances for each user:
- **Positive balance**: Others owe you money
- **Negative balance**: You owe others money
- **Zero balance**: You're settled up

## Debt Simplification

The smart debt simplification algorithm minimizes the number of transactions needed to settle all debts within a group. For example, if:
- Alice owes Bob $10
- Bob owes Charlie $10

The algorithm suggests: Alice pays Charlie $10 (eliminating the need for Bob to be involved).

## API Documentation

Full interactive API documentation is available at:
`http://localhost:8080/swagger/index.html`

## Environment Variables

```env
# Server
SERVER_ADDR=localhost
SERVER_PORT=8080

# MongoDB
MONGO_URI=mongodb://localhost:27017
MONGO_DATABASE=splitwise

# Redis (optional)
USE_REDIS=false
REDIS_DEFAULT_ADDR=localhost:6379

# JWT
JWT_SECRET=your-secret-key
JWT_ACCESS_EXPIRATION_MINUTES=1440
JWT_REFRESH_EXPIRATION_DAYS=7

# Mode
MODE=debug
```

## Database Models

### Users
- ID, Name, Email, Password (hashed)
- Email verification status
- User role

### Groups
- Name, Description, Currency
- Creator and Members list
- Active status for soft deletion

### Expenses
- Amount, Description, Category
- Payer and Split details
- Date and Settlement status

### Friendships
- Requester and Addressee
- Status (pending, accepted, rejected, blocked)
- Timestamps

### Settlements
- Payer, Payee, Amount
- Status and Settlement date
- Related expenses

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

MIT License - see LICENSE file for details.

## Support

For issues and questions, please open a GitHub issue or check the API documentation at `/swagger/index.html`.
