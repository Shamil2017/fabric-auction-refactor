package domain

import (
	"errors"
	"strings"
)

type Auction struct {
	Type         string             `json:"objectType"`
	ItemSold     string             `json:"item"`
	Seller       string             `json:"seller"`
	Quantity     int                `json:"quantity"`
	Orgs         []string           `json:"organizations"`
	PrivateBids  map[string]BidHash `json:"privateBids"`
	RevealedBids map[string]FullBid `json:"revealedBids"`
	Winners      []Winner           `json:"winners"`
	Price        int                `json:"price"`
	Status       string             `json:"status"`
	Auditor      bool               `json:"auditor"`
}

type FullBid struct {
	Type     string `json:"objectType"`
	Quantity int    `json:"quantity"`
	Price    int    `json:"price"`
	Org      string `json:"org"`
	Buyer    string `json:"buyer"`
}

type BidHash struct {
	Org  string `json:"org"`
	Hash string `json:"hash"`
}

type Winner struct {
	Buyer    string `json:"buyer"`
	Quantity int    `json:"quantity"`
}

const (
	BidKeyType    = "bid"
	ObjectAuction = "auction"

	StatusOpen   = "open"
	StatusClosed = "closed"
	StatusEnded  = "ended"

	WithAuditor = "withAuditor"
)

func NewAuction(
	auctionID string,
	itemSold string,
	quantity int,
	seller string,
	sellerOrg string,
	withAuditor string,
) (*Auction, error) {
	if strings.TrimSpace(auctionID) == "" {
		return nil, errors.New("auctionID is required")
	}

	if strings.TrimSpace(itemSold) == "" {
		return nil, errors.New("itemSold is required")
	}

	if quantity <= 0 {
		return nil, errors.New("quantity must be greater than zero")
	}

	if strings.TrimSpace(seller) == "" {
		return nil, errors.New("seller is required")
	}

	if strings.TrimSpace(sellerOrg) == "" {
		return nil, errors.New("seller organization is required")
	}

	auditor := false
	if withAuditor == WithAuditor {
		auditor = true
	}

	return &Auction{
		Type:         ObjectAuction,
		ItemSold:     itemSold,
		Quantity:     quantity,
		Price:        0,
		Seller:       seller,
		Orgs:         []string{sellerOrg},
		PrivateBids:  make(map[string]BidHash),
		RevealedBids: make(map[string]FullBid),
		Winners:      []Winner{},
		Status:       StatusOpen,
		Auditor:      auditor,
	}, nil
}
