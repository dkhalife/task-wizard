# End-to-End Tests

## Overview

This directory contains end-to-end (E2E) tests for the Task Wizard frontend application using [Playwright](https://playwright.dev/).

## Test Suites

### Authentication & Onboarding (`auth.spec.ts`)

Comprehensive tests covering the core authentication flows:

#### 1. Sign Up Flow (6 tests)
- Navigation to signup page and form display
- Validation errors for empty/invalid inputs
- Invalid email format validation
- Weak password validation
- Invalid display name validation (special characters)
- Successful signup with verification message

#### 2. Login/Logout Flow (8 tests)
- Login page display and form elements
- Error handling for invalid credentials
- Navigation between login, signup, and forgot-password pages
- Successful login with valid credentials
- Session persistence after login
- Logout functionality and session clearing

#### 3. Password Reset Flow (6 tests)
- Password reset form display
- Validation for empty email
- Validation for invalid email format
- Successful reset request submission
- Navigation back to login from reset page
- Navigation to login after successful reset

#### 4. Session Expiry (6 tests)
- Handling expired tokens
- Token refresh for near-expiry tokens
- Re-authentication redirect on token expiry
- Token clearing on navigation with expired session

## Running Tests

### Prerequisites

1. Install dependencies:
   ```bash
   yarn install
   ```

2. Install Playwright browsers (first time only):
   ```bash
   yarn playwright install chromium
   ```

### Running All Tests

```bash
yarn test:e2e
```

### Running Tests in UI Mode

For interactive debugging and test development:

```bash
yarn test:e2e:ui
```

### Running Tests in Headed Mode

To see the browser while tests run:

```bash
yarn test:e2e:headed
```

### Running Specific Test Files

```bash
yarn playwright test e2e/auth.spec.ts
```

### Running Specific Tests

```bash
yarn playwright test -g "should display login form"
```

## Test Configuration

The test configuration is in `playwright.config.ts`:

- **Base URL**: `http://localhost:5173`
- **Test Directory**: `./e2e`
- **Web Server**: Automatically starts Vite dev server
- **Retries**: 2 retries in CI, 0 locally
- **Trace**: Enabled on first retry for debugging

## Backend Requirements

Some tests require a running backend API:

- **Sign up tests**: Need backend to create accounts
- **Login tests**: Need valid test accounts or mocked responses
- **Password reset tests**: Need backend email service

### Options for Backend Testing

1. **Mock API responses**: Use Playwright's request interception
2. **Test database**: Seed test database with known accounts
3. **Test environment**: Use a dedicated staging/test backend
4. **Skip backend tests**: Use `test.skip()` for tests requiring backend

## Test Patterns

### Selectors

The tests use various selector strategies:

- **By role**: `page.getByRole('button', { name: /sign in/i })`
- **By text**: `page.getByText('Create an account')`
- **By locator**: `page.locator('input[type="password"]')`
- **By autocomplete**: `page.locator('input[autocomplete="email"]')`

### Helper Functions

- `waitForNavigation(page, path)`: Wait for navigation to a specific path
- `clearAuthStorage(page)`: Clear authentication tokens from localStorage

### Test Structure

```typescript
test.describe('Feature Name', () => {
  test.beforeEach(async ({ page }) => {
    // Setup code (e.g., clear auth)
  });

  test('should do something', async ({ page }) => {
    // Test implementation
  });
});
```

## Debugging Tests

### View Test Report

After running tests, view the HTML report:

```bash
yarn playwright show-report
```

### View Traces

For failed tests with traces:

```bash
yarn playwright show-trace test-results/[test-name]/trace.zip
```

### Debug Mode

Run tests in debug mode with Playwright Inspector:

```bash
yarn playwright test --debug
```

### Screenshots on Failure

Playwright automatically captures screenshots on test failure. They're saved in `test-results/`.

## CI/CD Integration

The tests are configured to run in CI with:

- 2 retries on failure
- Single worker for stability
- No reuse of existing server
- HTML report generation

Add to your CI workflow:

```yaml
- name: Install dependencies
  run: cd frontend && yarn install

- name: Install Playwright browsers
  run: cd frontend && yarn playwright install --with-deps chromium

- name: Run E2E tests
  run: cd frontend && yarn test:e2e
```

## Best Practices

1. **Test isolation**: Each test should be independent
2. **Clear state**: Use `beforeEach` to reset state
3. **Wait properly**: Use Playwright's auto-waiting instead of fixed timeouts
4. **Descriptive names**: Use clear, descriptive test names
5. **Page objects**: Consider using page object model for complex pages
6. **Data-testid**: Add `data-testid` attributes for stable selectors

## Troubleshooting

### Tests timing out
- Check if the web server is starting correctly
- Increase timeout in playwright.config.ts
- Check for network issues or slow backend

### Element not found
- Verify the selector is correct
- Check if the element is visible
- Add explicit waits if needed

### Flaky tests
- Use Playwright's auto-waiting features
- Avoid fixed timeouts
- Check for race conditions
- Increase retries in CI

## Future Improvements

- [ ] Add API mocking for backend-dependent tests
- [ ] Add visual regression testing
- [ ] Add accessibility testing
- [ ] Add performance testing
- [ ] Add mobile device testing
- [ ] Add cross-browser testing (Firefox, Safari)
- [ ] Improve test data management
- [ ] Add page object models for better maintainability
