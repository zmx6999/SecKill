package service

const (
	ProductStatusSoldOut = 2001

	NetworkBusyErr = 1001
	ProductNotFoundErr = 1003
	ProductSoldOutErr = 1004
	TimeoutErr = 1005
	AlreadyBuyErr = 1006
	SecKillNotStartErr = 1007
	SecKillAlreadyEndErr = 1008
	CloseRequestErr = 1009
	ProductForceSoldOutErr = 1010
	ProductNotEnoughErr = 1011

	SecKillSuccess = 200
)

var errMsg = map[int]string{
	NetworkBusyErr: "Network busy",
	ProductNotFoundErr: "Product not found",
	ProductSoldOutErr: "Product has been sold out",
	TimeoutErr: "Timeout",
	AlreadyBuyErr: "Purchase number exceed",
	SecKillNotStartErr: "Second kill has not started",
	SecKillAlreadyEndErr: "Second kill has already ended",
	SecKillSuccess: "ok",
	CloseRequestErr: "Request has been closed",
	ProductForceSoldOutErr: "Product has been sold out",
	ProductNotEnoughErr: "Product not enough",
}

func getErrMsg(code int) string {
	return errMsg[code]
}

func GetErrMsg(code int) string {
	return getErrMsg(code)
}
