package utils

import (
	"fmt"
	bindings "liquidation-bot/bindings/common"
	"reflect"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

type Multicall3Result struct {
	Success    bool    "json:\"success\""
	ReturnData []uint8 "json:\"returnData\""
}

func Aggregate3(callOpts *bind.CallOpts, multicall3 *bindings.Multicall3, calls []bindings.Multicall3Call3) ([]bindings.Multicall3Result, error) {
	var aggregate3Result []any
	raw := &bindings.Multicall3Raw{Contract: multicall3}
	if err := raw.Call(callOpts, &aggregate3Result, "aggregate3", &calls); err != nil {
		return nil, fmt.Errorf("failed to execute multicall: %w", err)
	}

	return ParseAggregate3Result(aggregate3Result)
}

func ParseAggregate3Result(aggregate3Result []any) ([]bindings.Multicall3Result, error) {
	if len(aggregate3Result) == 0 {
		return nil, fmt.Errorf("failed to get aggregate3 result")
	}

	// 解析 aggregate3 结果
	aggregate3Results, ok := aggregate3Result[0].([]struct {
		Success    bool    "json:\"success\""
		ReturnData []uint8 "json:\"returnData\""
	})
	if !ok {
		return nil, fmt.Errorf("failed to parse aggregate3 result: %v", reflect.TypeOf(aggregate3Result[0]))
	}

	// 转换为 Multicall3Result 类型
	results := make([]bindings.Multicall3Result, len(aggregate3Results))
	for i, v := range aggregate3Results {
		results[i] = bindings.Multicall3Result{
			Success:    v.Success,
			ReturnData: v.ReturnData,
		}
	}
	return results, nil
}
