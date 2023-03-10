package schedule

type ImmediatelyScheduler[T any] interface {
	Scheduler[T]
}

type immediatelyScheduler[T any] struct {
	scheduleRequests chan Request[T]
	next             chan Trigger[T]
}

var _ ImmediatelyScheduler[int] = (*immediatelyScheduler[int])(nil)

func NewImmediatelyScheduler[T any](scheduleRequests chan Request[T], next chan Trigger[T]) ImmediatelyScheduler[T] {
	return &immediatelyScheduler[T]{
		scheduleRequests: scheduleRequests,
		next:             next,
	}
}

func (s *immediatelyScheduler[T]) Schedule() chan<- Request[T] {
	return s.scheduleRequests
}

func (s *immediatelyScheduler[T]) Next() <-chan Trigger[T] {
	return s.next
}

func (s *immediatelyScheduler[T]) Run(ignored <-chan struct{}) {
	// nop
}
