package flowcontrol

type CanSkip interface {
	CanSkip() bool
}

func IsCanSkip(err error) bool {
	skip, ok := err.(CanSkip)
	return ok && skip.CanSkip()
}

type ShouldLog interface {
	ShouldLog() bool
}

func IsShouldLog(err error) bool {
	shouldLog, ok := err.(ShouldLog)
	return ok && shouldLog.ShouldLog()
}
