# Feature: Android App

Native Android client for Task Wizard with task and label management.

## Capabilities

- MSAL (Microsoft Entra ID) authentication matching the web frontend's auth flow
- Task management: create, edit, complete, uncomplete, skip, delete tasks
- Recurrence support: once, daily, weekly, monthly, yearly, custom
- Label management: create, edit, delete labels with color coding
- Label assignment to tasks via multi-select chips
- Real-time updates via WebSocket connection
- Configurable server endpoint
- Pull-to-refresh on task and label lists
- Material 3 design with dynamic color support

## Architecture

- **UI**: Jetpack Compose with Material 3
- **Navigation**: Navigation Compose with bottom bar (Tasks, Labels, Settings)
- **State**: ViewModels with StateFlow, Hilt injection
- **Network**: Retrofit for REST, OkHttp WebSocket for real-time sync
- **Auth**: MSAL single-account flow with in-memory token cache
- **DI**: Hilt with singleton-scoped managers and repositories

## Key Surfaces

- `auth/` — AuthManager, AuthTokenProvider (MSAL wrapper)
- `api/` — TaskWizardApi (Retrofit), AuthInterceptor, ApiEndpointProvider
- `ws/` — WebSocketManager, WSMessage, WebSocketActions
- `repo/` — TaskRepository, LabelRepository, UserRepository
- `viewmodel/` — AuthViewModel, TaskListViewModel, TaskFormViewModel, LabelViewModel
- `ui/screen/` — SignInScreen, TaskListScreen, TaskFormScreen, LabelsScreen, SettingsScreen
- `ui/navigation/` — AppNavigation (NavHost with bottom bar)
- `ui/components/` — Reusable Compose components (DateTimePickerRow, GroupHeader, LabelDialog, LabelItem, NotificationsSection, RecurrenceSection, SchedulingSection, TaskChips, TaskItem)
- `ui/theme/` — Material 3 theme definitions (Color, Theme, Type)
- `ui/utils/` — UI utility functions (DateTimeUtils, TaskFormatters)
- `data/` — Local data layer (AppPreferences, GroupingRepository, TaskGrouper, ThemeRepository)
- `model/` — Data model classes (ApiResponse, AuthConfig, Label, Notification, Recurrence, Task, User)
- `di/` — Hilt dependency injection modules (AppModule, NetworkModule)
- `utils/` — Utility classes (SoundManager)
