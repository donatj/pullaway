package pullaway

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"time"

	"golang.org/x/net/websocket"
)

var (
	ErrWebsocketConnectFail = fmt.Errorf("error connecting to WebSocket")
	ErrWebsocketLoginFail   = fmt.Errorf("error logging in to WebSocket")
	ErrWebsocketReadFail    = fmt.Errorf("error reading from WebSocket")
	ErrNeedReconnect        = fmt.Errorf("need to reconnect")
	ErrWebsocketPermanent   = fmt.Errorf("permanent error")
	ErrSessionIssue         = fmt.Errorf("session issue")
)

func ListenWithReconnect(deviceID string, secret string, ml MessageCallback) error {
connect:
	for {
		err := Listen(deviceID, secret, ml)
		if errors.Is(err, ErrWebsocketPermanent) || errors.Is(err, ErrSessionIssue) {
			return err
		}

		if errors.Is(err, ErrNeedReconnect) {
			time.Sleep(5 * time.Second)
			continue connect
		}

		time.Sleep(15 * time.Second)
		slog.Error("error listening to WebSocket", slog.String("error", err.Error()))
	}
}

type MessageCallback func() error

func Listen(deviceID string, secret string, ml MessageCallback) error {
	origin := "http://localhost/"
	url := "wss://client.pushover.net/push"

	// Establish WebSocket connection
	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		return errors.Join(ErrWebsocketConnectFail, err)
	}
	defer ws.Close()

	slog.Info("connected to WebSocket")

	loginMessage := fmt.Sprintf("login:%s:%s\n", deviceID, secret)
	_, err = ws.Write([]byte(loginMessage))
	if err != nil {
		return errors.Join(ErrWebsocketLoginFail, err)
	}

	for {
		var msg = make([]byte, 512)
		// Read message from WebSocket
		n, err := ws.Read(msg)
		if err != nil {
			return errors.Join(ErrWebsocketReadFail, err)
		}

		// Print the message received
		log.Printf("Received: %s\n", msg[:n])

		for _, m := range msg[:n] {
			switch m {
			case '#': // Heartbeat
				// PONG
			case '!': // Message
				if err := ml(); err != nil {
					return err
				}
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
