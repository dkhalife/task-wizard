# 📡 API Server

## 🔁 Inner loop

1. Navigate to the root of the repo
1. Ensure you have the latest packages installed with `go mod download`
1. Run the app `go run .`
1. (optional) For live reload, install air with
`go install github.com/cosmtrek/air@latest` then to run the app use `air`

## 🧹 Lint

Code must pass linting rules before it is merged into main.

1. Install `golangci-lint` if you don't have it: `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`
1. Run `go lint`

## 🧪 Testing

### 📃 Unit testing

1. Run `go test`
