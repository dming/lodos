package baserpc

import (
	"fmt"
)

const (
	ClientError           = -1000
	ClientParamNotAdapted = -1
	ClientFormatError     = -2
	//ClientSuccess = -3
	//client independence
	ClientUnAuthenticated = -100

	ServerError = 1000
	ServerParamMiss = -ClientParamNotAdapted
	ServerFormatError = -ClientFormatError
	//ServerSuccess = -ClientSuccess
	//server independence
	ServerDBError = 100
	ServerRpcInvokeError = 101
	ServerExpired = 102
)

var (
	// ErrParamsMiss is returned when client miss some params to invoke rpc
	ErrClientParamsMiss = &Error{Code: ClientParamNotAdapted, Reason: "client miss some params to invoke rpc"}

	// ErrParamsMiss is returned when client miss some params to invoke rpc
	ErrClientFormatError = &Error{Code: ClientFormatError, Reason: "client params format error"}

	// ErrParamsMiss is returned when client miss some params to invoke rpc
	ErrClientUnAuthenticated = &Error{Code: ClientUnAuthenticated, Reason: "client is unauthenticated"}

	// ErrParamsMiss is returned when client miss some params to invoke rpc
	ErrServerParamMiss = &Error{Code: ServerParamMiss, Reason: "server miss some params to invoke rpc"}

	// ErrParamsMiss is returned when client miss some params to invoke rpc
	ErrServerFormatError = &Error{Code: ServerFormatError, Reason: "server params format error"}

	// ErrParamsMiss is returned when client miss some params to invoke rpc
	ErrServerDBError = &Error{Code: ServerDBError, Reason: "server db error"}

)

// Error captures the code and reason a channel or connection has been closed
// by the server.
type Error struct {
	Code    int    // constant code from the specification
	Reason  string // description of the error
	//Server  bool   // true when initiated from the server, false when from this library
	//Recover bool   // true when this error can be recovered by retrying later or with different parameters
}

func NewError(code int, err interface{}) *Error {
	var errStr string
	switch e := err.(type) {
	case error:
		errStr = e.Error()
	case string:
		errStr = e
	default:
		errStr = "unknow error"
	}
	return &Error{
		Code:    code,
		Reason:  errStr,
		//Recover: isSoftExceptionCode(int(code)),
		//Server:  true,
	}
}

func (e Error) Error() string {
	return fmt.Sprintf("Exception (%d) Reason: %q", e.Code, e.Reason)
}