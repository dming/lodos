package rpcpb

func NewResultInfo(cid string, errCode int32, errStr string, resultsType []string, results [][]byte) *ResultInfo {
	return &ResultInfo{
		Cid: cid,
		Error: errStr,
		ErrCode: errCode,
		ResultsType: resultsType,
		Results: results,
	}
}