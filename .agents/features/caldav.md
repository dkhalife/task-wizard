# Feature: CalDAV Integration

An authenticated CalDAV endpoint that allows third-party calendar and task apps to sync with Task Wizard.

## Capabilities

- CalDAV endpoint available at `/dav/tasks`
- Supports PROPFIND, REPORT, GET, PUT, and HEAD methods
- Tasks exposed in iCalendar VTODO format
- Authentication via Basic Auth using the user's email and an API token (with dav scope) as the password
- Creating and updating tasks through CalDAV clients (parses VTODO summary and due date)
- Task changes made via CalDAV are broadcast over WebSocket to connected frontends
- XML-based WebDAV response generation with proper namespace handling
