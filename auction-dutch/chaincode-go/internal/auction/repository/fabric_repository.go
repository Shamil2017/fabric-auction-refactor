package repository

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	"github.com/hyperledger/fabric-protos-go-apiv2/msp"
	"google.golang.org/protobuf/proto"

	"github.com/hyperledger/fabric-samples/auction/dutch-auction/chaincode-go/internal/auction/domain"
)

type FabricRepository struct{}

func NewFabricRepository() *FabricRepository {
	return &FabricRepository{}
}

func (r *FabricRepository) GetAuction(
	ctx contractapi.TransactionContextInterface,
	auctionID string,
) (*domain.Auction, error) {
	auctionJSON, err := ctx.GetStub().GetState(auctionID)
	if err != nil {
		return nil, fmt.Errorf("read auction from world state: %w", err)
	}

	if auctionJSON == nil {
		return nil, errors.New("auction does not exist")
	}

	var auction domain.Auction
	if err := json.Unmarshal(auctionJSON, &auction); err != nil {
		return nil, fmt.Errorf("unmarshal auction: %w", err)
	}

	return &auction, nil
}

func (r *FabricRepository) SaveAuction(
	ctx contractapi.TransactionContextInterface,
	auctionID string,
	auction *domain.Auction,
) error {
	auctionJSON, err := json.Marshal(auction)
	if err != nil {
		return fmt.Errorf("marshal auction: %w", err)
	}

	if err := ctx.GetStub().PutState(auctionID, auctionJSON); err != nil {
		return fmt.Errorf("save auction to world state: %w", err)
	}

	return nil
}

func (r *FabricRepository) SetAuctionEndorsement(
	ctx contractapi.TransactionContextInterface,
	auctionID string,
	mspids []string,
	auditor bool,
) error {
	principals := make([]*msp.MSPPrincipal, len(mspids))
	participantPolicies := make([]*common.SignaturePolicy, len(mspids))

	for i, id := range mspids {
		principal, err := proto.Marshal(
			&msp.MSPRole{
				Role:          msp.MSPRole_PEER,
				MspIdentifier: id,
			},
		)
		if err != nil {
			return err
		}

		principals[i] = &msp.MSPPrincipal{
			PrincipalClassification: msp.MSPPrincipal_ROLE,
			Principal:               principal,
		}

		participantPolicies[i] = &common.SignaturePolicy{
			Type: &common.SignaturePolicy_SignedBy{
				SignedBy: int32(i),
			},
		}
	}

	var policy *common.SignaturePolicyEnvelope

	if !auditor {
		policy = &common.SignaturePolicyEnvelope{
			Version: 0,
			Rule: &common.SignaturePolicy{
				Type: &common.SignaturePolicy_NOutOf_{
					NOutOf: &common.SignaturePolicy_NOutOf{
						N:     int32(len(mspids)),
						Rules: participantPolicies,
					},
				},
			},
			Identities: principals,
		}
	} else {
		auditorMSP, err := proto.Marshal(
			&msp.MSPRole{
				Role:          msp.MSPRole_PEER,
				MspIdentifier: "Org3MSP",
			},
		)
		if err != nil {
			return err
		}

		principals = append(principals, &msp.MSPPrincipal{
			PrincipalClassification: msp.MSPPrincipal_ROLE,
			Principal:               auditorMSP,
		})

		auditorPolicy := &common.SignaturePolicy{
			Type: &common.SignaturePolicy_NOutOf_{
				NOutOf: &common.SignaturePolicy_NOutOf{
					N: 2,
					Rules: []*common.SignaturePolicy{
						{
							Type: &common.SignaturePolicy_SignedBy{
								SignedBy: int32(len(principals) - 1),
							},
						},
						{
							Type: &common.SignaturePolicy_NOutOf_{
								NOutOf: &common.SignaturePolicy_NOutOf{
									N:     1,
									Rules: participantPolicies,
								},
							},
						},
					},
				},
			},
		}

		allParticipantsPolicy := &common.SignaturePolicy{
			Type: &common.SignaturePolicy_NOutOf_{
				NOutOf: &common.SignaturePolicy_NOutOf{
					N:     int32(len(mspids)),
					Rules: participantPolicies,
				},
			},
		}

		policy = &common.SignaturePolicyEnvelope{
			Version: 0,
			Rule: &common.SignaturePolicy{
				Type: &common.SignaturePolicy_NOutOf_{
					NOutOf: &common.SignaturePolicy_NOutOf{
						N: 1,
						Rules: []*common.SignaturePolicy{
							auditorPolicy,
							allParticipantsPolicy,
						},
					},
				},
			},
			Identities: principals,
		}
	}

	policyBytes, err := proto.Marshal(policy)
	if err != nil {
		return err
	}

	if err := ctx.GetStub().SetStateValidationParameter(auctionID, policyBytes); err != nil {
		return fmt.Errorf("set auction endorsement policy: %w", err)
	}

	return nil
}
