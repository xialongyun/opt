package main

import (
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"time"
)

type TimeContract struct {
	contractapi.Contract
}

// CompareTime 比较time1 和 time2，如果time1 在time2 时间前面，返回true
func (t *TimeContract) CompareTime(time1 string, time2 string) bool {
	time1Obj, _ := time.Parse("2006-01-02 15:04:05", time1)
	time2Obj, _ := time.Parse("2006-01-02 15:04:05", time2)

	return time1Obj.Before(time2Obj)
}

// CompareWithNow time1 与现有时间比较，如果现有时间比time1晚，返回true
func (t *TimeContract) CompareWithNow(time1 string) bool {
	timeNow := time.Now().Format("2006-01-02 15:04:05")

	return t.CompareTime(time1, timeNow)
}
