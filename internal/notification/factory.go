package notification

import (
	"fmt"

	"task-scheduler/internal/domain"

	"go.uber.org/zap"
)

// NotifierFactory creates notifiers based on channel configuration
type NotifierFactory struct {
	logger *zap.Logger
}

// NewNotifierFactory creates a new notifier factory
func NewNotifierFactory(logger *zap.Logger) *NotifierFactory {
	return &NotifierFactory{
		logger: logger,
	}
}

// CreateNotifier creates a notifier based on the channel type and configuration
func (nf *NotifierFactory) CreateNotifier(channel domain.NotificationChannel) (domain.Notifier, error) {
	switch channel.Type {
	case domain.ChannelTypeEmail:
		return NewEmailNotifierFromChannel(channel, nf.logger)

	case domain.ChannelTypeWebhook:
		return NewWebhookNotifierFromChannel(channel, nf.logger)

	case domain.ChannelTypeSMS:
		return NewSMSNotifierFromChannel(channel, nf.logger)

	default:
		return nil, fmt.Errorf("unsupported channel type: %s", channel.Type)
	}
}

// CreateNotifiers creates multiple notifiers from a list of channels
func (nf *NotifierFactory) CreateNotifiers(channels []domain.NotificationChannel) (map[domain.ChannelType]domain.Notifier, error) {
	notifiers := make(map[domain.ChannelType]domain.Notifier)

	for _, channel := range channels {
		notifier, err := nf.CreateNotifier(channel)
		if err != nil {
			nf.logger.Warn("failed to create notifier", zap.String("type", string(channel.Type)), zap.Error(err))
			continue
		}
		notifiers[channel.Type] = notifier
	}

	return notifiers, nil
}
