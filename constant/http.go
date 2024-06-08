package constant

import (
	"net/http"
)

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
