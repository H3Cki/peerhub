package peerhub

import "errors"

type HubConfig struct {
	PeerService   PeerService
	SignalService SignalService
}

type Hub struct {
	peerSvc PeerService
	dealSvc SignalService
}

func NewHub(cfg HubConfig) *Hub {
	return &Hub{
		peerSvc: cfg.PeerService,
		dealSvc: cfg.SignalService,
	}
}

func (h *Hub) GetAnsweringPeersPrevies() ([]AnsweringPeerPreview, error) {
	aps, err := h.peerSvc.GetAnsweringPeers()
	if err != nil {
		return nil, err
	}
	apps := []AnsweringPeerPreview{}
	for _, ap := range aps {
		apps = append(apps, AnsweringPeerPreview{
			Name:      ap.Name,
			Protected: len(ap.AccessKeys) != 0,
		})
	}
	return apps, nil
}

func (h *Hub) CreateAnsweringPeer(req CreateAnsweringPeerRequest) (AnsweringPeer, error) {
	ap := AnsweringPeer(req)

	oldAP, err := h.peerSvc.GetAnsweringPeer(ap.Name)
	if err == nil && oldAP.ManagementKeyMatches(ap.ManagementKey) {
		err := h.peerSvc.UpdateAnsweringPeer(ap)
		if err != nil {
			return AnsweringPeer{}, err
		}
		return ap, nil
	}

	if !errors.Is(err, ErrAnsweringPeerNotFound) {
		return AnsweringPeer{}, err
	}

	err = h.peerSvc.CreateAnsweringPeer(ap)
	if err != nil {
		return AnsweringPeer{}, err
	}

	return ap, nil
}

func (h *Hub) CreateOfferingPeer(req CreateOfferingPeerRequest) (OfferingPeer, error) {
	op := OfferingPeer{
		Name:            req.Name,
		TargetName:      req.TargetName,
		TargetAccessKey: req.TargetAccessKey,
		ManagementKey:   req.ManagementKey,
		SDP:             req.SDP,
		Delete:          req.Delete,
	}

	oldOP, err := h.peerSvc.GetOfferingPeer(op.Name)
	if err == nil && oldOP.ManagementKeyMatches(op.ManagementKey) {
		err := h.peerSvc.UpdateOfferingPeer(op)
		if err != nil {
			return OfferingPeer{}, err
		}
		return op, nil
	}

	if !errors.Is(err, ErrOfferingPeerNotFound) {
		return OfferingPeer{}, err
	}

	if err := h.peerSvc.CreateOfferingPeer(op); err != nil {
		return OfferingPeer{}, err
	}

	return op, nil
}

// CreateAnswer creates an answer and returns the Offer which the answer relates to
func (h *Hub) CreateAnswer(req CreateAnswerRequest) (Answer, Offer, error) {
	offer, err := h.dealSvc.GetOffer(req.OfferID)
	if err != nil {
		return Answer{}, Offer{}, err
	}
	answer := NewAnswer(offer.ID, offer.AnsweringPeer, req.SDP)
	return answer, offer, nil
}

func (h *Hub) OffersForAnsweringPeer(ap AnsweringPeer) ([]Offer, []FailedOffer, error) {
	ops, err := h.peerSvc.GetOfferingPeersByTarget(ap.Name)
	if err != nil {
		return nil, nil, err
	}

	offers := []Offer{}
	fOffers := []FailedOffer{}

	for _, op := range ops {
		if !ap.AccessKeyMatches(op.TargetAccessKey) {
			fo := FailedOffer{
				OfferingPeer:  op.Name,
				AnsweringPeer: ap.Name,
				Error:         err,
			}
			fOffers = append(fOffers, fo)
			continue
		}

		offer := NewOffer(op.Name, op.SDP, ap.Name)
		offers = append(offers, offer)
	}

	return offers, fOffers, nil
}

func (h *Hub) OfferFromOfferingPeer(op OfferingPeer) (offer Offer, failedOffer FailedOffer, isOffer, isFailed bool, err error) {
	ap, err := h.peerSvc.GetAnsweringPeer(op.TargetName)
	if op.IgnoreNotFound && errors.Is(err, ErrAnsweringPeerNotFound) {
		return Offer{}, FailedOffer{}, false, false, nil
	}
	if err != nil {
		return Offer{}, FailedOffer{}, false, false, err
	}

	if !ap.AccessKeyMatches(op.TargetAccessKey) {
		fo := FailedOffer{
			OfferingPeer:  op.Name,
			AnsweringPeer: ap.Name,
			Error:         ErrInvalidAccessKey,
		}
		return Offer{}, fo, false, true, err
	}

	o := NewOffer(op.Name, op.SDP, ap.Name)
	if err := h.dealSvc.CreateOffer(o); err != nil {
		return Offer{}, FailedOffer{}, false, false, err
	}

	return o, FailedOffer{}, true, false, nil
}

func (h *Hub) DeleteAnsweringPeer(_ DeleteAnsweringPeerRequest) error {
	panic("not implemented") // TODO: Implement
}

func (h *Hub) DeleteOfferingPeer(_ DeleteOfferingPeerRequest) error {
	panic("not implemented") // TODO: Implement
}

type CreateAnsweringPeerRequest struct {
	Name          string   `json:"name"`
	AccessKeys    []string `json:"accesskey"`
	ManagementKey string   `json:"managementkey"`
}

type CreateAnswerRequest struct {
	OfferID string `json:"offerID"`
	SDP     string `json:"sdp"`
}

type DealForAnsweringPeerRequest struct {
	OfferingPeerName  string `json:"offeringpeername"`
	AnsweringPeerName string `json:"answeringpeername"`
	AccessKey         string `json:"accesskey"`
	SDP               string `json:"sdp"`
}

type DeleteAnsweringPeerRequest struct {
	PeerID string
}

type CreateOfferingPeerRequest struct {
	Name            string `json:"name"`
	TargetName      string `json:"targetname"`
	TargetAccessKey string `json:"targetaccesskey"`
	ManagementKey   string `json:"managementkey"`
	SDP             string `json:"sdp"`
	Delete          bool   `json:"delete"`
}

type DeleteOfferingPeerRequest struct {
	PeerID string
}
