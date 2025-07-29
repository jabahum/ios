# Redis Setup for Session Storage

## Overview
Your application has been configured to use Redis for session storage instead of in-memory storage. This provides persistent, scalable session management.

## Installation Options

### Option 1: Install Redis Locally (Recommended for Development)

#### Windows:
1. **Using Chocolatey:**
   ```bash
   choco install redis-64
   ```

2. **Using WSL (Windows Subsystem for Linux):**
   ```bash
   sudo apt update
   sudo apt install redis-server
   sudo systemctl start redis-server
   sudo systemctl enable redis-server
   ```

3. **Using Docker:**
   ```bash
   docker run -d --name redis -p 6379:6379 redis:alpine
   ```

#### macOS:
```bash
brew install redis
brew services start redis
```

#### Linux (Ubuntu/Debian):
```bash
sudo apt update
sudo apt install redis-server
sudo systemctl start redis-server
sudo systemctl enable redis-server
```

### Option 2: Use Docker (Recommended for Production)

Create a `docker-compose.yml` file:
```yaml
version: '3.8'
services:
  redis:
    image: redis:alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    restart: unless-stopped

volumes:
  redis_data:
```

Run with:
```bash
docker-compose up -d
```

## Configuration

The application is configured with these default Redis settings:
- **Host:** localhost
- **Port:** 6379
- **Database:** 0
- **Password:** (none)
- **Username:** (none)

## Testing Redis Connection

1. **Start your application:**
   ```bash
   go run cmd/web/main.go
   ```

2. **Check the logs** - you should see:
   ```
   INFO: Successfully connected to Redis for session storage
   ```

3. **If Redis is not running, you'll see:**
   ```
   WARNING: Failed to connect to Redis: dial tcp localhost:6379: connectex: No connection could be made
   WARNING: Sessions will not persist across restarts. Please ensure Redis is running on localhost:6379
   ```

## Customizing Redis Configuration

To change Redis settings, modify the `init()` function in `cmd/web/main.go`:

```go
redisStorage := redis.New(redis.Config{
    Host:     "your-redis-host",     // Change host
    Port:     6379,                  // Change port
    Username: "your-username",       // Add username
    Password: "your-password",       // Add password
    Database: 1,                     // Change database number
    Reset:    false,
    TLSConfig: nil,
})
```

## Benefits of Redis Session Storage

1. **Persistence:** Sessions survive application restarts
2. **Scalability:** Multiple application instances can share sessions
3. **Performance:** Fast in-memory storage with disk persistence
4. **Reliability:** Built-in replication and clustering options
5. **Monitoring:** Built-in tools for monitoring and debugging

## Troubleshooting

### Redis Connection Issues:
1. Ensure Redis is running: `redis-cli ping` (should return "PONG")
2. Check if port 6379 is available: `netstat -an | findstr 6379`
3. Verify firewall settings
4. Check Redis logs for errors

### Session Issues:
1. Clear browser cookies
2. Check Redis for session data: `redis-cli keys "*fiber_sess*"`
3. Restart both Redis and your application

## Production Considerations

1. **Security:** Set a strong Redis password
2. **Network:** Use Redis over TLS in production
3. **Monitoring:** Set up Redis monitoring
4. **Backup:** Configure Redis persistence and backups
5. **Clustering:** Consider Redis Cluster for high availability 