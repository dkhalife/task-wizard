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
      - /path/to/host/data:/data
      - /path/to/host/config:/config
      environment:
      - TW_ENV=selfhosted
      - TW_SQLITE_PATH=/data/tasks.db
```

2. Run the app with `docker compose up -d` 

Alternatively, you can use a `.env` file and reference it in the compose file using an `env_file` entry.

### Using Docker

1. Pull the latest image: `docker pull dkhalife/task-wizard`
1. Run the container:

```bash
docker run \
   -v /path/to/host/config:/config
   -v /path/to/host/data:/data
   -p 2021:2021 \
   -e TW_ENV=prod \
   -e TW_SQLITE_PATH=/data/tasks.db \
   dkhalife/task-wizard
```

Make sure to replace `/path/to/host` with your preferred root directory for data and config.

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
