package handler

import (
	"encoding/base64"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
)

func getSubmittingClientIdentity(ctx contractapi.TransactionContextInterface) (string, error) {
	b64ID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return "", fmt.Errorf("failed to read client ID: %w", err)
	}

	decodedID, err := base64.StdEncoding.DecodeString(b64ID)
	if err != nil {
		return "", fmt.Errorf("failed to decode client ID: %w", err)
	}

	return string(decodedID), nil
}
