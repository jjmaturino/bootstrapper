package api

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	"github.com/jjmaturino/bootstrapper/network"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	StandardErrorMessage         = "internal error"
	ResourceNotFoundErrorMessage = "resource not found"
	BadRequestErrorMessage       = "invalid request"
)

// Error RFC 7807 standard Problem Details for HTTP APIs.
// See https://www.rfc-editor.org/rfc/rfc7807.txt for details.

// ErrorResponse defines the output to the RFC 7807 format.
type ErrorResponse struct {

	// TypeUri is an absolute URI that identifies the problem type.  When de-referenced, it SHOULD provide
	// human-readable documentation for the type of problem (e.g., using HTML). This description is
	// not meant to be displayed to an application user; it is meant to help application developers.
	TypeUri url.URL `json:"type,omitempty"`

	// Title is a short, summary of the problem type. Written in English and readable by engineers
	// (usually not suited for non-technical stakeholders and not localized)
	Title string `json:"title"`

	// Status The HTTP status code generated by the origin server for this occurrence of the problem.
	HttpStatus int `json:"status"`

	// Detail is a human-readable explanation specific to this occurrence of the problem. This will
	// eventually be localised to respect the request's `Accept-Language` header, but currently it
	// is in English only.
	// example: `The startDate must be before the endDate`
	Details string `json:"detail,omitempty"`

	// Instance is an absolute URI that identifies the specific occurrence of the problem. It may or may not
	// yield further information if de-referenced.
	Instance *url.URL `json:"instance,omitempty"`

	// Note: Extensions to RFC 7807 are not controlled by any standards-setting organisation. If we need to
	//       add extra error information to any of the API responses it should be placed under a separate
	//       field to minimise the possibility of collision with any other API using 7807 to represent
	//       error results. There is no safe way to express that requirement here, given that the exact
	//       type of the extra error information is API specific.
}

func SendErrorResponse(c *gin.Context, httpStatus int, options ...func(*Response)) error {
	response := &Response{
		HttpStatus: httpStatus,
	}
	for _, o := range options {
		o(response)
	}

	if response.Error == nil {
		response.Error = errors.New(StandardErrorMessage)
	}

	c.JSON(response.HttpStatus, ErrorResponse{
		Title:      http.StatusText(response.HttpStatus),
		Details:    response.GetErrorDetails(),
		HttpStatus: response.HttpStatus,
	})

	c.Header(ContentTypeHeader, JSONProblemContentType)

	return response.Error
}

func SendNotFoundResponse(c *gin.Context, options ...func(*Response)) error {
	newOpts := func(r *Response) {
		r.Error = errors.New(ResourceNotFoundErrorMessage)
	}
	options = append(options, newOpts)

	return SendErrorResponse(c, http.StatusNotFound, options...)
}

func SendBadResponse(c *gin.Context, options ...func(*Response)) error {
	newOpts := func(r *Response) {
		r.Error = errors.New(BadRequestErrorMessage)
	}
	options = append(options, newOpts)

	return SendErrorResponse(c, http.StatusBadRequest, options...)
}

func SendInternalServerError(c *gin.Context, options ...func(*Response)) error {
	newOpts := func(r *Response) {
		r.Error = errors.New(StandardErrorMessage)
	}
	options = append(options, newOpts)

	return SendErrorResponse(c, http.StatusInternalServerError, options...)
}

// WebSocket Error Responses

func SendWSErrorResponse(ws network.Websocket, httpStatus int, options ...func(*Response)) error {
	response := &Response{
		HttpStatus: httpStatus,
	}
	for _, o := range options {
		o(response)
	}

	if response.Error == nil {
		response.Error = errors.New(StandardErrorMessage)
	}

	errorResp := ErrorResponse{
		Title:      http.StatusText(response.HttpStatus),
		Details:    response.GetErrorDetails(),
		HttpStatus: response.HttpStatus,
	}

	wsErrorMessage := WSMessage{
		Event: "error",
		Data:  errorResp,
	}

	message, _ := json.Marshal(wsErrorMessage)

	err := ws.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		return err
	}

	return response.Error
}

func SendWSNotFoundResponse(ws network.Websocket, options ...func(*Response)) error {
	newOpts := func(r *Response) {
		r.Error = errors.New(ResourceNotFoundErrorMessage)
	}
	options = append(options, newOpts)

	return SendWSErrorResponse(ws, http.StatusNotFound, options...)
}

func SendWSBadResponse(ws network.Websocket, options ...func(*Response)) error {
	newOpts := func(r *Response) {
		r.Error = errors.New(BadRequestErrorMessage)
	}
	options = append(options, newOpts)

	return SendWSErrorResponse(ws, http.StatusBadRequest, options...)
}

func SendWSInternalServerError(ws network.Websocket, options ...func(*Response)) error {
	newOpts := func(r *Response) {
		r.Error = errors.New(StandardErrorMessage)
	}
	options = append(options, newOpts)

	return SendWSErrorResponse(ws, http.StatusInternalServerError, options...)
}

func SendWSInternalServerErrorAndClose(ws network.Websocket, options ...func(*Response)) error {
	var sendErr error
	var closeErr error

	// Attempt to send the internal server error message.
	sendErr = SendWSInternalServerError(ws, options...)

	// Attempt to send a close frame regardless of sendErr.
	closeMessage := websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "Internal Server Error")
	closeErr = ws.WriteControl(websocket.CloseMessage, closeMessage, time.Now().Add(time.Second))
	if closeErr != nil {
		return closeErr
	}

	// Attempt to close the connection.
	closeErr = ws.Close()
	if closeErr != nil {
		return closeErr
	}

	return sendErr
}
