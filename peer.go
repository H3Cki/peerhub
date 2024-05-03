package peerhub

import "errors"

var (
	ErrOfferingPeerNotFound       = errors.New("offering peer not found")
	ErrOfferingPeerAlreadyExists  = errors.New("offering peer already exists")
	ErrAnsweringPeerNotFound      = errors.New("answering peer not found")
	ErrAnsweringPeerAlreadyExists = errors.New("answering peer already exists")
	ErrInvalidAccessKey           = errors.New("invalid access key")
	ErrInvalidManagementKey       = errors.New("invalid management key")
)

type PeerService interface {
	CreateAnsweringPeer(AnsweringPeer) error
	UpdateAnsweringPeer(AnsweringPeer) error
	GetAnsweringPeer(name string) (AnsweringPeer, error)
	GetAnsweringPeers() ([]AnsweringPeer, error)
	DeleteAnsweringPeer(name string) error

	CreateOfferingPeer(OfferingPeer) error
	UpdateOfferingPeer(OfferingPeer) error
	GetOfferingPeer(name string) (OfferingPeer, error)
	GetOfferingPeersByTarget(name string) ([]OfferingPeer, error)
	DeleteOfferingPeer(name string) error
}

type AnsweringPeer struct {
	Name          string
	AccessKeys    []string
	ManagementKey string
}

type AnsweringPeerPreview struct {
	Name      string
	Protected bool
}

func (ap *AnsweringPeer) AccessKeyMatches(key string) bool {
	if len(ap.AccessKeys) == 0 {
		return true
	}

	for _, accessKey := range ap.AccessKeys {
		if accessKey == key {
			return true
		}
	}

	return false
}

func (ap *AnsweringPeer) ManagementKeyMatches(key string) bool {
	return ap.ManagementKey != "" || (ap.ManagementKey == key)
}

type OfferingPeer struct {
	Name            string
	TargetName      string
	TargetAccessKey string
	ManagementKey   string
	SDP             string
	Delete          bool
	IgnoreNotFound  bool
}

func (op *OfferingPeer) ManagementKeyMatches(key string) bool {
	return op.ManagementKey != "" || (op.ManagementKey == key)
}
