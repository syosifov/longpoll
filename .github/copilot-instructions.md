# Long Polling Go-Gin Server

This is an experimental project implementing long polling in a Go-Gin environment. The server provides real-time event delivery through HTTP long polling with a 30-second timeout mechanism.

**Always reference these instructions first and fallback to search or bash commands only when you encounter unexpected information that does not match the info here.**

## Working Effectively

### Prerequisites
- Go 1.23+ is already installed and available
- No additional dependencies or external services required

### Bootstrap and Build
- Download dependencies: `go mod download` -- takes ~5 seconds
- Build the application: `go build -o longpoll .` -- takes ~30 seconds
- Run directly without building: `go run .` -- starts immediately

### Run the Application
- Start the server: `./longpoll` (if built) or `go run .`
- Server starts on port 8080 and shows debug output
- Look for "Server started on port 8080..." message to confirm startup

### Validation
- Always manually validate the long polling functionality after making changes to the core logic
- Test both endpoints with the validation scenarios below
- ALWAYS run `go fmt` and `go vet` before committing changes

## API Endpoints and Testing

### Long Polling Endpoint: GET /poll
Test the long polling behavior:
```bash
curl -X GET http://localhost:8080/poll
```
- **Expected behavior**: Waits up to 30 seconds for events
- **With events**: Returns immediately with JSON event data `{"message":"...", "time":"..."}`
- **Without events**: Returns after 30 seconds with `{"message":"No new events."}` and HTTP 204

### Event Publisher: POST /publish
Publish events to waiting clients:
```bash
curl -X POST http://localhost:8080/publish \
  -H "Content-Type: application/json" \
  -d '{"message": "Your test message here"}'
```
- **Expected response**: `{"message": "Event published."}` with HTTP 200
- **Effect**: Any waiting `/poll` requests immediately receive the event

### End-to-End Validation Scenario
Always test this complete workflow after making changes:

1. Start the server: `go run .`
2. In another terminal, start a long poll: `curl -X GET http://localhost:8080/poll &`
3. Wait 2-3 seconds, then publish an event: `curl -X POST http://localhost:8080/publish -H "Content-Type: application/json" -d '{"message": "E2E test"}'`
4. Verify the poll request immediately returns with the published message
5. Test timeout: start another poll request and wait 35 seconds - should return "No new events" after 30 seconds

## Development Workflow

### Code Quality
- Format code: `go fmt ./...`
- Check for issues: `go vet ./...`
- No linting tools configured - use Go standard practices

### Testing
- No unit tests exist (`go test` shows "no test files")
- All validation is manual using the API testing scenarios above
- Use the `test.rest` file for VS Code REST Client testing

### Debugging
- VS Code launch configuration exists in `.vscode/launch.json`
- Server runs in debug mode by default (shows all Gin debug output)
- Debug mode is hardcoded in `main.go` line 20 with `gin.SetMode(gin.DebugMode)`

## Repository Structure

```
.
├── .github/
│   └── copilot-instructions.md    # This file
├── .vscode/
│   └── launch.json               # VS Code debug configuration
├── .gitignore                    # Go standard gitignore
├── README.md                     # Basic project description
├── go.mod                        # Go module definition
├── go.sum                        # Dependency checksums
├── main.go                       # Complete application (85 lines)
└── test.rest                     # VS Code REST Client examples
```

### Key Code Locations
- **Event struct**: `main.go` lines 12-15 - defines message format
- **Long poll handler**: `main.go` lines 51-66 - implements 30s timeout logic
- **Publisher handler**: `main.go` lines 69-85 - accepts and broadcasts events
- **Event channel**: `main.go` line 17 - global channel for event communication

## Common Tasks

### Building and Running
```bash
# Quick development iteration
go run .

# Build for distribution
go build -o longpoll .
./longpoll

# Download dependencies (if needed)
go mod download
```

### Manual Testing Commands
```bash
# Test timeout behavior (should return after 30 seconds)
timeout 35s curl -X GET http://localhost:8080/poll

# Test immediate response with event
curl -X POST http://localhost:8080/publish -H "Content-Type: application/json" -d '{"message": "test"}' &
curl -X GET http://localhost:8080/poll

# Test invalid JSON (should return 400 error)
curl -X POST http://localhost:8080/publish -H "Content-Type: application/json" -d 'invalid json'
```

## Important Notes

### Application Behavior
- Only ONE event can be consumed per poll request (first-come-first-served)
- Events are NOT persisted - they exist only in memory channel
- Server must be restarted to clear any unconsumed events
- Concurrent polling works but each poll gets its own event copy

### Troubleshooting
- **Port 8080 already in use**: Kill existing processes with `pkill -f longpoll` or `lsof -ti:8080 | xargs kill`
- **Build errors**: Run `go mod tidy` to clean up dependencies
- **Missing events**: Events are consumed immediately - start poll before publishing
- **Timeout not working**: Check for network/proxy issues affecting the 30-second timeout

### When Making Changes
- Always test both the happy path (event received) and timeout path (no events)
- Pay attention to the channel communication in `main.go` - it's unbuffered
- Consider race conditions when modifying the event handling logic
- Test with multiple concurrent clients to ensure proper behavior

## Files NOT to Modify
- `.gitignore` - already configured for Go projects
- `go.sum` - managed automatically by Go
- `.vscode/launch.json` - working VS Code debug configuration