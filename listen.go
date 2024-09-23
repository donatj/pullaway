package pullaway

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"golang.org/x/net/websocket"
)

var (
	ErrWebsocketConnectFail = fmt.Errorf("error connecting to WebSocket")
	ErrWebsocketLoginFail   = fmt.Errorf("error logging in to WebSocket")
	ErrWebsocketReadFail    = fmt.Errorf("error reading from WebSocket")

	ErrNeedReconnect  = fmt.Errorf("need to reconnect")
	ErrPermanentIssue = fmt.Errorf("permanent error")
	ErrSessionIssue   = fmt.Errorf("session issue")
)

// LeveledLogger is an interface for loggers or logger wrappers that support leveled logging.
// The methods take a message string and optional variadic key-value pairs.
type LeveledLogger interface {
	Error(string, ...interface{})
	Info(string, ...interface{})
	Debug(string, ...interface{})
	Warn(string, ...interface{})
}

type Listener struct {
	Log LeveledLogger
}

func NewListener(l LeveledLogger) *Listener {
	if l == nil {
		l = slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	return &Listener{
		Log: l,
	}
}

func (l *Listener) ListenWithReconnect(deviceID string, secret string, ml MessageCallback) error {
connect:
	for {
		err := l.Listen(deviceID, secret, ml)
		if errors.Is(err, ErrPermanentIssue) || errors.Is(err, ErrSessionIssue) {
			return fmt.Errorf("listen error: %w", err)
		}

		if errors.Is(err, ErrNeedReconnect) {
			l.Log.Info("reconnecting on request", "error", err.Error())
			time.Sleep(5 * time.Second)
			continue connect
		}

		l.Log.Error("error listening to WebSocket", "error", err.Error())
		time.Sleep(15 * time.Second)
	}
}

type MessageCallback func() error

func (l *Listener) Listen(deviceID string, secret string, ml MessageCallback) error {
	origin := "http://localhost/"
	url := "wss://client.pushover.net/push"

	// Establish WebSocket connection
	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		return errors.Join(ErrWebsocketConnectFail, err)
	}
	defer ws.Close()

	l.Log.Debug("connected to WebSocket")

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

		l.Log.Debug("received message", "message", string(msg[:n]))

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
				return ErrPermanentIssue
			case 'A': // Error, Session Closed
				return ErrSessionIssue
			default:
				l.Log.Warn("unknown message", "message", string(m))
			}
		}
	}
}

var DefaultListener = &Listener{
	Log: slog.New(slog.NewTextHandler(io.Discard, nil)),
}

func ListenWithReconnect(deviceID string, secret string, ml MessageCallback) error {
	return DefaultListener.ListenWithReconnect(deviceID, secret, ml)
}

func Listen(deviceID string, secret string, ml MessageCallback) error {
	return DefaultListener.Listen(deviceID, secret, ml)
}

type AuthorizedListener struct {
	*AuthorizedClient
	*Listener
}

func NewAuthorizedListener(ac *AuthorizedClient, l LeveledLogger) *AuthorizedListener {
	return &AuthorizedListener{
		AuthorizedClient: ac,
		Listener:         NewListener(l),
	}
}

func (al *AuthorizedListener) ListenWithReconnect(ml MessageCallback) error {
	return al.Listener.ListenWithReconnect(al.DeviceID, al.UserSecret, ml)
}

func (al *AuthorizedListener) Listen(ml MessageCallback) error {
	return al.Listener.Listen(al.DeviceID, al.UserSecret, ml)
}
