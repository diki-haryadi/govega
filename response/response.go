package response

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"

	"github.com/diki-haryadi/govega/custerr"
	"github.com/diki-haryadi/govega/log"
)

type JSONResponse struct {
	Data        interface{}            `json:"data,omitempty"`
	Message     string                 `json:"message,omitempty"`
	Code        string                 `json:"code"`
	StatusCode  int                    `json:"statusCode"`
	ErrorString string                 `json:"error,omitempty"`
	Error       error                  `json:"-"`
	RealError   string                 `json:"-"`
	Latency     string                 `json:"latency"`
	Log         map[string]interface{} `json:"-"`
	HTMLPage    bool                   `json:"-"`
	Result      interface{}            `json:"result,omitempty"`
}

func NewJSONResponse() *JSONResponse {
	return &JSONResponse{
		Code:       StatusCodeGenericSuccess,
		StatusCode: GetHTTPCode(StatusCodeGenericSuccess),
		Log:        map[string]interface{}{},
	}
}

func (r *JSONResponse) SetData(data interface{}) *JSONResponse {
	r.Data = data
	return r
}

func (r *JSONResponse) SetHTML() *JSONResponse {
	r.HTMLPage = true
	return r
}

func (r *JSONResponse) SetResult(result interface{}) *JSONResponse {
	r.Result = result
	return r
}

func (r *JSONResponse) SetMessage(msg string) *JSONResponse {
	r.Message = msg
	return r
}

func (r *JSONResponse) SetLatency(latency float64) *JSONResponse {
	r.Latency = fmt.Sprintf("%.2f ms", latency)
	return r
}

func (r *JSONResponse) SetLog(key string, val interface{}) *JSONResponse {
	_, file, no, _ := runtime.Caller(1)
	log.WithFields(log.Fields{
		"code":            r.Code,
		"err":             val,
		"function_caller": fmt.Sprintf("file %v line no %v", file, no),
	}).Errorln("Error API")
	r.Log[key] = val
	return r
}

func getErrType(err error) error {
	switch err.(type) {
	case custerr.ErrChain:
		errType := err.(custerr.ErrChain).Type
		if errType != nil {
			err = errType
		}
	}
	return err
}

func (r *JSONResponse) SetError(err error, a ...string) *JSONResponse {
	r.Code = GetErrorCode(err)
	r.SetLog("error", err)
	r.RealError = fmt.Sprintf("%+v", err)
	err = getErrType(err)
	r.Error = err
	r.ErrorString = err.Error()
	r.StatusCode = GetHTTPCode(r.Code)

	if r.StatusCode == http.StatusInternalServerError {
		r.ErrorString = "Internal Server error"
	}
	if len(a) > 0 {
		r.ErrorString = a[0]
	}
	return r
}

func (r *JSONResponse) GetBody() []byte {
	b, _ := json.Marshal(r)
	return b
}

func (r *JSONResponse) Send(w http.ResponseWriter) {
	if r.HTMLPage {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(r.StatusCode)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(r.StatusCode)
		err := json.NewEncoder(w).Encode(r)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err.Error(),
			}).Errorln("[JSONResponse] Error encoding response")
		}
	}
}

// APIStatusSuccess for standard request api status success
func (r *JSONResponse) APIStatusSuccess() *JSONResponse {
	r.Code = StatusCode(StatusSuccess)
	r.Message = StatusText(StatusSuccess)
	return r
}

// APIStatusCreated
func (r *JSONResponse) APIStatusCreated() *JSONResponse {
	r.StatusCode = StatusCreated
	r.Code = StatusCode(StatusCreated)
	r.Message = StatusText(StatusCreated)
	return r
}

// APIStatusAccepted
func (r *JSONResponse) APIStatusAccepted() *JSONResponse {
	r.StatusCode = StatusAccepted
	r.Code = StatusCode(StatusAccepted)
	r.Message = StatusText(StatusAccepted)
	return r
}

// APIStatusErrorUnknown
func (r *JSONResponse) APIStatusErrorUnknown() *JSONResponse {
	r.StatusCode = StatusErrorUnknown
	r.Code = StatusCode(StatusErrorUnknown)
	r.Message = StatusText(StatusErrorUnknown)
	return r
}

// APIStatusInvalidAuthentication
func (r *JSONResponse) APIStatusInvalidAuthentication() *JSONResponse {
	r.StatusCode = StatusInvalidAuthentication
	r.Code = StatusCode(StatusInvalidAuthentication)
	r.Message = StatusText(StatusInvalidAuthentication)
	return r
}

// APIStatusUnauthorized
func (r *JSONResponse) APIStatusUnauthorized() *JSONResponse {
	r.StatusCode = StatusUnauthorized
	r.Code = StatusCode(StatusUnauthorized)
	r.Message = StatusText(StatusUnauthorized)
	return r
}

// APIStatusForbidden
func (r *JSONResponse) APIStatusForbidden() *JSONResponse {
	r.StatusCode = StatusForbidden
	r.Code = StatusCode(StatusForbidden)
	r.Message = StatusText(StatusForbidden)
	return r
}

// APIStatusBadRequest
func (r *JSONResponse) APIStatusBadRequest() *JSONResponse {
	r.StatusCode = StatusErrorForm
	r.Code = StatusCode(StatusErrorForm)
	r.Message = StatusText(StatusErrorForm)
	return r
}

// APIStatusNotFound
func (r *JSONResponse) APIStatusNotFound() *JSONResponse {
	r.StatusCode = StatusNotFound
	r.Code = StatusCode(StatusNotFound)
	r.Message = StatusText(StatusNotFound)
	return r
}
