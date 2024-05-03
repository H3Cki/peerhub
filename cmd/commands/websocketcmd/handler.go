package websocketcmd

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/H3Cki/peerhub"
	"github.com/H3Cki/peerhub/cmd/commands"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	HandshakeTimeout: 10 * time.Second,
}

type handler struct {
	hub *peerhub.Hub
	wc  *writerCache
}

func (h *handler) registerHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/answerings", commands.AnsweringsHandler(h.hub))
	mux.HandleFunc("/hub", h.wsHub)
}

func (h *handler) wsHub(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for {
		msg := message{}
		if err := conn.ReadJSON(&msg); err != nil {
			fmt.Println(err)
			return
		}

		if err := h.handleMessage(conn, msg); err != nil {
			fmt.Println(err)
		}
	}
}

func (h *handler) handleMessage(conn *websocket.Conn, msg message) error {
	w := newWriter(conn, msg.Conv)

	switch msg.Type {
	case messageTypeCreateAnsweringPeer:
		req := peerhub.CreateAnsweringPeerRequest{}
		if err := msg.UnmarshalData(&req); err != nil {
			return err
		}
		err := h.handleCreateAnsweringPeer(w, req)
		if err != nil {
			w.Error(err)
		}
		return err
	case messageTypeCreateOfferingPeer:
		req := peerhub.CreateOfferingPeerRequest{}
		if err := msg.UnmarshalData(&req); err != nil {
			return err
		}
		err := h.handleCreateOfferingPeer(w, req)
		if err != nil {
			w.Error(err)
		}
		return err
	case messageTypeOfferAnswer:
		req := peerhub.CreateAnswerRequest{}
		if err := msg.UnmarshalData(&req); err != nil {
			return err
		}
		err := h.handleCreateAnswer(w, req)
		if err != nil {
			w.Error(err)
		}
		return err
	}

	return nil
}

// handleCreateAnsweringPeer creates answering peer and sends all matching offers to it
func (h *handler) handleCreateAnsweringPeer(apWriter *writer, req peerhub.CreateAnsweringPeerRequest) error {
	// create ap
	ap, err := h.hub.CreateAnsweringPeer(req)
	if err != nil {
		return fmt.Errorf("error creating answering peer: %w", err)
	}

	if err := h.wc.setA(ap.Name, apWriter, true); err != nil {
		return fmt.Errorf("error caching peer's connection: %w", err)
	}

	err = apWriter.Write(messageTypeInfo, genericMessage{
		Message: fmt.Sprintf("answering peer %s created", ap.Name),
	})
	if err != nil {
		return err
	}

	// create ap deals
	offers, fOffers, err := h.hub.OffersForAnsweringPeer(ap)
	if err != nil {
		return fmt.Errorf("error getting offers for answering peer: %w", err)
	}

	if err := h.sendOffers(apWriter, offers, fOffers); err != nil {
		return err
	}

	return nil
}

func (h *handler) sendOffers(w *writer, offers []peerhub.Offer, fOffers []peerhub.FailedOffer) error {
	errs := []error{}
	// handle deals - write offer to answering peer's connection
	for _, offer := range offers {
		err := w.Write(messageOffer, offer)
		if err != nil {
			errs = append(errs, err)
		}
	}

	// handle failed deals
	for _, fdeal := range fOffers {
		err := w.Write(messageTypeOfferFailed, fdeal)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// handleCreateOfferingPeer creates offering peer and sends an offer to matched answering if such was found
func (h *handler) handleCreateOfferingPeer(opWriter *writer, req peerhub.CreateOfferingPeerRequest) error {
	// create op
	op, err := h.hub.CreateOfferingPeer(req)
	if err != nil {
		return fmt.Errorf("error creating answering peer: %w", err)
	}

	if err := h.wc.setO(op.Name, opWriter, true); err != nil {
		return fmt.Errorf("error caching peer's connection: %w", err)
	}

	err = opWriter.Write(messageTypeInfo, genericMessage{
		Message: fmt.Sprintf("offering peer %s created", op.Name),
	})
	if err != nil {
		return err
	}

	offer, failed, isOffer, isFailed, err := h.hub.OfferFromOfferingPeer(op)
	if err != nil {
		return err
	}

	if isOffer {
		err = opWriter.Info(fmt.Sprintf("offer for peer %s created", offer.AnsweringPeer))
		if err != nil {
			fmt.Println(err)
		}

		// send offer to ap
		apW, ok := h.wc.getA(offer.AnsweringPeer)
		if !ok {
			return fmt.Errorf("could not find connection to answering peer")
		}

		apwErr := apW.Conv("").Write(messageOffer, offer)
		if err != nil {
			opwErr := opWriter.Error(errors.New("error sending offer to answering peer"))
			return errors.Join(err, apwErr, opwErr)
		}

		opwErr := opWriter.Info(fmt.Sprintf("offer for peer %s sent", offer.AnsweringPeer))
		return errors.Join(err, opwErr)
	}

	if isFailed {
		err = opWriter.Write(messageTypeOfferFailed, failed)
		if err != nil {
			fmt.Println(err)
		}
	}

	return nil
}

// handleCreateAnswer creates an answer and sends it to the offering peer
func (h *handler) handleCreateAnswer(w *writer, req peerhub.CreateAnswerRequest) error {
	answer, offer, err := h.hub.CreateAnswer(req)
	if err != nil {
		return fmt.Errorf("error creating answer: %w", err)
	}

	if err := h.wc.setA(answer.AnsweringPeer, w, true); err != nil {
		return fmt.Errorf("error caching peer's connection: %w", err)
	}

	// send answer to op
	opW, ok := h.wc.getO(offer.OfferingPeer)
	if !ok {
		return fmt.Errorf("could not find connection to offering peer")
	}

	err = opW.Conv("").Write(messageTypeOfferAnswer, answer)
	if err != nil {
		wErr := w.Error(errors.New("error sending answer to offering peer"))
		return errors.Join(err, wErr)
	}

	wErr := w.Info(fmt.Sprintf("answer sent to %s", offer.OfferingPeer))

	return wErr
}
