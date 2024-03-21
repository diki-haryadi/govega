package response

import (
	"errors"
	"net/http"
	"strconv"
)

var (
	ErrBadRequest          = errors.New("bad request")
	ErrForbiddenResource   = errors.New("forbidden resource")
	ErrNotFound            = errors.New("not found")
	ErrPreConditionFailed  = errors.New("precondition failed")
	ErrInternalServerError = errors.New("internal server error")
	ErrTimeoutError        = errors.New("timeout error")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrConflict            = errors.New("conflict")
)

const (
	StatusCodeGenericSuccess            = "200000"
	StatusCodeBadRequest                = "400000"
	StatusCodeAlreadyRegistered         = "400001"
	StatusCodeUnauthorized              = "401000"
	StatusCodeForbidden                 = "403000"
	StatusCodeNotFound                  = "404000"
	StatusCodeConflict                  = "409000"
	StatusCodeGenericPreconditionFailed = "412000"
	StatusCodeOTPLimitReached           = "412550"
	StatusCodeNoLinkerExist             = "412553"
	StatusCodeInternalError             = "500000"
	StatusCodeFailedSellBatch           = "500100"
	StatusCodeFailedOTP                 = "503000"
	StatusCodeServiceUnavailable        = "503000"
	StatusCodeTimeoutError              = "504000"
)

func GetErrorCode(err error) string {
	err = getErrType(err)

	switch err {
	case ErrBadRequest:
		return StatusCodeBadRequest
	case ErrForbiddenResource:
		return StatusCodeForbidden
	case ErrNotFound:
		return StatusCodeNotFound
	case ErrConflict:
		return StatusCodeConflict
	case ErrUnauthorized:
		return StatusCodeUnauthorized
	case ErrForbiddenResource:
		return StatusCodeForbidden
	case ErrPreConditionFailed:
		return StatusCodeGenericPreconditionFailed
	case ErrInternalServerError:
		return StatusCodeInternalError
	case ErrTimeoutError:
		return StatusCodeTimeoutError
	case nil:
		return StatusCodeGenericSuccess
	default:
		return StatusCodeInternalError
	}
}

func GetHTTPCode(code string) int {
	s := code[0:3]
	i, _ := strconv.Atoi(s)
	return i
}

const (
	StatusCtxKey                = 0
	StatusSuccess               = http.StatusOK
	StatusErrorForm             = http.StatusBadRequest
	StatusErrorUnknown          = http.StatusBadGateway
	StatusInternalError         = http.StatusInternalServerError
	StatusUnauthorized          = http.StatusUnauthorized
	StatusCreated               = http.StatusCreated
	StatusAccepted              = http.StatusAccepted
	StatusForbidden             = http.StatusForbidden
	StatusInvalidAuthentication = http.StatusProxyAuthRequired
	StatusNotFound              = http.StatusNotFound
)

var statusMap = map[int][]string{
	StatusSuccess:               {"STATUS_OK", "Success"},
	StatusErrorForm:             {"STATUS_BAD_REQUEST", "Invalid data request"},
	StatusErrorUnknown:          {"STATUS_BAD_GATEWAY", "Oops something went wrong"},
	StatusInternalError:         {"INTERNAL_SERVER_ERROR", "Oops something went wrong"},
	StatusUnauthorized:          {"STATUS_UNAUTHORIZED", "Not authorized to access the service"},
	StatusCreated:               {"STATUS_CREATED", "Resource has been created"},
	StatusAccepted:              {"STATUS_ACCEPTED", "Resource has been accepted"},
	StatusForbidden:             {"STATUS_FORBIDDEN", "Forbidden access the resource "},
	StatusInvalidAuthentication: {"STATUS_INVALID_AUTHENTICATION", "The resource owner or authorization server denied the request"},
	StatusNotFound:              {"STATUS_NOT_FOUND", "Not Found"},
}

func StatusCode(code int) string {
	return statusMap[code][0]
}

func StatusText(code int) string {
	return statusMap[code][1]
}
