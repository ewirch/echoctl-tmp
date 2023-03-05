package can

import "echoctl/flowcontrol"

type sendBufferFullError struct {
}

var _ flowcontrol.ShouldRetry = sendBufferFullError{}
var _ error = sendBufferFullError{}

func (s sendBufferFullError) ShouldRetry() bool {
	return true
}

func (s sendBufferFullError) Error() string {
	return "socket send buffer full"
}
