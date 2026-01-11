# Exercises API Documentation

High-performance REST API for filtering exercises with intelligent caching.

## Setup & Running

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

---

### GET /muscle_group
Get the list of available muscle groups.

**Response:**
```json
{
  "muscle_groups": [
    "chest",
    "back",
    "shoulder",
    "arm",
    "abdominals",
    "human leg",
    "triceps",
    "biceps",
    "hamstring",
    "lower back"
  ]
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

## Example curl Requests

```bash
# Single muscle group, no equipment
curl "http://localhost:8080/exercises?muscle_group=chest"

# Multiple muscle groups with equipment
curl "http://localhost:8080/exercises?muscle_group=chest,back&equipment=barbell"

# With category filter
curl "http://localhost:8080/exercises?muscle_group=arm&equipment=dumbbell&category=strength"

# Get available muscle groups
curl "http://localhost:8080/muscle_group"

# Get available options
curl "http://localhost:8080/stats"

# Health check
curl "http://localhost:8080/health"
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
