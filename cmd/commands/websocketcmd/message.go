package websocketcmd

import (
	"encoding/json"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type messageType string

const (
	// Inbound
	messageTypeCreateOfferingPeer  messageType = "create_offering_peer"
	messageTypeCreateAnsweringPeer messageType = "create_answering_peer"

	messageTypeOfferAnswer        messageType = "offer_answer"
	messageTypeDealAnswerRejected messageType = "deal_answer_rejected"
	messageTypeDealAnswerError    messageType = "deal_answer_error"

	// Outbound
	messageOffer           messageType = "offer"
	messageTypeOfferFailed messageType = "offer_failed"

	messageTypeInfo  messageType = "info"
	messageTypeError messageType = "error"
)

type message struct {
	Type messageType `json:"type"`
	Conv string      `json:"conv"`
	Data any         `json:"data"`
}

func (m message) UnmarshalData(v any) error {
	bytes, err := json.Marshal(m.Data)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, v)
}

type writer struct {
	mu   sync.Mutex
	conn *websocket.Conn
	conv string
}

func newWriter(conn *websocket.Conn, conv string) *writer {
	return &writer{
		conn: conn,
		conv: conv,
	}
}

func (w *writer) Conv(c string) *writer {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.conv = c
	return w
}

func (w *writer) Write(mt messageType, data any) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	c := w.conv
	if c == "" {
		c = uuid.NewString()
	}
	msg := message{
		Type: mt,
		Conv: c,
		Data: data,
	}
	return w.conn.WriteJSON(msg)
}

func (w *writer) Info(msg string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	c := w.conv
	if c == "" {
		c = uuid.NewString()
	}
	m := message{
		Type: messageTypeInfo,
		Conv: c,
		Data: genericMessage{
			Message: msg,
		},
	}
	return w.conn.WriteJSON(m)
}

func (w *writer) Error(err error) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	c := w.conv
	if c == "" {
		c = uuid.NewString()
	}
	msg := message{
		Type: messageTypeError,
		Conv: c,
		Data: genericMessage{
			Message: err.Error(),
		},
	}
	return w.conn.WriteJSON(msg)
}

type genericMessage struct {
	Message string `json:"message"`
}
