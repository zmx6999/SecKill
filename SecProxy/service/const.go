package service

const (
	ProductStatusSoldOut = 2001

	NetworkBusy = 1001
	SecKillSuccess = 1002
	ProductNotFound = 1003
	ProductSoldOut = 1004
	ProductStockInsufficient = 1005
	ProductOnePersonBuyExceed = 1006
	Timeout = 1007
	SecKillNotStart = 1008
	SecKillEnd = 1009
	ProductForceSoldOut = 1010
	NotifyClosed = 1011
	InvalidRequest = 1012
)

var msg = map[int]string{
	NetworkBusy: "Network Busy",
	SecKillSuccess: "SecKill Success",
	ProductNotFound: "Product Not Found",
	ProductSoldOut: "Product Sold Out",
	ProductStockInsufficient: "Product Stock Insufficient",
	ProductOnePersonBuyExceed: "Product One Person Buy Exceed",
	Timeout: "Timeout",
	SecKillNotStart: "Sec Kill Not Start",
	SecKillEnd: "Sec Kill End",
	ProductForceSoldOut: "Product Sold Out",
	NotifyClosed: "Notify Closed",
	InvalidRequest: "Invalid Request",
}

func GetMsg(code int) string {
	m, ok := msg[code]
	if !ok {
		return "Unknown Error"
	}
	return m
}
