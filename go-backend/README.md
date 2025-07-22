# eSports Fantasy Backend

A high-performance, scalable GoLang backend for eSports fantasy gaming platform with real-time features and OTP authentication.

## 🚀 Features

### Core Features
- **OTP-based Authentication** - Secure phone number verification with console logging for development
- **Tournament & Match Management** - Complete admin panel for managing tournaments, matches, and teams  
- **Fantasy Team Creation** - Users can create teams with captain/vice-captain selection
- **Contest Management** - Entry fee-based contests with prize distribution
- **Real-time Leaderboards** - WebSocket-powered live leaderboard updates
- **Payment Integration** - Razorpay payment gateway integration (dummy for testing)
- **High Concurrency** - Built with Go's goroutines for optimal performance

### Technical Features
- **Clean Architecture** - Layered architecture with clear separation of concerns
- **OpenAPI Documentation** - Auto-generated Swagger docs at `/docs` and ReDoc at `/redoc`  
- **PostgreSQL Database** - Robust relational database with UUID primary keys
- **Redis Caching** - High-speed caching and real-time leaderboard storage
- **WebSocket Support** - Real-time updates for leaderboards and match status
- **JWT Authentication** - Secure token-based authentication
- **Database Migrations** - Automated database schema management

## 📋 Prerequisites

- Go 1.19 or higher
- PostgreSQL 12+
- Redis 6+
- Git

## 🛠 Installation

1. **Clone the repository**
   ```bash
   git clone <your-repo-url>
   cd go-backend
   ```

2. **Install dependencies**
   ```bash
   make deps
   ```

3. **Set up environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your database credentials
   ```

4. **Set up the database**
   ```bash
   # Make sure PostgreSQL is running
   # The application will auto-migrate tables on startup
   ```

5. **Start Redis**
   ```bash
   redis-server
   ```

## 🚀 Running the Application

### Development Mode
```bash
make run
```

### Production Build
```bash
make build
./build/esports-fantasy
```

### Using Docker
```bash
make docker-build
make docker-run
```

## 📚 API Documentation

Once the application is running:

- **Swagger UI**: http://localhost:8080/docs
- **ReDoc**: http://localhost:8080/redoc  
- **Health Check**: http://localhost:8080/health

## 🔐 Authentication Flow

### 1. Send OTP
```bash
POST /api/v1/auth/send-otp
{
  "phone_number": "+919999999999"
}
```

The OTP will be displayed in the console:
```
==========================================
📱 OTP SENT TO: +919999999999
🔢 YOUR OTP: 123456
⏰ Expires in: 5 minutes
==========================================
```

### 2. Verify OTP
```bash
POST /api/v1/auth/verify-otp
{
  "phone_number": "+919999999999",
  "otp": "123456"
}
```

### 3. Use JWT Token
Add to Authorization header: `Bearer <your-jwt-token>`

## 🏆 Core Workflows

### Admin Workflow
1. **Create Tournament**
   ```bash
   POST /api/v1/admin/tournaments
   {
     "name": "BGMI Championship 2024",
     "game_type": "BGMI",
     "start_date": "2024-01-15T10:00:00Z",
     "end_date": "2024-01-20T18:00:00Z"
   }
   ```

2. **Create eSports Teams & Players**
   ```bash
   POST /api/v1/admin/esports-teams
   POST /api/v1/admin/players
   ```

3. **Create Match**
   ```bash
   POST /api/v1/admin/matches
   {
     "tournament_id": "uuid",
     "name": "Grand Finals - Map 1",
     "map_name": "Erangel",
     "start_time": "2024-01-20T15:00:00Z"
   }
   ```

4. **Create Contest**
   ```bash
   POST /api/v1/admin/contests
   {
     "match_id": "uuid",
     "name": "Winner Takes All",
     "entry_fee": 50.00,
     "max_entries": 1000,
     "prize_pool": "{\"1\": 25000, \"2\": 15000, \"3\": 10000}"
   }
   ```

5. **Update Player Stats**
   ```bash
   PUT /api/v1/admin/stats/match/{matchId}/player/{playerId}
   {
     "kills": 8,
     "knockouts": 12,
     "revives": 3,
     "survival_time_minutes": 25,
     "is_mvp": true
   }
   ```

### User Workflow
1. **Create Fantasy Team**
   ```bash
   POST /api/v1/fantasy/teams
   {
     "contest_id": "uuid",
     "team_name": "Pro Gamers",
     "player_ids": ["uuid1", "uuid2", "uuid3", "uuid4", "uuid5"],
     "captain_id": "uuid1",
     "vice_captain_id": "uuid2"
   }
   ```

2. **View Leaderboard**
   ```bash
   GET /api/v1/contests/{contestId}/leaderboard?limit=100
   ```

## 🎮 Scoring System

| Action | Points |
|--------|---------|
| Kill | +10 |
| Knockout | +6 |  
| Revive | +5 |
| Survival (per minute) | +1 |
| Not Knocked (full match) | +15 |
| MVP | +20 |
| Team Kill | -2 |

**Multipliers:**
- Captain: 2x points
- Vice Captain: 1.5x points

## 🌐 WebSocket Real-time Updates

Connect to `ws://localhost:8080/api/v1/ws`

### Subscribe to Contest Updates
```json
{
  "action": "subscribe",
  "channel": "contest:uuid-here"
}
```

### Leaderboard Update Message
```json
{
  "type": "leaderboard_update",
  "payload": {
    "contest_id": "uuid",
    "rankings": [
      {
        "team_id": "uuid",
        "team_name": "Pro Gamers",  
        "user_name": "PlayerX",
        "points": 550.5,
        "rank": 1
      }
    ]
  }
}
```

## 💳 Payment Integration

### Create Payment Order (Dummy)
```bash
POST /api/v1/payment/create-order
{
  "amount": 100.0,
  "contest_id": "uuid"
}
```

### Handle Payment Success
```bash
POST /api/v1/payment/success
{
  "razorpay_payment_id": "pay_test123",
  "razorpay_order_id": "order_test123"
}
```

## 🗄 Database Schema

Key tables:
- `users` - User profiles and wallet balances
- `tournaments` - Tournament information
- `esports_teams` - Professional eSports teams  
- `players` - Individual players with credit values
- `matches` - Match details and status
- `contests` - Contest configuration and prize pools
- `fantasy_teams` - User-created teams
- `fantasy_team_players` - Team compositions with captain info
- `player_match_stats` - Match statistics and points
- `transactions` - Payment and wallet transactions

## 🧪 Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Test specific package
go test -v ./internal/services/
```

## 📊 Monitoring & Health

- **Health Check**: `GET /health`
- **Metrics**: Built-in logging for all operations
- **Error Handling**: Comprehensive error responses

## 🔧 Development Commands

```bash
# Install dev tools
make tools

# Format code  
make fmt

# Run linter
make lint

# Generate swagger docs
make swagger

# Clean build artifacts
make clean
```

## 🌟 High Concurrency Features

- **Goroutine-based Processing** - Async scoring calculations
- **WebSocket Broadcasting** - Real-time updates to multiple clients
- **Redis Leaderboards** - High-speed sorted sets for rankings
- **Database Connection Pooling** - Efficient database connections
- **Non-blocking Operations** - Background processing for stats updates

## 🔒 Security

- JWT token authentication
- Phone number verification
- Input validation and sanitization
- CORS configuration
- Admin role-based access control

## 🚀 Deployment

The application is containerized and ready for deployment with:
- Docker support
- Environment-based configuration
- Health checks for monitoring
- Graceful shutdown handling

---

**Built with ❤️ using Go, Gin, GORM, PostgreSQL, Redis, and WebSockets for the ultimate eSports fantasy experience!**