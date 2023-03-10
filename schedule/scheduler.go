package schedule

import (
	"github.com/benbjohnson/clock"
	"time"
)

// A Scheduler accepts schedule requests with a payload and a duration after which the payload is returned again. Use the Schedule channel to submit schedule requests. Listen to the Next channel for due triggers. You have to call Run for the scheduler to process inputs and send outputs. Run() does only return when cancelled, call it in a go routine.
type Scheduler[T any] interface {
	// Run starts the scheduler. It only returns when the cancel channel is closed. Call Run in a go routine.
	Run(cancel <-chan struct{})

	// Schedule returns a channel which accepts schedule requests.
	Schedule() chan<- Request[T]

	// Next returns a channel, which sends the submitted payloads after they are due.
	Next() <-chan Trigger[T]
}

// A Request represents a schedule request. It holds a payload and a duration after which the payload should be returned again.
type Request[T any] struct {
	// Data is the payload which will be returned when the duration passed.
	Data *T

	// TriggerIn is the duration after which the payload will be returned.
	TriggerIn time.Duration
}

// A Trigger is issued, when a payload is due to be returned. It contains the payload itself, and the time when it was actually triggered.
type Trigger[T any] struct {
	// Submitted schedule payload.
	Data *T

	// Time when this trigger triggered.
	TriggeredAt time.Time
}

type scheduledItem[T any] struct {
	data      *T
	triggerAt time.Time
}

type scheduler[T any] struct {
	in    chan Request[T]
	next  chan Trigger[T]
	clock clock.Clock
	items []*scheduledItem[T]
}

// Interface implementation check.
var _ Scheduler[int64] = (*scheduler[int64])(nil)

// NewScheduler creates a new Scheduler using the given clock.
func NewScheduler[T any](clock clock.Clock) Scheduler[T] {
	sched := scheduler[T]{
		in:    make(chan Request[T]),
		next:  make(chan Trigger[T]),
		clock: clock,
	}
	return &sched
}

func (s *scheduler[T]) Schedule() chan<- Request[T] {
	return s.in
}

func (s *scheduler[T]) Next() <-chan Trigger[T] {
	return s.next
}

func (s *scheduler[T]) Run(cancel <-chan struct{}) {
	for !cancelled(cancel) {
		item, timer := s.getNextItem()
		select {
		case triggeredAt := <-timer.c():
			s.next <- Trigger[T]{item.data, triggeredAt}
			s.removeItem(item)
		case scheduleRequest := <-s.in:
			s.addItem(scheduleRequest)
		case <-cancel:
		}
		timer.stop()
	}
}

// getNextItem returns the item from items[] with the smallest trigger time, and a timer which will trigger after the item trigger time passes. The returned item is not removed from the items[] array.
func (s *scheduler[T]) getNextItem() (*scheduledItem[T], timer) {
	items := s.items
	if len(items) == 0 {
		// When there are no items to select from, we use a trick. We return a fake item, and a timer which never triggers. This way calling code can use getNextItem() without any nil checks.
		return dummyItem[T](), newForEverTimer()
	}
	nextItem := items[0]
	for i := range items {
		if items[i].triggerAt.Before(nextItem.triggerAt) {
			nextItem = items[i]
		}
	}
	return nextItem, newClockTimer(s.clock, nextItem.triggerAt)
}

func (s *scheduler[T]) addItem(request Request[T]) {
	item := scheduledItem[T]{data: request.Data, triggerAt: s.clock.Now().Add(request.TriggerIn)}
	s.items = append(s.items, &item)
}

func (s *scheduler[T]) removeItem(item *scheduledItem[T]) {
	for i := range s.items {
		if s.items[i] == item {
			s.items = removeIdx(s.items, i)
			break
		}
	}
}

func dummyItem[T any]() *scheduledItem[T] {
	tmpItem := scheduledItem[T]{}
	return &tmpItem
}

func removeIdx[T any](items []*scheduledItem[T], i int) []*scheduledItem[T] {
	// If `i` is the last item.
	if len(items) == i+1 {
		return items[0:i]
	}
	return append(items[0:i], items[i+1:]...)
}

func cancelled(cancel <-chan struct{}) bool {
	var ok bool
	select {
	case _, ok = <-cancel:
	default:
		return false
	}
	return !ok
}
