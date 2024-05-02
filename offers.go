package sdphub

import (
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Agreement struct {
	ID       string
	Offerer  Offerer
	Answerer Answerer
}

func NewAgreement(offerer Offerer, answerer Answerer) Agreement {
	return Agreement{
		ID:       uuid.NewString(),
		Offerer:  offerer,
		Answerer: answerer,
	}
}

func (a *Agreement) OfferToAnswerer() error {
	ctx := MessageContext{
		Conn:   a.Answerer.Conn,
		ConvID: a.Answerer.ConvID,
	}
	return Send(ctx, MessageTypeAgreementOffer, OfferMessage{
		AgreementID: a.ID,
		Name:        a.Offerer.Name,
		SDP:         a.Offerer.SDP,
	})
}

func (a *Agreement) AnswerToOfferer(sdp string) error {
	ctx := MessageContext{
		Conn:   a.Offerer.Conn,
		ConvID: a.Offerer.ConvID,
	}
	return Send(ctx, MessageTypeAgreementAnswer, AnswerMessage{
		AgreementID: a.ID,
		Name:        a.Answerer.Name,
		SDP:         sdp,
	})
}

type Offerer struct {
	Name          string
	AnswererName  string
	AccessKey     string
	ManagementKey string
	SDP           string
	ConvID        string

	Conn        *websocket.Conn
	Address     string
	LastMessage time.Time
}

func (o *Offerer) ManagementKeyMatches(key string) bool {
	return o.ManagementKey == "" || (o.ManagementKey == key)
}

type Answerer struct {
	Name          string
	Description   string
	AccessKey     string
	ManagementKey string
	ConvID        string

	Conn        *websocket.Conn
	Address     string
	LastMessage time.Time
}

func (r *Answerer) ManagementKeyMatches(key string) bool {
	return r.ManagementKey == "" || (r.ManagementKey == key)
}

func (r *Answerer) AccessKeyMatches(key string) bool {
	return r.AccessKey == "" || (r.AccessKey == key)
}
