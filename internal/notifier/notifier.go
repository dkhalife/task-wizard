package notifier

import (
	"context"

	nModel "donetick.com/core/internal/notifier/model"
)

type Notifier struct {
}

func NewNotifier() *Notifier {
	return &Notifier{}
}

func (n *Notifier) SendNotification(c context.Context, notification *nModel.Notification) error {
	return nil
}
