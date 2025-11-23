# SLI Test: Users Get Review

**Endpoint**: `GET /users/getReview?user_id=<user_id>`
**Test Configuration**:
- Rate: 5 RPS
- Duration: 1 minute
- SLI thresholds: p95 < 300ms, 99.9% success rate
