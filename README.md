# Integrated Outbreak System (IOS)

A comprehensive web application for managing health-related data, including patient encounters, VHF cases, MPOX cases, outbreak assignments, and inventory management.

## Features

- **Patient Management**: Complete patient registration and encounter tracking
- **VHF Case Management**: Comprehensive VHF case investigation forms
- **MPOX Case Management**: MPOX case investigation and follow-up
- **Outbreak Management**: Outbreak assignment and tracking
- **Inventory Management**: Complete inventory system for treatment sites
- **RBAC System**: Role-based access control
- **Session Management**: Redis-based session storage for persistence

## Prerequisites

- Go 1.21 or higher
- PostgreSQL 12 or higher
- Redis 6.0 or higher (optional, for session persistence)

## Installation

### 1. Install Go Dependencies

```bash
go mod download
```

### 2. Install PostgreSQL

#### Ubuntu/Debian:
```bash
sudo apt update
sudo apt install postgresql postgresql-contrib
```

#### Windows:
Download and install from [PostgreSQL official website](https://www.postgresql.org/download/windows/)

### 3. Install Redis (Optional, for session persistence)

#### Ubuntu/Debian:
```bash
sudo apt update
sudo apt install redis-server
sudo systemctl start redis-server
sudo systemctl enable redis-server
```

#### Windows:
Download and install from [Redis official website](https://redis.io/download) or use WSL2 with Ubuntu.

#### macOS:
```bash
brew install redis
brew services start redis
```

### 4. Database Setup

1. Create the database:
```sql
CREATE DATABASE ios;
```

2. Create a user (optional):
```sql
CREATE USER pwaiswa WITH PASSWORD 'pwaiswa';
GRANT ALL PRIVILEGES ON DATABASE ios TO pwaiswa;
```

3. Run the database migrations:
```bash
# The application will automatically run migrations on startup
```

### 5. Configuration

1. Copy the configuration file:
```bash
cp cmd/web/config.json.example cmd/web/config.json
```

2. Update the configuration in `cmd/web/config.json`:
```json
{
    "Port": "0.0.0.0:3000",
    "Ux": "postgres",
    "Px": "your_password",
    "Dx": "ios",
    "RedisEnabled": true,
    "RedisHost": "localhost",
    "RedisPort": 6379,
    "RedisPassword": "",
    "RedisDatabase": 0
}
```

## Running the Application

### Development Mode

```bash
cd cmd/web
go run main.go
```

### Production Mode

```bash
go build -o ios cmd/web/main.go
./ios
```

## Session Management

The application supports both in-memory and Redis-based session storage:

- **In-memory**: Sessions are lost when the application restarts
- **Redis**: Sessions persist across application restarts

### Redis Configuration

To enable Redis session storage:

1. Ensure Redis is running
2. Set `"RedisEnabled": true` in `config.json`
3. Configure Redis connection details:
   - `RedisHost`: Redis server hostname (default: localhost)
   - `RedisPort`: Redis server port (default: 6379)
   - `RedisPassword`: Redis password (if required)
   - `RedisDatabase`: Redis database number (default: 0)

### Session Health Check

The application provides endpoints for monitoring session health:

- `GET /session/health`: Check session storage status
- `GET /session/clear`: Clear current session (for debugging)

## API Endpoints

### Authentication
- `POST /login`: User login
- `GET /logout`: User logout

### VHF Management
- `GET /vhf-cif`: VHF case investigation form
- `POST /vhf-cif/patient/:id`: Submit patient information
- `GET /vhf-lab/:id`: VHF laboratory form
- `POST /vhf-lab/save/:id`: Save laboratory data

### Inventory Management
- `GET /inventory`: Inventory dashboard
- `GET /inventory/items`: Inventory items list
- `GET /inventory/items/new`: Add new inventory item
- `POST /inventory/items/save`: Save inventory item

## Troubleshooting

### Redis Connection Issues

If Redis is not available, the application will automatically fall back to in-memory session storage. Check the logs for connection status.

### Database Connection Issues

Ensure PostgreSQL is running and the connection details in `config.json` are correct.

### Session Issues

1. Check Redis connection: `GET /session/health`
2. Clear session if needed: `GET /session/clear`
3. Check application logs for session-related errors

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License.
