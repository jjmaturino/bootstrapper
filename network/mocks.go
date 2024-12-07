package network

import (
	"sync"
	"time"
)

// MockWebSocket is a mock implementation of a websocket connection
type mockWebSocketMessage struct {
	MessageType int
	Data        []byte
	TimeStamp   time.Time
}

type MockWebSocket struct {
	Messages     []mockWebSocketMessage
	Writes       []mockWebSocketMessage
	TimesWritten int
	Errors       map[string]error
	Closed       bool
	mu           sync.Mutex
}

func NewMockWebSocket() *MockWebSocket {
	return &MockWebSocket{
		Messages:     []mockWebSocketMessage{},
		Writes:       []mockWebSocketMessage{},
		TimesWritten: 0,
		Errors:       map[string]error{},
		Closed:       false,
		mu:           sync.Mutex{},
	}
}

var _ Websocket = (*MockWebSocket)(nil)

func (m *MockWebSocket) ReadMessage() (messageType int, p []byte, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err, ok := m.Errors["ReadMessage"]; ok {
		return 0, nil, err
	}

	if len(m.Messages) == 0 {
		return 0, nil, nil
	}

	message := m.Messages[0]
	m.Messages = m.Messages[1:]
	return message.MessageType, message.Data, nil
}

func (m *MockWebSocket) WriteMessage(messageType int, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err, ok := m.Errors["WriteMessage"]; ok {
		return err
	}

	msg := mockWebSocketMessage{
		MessageType: messageType,
		Data:        data,
		TimeStamp:   time.Now(),
	}

	m.Messages = append(m.Messages, msg)
	m.Writes = append(m.Writes, msg)
	m.TimesWritten++

	return nil
}

func (m *MockWebSocket) SetWriteDeadline(t time.Time) error {
	return nil
}

func (m *MockWebSocket) WriteControl(messageType int, data []byte, deadline time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err, ok := m.Errors["WriteControl"]; ok {
		return err
	}

	// Simulate writing the control message
	msg := mockWebSocketMessage{
		MessageType: messageType,
		Data:        data,
		TimeStamp:   time.Now(),
	}

	m.Writes = append(m.Writes, msg)
	m.TimesWritten++

	return nil
}

func (m *MockWebSocket) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err, ok := m.Errors["Close"]; ok {
		return err
	}
	m.Closed = true
	return nil
}

// Helper methods for testing
func (m *MockWebSocket) SetError(method string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Errors[method] = err
}

func (m *MockWebSocket) ClearError(method string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.Errors, method)
}

func (m *MockWebSocket) IsClosed() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.Closed
}

func (m *MockWebSocket) GetWrites() []mockWebSocketMessage {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]mockWebSocketMessage(nil), m.Writes...)
}

func (m *MockWebSocket) MessagesReceivedSimultaneously(tolerance time.Duration) bool {
	if len(m.Messages) < 2 {
		return true
	}
	baseTime := m.Messages[0].TimeStamp
	for _, msg := range m.Messages[1:] {
		if msg.TimeStamp.Sub(baseTime) > tolerance {
			return false
		}
	}
	return true
}

func MessagesReceivedSimultaneously(tolerance time.Duration, websockets []*MockWebSocket) bool {
	if len(websockets) == 0 {
		return true // No websockets to compare
	}

	for i := 0; i < websockets[0].TimesWritten; i++ { // Iterate based on the times messages were written
		baseTime := websockets[0].Writes[i].TimeStamp
		for _, ws := range websockets {
			if i >= len(ws.Writes) || ws.Writes[i].TimeStamp.Sub(baseTime) > tolerance {
				return false
			}
		}
	}
	return true
}
