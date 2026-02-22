# Feature: Task Management

The core feature of Task Wizard. Users can create, edit, complete, skip, reschedule, and delete tasks.

## Capabilities

- CRUD operations on tasks with title, due date, labels, and notification settings
- Mark tasks as complete or uncomplete them
- Skip a task occurrence (advances to next due date without recording completion)
- Reschedule a task by updating its due date
- Group tasks by due date or by label in the UI
- Search/filter tasks in a table overview
- Quick-complete button on task cards with audio feedback
- Context menu on tasks for fast actions (edit, delete, complete, skip, reschedule, view history)
- Keyboard shortcut (`+`) to create a new task from the overview

## Data Model

A task has a title, optional next due date, optional end date, active/inactive flag, and associations to labels and notification triggers. Tasks are owned by the user who created them.
