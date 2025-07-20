[![Build](https://github.com/dkhalife/tasks-backend/actions/workflows/go-build.yml/badge.svg)](https://github.com/dkhalife/tasks-backend/actions/workflows/go-build.yml) [![codecov](https://codecov.io/gh/dkhalife/tasks-backend/graph/badge.svg?token=UQ4DTE3WI1)](https://codecov.io/gh/dkhalife/tasks-backend) [![CodeQL](https://github.com/dkhalife/tasks-backend/actions/workflows/github-code-scanning/codeql/badge.svg)](https://github.com/dkhalife/tasks-backend/actions/workflows/github-code-scanning/codeql) [![Dependabot Updates](https://github.com/dkhalife/tasks-backend/actions/workflows/dependabot/dependabot-updates/badge.svg)](https://github.com/dkhalife/tasks-backend/actions/workflows/dependabot/dependabot-updates)

# Task Wizard

**Privacy First, Productivity Always!**

Task Wizard is a free and open-source app designed to help manage tasks effectively. Its primary focus is to give users control over data and allow them to build integrations around it however they choose to.

This repo contains the backend logic for the app. You can find the default frontend implementation that is released with it in the [tasks-frontend](https://github.com/dkhalife/tasks-frontend) repo. This repo started as a fork of [DoneTick](https://github.com/donetick/donetick) but has since diverged from the original source code in order to accomplish different goals. Kudos to the contributors of [DoneTick](https://github.com/donetick/donetick) for helping kickstart this project.

## 🎯 Goals and principles

Task Wizard's primary goal is to allow users to own and protect their data and the following principles are ways to accomplish that:

* All the user data sent by this frontend only ever goes to a single backend
* 🔜 When data is stored, it is encrypted with a user key
* The code is continuously scanned by a CI that runs CodeQL
* Dependencies are kept to a minimum
* When vulnerabilities are detected in dependencies they are auto updated with Dependabot

## ✨ Features

✅ Fast and simple task creation and completion for those times you are in a hurry

🏷️ Label assignment to help you categorize and recall tasks efficiently

📅 Due and completion dates tracking for users who need historical records

🔁 Recurring patterns for those chores you don't want to forget

📧 Notifications for important deadlines you don't want to miss

🗝️ Fine-grained access tokens for endless integration possibilities

🌐 Authenticated CalDAV endpoint at `/dav/tasks` with app token as the password

## ⌨️ Keyboard Shortcuts

| Context/Screen                | Shortcut                           | Action or Result                                                   |
|-------------------------------|------------------------------------|-------------------------------------------------------------------|
| Tasks Overview                | `Ctrl + F`                         | Focuses the search box.                                           |
| Tasks Overview                | `+` (outside of inputs)            | Opens the “Add Task” screen.                                      |
| Forgot Password, Task Edit, and Password/Date modals | `Enter` in text input fields, password fields, or date fields | Submits or saves the form or dialog.                              |

## 🚀 Installation

### 🚢 Using Docker Compose (recommended)

1. In a compose.yml file, paste the following:

```yaml
services:
   tasks:
      image: dkhalife/task-wizard
      container_name: tasks
      restart: unless-stopped
      ports:
      - 2021:2021
      volumes:
      - /path/to/host/config:/config
```

2. Run the app with `docker compose up -d` 

Alternatively, you can use a `.env` file and reference it in the compose file using an `env_file` entry.

### 🛳️ Using Docker

1. Pull the latest image: `docker pull dkhalife/task-wizard`
1. Run the container:

```bash
docker run \
   -v /path/to/host/config:/config
   -p 2021:2021 \
   dkhalife/task-wizard
```

Make sure to replace `/path/to/host` with your preferred root directory for config.

## ⚙️ Configuration

In the [config](./config/) directory are a couple of starter configuration files for prod and dev environments. The server expects a config.yaml in the config directory and will load settings from it when started.

**Note:** You can set `email.host`, `email.port`, `email.email`, `email.password` and `jwt.secret` using environment variables `TW_EMAIL_HOST`, `TW_EMAIL_PORT`, `TW_EMAIL_SENDER`, `TW_EMAIL_PASSWORD` and `TW_JWT_SECRET` for improved security and flexibility. The server will fail to start if `jwt.secret` is left as `"secret"`, so be sure to set `TW_JWT_SECRET` or edit `config.yaml`.

The configuration files are yaml mappings with the following values:

| Configuration Entry                      | Default Value                                       | Description                                                                 |
|------------------------------------------|-----------------------------------------------------|-----------------------------------------------------------------------------|
| `name`                                   | `"prod"`                                            | The name of the environment configuration.                                  |
| `database.migration`                     | `true`                                              | Indicates if database migration should be performed.                        |
| `database.path`                          | `/config/task-wizard.db`                            | The path at which to store the SQLite database.                             |
| `jwt.secret`                             | `"secret"`                                          | The secret key used for signing JWT tokens. **Make sure to change that or set `TW_JWT_SECRET`.**   |
| `jwt.session_time`                       | `168h`                                              | The duration for which a JWT session is valid.                              |
| `jwt.max_refresh`                        | `168h`                                              | The maximum duration for refreshing a JWT session.                          |
| `server.host_name`                       | `localhost`                                         | The hostname to use for external links.                                     |
| `server.port`                            | `2021`                                              | The port on which the server listens.                                       |
| `server.read_timeout`                    | `2s`                                                | The maximum duration for reading the entire request.                        |
| `server.write_timeout`                   | `1s`                                                | The maximum duration before timing out writes of the response.              |
| `server.rate_period`                     | `60s`                                               | The period for rate limiting.                                               |
| `server.rate_limit`                      | `300`                                               | The maximum number of requests allowed within the rate period.              |
| `server.serve_frontend`                  | `true`                                              | Indicates if the frontend should be served by the backend server.           |
| `server.registration`                    | `true`                                              | Indicates whether new accounts can be created on the backend server.        |
| `server.log_level`                       | `debug` when `server.debug` = `true`, else `warn`   | The min level to log (debug, info, warn, error, dpanic, panic, fatal).      |
| `server.allowed_origins`                 | `(empty)`                                           | Origins allowed to issue cross-domain requests.                             |
| `server.allow_credentials`               | `false`                                             | Whether cross-domain requests can include credentials.                      |
| `scheduler_jobs.due_frequency`           | `5m`                                                | The interval for sending regular notifications.                             |
| `scheduler_jobs.overdue_frequency`       | `24h`                                               | The interval for sending overdue notifications.                             |
| `scheduler_jobs.password_reset_validity` | `24h`                                               | How long password reset tokens are valid for.                               |
| `scheduler_jobs.notification_cleanup`    | `10m`                                               | The interval for cleaning up sent notifications.                            |
| `scheduler_jobs.token_expiration_cleanup`| `24h`                                               | The interval for cleaning up expired tokens.                                |
|`scheduler_jobs.token_expiration_reminder`| `72h`                                               | How long before an app token expiration to send a reminder for it.          |
| `email.host`                             | (empty)                                             | The email server host.                                                      |
| `email.port`                             | (empty)                                             | The email server port.                                                      |
| `email.email`                            | (empty)                                             | The email address used for sending emails.                                  |
| `email.password`                         | (empty)                                             | The password for authenticating with the email server.                      |

## 🛠️ Development

### 🖥️ Devcontainer

A [devcontainer](./.devcontainer/devcontainer.json) configuration is set up in this repo to help jumpstart development with all the required dependencies available for both the frontend and backend. You can use this configuration alongside
GitHub codespaces to jump into a remote development environment without installing anything on your local machine. For the best experience make sure your codespace has both repos cloned in it. Ports can be forwarded from within the container so that you are able to test changes locally through the VS Code tunnel.

### 📃 Requirements

* GoLang

### 🔁 Inner loop

1. Navigate to the root of the repo
1. Ensure you have the latest packages installed with `go mod download`
1. Run the app `go run .`
1. (optional) For live reload, install air with
`go install github.com/cosmtrek/air@latest` then to run the app use `air`

## 🤝 Contributing

Contributions are welcome! If you would like to contribute to this repo, feel free to fork the repo and submit pull requests.
If you have ideas but aren't familiar with code, you can also [open issues](https://github.com/dkhalife/tasks-backend/issues).

## 🔒 License

See the [LICENSE](LICENSE) file for more details.
