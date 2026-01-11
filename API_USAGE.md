# Exercises API Documentation

High-performance REST API for filtering exercises with intelligent caching.

## Available Implementations

- **Go** (Recommended for production) - Superior performance and concurrency
- **Node.js** - Easier deployment and development

## Setup & Running

### Node.js Setup
```bash
npm install
npm start        # Production
npm run dev      # Development with auto-reload
```

Server runs on `http://localhost:3000`

### Go Setup
```bash
go mod download
go run main.go
```

Server runs on `http://localhost:8080`

## API Endpoints

### GET /health
Health check endpoint.

**Response:**
```json
{
  "status": "ok",
  "timestamp": "2026-01-11T12:00:00Z"
}
```

---

### GET /exercises
Filter exercises based on query parameters.

**Query Parameters:**
- `muscle_group` (required): One or more muscle groups, comma-separated
  - Valid values: `chest`, `back`, `shoulder`, `arm`, `abdominals`, `human leg`, `triceps`, `biceps`, `hamstring`, `lower back`
- `equipment` (optional, default: `none`): Type of equipment
  - Valid values: `none`, `ez curl bar`, `barbell`, `dumbbell`, `gym mat`, `exercise ball`, `medicine ball`, `pull-up bar`, `bench`, `incline bench`, `kettlebell`, `machine`, `cable`, `bands`, `foam roll`, `other`
- `category` (optional): Exercise category
  - Valid values: `strength`, `stretching`, `plyometrics`, `strongman`, `cardio`, `olympic weightlifting`, `crossfit`, `calisthenics`

**Examples:**

Single muscle group:
```
GET /exercises?muscle_group=chest&equipment=barbell
```

Multiple muscle groups (comma-separated):
```
GET /exercises?muscle_group=chest,back&equipment=dumbbell&category=strength
```

With category:
```
GET /exercises?muscle_group=arm&equipment=none&category=stretching
```

**Response:**
```json
{
  "data": [
    {
      "category": "strength",
      "description": "...",
      "equipment": ["barbell"],
      "instructions": [...],
      "name": "Barbell Bench Press",
      "primary_muscles": ["chest"],
      "secondary_muscles": ["triceps"],
      "video": "https://...",
      "muscle_group": "chest"
    }
  ],
  "count": 1,
  "cached": false,
  "timestamp": "2026-01-11T12:00:00Z"
}
```

---

### GET /stats
Get available filters and statistics.

**Response:**
```json
{
  "categories": ["strength", "stretching", ...],
  "equipment": ["none", "barbell", ...],
  "muscle_groups": ["chest", "back", ...],
  "total_exercises": 1234
}
```

## Caching Strategy

Both implementations use intelligent caching with:

- **TTL**: 5 minutes per cache entry
- **Key Generation**: Based on combination of `muscle_group`, `equipment`, and `category`
- **Automatic Cleanup**: Expired entries are removed periodically
- **Cache Hit Indicator**: Response includes `"cached": true/false` flag

## Performance Characteristics

### Go Implementation
- **Concurrency**: Handles thousands of concurrent requests
- **Memory**: ~50-100MB depending on data size
- **Latency**: <5ms average response time
- **Cache Hit**: <1ms response time
- **Best for**: High-traffic production environments

### Node.js Implementation
- **Concurrency**: Handles 100s of concurrent requests comfortably
- **Memory**: ~100-150MB
- **Latency**: ~5-10ms average response time
- **Cache Hit**: ~2-5ms response time
- **Best for**: Development, prototyping, moderate traffic

## Example curl Requests

```bash
# Single muscle group, no equipment
curl "http://localhost:3000/exercises?muscle_group=chest"

# Multiple muscle groups with equipment
curl "http://localhost:3000/exercises?muscle_group=chest,back&equipment=barbell"

# With category filter
curl "http://localhost:3000/exercises?muscle_group=arm&equipment=dumbbell&category=strength"

# Get available options
curl "http://localhost:3000/stats"

# Health check
curl "http://localhost:3000/health"
```

## Error Handling

**Missing required parameter:**
```json
{
  "error": "muscle_group parameter is required"
}
```

**Status Code**: 400

## Performance Tips

1. **Use caching effectively**: Queries with same parameters will return cached results
2. **Go over Node.js**: For high-traffic scenarios, use Go implementation
3. **Batch requests**: Group related queries to maximize cache hits
4. **Monitor cache**: Check `cached` flag in responses to understand cache effectiveness
