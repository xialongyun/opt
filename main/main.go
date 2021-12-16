package main

import (
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {

	chaincode, err := contractapi.NewChaincode(
		new(RoleContract),
		new(PowerTXContract),
		new(ElectionContract),
		new(BallotContract),
		new(VarChangeContract),
		new(TimeContract))

	if err != nil {
		fmt.Printf("Error create Contract chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting Contract chaincode: %s", err.Error())
	}
}
