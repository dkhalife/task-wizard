# Feature: Task History & Activity

Tracks completion and skip records for tasks. Powers both per-task analytics and a
global, reverse-chronological Activity feed with quick revert.

## Per-task history (analytics)

- Every completion is recorded in `task_histories` with a timestamp; skips are recorded
  with a `NULL` completed date.
- View completion history for a single task from the task context menu.
- Summary statistics: total completions, average delay, maximum delay.
- Performance metrics showing how early or late completions were relative to due dates.

## Activity view (global recent actions)

- A dedicated Activity view lists recent completions and skips across **all** of the
  user's tasks in reverse-chronological order (newest first).
- Built from `task_histories` rather than the task active/inactive flag, so completions of
  **recurring** tasks appear here even though the task stays active.
- Each entry shows the task title, whether it was completed or skipped, an on-time/late
  indicator, and when it happened.
- Cursor-based pagination (`before_id`) keeps the feed stable while new actions arrive.
- Available on both the web frontend (`/activity`) and the Android app (bottom-nav
  destination), each as a separate destination alongside the per-task history screen.

### Revert

- Each entry that is the **most recent action for its task** (`is_latest`) can be reverted.
  Older entries are shown for context but are not revertible, which keeps recurring
  schedules coherent.
- Revert targets a specific history id via `POST /tasks/{id}/undo?history_id=`. In one
  transaction the backend verifies the row belongs to the task and is still the latest;
  if a newer action landed in the meantime it returns **409 Conflict** and the client
  refreshes the feed and shows a message.
- Reverting deletes the history row and restores the task's previous due date and active
  state, rolling a recurring task back to the occurrence that was completed.
- The web client reverts over WebSocket/HTTP and refreshes via broadcasts; Android issues
  the revert as a direct online API call (the Activity feed is online-only and not part of
  the offline outbox). The MCP `uncomplete` tool resolves the task's latest history id
  before calling undo.
