package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type BallotContract struct {
	contractapi.Contract
}

// BallotProposal 投票提案
type BallotProposal struct {
	BallotProposalName string           `json:"ballot_proposal_name"`
	ProposerName       string           `json:"proposer_name"`
	ProposalType       string           `json:"proposal_type"`
	VoterMap           map[string]Voter `json:"voter_map"`
	UpVotes            int              `json:"up_votes"`
	NegativeVotes      int              `json:"negative_votes"`
	NumberOfVoter      int              `json:"number_of_voter"`
	NumberOfVoted      int              `json:"number_of_voted"`
	State              string           `json:"state"`
	StartTime          string           `json:"start_time"`
	EndTime            string           `json:"end_time"`
	Variable           string           `json:"variable"`
	Value              int              `json:"value"`
	Result             bool             `json:"result"`
}

// CreateBallotProposal 创建投票提案
func (b *BallotContract) CreateBallotProposal(
	ctx contractapi.TransactionContextInterface,
	ballotProposalName string,
	proposerName string,
	proposalType string,
	startTime string,
	endTime string) (*BallotProposal, error) {
	// 1.判断投票提案是否存在
	if b.BallotProposalExist(ctx, ballotProposalName) {
		return nil, fmt.Errorf("Ballot proposal is existed ! ")
	}

	// 2.判断时间是否符合规范
	var t TimeContract
	if !t.CompareTime(startTime, endTime) {
		return nil, fmt.Errorf("End time earlier than start time ! ")
	}

	// 3.查看proposer是否存在
	var r RoleContract
	proposer, err := r.QueryUser(ctx, proposerName)

	if err != nil {
		return nil, fmt.Errorf("query proposer false, %s", err.Error())
	}

	// 4.查看powerUser信用值， 若小于某个额度，则拒绝发起提案
	if proposer.UserCredit-CreditBorder < 0 {
		return nil, fmt.Errorf("proposer credit less than %d", CreditBorder)
	}

	// 5.获取候选人
	var e ElectionContract
	voterMap := make(map[string]Voter)

	// 6.获取全体用户的列表， 提案由全体用户公投
	publicUserList := r.QueryUserList(ctx)
	if proposalType == "Public" {
		for _, userName := range publicUserList.Users {
			//获取用户
			user, _ := r.QueryUser(ctx, userName)

			voterMap[user.UserName] = Voter{
				VoterName:  user.UserName,
				UserCredit: user.UserCredit,
				Power:      user.Power,
				Voted:      false,
			}
		}
	}

	// 7.获取委员会委员的列表，提案由全体委员投票
	leagueUserList := e.QueryCommittee(ctx)
	if proposalType == "League" {
		for _, userName := range leagueUserList.Users {
			//获取用户
			user, _ := r.QueryUser(ctx, userName)

			voterMap[user.UserName] = Voter{
				VoterName:  user.UserName,
				UserCredit: user.UserCredit,
				Power:      user.Power,
				Voted:      false,
			}
		}
	}

	// 8.赋值
	ballotProposal := BallotProposal{
		BallotProposalName: ballotProposalName,
		ProposerName:       proposerName,
		VoterMap:           voterMap,
		UpVotes:            0,
		NegativeVotes:      0,
		NumberOfVoter:      len(voterMap),
		NumberOfVoted:      0,
		State:              "Voting",
		StartTime:          startTime,
		EndTime:            endTime,
		Result:             false,
	}

	ballotProposalAsBytes, _ := json.Marshal(ballotProposal)

	// 9.上链
	err = ctx.GetStub().PutState(ballotProposalName, ballotProposalAsBytes)

	if err != nil {
		return nil, err
	}

	return &ballotProposal, nil
}

// VoteBallotProposal 投票
func (b *BallotContract) VoteBallotProposal(
	ctx contractapi.TransactionContextInterface,
	ballotProposalName string,
	voterName string,
	vote bool) (*BallotProposal, error) {
	var r RoleContract
	// 1.判断投票提案是否存在
	if !b.BallotProposalExist(ctx, ballotProposalName) {
		return nil, fmt.Errorf("Ballot proposal is not existed ! ")
	}

	// 2.获取投票提案
	ballotProposal, err := b.QueryBallotProposal(ctx, ballotProposalName)

	if err != nil {
		return nil, err
	}

	// 3.判断是否到达选举时间
	//var t TimeContract
	//if !t.CompareWithNow(ballotProposal.StartTime) || t.CompareWithNow(ballotProposal.EndTime) {
	//	return nil, fmt.Errorf("The proposal is not voting ! ")
	//}

	// 3.判断投票人是否存在
	if !r.UserExist(ctx, voterName) {
		return nil, fmt.Errorf("%s is not exist ! ", voterName)
	}

	// 4.获取投票人信息
	voter := ballotProposal.VoterMap[voterName]

	// 5.判断投票人是否投过票，如果投过票返回
	if voter.Voted == true {
		return nil, fmt.Errorf("voter had voted")
	}

	// 6.修改投票信息
	ballotProposal.VoterMap[voterName] = Voter{
		voterName,
		voter.UserCredit,
		voter.Power,
		true}

	// 7.更改投票状态
	if vote == true {
		ballotProposal.UpVotes += voter.UserCredit
	} else {
		ballotProposal.NegativeVotes += voter.UserCredit
	}

	// 8.更改已完成投票的人数
	ballotProposal.NumberOfVoted++

	// 9.投票提案上链
	ballotProposalAsBytes, _ := json.Marshal(ballotProposal)
	err = ctx.GetStub().PutState(ballotProposalName, ballotProposalAsBytes)

	if err != nil {
		return nil, err
	}

	// 9.更新信用值
	_ = r.ChangeCredit(ctx, voterName, BallotAwardCredit)

	return ballotProposal, nil
}

// CheckBallotProposal 检查投票提案结果
func (b *BallotContract) CheckBallotProposal(
	ctx contractapi.TransactionContextInterface,
	ballotProposalName string) (*BallotProposal, error) {
	// 1.判断投票提案是否存在
	if !b.BallotProposalExist(ctx, ballotProposalName) {
		return nil, fmt.Errorf("Ballot proposal is not existed ! ")
	}

	// 2.获取投票提案
	ballotProposal, err := b.QueryBallotProposal(ctx, ballotProposalName)

	if err != nil {
		return nil, fmt.Errorf("Failed to query election proposal Info from world state. %s ", err.Error())
	}

	// 3.判断是否到达投票时间
	//var t TimeContract
	//if !t.CompareWithNow(ballotProposal.EndTime) {
	//	return nil, fmt.Errorf("The proposal is voting ! ")
	//}

	// 4.更改投票提案的状态
	ballotProposal.State = "Done"

	if ballotProposal.NegativeVotes-ballotProposal.UpVotes < 0 &&
		ballotProposal.NumberOfVoted*2-ballotProposal.NumberOfVoter > 0 {
		ballotProposal.Result = true
	} else {
		ballotProposal.Result = false
	}

	// 5.投票提案上链
	ballotProposalAsBytes, _ := json.Marshal(ballotProposal)
	err = ctx.GetStub().PutState(ballotProposalName, ballotProposalAsBytes)

	if err != nil {
		return nil, err
	}

	return ballotProposal, nil
}

// QueryBallotProposal 获取投票提案
func (b *BallotContract) QueryBallotProposal(
	ctx contractapi.TransactionContextInterface,
	ballotProposalName string) (*BallotProposal, error) {
	// 1.获取投票提案信息
	ballotProposalAsBytes, err := ctx.GetStub().GetState(ballotProposalName)

	if err != nil {
		return nil, fmt.Errorf("Failed to query User Info from world state. %s", err.Error())
	}

	if ballotProposalAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", ballotProposalName)
	}

	// 2.赋值
	ballotProposal := new(BallotProposal)
	_ = json.Unmarshal(ballotProposalAsBytes, ballotProposal)

	return ballotProposal, nil
}

// BallotProposalExist 判断投票提案是否存在
func (b *BallotContract) BallotProposalExist(
	ctx contractapi.TransactionContextInterface,
	ballotProposalName string) bool {
	ballotProposalAsBytes, err := ctx.GetStub().GetState(ballotProposalName)
	// 1.获取投票提案
	if err != nil {
		return false
	}

	// 2.如果投票提案不存在，或者获取选举提案失败，返回false
	if ballotProposalAsBytes == nil {
		return false
	}

	// 3.如果投票提案存在，返回true
	return true
}
