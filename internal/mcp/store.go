package mcp

import "sync"

// Message represents a message between agents
type Message struct {
	ID        string `json:"id"`
	From      string `json:"from"`
	To        string `json:"to"`
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"`
}

// MessageOutput is the common output for message operations
type MessageOutput struct {
	Success   bool     `json:"success"`
	MessageID string   `json:"message_id,omitempty"`
	Message   *Message `json:"message,omitempty"`
	Error     string   `json:"error,omitempty"`
	Hint      string   `json:"hint,omitempty"`
}

// MessageStore manages messages between agents
type MessageStore struct {
	mu       sync.RWMutex
	messages map[string]*Message      // key: message ID (for ack/removal)
	channels map[string]chan *Message // key: agent name (receiver)
}

func NewMessageStore() *MessageStore {
	return &MessageStore{
		messages: make(map[string]*Message),
		channels: make(map[string]chan *Message),
	}
}

func (s *MessageStore) StoreMessage(msg *Message) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages[msg.ID] = msg
}

func (s *MessageStore) DeleteMessage(msgID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.messages, msgID)
}

func (s *MessageStore) MessageExists(msgID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.messages[msgID]
	return exists
}

func (s *MessageStore) GetOrCreateChannel(agentName string) chan *Message {
	s.mu.Lock()
	defer s.mu.Unlock()

	ch, exists := s.channels[agentName]
	if !exists {
		ch = make(chan *Message, 100)
		s.channels[agentName] = ch
	}
	return ch
}
