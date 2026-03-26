# Feature: Calendar Sync (Android)

Syncs active tasks with due dates into a read-only local calendar on the device, visible in Google Calendar or any calendar app.

## Capabilities

- Creates a local "Task Wizard" calendar on the device via Calendar Provider
- Syncs active tasks as 15-minute events at their due date/time with no reminders
- Background sync every 15 minutes via WorkManager (battery-friendly, survives app kill)
- Togglable in Settings — OFF removes calendar and stops background sync
- Runtime permission request for `READ_CALENDAR` + `WRITE_CALENDAR`

## Architecture

Four modular layers, each reusable independently:

### 1. `CalendarProviderClient` (low-level, reusable)
- Thin wrapper around `ContentResolver` + `CalendarContract`
- CRUD operations: calendars (create/delete/find) and events (insert/update/delete/query)
- No domain logic — can be used by any feature that needs calendar access

### 2. `CalendarSyncEngine` (domain-layer, reusable)
- Diff-based sync: compares task list against existing calendar events
- Uses task ID in `SYNC_DATA1` to match events to tasks
- Inserts new events, updates changed events, removes stale events
- Stateless — accepts task list and calendar ID, does the rest

### 3. `CalendarSyncWorker` (WorkManager)
- `@HiltWorker` `CoroutineWorker` that fetches tasks via `TaskWizardApi`
- Passes fetched tasks to `CalendarSyncEngine`
- Returns `Result.retry()` on failure for automatic backoff

### 4. `CalendarRepository` (orchestration)
- `enableCalendarSync()` — creates calendar, schedules periodic WorkManager job
- `disableCalendarSync()` — cancels worker, deletes calendar
- Persists enabled/disabled state in `SharedPreferences`

## Key Surfaces

- `data/calendar/CalendarProviderClient.kt` — ContentProvider wrapper
- `data/calendar/CalendarSyncEngine.kt` — Diff-based sync logic
- `data/calendar/CalendarSyncWorker.kt` — WorkManager worker
- `data/calendar/CalendarRepository.kt` — Feature orchestration + preferences
- `data/AppPreferences.kt` — `KEY_CALENDAR_SYNC` constant
- `ui/screen/SettingsScreen.kt` — Calendar sync toggle with permission handling

## Permissions

- `android.permission.READ_CALENDAR` — required to query existing events for diff sync
- `android.permission.WRITE_CALENDAR` — required to create calendar and manage events
- Requested at runtime when user enables the feature

## Event Mapping

| Task Field | Calendar Event Field |
|------------|---------------------|
| `id` | `SYNC_DATA1` (for matching) |
| `title` | `TITLE` |
| `next_due_date` | `DTSTART` |
| `next_due_date + 15min` | `DTEND` |
| — | `HAS_ALARM = 0` (no reminders) |
| — | `EVENT_TIMEZONE = UTC` |
