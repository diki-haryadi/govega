package httprq

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
)

type Response struct {
	StatusCode int
	Body       *bytes.Buffer
	Error      error
}

var (
	ErrEmptyResponseBody = errors.New("Response body is empty")
)

func (r *Response) Consume(v interface{}) error {
	if r.Error != nil {
		return r.Error
	}

	if r.StatusCode < 200 || r.StatusCode > 299 {
		log.Println("statusCode", r.StatusCode)
		log.Println("body", r.Body)
		log.Println("Error when make request")

		body := ""
		if r.Body != nil {
			body = r.Body.String()
		}

		return fmt.Errorf("Response return status not OK, with status code %d, and body %s",
			r.StatusCode,
			body,
		)
	}

	if r.Body == nil {
		return ErrEmptyResponseBody
	}

	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		bodyBytes, _ := ioutil.ReadAll(r.Body)
		return fmt.Errorf("failed copying response body to interface, cause %s, responseBody %s",
			err.Error(),
			string(bodyBytes),
		)
	}

	return nil
}
