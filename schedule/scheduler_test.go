package schedule_test

import (
	"echoctl/schedule"
	"github.com/benbjohnson/clock"
	"github.com/stretchr/testify/assert"
	"gopkg.in/tomb.v2"
	"testing"
	"time"
)

func TestNewPublisher(t *testing.T) {
	t.Run("scheduler exits on cancel", func(t *testing.T) {
		t.Parallel()

		c := clock.NewMock()
		cancel := make(chan struct{})
		var tmb tomb.Tomb

		scheduler := schedule.NewScheduler[int](c)
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

		c := clock.NewMock()
		c.Set(parse("1970-01-01 00:00:00"))
		cancel := make(chan struct{})
		defer close(cancel)
		var tmb tomb.Tomb

		scheduler := schedule.NewScheduler[int](c)
		tmb.Go(func() error {
			scheduler.Run(cancel)
			return nil
		})
		scheduler.Schedule() <- schedule.Request[int]{Data: intPtr(1), TriggerIn: time.Second}
		c.Add(time.Second)

		select {
		case trigger := <-scheduler.Next():
			assert.Equal(t, 1, *trigger.Data, "expected the payload in the trigger to be 1")
			assert.Equal(t, parse("1970-01-01 00:00:01"), trigger.TriggeredAt, "expected the payload in the trigger to be 1")
		case <-time.After(time.Second):
			t.Errorf("Expected scheduler to trigger, but it did not.")
		}

	})
	t.Run("schedule request are only returned after specified duration", func(t *testing.T) {
		t.Parallel()

		c := clock.NewMock()
		c.Set(parse("1970-01-01 00:00:00"))
		cancel := make(chan struct{})
		defer close(cancel)
		var tmb tomb.Tomb

		scheduler := schedule.NewScheduler[int](c)
		tmb.Go(func() error {
			scheduler.Run(cancel)
			return nil
		})
		scheduler.Schedule() <- schedule.Request[int]{Data: intPtr(1), TriggerIn: 1 * time.Second}
		scheduler.Schedule() <- schedule.Request[int]{Data: intPtr(3), TriggerIn: 3 * time.Second}
		scheduler.Schedule() <- schedule.Request[int]{Data: intPtr(6), TriggerIn: 6 * time.Second}

		var actual []schedule.Trigger[int]
		for i := 0; i < 6; i++ {
			c.Add(time.Second)
			select {
			case trigger := <-scheduler.Next():
				actual = append(actual, trigger)
			default:
			}
		}
		expected := []schedule.Trigger[int]{
			{intPtr(1), parse("1970-01-01 00:00:01")},
			{intPtr(3), parse("1970-01-01 00:00:03")},
			{intPtr(6), parse("1970-01-01 00:00:06")},
		}
		assert.Equal(t, expected, actual, "received triggers have to match in value and time")
	})
}

func parse(timeStr string) time.Time {
	t, err := time.Parse(time.DateTime, timeStr)
	if err != nil {
		panic(err)
	}
	return t
}

func intPtr(i int) *int {
	return &i
}
