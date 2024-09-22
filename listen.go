package pullaway

import (
	"errors"
	"fmt"
	"log"

	"golang.org/x/net/websocket"
)

var (
	ErrNeedReconnect      = fmt.Errorf("need to reconnect")
	ErrWebsocketPermanent = fmt.Errorf("permanent error")
	ErrSessionIssue       = fmt.Errorf("session issue")
)

func ListenWithReconnect(deviceID string, secret string, ac *AuthorizedClient, ml MessageCallback) error {
connect:
	for {
		err := Listen(deviceID, secret, ac, ml)
		if errors.Is(err, ErrNeedReconnect) {
			continue connect
		}

		return err
	}
}

type MessageListener func(*DownloadResponse) error

type MessageCallback func() error

func Listen(deviceID string, secret string, ac *AuthorizedClient, ml MessageCallback) error {
	origin := "http://localhost/"
	url := "wss://client.pushover.net/push"

	// Establish WebSocket connection
	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		log.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer ws.Close()

	log.Println("Connected to WebSocket.")

	loginMessage := fmt.Sprintf("login:%s:%s\n", deviceID, secret)
	_, err = ws.Write([]byte(loginMessage))
	if err != nil {
		log.Fatalf("Error writing login message: %v", err)
	}

	for {
		var msg = make([]byte, 512)
		// Read message from WebSocket
		n, err := ws.Read(msg)
		if err != nil {
			return fmt.Errorf("error reading from WebSocket: %w", err)
		}

		// Print the message received
		log.Printf("Received: %s\n", msg[:n])

		for _, m := range msg[:n] {
			switch m {
			case '#': // Heartbeat
				log.Println("PONG")
			case '!': // Message
				if err := ml(); err != nil {
					return err
				}

				// messages, _, err := ac.DownloadAndDeleteMessages()
				// if err != nil {
				// 	log.Printf("error downloading messages: %v", err)
				// } else {
				// 	if err := ml(messages); err != nil {
				// 		return err
				// 	}
				// }
			case 'R': // Reconnect
				return ErrNeedReconnect
			case 'E': // Error, Permanent
				return ErrWebsocketPermanent
			case 'A': // Error, Session Closed
				return ErrSessionIssue
			default:
				log.Printf("Unknown message: %s\n", string(m))
			}
		}
	}
}
