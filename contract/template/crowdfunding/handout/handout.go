package main

import (
	"encoding/json"
	"errors"
	"github.com/ssbcV2/contract"
	"github.com/ssbcV2/meta"
	"github.com/ssbcV2/util"
)

var Statistics map[string]int         // 接收来自A链的统计数据
var Publisher string				  // 合约发布人
var Ready bool                        // 数据是否已经准备好

func init() {
	Statistics = map[string]int{}
	Publisher = contract.Caller()
	Ready = false
}

func GetCoin(args map[string]interface{}) (interface{}, error) {
	if !Ready {
		cb := meta.Callback{
			Caller:   "",
			Value:    0,
			Contract: contract.Name(),
			Method:   "ReceiveData",
			Args:     nil,
			Address:  "",
		}
		cbBytes, _ := json.Marshal(cb)
		reqArgs := map[string]interface{}{
			"type":     "chain", // "api":第三方接口，"chain":"跨链数据"
			"callback": string(cbBytes),
			"name": "ssbc2",
			"dataType": "contractData",
			"params": "deposit,Money",
		}
		// 日志事件
		recordArgs := map[string]interface{}{
			"state": "success",
		}
		// 调用QueryData预言机合约请求外部数据
		res, err := contract.Call("oracle", "QueryData", reqArgs)
		if err != nil {
			contract.Info("call QueryData contract error: %s", err)
			recordArgs["state"] = "fail"
			_, err = contract.Call("oracle", "RecordEvent", recordArgs)
			return meta.ContractUpdateData{}, err
		}
		resBytes, _ := json.Marshal(res)
		recordArgs["res"] = string(resBytes)
		_, err = contract.Call("oracle", "RecordEvent", recordArgs)
		return res, nil
	}

	a_addr, ok := args["a_addr"].(string)
	if !ok {
		return nil, errors.New("缺少a_addr参数")
	}

	if len(Statistics) == 0 {
		return nil, errors.New("没有数据或数据尚未准备好")
	}

	amount, ok := Statistics[a_addr]
	if !ok {
		return nil, errors.New("A链地址不存在")
	}

	err := contract.TransferFrom(Publisher, contract.Caller(), amount)
	if err != nil {
		return nil, err
	}

	Statistics[a_addr] = 0
	return nil, nil
}

func ReceiveData(args map[string]interface{}) (interface{}, error) {
	recordArgs := map[string]interface{}{
		"state": "success",
	}
	contract.Info("ReceiveData 方法收到参数：%+v", args)

	data, ok := args["data"].(string)
	if !ok {
		recordArgs["state"] = "fail"
		_, err := contract.Call("oracle", "RecordEvent", recordArgs)
		return meta.ContractUpdateData{}, err
	}
	contract.Info("ReceiveData 收到跨链数据：%s", data)
	Statistics = getStatistics(util.JsonToMap(data))
	Ready = true
	_, _ = contract.Call("oracle", "RecordEvent", recordArgs)
	return nil, nil
}

func getStatistics(args map[string]interface{}) map[string]int {
	d := args["Money"].(map[string]interface{})
	res := map[string]int{}

	for k, v := range d {
		tmp := v.(float64)
		res[k] = int(tmp)
	}
	return res
}
