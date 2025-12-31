package fcm

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"go.uber.org/zap"
	"google.golang.org/api/option"
)

type Client struct {
	msgClient *messaging.Client
	logger    *zap.Logger
}

func NewClient(ctx context.Context, logger *zap.Logger, credentialsFile string) (*Client, error) {
	var opts []option.ClientOption
	if credentialsFile != "" {
		opts = append(opts, option.WithCredentialsFile(credentialsFile))
	} else {
		logger.Warn("No Firebase credentials file provided. FCM will utilize environment variable GOOGLE_APPLICATION_CREDENTIALS or default credentials.")
	}

	app, err := firebase.NewApp(ctx, nil, opts...)
	if err != nil {
		return nil, fmt.Errorf("error initializing firebase app: %w", err)
	}

	msgClient, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting messaging client: %w", err)
	}

	return &Client{
		msgClient: msgClient,
		logger:    logger,
	}, nil
}

func (c *Client) Send(ctx context.Context, token string, title, body string, data map[string]string) error {
	if token == "" {
		return nil // No token, skip
	}

	message := &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
	}

	_, err := c.msgClient.Send(ctx, message)
	if err != nil {
		c.logger.Error("Failed to send FCM message", zap.String("token", token), zap.Error(err))
		return err
	}
	return nil
}
