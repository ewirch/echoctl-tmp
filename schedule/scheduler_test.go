package schedule_test

import (
	"echoctl/schedule"
	"github.com/stretchr/testify/assert"
	"gopkg.in/tomb.v2"
	"testing"
	"time"
)

func TestNewPublisher(t *testing.T) {
	t.Run("scheduler exits on cancel", func(t *testing.T) {
		t.Parallel()

		cancel := make(chan struct{})
		var tmb tomb.Tomb

		scheduler := schedule.NewScheduler[int]()
		tmb.Go(func() error {
			scheduler.Run(cancel)
			return nil
		})
		close(cancel)

		select {
		case <-tmb.Dead():
		case <-time.After(time.Second):
			t.Errorf("Expected scheduler to exit, but did not exit in 1s")
		}
	})
	t.Run("schedule request with duration 0 is immediately returned", func(t *testing.T) {
		t.Parallel()

		cancel := make(chan struct{})
		defer close(cancel)
		var tmb tomb.Tomb

		scheduler := schedule.NewScheduler[int]()
		tmb.Go(func() error {
			scheduler.Run(cancel)
			return nil
		})
		scheduler.Schedule() <- schedule.Request[int]{Data: intPtr(1), TriggerIn: time.Millisecond}

		select {
		case trigger := <-scheduler.Next():
			assert.Equal(t, 1, *trigger.Data, "expected the payload in the trigger to be 1")
		case <-time.After(time.Second):
			t.Errorf("Expected scheduler to trigger, but it did not.")
		}

	})
	t.Run("schedule request are only returned after specified duration", func(t *testing.T) {
		t.Parallel()

		cancel := make(chan struct{})
		defer close(cancel)
		var tmb tomb.Tomb

		scheduler := schedule.NewScheduler[int]()
		tmb.Go(func() error {
			scheduler.Run(cancel)
			return nil
		})
		scheduler.Schedule() <- schedule.Request[int]{Data: intPtr(1), TriggerIn: 1 * time.Millisecond}
		scheduler.Schedule() <- schedule.Request[int]{Data: intPtr(3), TriggerIn: 3 * time.Millisecond}
		scheduler.Schedule() <- schedule.Request[int]{Data: intPtr(6), TriggerIn: 6 * time.Millisecond}

		actual := make([]schedule.Trigger[int], 0)
		for i := 0; i < 3; i++ {
			next := readWithTimeout(t, scheduler.Next())
			actual = append(actual, *next)
		}
		assert.NotEmpty(t, actual, "actual is expected to contain elements")
		assert.Equal(t, 1, *actual[0].Data, "received triggers have to match in value and order")
		assert.Equal(t, 3, *actual[1].Data, "received triggers have to match in value and order")
		assert.Equal(t, 6, *actual[2].Data, "received triggers have to match in value and order")
	})

	// It is possible that multiple (maybe even "many") items trigger at the same time, or close to each other, and the consuming go routine is reading the Next channel slowly, and maybe even depends on scheduler to consume from the Schedule channel at the same time. Therefore, scheduler should not block when sending to the Next channel.
	t.Run("does not block on full next-channel", func(t *testing.T) {
		t.Parallel()

		cancel := make(chan struct{})
		defer close(cancel)
		var tmb tomb.Tomb

		scheduler := schedule.NewScheduler[int]()
		tmb.Go(func() error {
			scheduler.Run(cancel)
			return nil
		})
		for i := 0; i < 10; i++ {
			scheduler.Schedule() <- schedule.Request[int]{Data: intPtr(i), TriggerIn: time.Duration(i) * 10 * time.Millisecond}
		}
		for i := 0; i < 10; i++ {
			trigger := readWithTimeout(t, scheduler.Next())
			assert.Equal(t, i, *trigger.Data, "received triggers have to match in value and order")
		}
	})
}

func intPtr(i int) *int {
	return &i
}

func readWithTimeout[T any](t *testing.T, ch <-chan T) *T {
	select {
	case value := <-ch:
		return &value
	case <-time.After(time.Second):
		assert.Fail(t, "Poller failed to send command in 1s")
		return nil
	}
}
