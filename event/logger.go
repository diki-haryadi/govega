package event

import (
	"context"

	"github.com/diki-haryadi/govega/log"
	"github.com/sirupsen/logrus"
)

type EventLogger struct {
	log *logrus.Entry
}

func EventLoggerSender(ctx context.Context, config interface{}) (Sender, error) {
	return NewEventLogger(ctx, config)
}

func EventLoggerWriter(ctx context.Context, config interface{}) (Writer, error) {
	return NewEventLogger(ctx, config)
}

func EventLoggerListener(ctx context.Context, config interface{}) (Listener, error) {
	return NewEventLogger(ctx, config)
}

func NewEventLogger(ctx context.Context, config interface{}) (*EventLogger, error) {
	return &EventLogger{log: log.GetLogger(ctx, "event", "logger")}, nil
}

func (e *EventLogger) Send(ctx context.Context, message *EventMessage) error {
	h, _ := message.Hash()
	e.log.WithFields(logrus.Fields{
		"topic":    message.Topic,
		"key":      message.Key,
		"data":     message.Data,
		"metadata": message.Metadata,
		"hash":     h,
	}).Info()
	return nil
}

func (e *EventLogger) Delete(ctx context.Context, message *EventMessage) error {
	h, _ := message.Hash()
	e.log.WithFields(logrus.Fields{
		"hash": h,
	}).Info("message succesfully sent")
	return nil
}

func (e *EventLogger) Listen(ctx context.Context, topic, group string) (Iterator, error) {
	e.log.WithFields(logrus.Fields{
		"topic": topic,
		"group": group,
	}).Println("listen request")
	return e, nil
}

func (e *EventLogger) Next(ctx context.Context) (ConsumeMessage, error) {
	<-ctx.Done()
	return nil, nil
}
