[![Build](https://github.com/dkhalife/tasks-backend/actions/workflows/go-build.yml/badge.svg)](https://github.com/dkhalife/tasks-backend/actions/workflows/go-build.yml) [![CodeQL](https://github.com/dkhalife/tasks-backend/actions/workflows/github-code-scanning/codeql/badge.svg)](https://github.com/dkhalife/tasks-backend/actions/workflows/github-code-scanning/codeql) [![Dependabot Updates](https://github.com/dkhalife/tasks-backend/actions/workflows/dependabot/dependabot-updates/badge.svg)](https://github.com/dkhalife/tasks-backend/actions/workflows/dependabot/dependabot-updates)

# Task Wizard

**Privacy First, Productivity Always!**

Task Wizard is a free and open-source app designed to help manage tasks effectively. Its primary focus is to give users control over data and allow them to build integrations around it however they choose to.

This repo contains the backend logic for the app. You can find the default frontend implementation that is released with it in the [tasks-frontend](https://github.com/dkhalife/tasks-frontend) repo. This repo started as a fork of [DoneTick](https://github.com/donetick/donetick) but has since diverged from the original source code in order to accomplish different goals. Kudos to the contributors of [DoneTick](https://github.com/donetick/donetick) for helping kickstart this project.

## 🎯 Goals and principles

Task Wizard's primary goal is to allow users to own and protect their data and the following principles are ways to accomplish that:

* All the user data sent by this frontend only ever goes to a single backend
* 🔜 When data is stored, it is encrypted with a user key
* The code is continously scanned by a CI that runs CodeQL
* Dependencies are kept to a minimum
* When vulnerabilities are detected in dependencies they are auto updated with Dependabot

## ✨ Features

✅ Fast and simple task creation and completion for those times you are in a hurry

🏷️ Label assignment to help you categorize and recall tasks efficiently

📅 Due and completion dates tracking for users who need historical records

🔁 Recurring patterns for those chores you don't want to forget

📧 Notifications for important deadlines you don't want to miss

## 🚀 Installation

### Using Docker Compose (recommended)

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
      environment:
      - TW_ENV=prod
```

2. Run the app with `docker compose up -d` 

Alternatively, you can use a `.env` file and reference it in the compose file using an `env_file` entry.

### Using Docker

1. Pull the latest image: `docker pull dkhalife/task-wizard`
1. Run the container:

```bash
docker run \
   -v /path/to/host/config:/config
   -p 2021:2021 \
   -e TW_ENV=prod \
   dkhalife/task-wizard
```

Make sure to replace `/path/to/host` with your preferred root directory for config.

## ⚙️ Configuration

In the [config](./config/) directory are a couple of starter configuration files for a `debug` and a `prod` environment. The environment variable `TW_ENV` helps toggle between those configuration files as well as set the verbosity of the logs printed at runtime.

The configuration files are yaml mappings with the following values:

| Configuration Entry                      | Default Value                                       | Description                                                                 |
|------------------------------------------|-----------------------------------------------------|-----------------------------------------------------------------------------|
| `name`                                   | `"prod"`                                            | The name of the environment configuration.                                  |
| `database.migration`                     | `true`                                              | Indicates if database migration should be performed.                        |
| `database.path`                          | `/config/task-wizard.db`                            | The path at which to store the SQLite database.                             |
| `jwt.secret`                             | `"secret"`                                          | The secret key used for signing JWT tokens. **Make sure to change that.**   |
| `jwt.session_time`                       | `168h`                                              | The duration for which a JWT session is valid.                              |
| `jwt.max_refresh`                        | `168h`                                              | The maximum duration for refreshing a JWT session.                          |
| `server.port`                            | `2021`                                              | The port on which the server listens.                                       |
| `server.read_timeout`                    | `2s`                                                | The maximum duration for reading the entire request.                        |
| `server.write_timeout`                   | `1s`                                                | The maximum duration before timing out writes of the response.              |
| `server.rate_period`                     | `60s`                                               | The period for rate limiting.                                               |
| `server.rate_limit`                      | `300`                                               | The maximum number of requests allowed within the rate period.              |
| `server.serve_frontend`                  | `true`                                              | Indicates if the frontend should be served by the backend server.           |
| `server.debug`                           | `false`                                             | Indicates if the server should run in debug mode (only use for development) |
| `scheduler_jobs.due_frequency`           | `5m`                                                | The interval for sending regular notifications.                             |
| `scheduler_jobs.overdue_frequency`       | `24h`                                               | The interval for sending overdue notifications.                             |
| `scheduler_jobs.password_reset_validity` | `24h`                                               | How long password reset tokens are valid for.                               |
| `email.host`                             | (empty)                                             | The email server host.                                                      |
| `email.port`                             | (empty)                                             | The email server port.                                                      |
| `email.key`                              | (empty)                                             | The key for authenticating with the email server.                           |
| `email.email`                            | (empty)                                             | The email address used for sending emails.                                  |
| `email.appHost`                          | (empty)                                             | The application host URL used in email communications.                      |

## 🛠️ Development Environment

1. Clone the repo: `git clone https://github.com/dkhalife/tasks-backend.git`
1. Navigate to the project directory: `cd path/to/cloned/repo`
1. Install dependencies: `go mod download`
1. Run the app: `go run .`
1. (optional) For live reload, install air with
`go install github.com/cosmtrek/air@latest` then to run the app use `air`

## 🤝 Contributing

Contributions are welcome! If you would like to contribute to this repo, please follow these steps:

1. Fork the repository
1. Create a new branch: `git checkout -b feature/your-feature-name`
1. Make your changes and commit them: `git commit -m 'Add some feature'`
1. Push to the branch: `git push origin feature/your-feature-name`
1. Submit a pull request

If you have ideas but aren't familiar with code, you can also [open issues](https://github.com/dkhalife/tasks-backend/issues).

## 🔒 License

See the [LICENSE](LICENSE) file for more details.
