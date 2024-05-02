package sdphub

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

var (
	validate = validator.New()

	ErrOffererAlreadyExists = errors.New("offerer already exists")
	ErrOfferNotFound        = errors.New("offer not found")
	ErrAnswererNotFound     = errors.New("answerer not found")
	ErrAgreementNotFound    = errors.New("match not found")
	ErrInvalidAccessKey     = errors.New("invalid access key")
	ErrUnknownMessageType   = errors.New("unknown message type")
)

type Hub struct {
	reg *Registry
}

func NewHub() *Hub {
	return &Hub{
		reg: NewRegistry(),
	}
}

// Answerers lists all answerers
func (h *Hub) Answerers() []Answerer {
	return h.reg.Answerers()
}

// CreateAnswerer creates an answerer
func (h *Hub) CreateAnswerer(ctx MessageContext, req CreateAnswererRequest) error {
	a := Answerer{
		Name:          req.Name,
		Description:   req.Description,
		AccessKey:     req.AccessKey,
		ManagementKey: req.ManagementKey,
		Conn:          ctx.Conn,
		Address:       ctx.Addr,
		LastMessage:   ctx.MessageTime,
		ConvID:        ctx.ConvID,
	}

	replaced, err := h.reg.CreateAnswerer(a)
	if err != nil {
		return err
	}

	if replaced != nil && replaced.Conn != a.Conn {
		if err := replaced.Conn.Close(); err != nil {
			fmt.Println("error closing replaced registrants connection")
		}
	}

	SendInfo(ctx, "registration complete")

	go func() {
		err := h.handleOfferers(a)
		if err != nil {
			fmt.Println(err)
		}
	}()

	return nil
}

// SendOffer attempts to make an agreement with an answerer, it fails if the answerer
// does not exist or access key doesn't match.
func (h *Hub) SendOffer(ctx MessageContext, req CreateOffererRequest) error {
	a, ok := h.reg.Answerer(req.AnswererName)
	if !ok {
		return ErrAnswererNotFound
	}

	o := Offerer{
		Name:          req.Name,
		AnswererName:  req.AnswererName,
		AccessKey:     req.AccessKey,
		ManagementKey: req.ManagementKey,
		SDP:           req.SDP,
		Conn:          ctx.Conn,
		Address:       ctx.Addr,
		LastMessage:   ctx.MessageTime,
		ConvID:        ctx.ConvID,
	}

	return h.createAgreement(o, a)
}

// CreateOfferer attempts to make an agreement with an answerer, it fails if access key
// doesn't match and adds an offerer which will be answered as soon as matching answerer appears.
// In case of an access key mismatch with matching answerer error message will be sent.
func (h *Hub) CreateOfferer(ctx MessageContext, req CreateOffererRequest) error {
	o := Offerer{
		Name:          req.Name,
		AnswererName:  req.AnswererName,
		AccessKey:     req.AccessKey,
		ManagementKey: req.ManagementKey,
		SDP:           req.SDP,
		ConvID:        ctx.ConvID,
		Conn:          ctx.Conn,
		Address:       ctx.Addr,
		LastMessage:   ctx.MessageTime,
	}

	a, ok := h.reg.Answerer(req.AnswererName)
	if !ok {
		existingOfferer, ok := h.reg.Offerer(o.AnswererName, o.Name)
		if ok {
			if !existingOfferer.ManagementKeyMatches(o.ManagementKey) {
				return ErrOffererAlreadyExists
			}
			return h.reg.CreateOfferer(o)
		}

		h.reg.CreateOfferer(o)
		return nil
	}

	return h.createAgreement(o, a)
}

func (h *Hub) AcceptAgreement(ctx MessageContext, req AcceptAgreementRequest) error {
	agr, ok := h.reg.Agreement(req.AgreementID)
	if !ok {
		return ErrAgreementNotFound
	}

	//todo notify answerer about failures

	defer func() {
		h.reg.DeleteAgreement(agr.ID)
	}()

	return agr.AnswerToOfferer(req.SDP)
}

func (h *Hub) RejectAgreement(tx MessageContext, req RejectAgreementRequest) error {
	agr, ok := h.reg.Agreement(req.AgreementID)
	if !ok {
		return ErrAgreementNotFound
	}

	//todo notify answerer about failures

	defer func() {
		h.reg.DeleteAgreement(agr.ID)
	}()

	return SendError(MessageContext{
		Conn:   agr.Offerer.Conn,
		ConvID: agr.Offerer.ConvID,
	}, fmt.Errorf("agreement %s rejected: %s", agr.ID, req.Reason))
}

func (h *Hub) handleOfferers(a Answerer) error {
	offerers := h.reg.Offerers(a.Name)
	errs := []error{}
	for _, o := range offerers {
		err := h.createAgreement(o, a)
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

func (h *Hub) createAgreement(o Offerer, a Answerer) error {
	if !a.AccessKeyMatches(o.AccessKey) {
		return ErrInvalidAccessKey
	}

	agr := NewAgreement(o, a)
	h.reg.CreateAgreement(agr)
	return agr.OfferToAnswerer()
}
