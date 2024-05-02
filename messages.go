package sdphub

import (
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type MessageType string

var (
	MessageTypeCreateAnswerer  MessageType = "create_answerer"
	MessageTypeFindAnswerer    MessageType = "find_answerer"
	MessageTypeCreateOfferer   MessageType = "create_offerer"
	MessageTypeAcceptAgreement MessageType = "accept_agreement"
	MessageTypeRejectAgreement MessageType = "reject_agreement"

	// Outgoing
	MessageTypeAgreementOffer  MessageType = "agreement_offer"
	MessageTypeAgreementAnswer MessageType = "agreement_answer"
	MessageTypeInfo            MessageType = "info"
	MessageTypeError           MessageType = "error"
)

type Message struct {
	Type   MessageType `json:"type"`
	ConvID string      `json:"convid"`
	Data   any         `json:"data"`
}

type MessageContext struct {
	Conn        *websocket.Conn
	Addr        string
	MessageTime time.Time
	ConvID      string
}

type InfoMessage struct {
	Message string `json:"message"`
}

type ErrorMessage struct {
	Error string `json:"error"`
}

func Send(ctx MessageContext, mt MessageType, data any) error {
	return ctx.Conn.WriteJSON(Message{
		ConvID: convID(ctx.ConvID),
		Type:   mt,
		Data:   data,
	})
}

func SendInfo(ctx MessageContext, msg string) error {
	return ctx.Conn.WriteJSON(Message{
		ConvID: convID(ctx.ConvID),
		Type:   MessageTypeInfo,
		Data: InfoMessage{
			Message: msg,
		},
	})
}

func SendError(ctx MessageContext, err error) error {
	return ctx.Conn.WriteJSON(Message{
		ConvID: convID(ctx.ConvID),
		Type:   MessageTypeError,
		Data: ErrorMessage{
			Error: err.Error(),
		},
	})
}

func convID(cID string) string {
	if cID != "" {
		return cID
	}
	return uuid.NewString()
}

// peer-sent messages
type CreateAnswererRequest struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	AccessKey     string `json:"accesskey"`
	ManagementKey string `json:"managementkey"`
}

type CreateOffererRequest struct {
	Name          string `json:"name"`
	AnswererName  string `json:"answerername"`
	AccessKey     string `json:"accesskey"`
	ManagementKey string `json:"managementkey"`
	SDP           string `json:"sdp"`
}

type AcceptAgreementRequest struct {
	AgreementID string `json:"agreementid"`
	SDP         string `json:"sdp"`
}

type RejectAgreementRequest struct {
	AgreementID string `json:"agreementid"`
	Reason      string `json:"reason"`
}

// hub-sent messages
type OfferMessage struct {
	AgreementID string `json:"agreementid"`
	Name        string `json:"offerername"`
	SDP         string `json:"sdp"`
}

type AnswerMessage struct {
	AgreementID string `json:"agreementid"`
	Name        string `json:"answerername"`
	SDP         string `json:"sdp"`
}
