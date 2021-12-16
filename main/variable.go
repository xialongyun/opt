package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// InitCredit 初始信用值
var InitCredit int = 100

// CreditBorder 初始信用值边界
var CreditBorder int = 50

// TxAwardCredit 初始交易奖励信用值
var TxAwardCredit int = 5

// PowerBorder 初始电能分档边界
var PowerBorder int = 50

// AwardCredit 奖励分
func (v *VarChangeContract) AwardCredit(power int) int {
	return (power/PowerBorder + 1) * TxAwardCredit
}

// BallotAwardCredit 初始化投票奖励信用值
var BallotAwardCredit int = 6

// CommitteeMemberNumber 初始化委员会成员数量
var CommitteeMemberNumber = 5

type VarChangeContract struct {
	contractapi.Contract
}

// CreateChangeVariableProposal 创建更改变量投票提案
func (v *VarChangeContract) CreateChangeVariableProposal(
	ctx contractapi.TransactionContextInterface,
	ballotProposalName string,
	proposerName string,
	proposalType string,
	startTime string,
	endTime string,
	variable string,
	value int) (*BallotProposal, error) {
	var b BallotContract
	// 1.检查要更改的变量的名称是否准确
	if variable != "InitCredit" &&
		variable != "CreditBorder" &&
		variable != "TxAwardCredit" &&
		variable != "PowerBorder" &&
		variable != "BallotAwardCredit" &&
		variable != "CommitteeMemberNumber" {
		return nil, fmt.Errorf("The variable is not right ! ")
	}

	// 2.发起提案
	ballotProposal, err := b.CreateBallotProposal(ctx, ballotProposalName, proposerName, proposalType, startTime, endTime)

	if err != nil {
		return nil, err
	}

	ballotProposal.Variable = variable
	ballotProposal.Value = value

	ballotProposalAsBytes, _ := json.Marshal(ballotProposal)

	// 3.上链
	err = ctx.GetStub().PutState(ballotProposalName, ballotProposalAsBytes)

	if err != nil {
		return nil, err
	}

	return ballotProposal, nil
}

// CheckChangeVariableProposal 检查投票结果，并更改变量值
func (v *VarChangeContract) CheckChangeVariableProposal(
	ctx contractapi.TransactionContextInterface,
	ballotProposalName string) (*BallotProposal, error) {
	// 1.检查结果
	var b BallotContract
	ballotProposal, err := b.CheckBallotProposal(ctx, ballotProposalName)

	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	variable := ballotProposal.Variable

	// 2.如果提案结果为true,更改对应变量的值
	if ballotProposal.Result {
		if variable == "InitCredit" {
			InitCredit = ballotProposal.Value
		} else if variable == "CreditBorder" {
			CreditBorder = ballotProposal.Value
		} else if variable == "TxAwardCredit" {
			TxAwardCredit = ballotProposal.Value
		} else if variable == "PowerBorder" {
			PowerBorder = ballotProposal.Value
		} else if variable == "BallotAwardCredit" {
			BallotAwardCredit = ballotProposal.Value
		} else if variable == "CommitteeMemberNumber" {
			CommitteeMemberNumber = ballotProposal.Value
		}
	}

	return ballotProposal, nil
}
