package peerhub

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrOfferNotFound  = errors.New("offer not found")
	ErrAnswerNotFound = errors.New("answer not found")
)

type SignalService interface {
	CreateOffer(Offer) error
	GetOffer(offerID string) (Offer, error)
	DeleteOffer(offerID string) error

	CreateAnswer(Answer) error
	GetAnswer(answerID string) (Answer, error)
	DeleteAnswer(answerID string) error
}

type Offer struct {
	ID            string `json:"id"`
	OfferingPeer  string `json:"offeringpeer"`
	AnsweringPeer string `json:"answeringpeer"`
	SDP           string `json:"sdp"`
}

func NewOffer(opName, sdp, apName string) Offer {
	return Offer{
		ID:            uuid.NewString(),
		OfferingPeer:  opName,
		AnsweringPeer: apName,
		SDP:           sdp,
	}
}

type Answer struct {
	ID            string `json:"id"`
	OfferID       string `json:"offerid"`
	AnsweringPeer string `json:"answeringpeer"`
	SDP           string `json:"sdp"`
}

func NewAnswer(offerID, apName, apSDP string) Answer {
	return Answer{
		ID:            uuid.NewString(),
		OfferID:       offerID,
		AnsweringPeer: apName,
		SDP:           apSDP,
	}
}

type FailedOffer struct {
	OfferingPeer  string `json:"offeringpeer"`
	AnsweringPeer string `json:"answeringpeer"`
	Error         error  `json:"error"`
}
