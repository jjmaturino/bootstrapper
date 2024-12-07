package api

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Test function for GetErrorDetails
func TestGetErrorDetails(t *testing.T) {
	tests := []struct {
		name     string
		response Response
		expected string
	}{
		{
			name: "No error and no error details",
			response: Response{
				Error:        nil,
				ErrorDetails: map[string]any{},
			},
			expected: "",
		},
		{
			name: "Error but no error details",
			response: Response{
				Error:        errors.New("An error occurred"),
				ErrorDetails: map[string]any{},
			},
			expected: "An error occurred",
		},
		{
			name: "No error but has error details",
			response: Response{
				Error: nil,
				ErrorDetails: map[string]any{
					"code":    "404",
					"message": "Not Found",
				},
			},
			expected: " [code:404;message:Not Found]",
		},
		{
			name: "Error and error details",
			response: Response{
				Error: errors.New("An error occurred"),
				ErrorDetails: map[string]any{
					"code":    "500",
					"message": "Internal Server Error",
				},
			},
			expected: "An error occurred [code:500;message:Internal Server Error]",
		},
		{
			name: "Error and multiple error details",
			response: Response{
				Error: errors.New("Validation failed"),
				ErrorDetails: map[string]any{
					"field":   "username",
					"issue":   "required",
					"attempt": "first",
				},
			},
			expected: "Validation failed [attempt:first;field:username;issue:required]",
		},
		{
			name: "No error but empty error details",
			response: Response{
				Error:        nil,
				ErrorDetails: map[string]any{},
			},
			expected: "",
		},
		{
			name: "Error with empty error details",
			response: Response{
				Error:        errors.New("An error occurred"),
				ErrorDetails: map[string]any{},
			},
			expected: "An error occurred",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.response.GetErrorDetails()
			if actual != tt.expected {
				t.Errorf("Test %q failed: expected %q, got %q", tt.name, tt.expected, actual)
			}
		})
	}
}

func TestSendOKResponse_NoOptions(t *testing.T) {
	// Set up Gin in test mode
	gin.SetMode(gin.TestMode)

	// Create a response recorder
	w := httptest.NewRecorder()
	// Create a Gin context with the response recorder
	c, _ := gin.CreateTestContext(w)

	// Call the function under test
	SendOKResponse(c)

	// Assert that the status code is 200 OK
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// Assert that the response body is empty
	if w.Body.Len() != 0 {
		t.Errorf("Expected empty body, got %s", w.Body.String())
	}
}

func TestSendOKResponse_WithContents(t *testing.T) {
	// Set up Gin in test mode
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Define contents
	contents := map[string]string{"message": "Hello, world!"}

	// Call the function with contents
	SendOKResponse(c, WithContents(contents))

	// Assert status code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// Assert response body
	expectedBody, _ := json.Marshal(contents)
	if w.Body.String() != string(expectedBody) {
		t.Errorf("Expected body %s, got %s", string(expectedBody), w.Body.String())
	}
}

func TestSendSuccessfulResponse_CustomStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	statusCode := http.StatusCreated // 201

	SendSuccessfulResponse(c, statusCode)

	if w.Code != statusCode {
		t.Errorf("Expected status code %d, got %d", statusCode, w.Code)
	}

	if w.Body.Len() != 0 {
		t.Errorf("Expected empty body, got %s", w.Body.String())
	}
}

func TestSendSuccessfulResponse_WithLocationHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	location := "http://example.com/resource/1"
	option := func(r *Response) {
		r.LocationHeader = location
	}

	SendSuccessfulResponse(c, http.StatusCreated, option)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
	}

	if w.Header().Get(LocationHeader) != location {
		t.Errorf("Expected Location header %s, got %s", location, w.Header().Get(LocationHeader))
	}

	if w.Body.Len() != 0 {
		t.Errorf("Expected empty body, got %s", w.Body.String())
	}
}

func TestSendSuccessfulResponse_NoContents(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	SendNoContentResponse(c)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status code %d, got %d", http.StatusNoContent, w.Code)
	}

	if w.Body.Len() != 0 {
		t.Errorf("Expected empty body, got %s", w.Body.String())
	}
}

func TestSendSuccessfulResponse_WithContents(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	contents := map[string]string{"status": "success"}

	SendSuccessfulResponse(c, http.StatusOK, WithContents(contents))

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	expectedBody, _ := json.Marshal(contents)
	if w.Body.String() != string(expectedBody) {
		t.Errorf("Expected body %s, got %s", string(expectedBody), w.Body.String())
	}
}
