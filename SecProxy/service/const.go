package service

const (
	ProductStatusSoldOut = 2001

	NetworkBusy = 1001
	SecKillSuccess = 1002
	ProductNotFound = 1003
	SoldOut = 1004
	ProductNotEnough = 1005
	PurchaseExceed = 1006
	Timeout = 1007
	SecKillNotStart = 1008
	SecKillEnd = 1009
	ForceSoldOut = 1010
	RequestClose = 1011
)

var errMsg = map[int]string{
	NetworkBusy: "Network busy",
	SecKillSuccess: "ok",
	ProductNotFound: "Product not found",
	SoldOut: "Sold out",
	ProductNotEnough: "Product not enough",
	PurchaseExceed: "Purchase exceed",
	Timeout: "Timeout",
	SecKillNotStart: "Second kill has not started",
	SecKillEnd: "Second kill has ended",
	ForceSoldOut: "Sold out",
	RequestClose: "Request close",
}

func GetErrMsg(code int) string {
	msg, ok := errMsg[code]
	if !ok {
		return "UNKNOWN ERROR"
	}
	return msg
}
