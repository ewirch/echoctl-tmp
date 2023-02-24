package main

import (
	"context"
	"echoctl/can"
	"echoctl/mqtt"
	"fmt"
	phaoMqtt "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	"gopkg.in/tomb.v2"
	"reflect"
	"time"
)

func daemonize(lc fx.Lifecycle, shutdowner fx.Shutdowner, stopTimeout time.Duration, log *zap.Logger, client phaoMqtt.Client, socket can.Socket, tombs ...*tomb.Tomb) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			killAllAndWait(ctx, tombs)
			client.Disconnect(mqtt.GetQuiesce(ctx))
			_ = socket.Close()
			return nil
		},
	})

	go shutdownOnDeath(tombs, log, stopTimeout, shutdowner)
}

func shutdownOnDeath(tombs []*tomb.Tomb, log *zap.Logger, stopTimeout time.Duration, shutdowner fx.Shutdowner) {
	// Wait for the first tomb to die.
	diedIdx, died := selectOnSlice((*tomb.Tomb).Dying, tombs)

	// Eagerly print error. Prevent that app panics, and the original error will not be exposed.
	printTombError(died, log)

	killCtx, killCtxCancel := context.WithTimeout(context.Background(), stopTimeout)
	defer killCtxCancel()
	killAllAndWait(killCtx, tombs)

	// Avoid double-printing error
	tombs = slices.Delete(tombs, diedIdx, diedIdx+1)
	printTombErrors(tombs, log)

	err := shutdowner.Shutdown()
	if err != nil {
		panic(fmt.Errorf("non-graceful shutdown: %w", err))
	}
}

func printTombErrors(tombs []*tomb.Tomb, log *zap.Logger) {
	for _, t := range tombs {
		printTombError(t, log)
	}
}

func printTombError(t *tomb.Tomb, log *zap.Logger) {
	err := t.Err()
	if err != nil && err != tomb.ErrDying {
		log.Error("Error", zap.Error(err))
	}
}

func killAllAndWait(ctx context.Context, tombs []*tomb.Tomb) {
	for _, t := range tombs {
		t.Kill(nil)
	}

	for _, t := range tombs {
		select {
		case <-t.Dead():
		case <-ctx.Done():
		}
	}
}

func selectOnSlice(channelFunction func(*tomb.Tomb) <-chan struct{}, tombs []*tomb.Tomb) (chosenIdx int, chosen *tomb.Tomb) {
	cases := make([]reflect.SelectCase, len(tombs))
	for i, t := range tombs {
		cases[i] = reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(channelFunction(t)),
		}
	}
	chosenIdx, _, _ = reflect.Select(cases)
	return chosenIdx, tombs[chosenIdx]
}
