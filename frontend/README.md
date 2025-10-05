[![Node.js CI](https://github.com/dkhalife/tasks-frontend/actions/workflows/ci.yml/badge.svg)](https://github.com/dkhalife/tasks-frontend/actions/workflows/ci.yml) [![CodeQL](https://github.com/dkhalife/tasks-frontend/actions/workflows/github-code-scanning/codeql/badge.svg)](https://github.com/dkhalife/tasks-frontend/actions/workflows/github-code-scanning/codeql) 
[![Dependabot Updates](https://github.com/dkhalife/tasks-frontend/actions/workflows/dependabot/dependabot-updates/badge.svg)](https://github.com/dkhalife/tasks-frontend/actions/workflows/dependabot/dependabot-updates)

## Task Wizard Frontend

This repo is a frontend for [tasks-backend](https://github.com/dkhalife/tasks-backend), a free and open source solution for task management. It started as a fork of [DoneTick](https://github.com/donetick/frontend) but has since diverged from the original source code in order to accomplish different goals. Kudos to the contributors of [DoneTick](https://github.com/donetick/frontend) for helping kickstart this project.

## 🎯 Goals and principles

Task Wizard's primary goal is to allow users to own and protect their data and the following principles are ways to accomplish that:

* All the user data sent by this frontend only ever goes to a single backend
* 🔜 When data is cached on the local machine, it is encrypted with a user key
* The repo uses TypeScript with a strict set of rules designed to reduce bugs closest to development
* The code is continously scanned by a CI that runs CodeQL
* Dependencies are kept to a minimum
* When vulnerabilities are detected in dependencies they are auto updated with Dependabot

## 🚀 Installation

This app is meant to be self-hosted by individuals and groups who wish to own their data and those who prioritize their data privacy. For setup instructions check out the [tasks-backend](https://github.com/dkhalife/tasks-backend) README.

## 🛠️ Development

### 🖥️ Devcontainer

A [devcontainer](./.devcontainer/devcontainer.json) configuration is set up in this repo to help jumpstart development with all the required dependencies available for both the frontend and backend. You can use this configuration alongside
GitHub codespaces to jump into a remote development environment without installing anything on your local machine. For the best experience make sure your codespace has both repos cloned in it. Ports can be forwarded from within the container so that you are able to test changes locally through the VS Code tunnel.

### 📃 Requirements

* NodeJS 20+
* `yarn`

### 🔁 Inner loop

1. Navigate to the root of the repo
1. Ensure you have the latest packages installed with `yarn install`
1. Run `yarn start`. The output will contain instructions on how to browse the frontend.
1. Separately follow instructions from [tasks-backend](https://github.com/dkhalife/tasks-backend)
1. (optionally) If using a different host for the backend, update `VITE_APP_API_URL` in [.env](./.env)

### 🧪 E2E Testing

End-to-end tests are written using [Playwright](https://playwright.dev/). To run the tests:

1. Ensure dependencies are installed: `yarn install`
1. Install Playwright browsers: `npx playwright install chromium`
1. Run tests: `yarn test:e2e`

Additional test commands:
* `yarn test:e2e:ui` - Run tests in interactive UI mode
* `yarn test:e2e:headed` - Run tests in headed mode (see the browser)

Tests are located in the `e2e/` directory and automatically run in CI on every pull request.

## 🤝 Contributing

Contributions are welcome! If you would like to contribute to this repo, feel free to fork the repo and submit pull requests.
If you have ideas but aren't familiar with code, you can also [open issues](https://github.com/dkhalife/tasks-frontend/issues).

## 🔒 License

See the [LICENSE](LICENSE) file for more details.
