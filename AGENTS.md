# AI Agent Instructions for RSSX Project

## 🌍 PRIMARY RULE: ENGLISH ONLY

**ALL content generated MUST be in English:**
- Code comments and documentation
- Function, variable, class, and file names
- Commit messages and PR descriptions
- Error messages and log output
- Test descriptions and assertions
- Configuration comments
- API documentation

No exceptions. This applies to all AI-generated and human-written content.

## Project Overview

RSSX is a modern RSS reader application consisting of:

- **rssx-api/** - Go backend service (REST API, RSS feed processing, user management)
- **rssx-ui/** - Vue.js 3 frontend (TypeScript, Vuetify, responsive UI)

### Technology Stack

**Backend:**
- Go 1.20+
- SQLite (primary storage)
- Redis (caching)
- JWT authentication
- Standard library routing

**Frontend:**
- Vue.js 3 (Composition API)
- TypeScript 4.5+
- Vuetify 3 (Material Design)
- Webpack 5
- Jest + Cypress (testing)

**Deployment:**
- Docker/Podman containers
- Kubernetes support
- GitHub Actions (CI/CD)

## Repository Structure

```
rssx/
├── rssx-api/              # Backend service
│   ├── feed/              # RSS feed management
│   ├── news/              # News article handling
│   ├── user/              # Authentication & user management
│   ├── rss/               # RSS parsing and sync
│   ├── storage/           # Data persistence (SQLite, Redis)
│   ├── utils/             # Shared utilities
│   └── common/            # Common database functions
├── rssx-ui/               # Frontend application
│   ├── src/
│   │   ├── components/    # Vue components
│   │   ├── views/         # Page views
│   │   ├── router/        # Vue Router config
│   │   ├── store/         # Vuex state management
│   │   └── utils/         # Frontend utilities
│   └── tests/             # Frontend tests
└── deploy/                # Deployment configs
    ├── docker/            # Docker build scripts
    └── k8s/               # Kubernetes manifests
```

## Code Style Guidelines

### Go Backend Standards

#### Naming and Documentation
```go
// ✅ CORRECT: Clear English names with proper documentation
// FetchUserFeeds retrieves all RSS feeds for the specified user.
// It returns an empty slice if the user has no feeds.
func FetchUserFeeds(ctx context.Context, userID string) ([]Feed, error) {
    // Validate user ID format
    if !isValidUserID(userID) {
        return nil, ErrInvalidUserID
    }
    
    // Query database with context timeout
    feeds, err := db.QueryFeeds(ctx, userID)
    if err != nil {
        return nil, fmt.Errorf("fetch user feeds: %w", err)
    }
    
    return feeds, nil
}

// ❌ WRONG: Non-English or missing documentation
func 获取订阅(id string) []Feed { }  // Wrong: Non-English names
func getFeed(id string) { }          // Wrong: No documentation, unclear return
```

#### Error Handling
```go
// ✅ CORRECT: Explicit error handling with context
if err != nil {
    return fmt.Errorf("failed to parse RSS feed %s: %w", feedURL, err)
}

// Log errors with structured fields
log.Error("database query failed",
    "operation", "fetch_feeds",
    "user_id", userID,
    "error", err)

// ❌ WRONG: Ignoring errors or non-English messages
db.Query() // Wrong: Ignoring error
return errors.New("数据库错误") // Wrong: Non-English
```

#### Testing
```go
// ✅ CORRECT: Table-driven tests with English descriptions
func TestFetchFeeds(t *testing.T) {
    tests := []struct {
        name    string
        userID  string
        want    []Feed
        wantErr bool
    }{
        {
            name:    "valid user with feeds",
            userID:  "user123",
            want:    []Feed{{ID: "1", Title: "Tech News"}},
            wantErr: false,
        },
        {
            name:    "user with no feeds returns empty slice",
            userID:  "user456",
            want:    []Feed{},
            wantErr: false,
        },
        {
            name:    "invalid user ID returns error",
            userID:  "",
            want:    nil,
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### TypeScript/Vue.js Frontend Standards

#### Component Documentation
```typescript
/**
 * FeedList displays a paginated list of RSS feeds with filtering.
 * 
 * @component
 * @example
 * <FeedList 
 *   :feeds="userFeeds" 
 *   @select="handleFeedSelect"
 *   @refresh="refreshFeeds" 
 * />
 */
export default defineComponent({
    name: 'FeedList',
    props: {
        feeds: {
            type: Array as PropType<Feed[]>,
            required: true,
            default: () => []
        }
    },
    emits: ['select', 'refresh'],
    setup(props, { emit }) {
        // Implementation
    }
});
```

#### Type Safety
```typescript
// ✅ CORRECT: Proper TypeScript types
interface Feed {
    id: string;
    title: string;
    url: string;
    category?: string;
    updatedAt: number;
}

async function fetchFeeds(): Promise<Feed[]> {
    const response = await api.get<Feed[]>('/api/feeds');
    return response.data;
}

// ❌ WRONG: Using any or missing types
function fetchFeeds(): any { }  // Wrong: Using any
let feeds = [];  // Wrong: No type annotation
```

#### Comments and Error Handling
```typescript
// ✅ CORRECT: Clear English comments
/**
 * Syncs all feeds and updates the UI state.
 * Shows error notification if sync fails.
 */
async function syncAllFeeds(): Promise<void> {
    try {
        // Set loading state before sync
        isLoading.value = true;
        
        // Fetch latest feed data from API
        const feeds = await feedService.syncAll();
        
        // Update store with new data
        store.commit('setFeeds', feeds);
        
        // Show success notification
        showNotification('Feeds synced successfully', 'success');
    } catch (error) {
        // Log error and show user-friendly message
        console.error('Feed sync failed:', error);
        showNotification('Failed to sync feeds. Please try again.', 'error');
    } finally {
        isLoading.value = false;
    }
}

// ❌ WRONG: Non-English or missing error handling
async function sync() {
    const data = await api.get('/feeds'); // Wrong: No error handling
    // 更新数据 - Wrong: Non-English comment
}
```

## Build and Test Commands

### Backend (rssx-api/)
```bash
# Build the application
go build -o rssx-api

# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Format code
go fmt ./...

# Lint code (requires golangci-lint)
golangci-lint run

# Run the server locally
./rssx-api -config config-local.toml
```

### Frontend (rssx-ui/)
```bash
# Install dependencies
pnpm install

# Run development server
pnpm serve

# Build for production
pnpm build

# Run unit tests
pnpm test:unit

# Run e2e tests
pnpm test:e2e

# Lint and fix
pnpm lint
```

## Git Commit Conventions

Use [Conventional Commits](https://www.conventionalcommits.org/) format:

```bash
# Feature additions
feat: add RSS feed category filtering
feat(api): implement JWT token refresh endpoint

# Bug fixes
fix: resolve feed sync timeout issue
fix(ui): correct feed list sorting order

# Documentation
docs: update API endpoint documentation
docs(readme): add development setup instructions

# Code refactoring
refactor: simplify feed parsing logic
refactor(storage): optimize database queries

# Tests
test: add unit tests for user authentication
test(e2e): add feed sync integration tests

# Chores and maintenance
chore: update dependencies
chore(ci): configure GitHub Actions workflow
```

## Security Best Practices

1. **Never commit sensitive data**
   - No passwords, API keys, or tokens in code
   - Use environment variables for secrets
   - Add sensitive files to `.gitignore`

2. **Input validation**
   ```go
   // Validate and sanitize all user inputs
   func ValidateURL(rawURL string) error {
       u, err := url.Parse(rawURL)
       if err != nil {
           return ErrInvalidURL
       }
       if u.Scheme != "http" && u.Scheme != "https" {
           return ErrInvalidScheme
       }
       return nil
   }
   ```

3. **SQL injection prevention**
   ```go
   // Use parameterized queries
   db.Query("SELECT * FROM feeds WHERE user_id = ?", userID)
   ```

4. **Authentication**
   - Implement proper JWT token validation
   - Use HTTPS in production
   - Implement rate limiting
   - Hash passwords with bcrypt

## Testing Guidelines

### Backend Testing
- Write unit tests for all business logic
- Use table-driven tests for multiple scenarios
- Mock external dependencies (database, HTTP clients)
- Test error cases and edge conditions
- Place tests in `*_test.go` files next to implementation

### Frontend Testing
- Unit test utilities and pure functions with Jest
- Integration test components with Vue Test Utils
- E2E test critical user flows with Cypress
- Mock API calls in tests
- Test error states and loading states

## Performance Considerations

- Implement database connection pooling
- Use Redis for caching frequently accessed data
- Implement pagination for large datasets
- Optimize database queries with proper indexes
- Use lazy loading for frontend routes and components
- Implement request debouncing for search inputs

## Deployment

### Docker/Podman
```bash
# Build backend image
cd rssx-api
podman build -t rssx-api:latest -f Containerfile .

# Build frontend image
cd rssx-ui
podman build -t rssx-ui:latest -f Dockerfile .

# Run with environment variables
podman run -d \
  -p 8080:8080 \
  -e RSSX_SECURITY_KEY="your-secret-key" \
  rssx-api:latest
```

### Kubernetes
- Use provided manifests in `deploy/k8s/`
- Configure ConfigMaps for environment-specific settings
- Use Secrets for sensitive data
- Implement health check endpoints
- Configure resource limits

## API Design Principles

- Use RESTful conventions
- Return appropriate HTTP status codes
  - 200: Success
  - 201: Created
  - 400: Bad Request
  - 401: Unauthorized
  - 404: Not Found
  - 500: Internal Server Error
- Use JSON for request/response bodies
- Implement API versioning (e.g., `/api/v1/`)
- Provide clear error messages in responses
- Document all endpoints

## Configuration Management

- Use separate config files per environment:
  - `config-local.toml` - Local development
  - `config-k8s.toml` - Kubernetes deployment
  - `config.toml` - Production
- Document all configuration options in `CONFIG.md`
- Use environment variables for secrets
- Provide sensible defaults

## Logging Standards

```go
// Use structured logging
log.Info("RSS feed synced",
    "feed_id", feedID,
    "articles_added", count,
    "duration_ms", duration.Milliseconds())

log.Error("database connection failed",
    "host", dbHost,
    "error", err,
    "retry_attempt", retryCount)

// All log messages must be in English
// ❌ WRONG: log.Info("同步完成") 
```

## Dependencies Management

### Backend (Go)
- Use `go.mod` for dependency management
- Run `go mod tidy` to clean dependencies
- Pin versions for production
- Review security advisories regularly

### Frontend (Node.js)
- Use `pnpm` for package management
- Keep `pnpm-lock.yaml` in version control
- Update dependencies regularly
- Audit for security vulnerabilities: `pnpm audit`

## AI Assistant Guidelines

When generating code or suggestions:

1. ✅ **Always use English** for all text content
2. ✅ Follow existing code patterns and architecture
3. ✅ Include proper error handling
4. ✅ Add documentation comments for public APIs
5. ✅ Write tests for new functionality
6. ✅ Consider security implications
7. ✅ Use appropriate logging
8. ✅ Follow the technology stack already in use
9. ✅ Optimize for readability over cleverness
10. ✅ Suggest improvements when patterns can be enhanced

## Common Patterns

### Backend: Database Query Pattern
```go
func (r *Repository) GetFeed(ctx context.Context, id string) (*Feed, error) {
    query := "SELECT id, title, url, created_at FROM feeds WHERE id = ?"
    
    var feed Feed
    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &feed.ID,
        &feed.Title,
        &feed.URL,
        &feed.CreatedAt,
    )
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, ErrFeedNotFound
        }
        return nil, fmt.Errorf("query feed: %w", err)
    }
    
    return &feed, nil
}
```

### Frontend: API Call Pattern
```typescript
// services/feedService.ts
class FeedService {
    async getFeeds(): Promise<Feed[]> {
        try {
            const response = await api.get<Feed[]>('/api/v1/feeds');
            return response.data;
        } catch (error) {
            if (error.response?.status === 401) {
                throw new AuthenticationError('Please login again');
            }
            throw new ApiError('Failed to fetch feeds', error);
        }
    }
}
```

## Documentation Requirements

- Update README.md when adding major features
- Document API changes in API documentation
- Add inline comments for complex logic
- Keep configuration documentation current
- Document deployment procedures
- Include troubleshooting guides

## Review Checklist

Before submitting code:
- [ ] All code comments are in English
- [ ] Tests are written and passing
- [ ] Error handling is implemented
- [ ] Code follows project conventions
- [ ] Documentation is updated
- [ ] No sensitive data in code
- [ ] Commit messages follow conventions
- [ ] Code is properly formatted
- [ ] No linter warnings

---

**Remember:** This project prioritizes clarity, maintainability, and international collaboration. Using English consistently throughout the codebase is essential for achieving these goals.
