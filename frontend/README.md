# 💻 Frontend

## 🔁 Inner loop

1. Ensure you have the latest packages installed with `yarn install`
1. Run `yarn start`. The output will contain instructions on how to browse the frontend.
1. Separately follow instructions from [apiserver](../apiserver)
1. (optionally) If using a different host for the backend, update `VITE_APP_API_URL` in [.env](./.env)

## 🧹 Linting

Code must pass linting rules before it is merged into main.

1. Run `yarn lint`

## 🧪 Testing

### 📃 Unit testing

1. Run `yarn test`

### 🧑🏼‍🔬 E2E Testing

End-to-end tests are written using [Playwright](https://playwright.dev/). 

To run the tests:
1. Ensure playwright dependencies are installed: `npx playwright install chromium`
1. Run tests: `yarn test:e2e`

Additional test commands:
* `yarn test:e2e:ui` - Run tests in interactive UI mode
* `yarn test:e2e:headed` - Run tests in headed mode (see the browser)

Tests are located in the `e2e/` directory and automatically run in CI on every pull request.
