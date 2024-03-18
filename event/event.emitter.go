package event

import (
	"context"
	"errors"
	"time"

	"github.com/sirupsen/logrus"

	"bitbucket.org/rctiplus/vegapunk/log"
)

type (
	Emitter struct {
		sender      Sender
		writer      Writer
		channel     chan *EventMessage
		eventConfig *EventConfig
	}

	EmitterConfig struct {
		Sender      *DriverConfig `json:"sender" mapstructure:"sender"`
		Writer      *DriverConfig `json:"writer" mapstructure:"writer"`
		EventConfig *EventConfig  `json:"event_config" mapstructure:"event_config"`
	}
)

func New(ctx context.Context, config *EmitterConfig) (*Emitter, error) {
	if config == nil {
		return nil, errors.New("[event/emitter] missing config")
	}

	em := &Emitter{
		channel:     make(chan *EventMessage),
		eventConfig: config.EventConfig,
	}

	if config.Sender == nil {
		config.Sender = &DriverConfig{Type: "log"}
		log.GetLogger(ctx, "event/emitter", "New").Info("empty sender, using log by default")
		//return nil, errors.New("[event/emitter] missing sender driver config")
	}

	if em.eventConfig == nil {
		em.eventConfig = NewEventConfig()
	}

	sf, ok := senders[config.Sender.Type]
	if !ok {
		return nil, errors.New("[event/emitter] unsupported sender driver")
	}

	sd, err := sf(ctx, config.Sender.Config)
	if err != nil {
		return nil, err
	}
	em.sender = sd

	if config.Writer != nil {

		wf, ok := writers[config.Writer.Type]
		if !ok {
			return nil, errors.New("[event/emitter] unsupported writer driver")
		}

		wr, err := wf(ctx, config.Writer.Config)
		if err != nil {
			return nil, err
		}

		em.writer = wr
		log.GetLogger(ctx, "event/emitter", "New").Info("enable hybrid mode")
		//running hybrid mode
		//don't use parent context on routine
		//because it might be canceled from parent routine when they finish
		//causing whatever logic inside the routine to be canceled right away
		//when they checking if the context is done
		go em.worker(context.Background())
	}

	return em, nil
}

func (e *Emitter) Publish(ctx context.Context, event, key string, message interface{}, metadata map[string]interface{}) error {
	log := log.WithContext(ctx).WithFields(logrus.Fields{
		"pkg":      "event",
		"function": "Publish",
	})

	if e.sender == nil {
		log.Error("[event/emitter] driver is not set")
		return errors.New("[event/emitter] driver is not set")
	}

	topic := e.eventConfig.getTopic(event)
	md := e.eventConfig.getMetadata(event)

	if metadata == nil {
		metadata = map[string]interface{}{}
	}

	for k, v := range md {
		metadata[k] = v
	}

	mhash, err := hash(message)
	if err != nil {
		return err
	}

	metadata[MetaHash] = mhash
	metadata[MetaTime] = time.Now()

	msg := &EventMessage{
		Topic:    topic,
		Key:      key,
		Data:     message,
		Metadata: metadata,
	}

	if e.writer != nil {
		// usinghybrid mode
		if err := e.writer.Send(ctx, msg); err != nil {
			return err
		}

		e.channel <- msg
	}

	return e.sender.Send(ctx, msg)
}

func (e *Emitter) worker(ctx context.Context) {
	log := log.WithContext(ctx).WithFields(logrus.Fields{
		"pkg":      "event",
		"function": "worker",
	})
	log.Info("Running event emitter in hybrid mode")

	for msg := range e.channel {
		if err := e.sender.Send(ctx, msg); err != nil {
			log.WithError(err).Error("[event/emitter] Error sending message through sender")
			continue
		}

		if err := e.writer.Delete(ctx, msg); err != nil {
			log.WithError(err).Error("[event/emitter] Error deleting message through writer")
			continue
		}
	}
}
