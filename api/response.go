package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"sort"
)

type Response struct {
	LocationHeader string
	HttpStatus     int
	Contents       interface{}
	ErrorDetails   map[string]any
	Error          error
}

func (r Response) GetErrorDetails() string {
	var details string
	if r.Error != nil {
		details = r.Error.Error()
	}

	if len(r.ErrorDetails) > 0 {
		keys := make([]string, 0, len(r.ErrorDetails))
		for k := range r.ErrorDetails {
			keys = append(keys, k)
		}

		sort.Strings(keys)
		first := true

		for _, name := range keys {
			value := r.ErrorDetails[name]

			prefix := ";"
			if first {
				prefix = " ["
			}

			details += fmt.Sprintf("%s%s:%s", prefix, name, value)
			first = false
		}
		if !first {
			details += "]"
		}
	}

	return details
}

func SendOKResponse(c *gin.Context, options ...func(*Response)) {
	SendSuccessfulResponse(c, http.StatusOK, options...)
}

func SendNoContentResponse(c *gin.Context) {
	SendSuccessfulResponse(c, http.StatusNoContent)
}

func SendSuccessfulResponse(c *gin.Context, httpStatus int, options ...func(*Response)) {
	response := &Response{
		HttpStatus: httpStatus,
	}

	for _, o := range options {
		o(response)
	}

	if response.LocationHeader != "" {
		c.Header(LocationHeader, response.LocationHeader)
	}

	if response.Contents == nil {
		c.Status(response.HttpStatus)
		c.Writer.WriteHeaderNow() // Ensure headers are flushed
		return
	}
	c.JSON(response.HttpStatus, response.Contents)
}

func WithContents(contents interface{}) func(*Response) {
	return func(r *Response) {
		r.Contents = contents
	}
}

// WSMessage defines the structure of WebSocket messages.
type WSMessage struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}
