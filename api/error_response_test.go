package api

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jjmaturino/bootstrapper/network"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSendErrorResponse(t *testing.T) {
	// Set Gin to Test Mode
	gin.SetMode(gin.TestMode)

	// Test cases
	tests := []struct {
		name           string
		httpStatus     int
		options        []func(*Response)
		expectedStatus int
		expectedBody   ErrorResponse
		expectedError  error
	}{
		{
			name:           "No options provided",
			httpStatus:     http.StatusInternalServerError,
			options:        nil,
			expectedStatus: http.StatusInternalServerError,
			expectedBody: ErrorResponse{
				Title:      http.StatusText(http.StatusInternalServerError),
				Details:    StandardErrorMessage,
				HttpStatus: http.StatusInternalServerError,
			},
			expectedError: errors.New(StandardErrorMessage),
		},
		{
			name:       "Option sets Error",
			httpStatus: http.StatusBadRequest,
			options: []func(*Response){
				func(r *Response) {
					r.Error = errors.New("invalid input")
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: ErrorResponse{
				Title:      http.StatusText(http.StatusBadRequest),
				Details:    "invalid input",
				HttpStatus: http.StatusBadRequest,
			},
			expectedError: errors.New("invalid input"),
		},
		{
			name:       "Option sets ErrorDetails",
			httpStatus: http.StatusNotFound,
			options: []func(*Response){
				func(r *Response) {
					r.ErrorDetails = map[string]any{
						"resource": "User",
						"id":       "123",
					}
				},
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: ErrorResponse{
				Title:      http.StatusText(http.StatusNotFound),
				Details:    "internal error [id:123;resource:User]",
				HttpStatus: http.StatusNotFound,
			},
			expectedError: errors.New(StandardErrorMessage),
		},
		{
			name:       "Options set Error and ErrorDetails",
			httpStatus: http.StatusUnauthorized,
			options: []func(*Response){
				func(r *Response) {
					r.Error = errors.New("unauthorized access")
				},
				func(r *Response) {
					r.ErrorDetails = map[string]any{
						"token": "expired",
					}
				},
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody: ErrorResponse{
				Title:      http.StatusText(http.StatusUnauthorized),
				Details:    "unauthorized access [token:expired]",
				HttpStatus: http.StatusUnauthorized,
			},
			expectedError: errors.New("unauthorized access"),
		},
		{
			name:       "Option modifies HttpStatus",
			httpStatus: http.StatusOK,
			options: []func(*Response){
				func(r *Response) {
					r.HttpStatus = http.StatusForbidden // Should not change the response status
					r.Error = errors.New("forbidden")
				},
			},
			expectedStatus: http.StatusForbidden, // Expected to remain as passed
			expectedBody: ErrorResponse{
				Title:      http.StatusText(http.StatusForbidden),
				Details:    "forbidden",
				HttpStatus: http.StatusForbidden,
			},
			expectedError: errors.New("forbidden"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a response recorder
			w := httptest.NewRecorder()
			// Create a new Gin context
			c, _ := gin.CreateTestContext(w)

			// Call the function under test
			err := SendErrorResponse(c, tt.httpStatus, tt.options...)

			// Assertions
			if tt.expectedError != nil {
				require.Equal(t, tt.expectedError.Error(), err.Error(), "Returned error should match expected error")
			}

			// Check the response status code
			assert.Equal(t, tt.expectedStatus, w.Code, "HTTP status code should match expected status")

			// Check the Content-Type header
			assert.Equal(t, JSONProblemContentType, w.Header().Get(ContentTypeHeader), "Content-Type header should be set correctly")

			// Parse the response body
			var responseBody ErrorResponse
			err = json.Unmarshal(w.Body.Bytes(), &responseBody)
			require.NoError(t, err, "Response body should be valid JSON")

			// Check the response body
			assert.Equal(t, tt.expectedBody, responseBody, "Response body should match expected body")
		})
	}
}

func TestSendNotFoundResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	err := SendNotFoundResponse(c)

	assert.NotNil(t, err)
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusText(http.StatusNotFound), response.Title)
	assert.Equal(t, http.StatusNotFound, response.HttpStatus)
	assert.Equal(t, ResourceNotFoundErrorMessage, response.Details)
}

func TestSendBadResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	err := SendBadResponse(c)

	assert.NotNil(t, err)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusText(http.StatusBadRequest), response.Title)
	assert.Equal(t, http.StatusBadRequest, response.HttpStatus)
	assert.Equal(t, BadRequestErrorMessage, response.Details)
}

func TestSendInternalServerError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	err := SendInternalServerError(c)

	assert.NotNil(t, err)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusText(http.StatusInternalServerError), response.Title)
	assert.Equal(t, http.StatusInternalServerError, response.HttpStatus)
	assert.Equal(t, StandardErrorMessage, response.Details)
}

// Unit tests

func TestSendWSErrorResponse(t *testing.T) {
	// Create a MockWebSocket
	mockWS := network.NewMockWebSocket()

	// Call the function under test
	err := SendWSErrorResponse(mockWS, http.StatusInternalServerError)

	// Assert that an error is returned
	assert.Error(t, err, "Expected an error to be returned")
	assert.Equal(t, StandardErrorMessage, err.Error())

	// Retrieve the messages sent through the mock WebSocket
	writes := mockWS.GetWrites()
	assert.Len(t, writes, 1, "Expected one message to be written")

	// write should be the error message
	write := writes[0]
	assert.Equal(t, websocket.TextMessage, write.MessageType, "Expected message type to be TextMessage")

	// Unmarshal the message
	var wsMessage WSMessage
	err = json.Unmarshal(write.Data, &wsMessage)
	assert.NoError(t, err, "Failed to unmarshal message")

	// Assert the message content
	assert.Equal(t, "error", wsMessage.Event, "Expected event to be 'error'")

	// Unmarshal the Data field into ErrorResponse
	dataBytes, err := json.Marshal(wsMessage.Data)
	assert.NoError(t, err, "Failed to marshal Data field")
	var errorResponse ErrorResponse
	err = json.Unmarshal(dataBytes, &errorResponse)
	assert.NoError(t, err, "Failed to unmarshal Data field to ErrorResponse")

	assert.Equal(t, "Internal Server Error", errorResponse.Title)
	assert.Equal(t, http.StatusInternalServerError, errorResponse.HttpStatus)
	assert.Equal(t, StandardErrorMessage, errorResponse.Details)
}

func TestSendWSErrorResponseWithWriteError(t *testing.T) {
	// Create a MockWebSocket
	mockWS := network.NewMockeryMockWebsocket(t)

	mockWS.On("WriteMessage", mock.Anything, mock.Anything).Return(errors.New("mock error"))

	// Call the function under test
	err := SendWSErrorResponse(mockWS, http.StatusInternalServerError)
	assert.Error(t, err, "Expected an error to be returned")
}

func TestSendWSNotFoundResponse(t *testing.T) {
	// Create a MockWebSocket
	mockWS := network.NewMockWebSocket()

	// Call the function under test
	err := SendWSNotFoundResponse(mockWS)

	// Assert that an error is returned
	assert.Error(t, err, "Expected an error to be returned")
	assert.Equal(t, ResourceNotFoundErrorMessage, err.Error())

	// Retrieve the messages sent through the mock WebSocket
	writes := mockWS.GetWrites()
	assert.Len(t, writes, 1, "Expected one message to be written")

	// write should be the error message
	write := writes[0]
	assert.Equal(t, websocket.TextMessage, write.MessageType, "Expected message type to be TextMessage")

	// Unmarshal the message
	var wsMessage WSMessage
	err = json.Unmarshal(write.Data, &wsMessage)
	assert.NoError(t, err, "Failed to unmarshal message")

	// Assert the message content
	assert.Equal(t, "error", wsMessage.Event, "Expected event to be 'error'")

	// Unmarshal the Data field into ErrorResponse
	dataBytes, err := json.Marshal(wsMessage.Data)
	assert.NoError(t, err, "Failed to marshal Data field")
	var errorResponse ErrorResponse
	err = json.Unmarshal(dataBytes, &errorResponse)
	assert.NoError(t, err, "Failed to unmarshal Data field to ErrorResponse")

	assert.Equal(t, "Not Found", errorResponse.Title)
	assert.Equal(t, http.StatusNotFound, errorResponse.HttpStatus)
	assert.Equal(t, ResourceNotFoundErrorMessage, errorResponse.Details)
}

func TestSendWSBadResponse(t *testing.T) {
	// Create a MockWebSocket
	mockWS := network.NewMockWebSocket()

	// Call the function under test
	err := SendWSBadResponse(mockWS)

	// Assert that an error is returned
	assert.Error(t, err, "Expected an error to be returned")
	assert.Equal(t, BadRequestErrorMessage, err.Error())

	// Retrieve the messages sent through the mock WebSocket
	writes := mockWS.GetWrites()
	assert.Len(t, writes, 1, "Expected one message to be written")

	// write should be the error message
	write := writes[0]
	assert.Equal(t, websocket.TextMessage, write.MessageType, "Expected message type to be TextMessage")

	// Unmarshal the message
	var wsMessage WSMessage
	err = json.Unmarshal(write.Data, &wsMessage)
	assert.NoError(t, err, "Failed to unmarshal message")

	// Assert the message content
	assert.Equal(t, "error", wsMessage.Event, "Expected event to be 'error'")

	// Unmarshal the Data field into ErrorResponse
	dataBytes, err := json.Marshal(wsMessage.Data)
	assert.NoError(t, err, "Failed to marshal Data field")
	var errorResponse ErrorResponse
	err = json.Unmarshal(dataBytes, &errorResponse)
	assert.NoError(t, err, "Failed to unmarshal Data field to ErrorResponse")

	assert.Equal(t, "Bad Request", errorResponse.Title)
	assert.Equal(t, http.StatusBadRequest, errorResponse.HttpStatus)
	assert.Equal(t, BadRequestErrorMessage, errorResponse.Details)
}

func TestSendWSInternalServerError(t *testing.T) {
	// Create a MockWebSocket
	mockWS := network.NewMockWebSocket()

	// Call the function under test
	err := SendWSInternalServerError(mockWS)

	// Assert that an error is returned
	assert.Error(t, err, "Expected an error to be returned")
	assert.Equal(t, StandardErrorMessage, err.Error())

	// Retrieve the messages sent through the mock WebSocket
	writes := mockWS.GetWrites()
	assert.Len(t, writes, 1, "Expected one message to be written")

	// write should be the error message
	write := writes[0]
	assert.Equal(t, websocket.TextMessage, write.MessageType, "Expected message type to be TextMessage")

	// Unmarshal the message
	var wsMessage WSMessage
	err = json.Unmarshal(write.Data, &wsMessage)
	assert.NoError(t, err, "Failed to unmarshal message")

	// Assert the message content
	assert.Equal(t, "error", wsMessage.Event, "Expected event to be 'error'")

	// Unmarshal the Data field into ErrorResponse
	dataBytes, err := json.Marshal(wsMessage.Data)
	assert.NoError(t, err, "Failed to marshal Data field")
	var errorResponse ErrorResponse
	err = json.Unmarshal(dataBytes, &errorResponse)
	assert.NoError(t, err, "Failed to unmarshal Data field to ErrorResponse")

	assert.Equal(t, "Internal Server Error", errorResponse.Title)
	assert.Equal(t, http.StatusInternalServerError, errorResponse.HttpStatus)
	assert.Equal(t, StandardErrorMessage, errorResponse.Details)
}

func TestSendWSInternalServerErrorAndClose(t *testing.T) {
	// Create a MockWebSocket
	mockWS := network.NewMockWebSocket()

	// Call the function under test
	err := SendWSInternalServerErrorAndClose(mockWS)

	// Since SendWSInternalServerErrorAndClose returns the error from SendWSInternalServerError
	// unless WriteControl or Close fail, we expect it to return StandardErrorMessage
	assert.Error(t, err, "Expected an error to be returned")
	assert.Equal(t, StandardErrorMessage, err.Error())

	// Retrieve the messages sent through the mock WebSocket
	writes := mockWS.GetWrites()
	assert.Len(t, writes, 2, "Expected two messages to be written")

	// First write should be the error message
	write1 := writes[0]
	assert.Equal(t, websocket.TextMessage, write1.MessageType, "Expected first message to be TextMessage")

	// Unmarshal the error message
	var wsMessage WSMessage
	err = json.Unmarshal(write1.Data, &wsMessage)
	assert.NoError(t, err, "Failed to unmarshal error message")

	// Assert the error message content
	assert.Equal(t, "error", wsMessage.Event, "Expected event to be 'error'")

	// Unmarshal the Data field into ErrorResponse
	dataBytes, err := json.Marshal(wsMessage.Data)
	assert.NoError(t, err, "Failed to marshal Data field")
	var errorResponse ErrorResponse
	err = json.Unmarshal(dataBytes, &errorResponse)
	assert.NoError(t, err, "Failed to unmarshal Data field to ErrorResponse")

	assert.Equal(t, "Internal Server Error", errorResponse.Title)
	assert.Equal(t, http.StatusInternalServerError, errorResponse.HttpStatus)
	assert.Equal(t, StandardErrorMessage, errorResponse.Details)

	// Second write should be the close control message
	write2 := writes[1]
	assert.Equal(t, websocket.CloseMessage, write2.MessageType, "Expected second message to be CloseMessage")

	// Check that the WebSocket was closed
	assert.True(t, mockWS.IsClosed(), "Expected the WebSocket to be closed")
}

func TestSendWSInternalServerErrorAndCloseWithWriteError(t *testing.T) {
	// Create a MockWebSocket
	mockWS := network.NewMockeryMockWebsocket(t)

	mockWS.On("WriteMessage", mock.Anything, mock.Anything).Return(nil)
	mockWS.On("WriteControl", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("mock error"))

	// Call the function under test
	err := SendWSInternalServerErrorAndClose(mockWS)
	assert.Error(t, err, "Expected an error to be returned")
}

func TestSendWSInternalServerErrorAndCloseWithCloseError(t *testing.T) {
	// Create a MockWebSocket
	mockWS := network.NewMockeryMockWebsocket(t)

	mockWS.On("WriteMessage", mock.Anything, mock.Anything).Return(nil)
	mockWS.On("WriteControl", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockWS.On("Close").Return(errors.New("mock error"))

	// Call the function under test
	err := SendWSInternalServerErrorAndClose(mockWS)
	assert.Error(t, err, "Expected an error to be returned")
}
