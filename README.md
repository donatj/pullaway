# Pullaway

**Pullaway** is a lightweight Go [Pushover](https://pushover.net/) client cli and library. It  allows you to receive messages from Pushover in real-time via WebSocket. You can use it as a command-line tool or as a library in your Go projects.

**Note:** Pullaway is a Pushover **client** for receiving messages from Pushover. It is **not** a library for sending messages via Pushover.

> [!IMPORTANT]
> This tool requires a [Pushover account](https://pushover.net/) and a **Pushover Desktop license**. You can purchase a license at [https://pushover.net/clients/desktop](https://pushover.net/clients/desktop).

## Features

- **Login and Device Registration**: Securely log in to your Pushover account and register your device.
- **Message Retrieval**: Download and delete messages from your Pushover account.
- **Real-Time Listening**: Listen for incoming messages via WebSocket with automatic reconnection.
- **Secure Credential Storage**: Uses `keyring` to securely store your Pushover credentials.
- **Library Support**: Use Pullaway as a Go library in your own projects.

## Installation

### As a Command-Line Tool

Install **pullaway** using `go install`:

```bash
go install github.com/donatj/pullaway/cmd/pullaway@latest
```

This will download and install the `pullaway` command-line tool to your `$GOPATH/bin` directory.

### As a Library

To use **pullaway** as a library in your Go project, get it using `go get`:

```bash
go get github.com/donatj/pullaway
```

Then import it in your code:

```go
import "github.com/donatj/pullaway"
```

## Usage

### Prerequisites

- A [Pushover account](https://pushover.net/).
- A **Pushover Desktop license**: Purchase at [https://pushover.net/clients/desktop](https://pushover.net/clients/desktop).

### Command-Line Tool

#### Initialization

On first setup, you will need to run `pullaway init` to authorize your account and register the client as a Pushover device.

This only needs to be done once.

```bash
pullaway init
```

You will be prompted to enter:

- **Email**: Your Pushover account email.
- **Password**: Your Pushover account password.
- **Two-Factor Authentication**: If enabled, provide your 2FA code.
- **Device ShortName**: A name for your device (up to 25 characters).

Your credentials and device information will be securely stored using `keyring`.

#### Listening for Messages

After initialization, start listening for incoming messages:

```bash
pullaway listen
```

**Pullaway** will connect to Pushover's WebSocket server to receive messages in real-time. It automatically handles reconnections in case of network interruptions.

### Library Usage

You can use **pullaway** as a library in your Go projects to interact with Pushover. Below is a basic example:

```go
package main

import (
    "log"

    "github.com/donatj/pullaway"
)

func main() {
    // Initialize a new Pushover client
    pc := &pullaway.PushoverClient{}

    // Log in to Pushover
    loginResp, err := pc.Login("your-email@example.com", "your-password", "your-2fa-code")
    if err != nil {
        log.Fatalf("Login failed: %v", err)
    }

    // Register the device
    regResp, err := pc.Register(loginResp.Secret, "your-device-name")
    if err != nil {
        log.Fatalf("Registration failed: %v", err)
    }

    // Create an authorized client
    ac := pullaway.NewAuthorizedClient(loginResp.Secret, regResp.ID)

    // Download messages
    messages, err := ac.DownloadMessages()
    if err != nil {
        log.Fatalf("Failed to download messages: %v", err)
    }

    // Process messages
    for _, msg := range messages.Messages {
        log.Printf("Received message: %s", msg.Message)
    }
}
```

#### Listening for Messages with Reconnection

```go
package main

import (
    "log"

    "github.com/donatj/pullaway"
    "log/slog"
    "os"
)

func main() {
    // Assuming you have the user secret and device ID stored securely
    userSecret := "your-user-secret"
    deviceID := "your-device-id"

    // Create an authorized client
    ac := pullaway.NewAuthorizedClient(userSecret, deviceID)

    // Get an authorized listener with a logger
    listener := ac.GetAuthorizedListener(slog.New(slog.NewTextHandler(os.Stdout, nil)))

    // Define your message callback
    messageCallback := func() error {
        // Download and process messages
        messages, _, err := ac.DownloadAndDeleteMessages()
        if err != nil {
            return err
        }

        for _, msg := range messages.Messages {
            log.Printf("Received message: %s", msg.Message)
        }

        return nil
    }

    // Start listening with automatic reconnection
    err := listener.ListenWithReconnect(messageCallback)
    if err != nil {
        log.Fatalf("Error listening: %v", err)
    }
}
```

## Configuration

**Pullaway** securely stores your Pushover secret and device ID using the `keyring` library. This ensures that your sensitive information remains protected across sessions.

## Logging

By default, **pullaway** logs informational messages to the console. You can customize logging behavior by implementing your own `LeveledLogger` interface if needed.

## Important Note

**Pullaway** is intended solely for receiving messages from Pushover. It does **not** support sending messages. If you are looking for a library to send messages via Pushover, please refer to other available libraries.

## Contributing

Contributions are welcome! Feel free to open issues or submit pull requests for enhancements or bug fixes.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE.md) file for details.
