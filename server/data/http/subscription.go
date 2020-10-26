package http

import (
	"context"
	"fmt"
	"time"

	"github.com/mr-tron/base58"

	"gitlab.com/beneath-hq/beneath/models"
	"gitlab.com/beneath-hq/beneath/pkg/log"
	"gitlab.com/beneath-hq/beneath/pkg/secrettoken"
	"gitlab.com/beneath-hq/beneath/pkg/ws"
	"gitlab.com/beneath-hq/beneath/services/data"
)

// KeepAlive implements ws.Server
func (a *app) KeepAlive(numClients int, elapsed time.Duration) {
	// log state
	log.S.Infow(
		"ws keepalive",
		"clients", numClients,
		"elapsed", elapsed,
	)
}

// InitClient implements ws.Server
func (a *app) InitClient(client *ws.Client, payload map[string]interface{}) error {
	// get secret
	var secret models.Secret
	tokenStr, ok := payload["secret"].(string)
	if ok {
		// parse token
		token, err := secrettoken.FromString(tokenStr)
		if err != nil {
			return fmt.Errorf("malformed secret")
		}

		// authenticate
		secret = a.SecretService.AuthenticateWithToken(client.Context, token)
		if secret == nil {
			return fmt.Errorf("couldn't authenticate secret")
		}
	}

	if secret == nil {
		secret = &models.AnonymousSecret{}
	}

	// set secret as state
	client.SetState(secret)

	// log
	a.logWithSecret(secret, "ws init client",
		"ip", client.GetRemoteAddr(),
	)

	return nil
}

// CloseClient implements ws.Server
func (a *app) CloseClient(client *ws.Client) {
	// the client state is the secret
	secret := client.GetState().(models.Secret)

	// log
	a.logWithSecret(secret, "ws close client",
		"ip", client.GetRemoteAddr(),
		"reads", client.MessagesRead,
		"bytes_read", client.BytesRead,
		"writes", client.MessagesWritten,
		"bytes_written", client.BytesWritten,
		"elapsed", time.Since(client.StartTime),
		"error", client.Err,
	)
}

// StartQuery implements ws.Server
func (a *app) StartQuery(client *ws.Client, id ws.QueryID, payload map[string]interface{}) error {
	// get secret
	secret := client.GetState().(models.Secret)

	// get cursor bytes
	cursorStr, ok := payload["cursor"].(string)
	if !ok {
		return fmt.Errorf("payload must contain key 'cursor'")
	}
	cursorBytes, err := base58.Decode(cursorStr)
	if err != nil {
		return fmt.Errorf("payload key 'cursor' must contain a base58-encoded cursor")
	}

	// start subscription
	res, err := a.DataService.HandleSubscribe(client.Context, &data.SubscribeRequest{
		Secret: secret,
		Cursor: cursorBytes,
		Cb: func(msg data.SubscriptionMessage) {
			// TODO: For now, just sending empty messages
			client.SendData(id, "")
		},
	})
	if err != nil {
		return err
	}

	// set cancel as query state
	client.SetQueryState(id, res.Cancel)

	// log
	a.logWithSecret(secret, "ws start query",
		"ip", client.GetRemoteAddr(),
		"id", id,
		"instance", res.InstanceID.String(),
	)

	return nil
}

// StopQuery implements ws.Server
func (a *app) StopQuery(client *ws.Client, id ws.QueryID) error {
	// the query state is the cancel func
	cancel := client.GetQueryState(id).(context.CancelFunc)
	cancel()

	// the client state is the secret
	secret := client.GetState().(models.Secret)

	// log
	a.logWithSecret(secret, "ws stop query",
		"ip", client.GetRemoteAddr(),
		"id", id,
	)

	return nil
}

func (a *app) logWithSecret(sec models.Secret, msg string, keysAndValues ...interface{}) {
	l := log.S

	if sec.IsUser() {
		l = l.With(
			"secret", sec.GetSecretID().String(),
			"user", sec.GetOwnerID().String(),
		)
	} else if sec.IsService() {
		l = l.With(
			"secret", sec.GetSecretID().String(),
			"service", sec.GetOwnerID().String(),
		)
	}

	l.Infow(msg, keysAndValues...)
}
