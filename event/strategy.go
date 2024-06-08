package event

import (
	"context"
	"fmt"
)

var (
	consumeStrategy = map[string]ConsumeStrategyFactory{
		"always_commit":     NoConfigConsumeStrategyFactory(AlwaysCommitStrategy),
		"commit_on_success": NoConfigConsumeStrategyFactory(CommitOnSuccessStrategy),
	}
)

type (
	ConsumeStrategy func(ctx context.Context, message ConsumeMessage, handler EventHandler) error

	ConsumeStrategyFactory func(ctx context.Context, config interface{}) (ConsumeStrategy, error)
)

// RegisterConsumeStrategy register consumestrategy
func RegisterConsumeStrategy(name string, fn ConsumeStrategyFactory) {
	consumeStrategy[name] = fn
}

// AlwaysCommitStrategy will always commit the message no matter what is the handler result
func AlwaysCommitStrategy(ctx context.Context, message ConsumeMessage, handler EventHandler) error {
	em, err := message.GetEventConsumeMessage(ctx)
	if err != nil {
		return fmt.Errorf("[event/alwaysCommitStrategy] failed to get event message: %w", err)
	}

	if err := message.Commit(ctx); err != nil {
		return fmt.Errorf("[event/alwaysCommitStrategy] failed to commit: %w", err)
	}

	return handler(ctx, em)
}

// CommitOnSuccessStrategy will only commit the message if handler doesn't return error
func CommitOnSuccessStrategy(ctx context.Context, message ConsumeMessage, handler EventHandler) error {
	em, err := message.GetEventConsumeMessage(ctx)
	if err != nil {
		return fmt.Errorf("[event/commitOnSuccessStrategy] failed to get event message: %w", err)
	}

	if err := handler(ctx, em); err != nil {
		return fmt.Errorf("[event/commitOnSuccessStrategy] handler failed to process message: %w", err)
	}

	if err := message.Commit(ctx); err != nil {
		return fmt.Errorf("[event/commitOnSuccessStrategy] failed to commit: %w", err)
	}

	return nil
}

// NoConfigConsumeStrategyFactory util function to create factory for consume strategy
// which doesn't need any config
func NoConfigConsumeStrategyFactory(strategy ConsumeStrategy) ConsumeStrategyFactory {
	return func(ctx context.Context, config interface{}) (ConsumeStrategy, error) {
		return strategy, nil
	}
}
