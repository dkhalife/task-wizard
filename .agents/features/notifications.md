# Feature: Notifications

Users can configure notifications to be reminded about task deadlines.

## Capabilities

- Notification triggers: due date, pre-due (configurable hours before), overdue
- Notification providers: Webhook (GET/POST) and Gotify push notifications
- Per-task notification configuration when creating or editing a task
- Background scheduler generates and sends notifications at configurable intervals
- Overdue notifications sent on a separate schedule (default 24h)
- Sent notifications are automatically cleaned up
- Desktop browser notifications toggle in settings
