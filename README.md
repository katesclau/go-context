
# Go Context and net/http Exploration with Bluesky API

This repository contains an example of using Go's `context` package and the `net/http` package to interact with the Bluesky API. The example demonstrates how to manage request timeouts and cancellations effectively using contexts.

## Prerequisites

- Go 1.16 or later
- Access to the Bluesky API (API key or other authentication if required)

## Installation

1. Clone the repository:
    ```sh
    git clone https://github.com/katesclau/go-context.git
    cd go-context-bluesky
    ```

2. Install dependencies:
    ```sh
    go mod tidy
    ```

## Usage

1. Set up your environment variables on `.env`:

2. Run the example:
    ```sh
    go run main.go
    ```

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Bluesky API](https://docs.bsky.app/docs/api)
- [Go Documentation](https://pkg.go.dev)
