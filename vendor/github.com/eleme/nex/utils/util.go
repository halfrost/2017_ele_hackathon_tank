package utils

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/apache/thrift/lib/go/thrift"
	json "github.com/json-iterator/go"
)

// IsExist check filepath is exists.
func IsExist(filepath string) bool {
	_, err := os.Stat(filepath)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

// Must make error panic.
func Must(err error, ctxinfo ...interface{}) {
	if err == nil {
		return
	}
	if len(ctxinfo) > 0 {
		info := []string{}
		for _, a := range ctxinfo { // XXX: fmt.Sprint is not good enough..
			info = append(info, fmt.Sprintf("%v", a))
		}
		panic(fmt.Errorf("%v: %+v", strings.Join(info, " "), err))
	} else {
		panic(err)
	}
}

// MarshalThriftError is used to marshal the Thrift error, since Thrift struct may
// has pointer field and the native format cannot pretty print the error message.
func MarshalThriftError(err error) string {
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	if err1 := enc.Encode(err); err1 != nil {
		return err.Error()
	}
	return strings.Trim(string(buf.Bytes()), " \n")
}

var defaultTApplicationExceptionMessage = map[int32]string{
	thrift.UNKNOWN_APPLICATION_EXCEPTION:  "unknown application exception",
	thrift.UNKNOWN_METHOD:                 "unknown method",
	thrift.INVALID_MESSAGE_TYPE_EXCEPTION: "invalid message type",
	thrift.WRONG_METHOD_NAME:              "wrong method name",
	thrift.BAD_SEQUENCE_ID:                "bad sequence ID",
	thrift.MISSING_RESULT:                 "missing result",
	thrift.INTERNAL_ERROR:                 "unknown internal error",
	thrift.PROTOCOL_ERROR:                 "unknown protocol error",
}

// DefaultThriftErrorMessage returns the default error message for thrift.TApplicationException,
// since some implementations may not set the 1:message field, such as thriftpy.
func DefaultThriftErrorMessage(err error) string {
	terr, ok := err.(thrift.TApplicationException)
	if !ok {
		return err.Error()
	}
	if msg := terr.Error(); msg != "" {
		return msg
	}
	return defaultTApplicationExceptionMessage[terr.TypeId()]
}
