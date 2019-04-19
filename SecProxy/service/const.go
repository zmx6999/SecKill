package service

const (
	ProductStatusNormal = 0
	ProductStatusSoldOut = 1
	ProductStatusForceSoldOut = 2

	ProductProcessNotStart = 0
	ProductProcessStart = 1
	ProductProcessEnd = 2

	OK = 0
	ProductNotStartErr = 3001
	ProductAlreadyEndErr = 3002
	ProductSoldOutErr = 3003
	ProductNotFoundErr = 3004
	InvalidRequestParamErr = 3005
	UserValidationErr = 3006
	NetworkBusyErr = 3007
	IPBlockErr = 3008
	UserBlockErr = 3009
	TimeoutErr = 3010
	ClientClosedErr = 3011
)

var (
	ErrMsg = map[int]string {
		OK: "OK",
		ProductNotStartErr: "Second kill has not started",
		ProductAlreadyEndErr: "Second kill has already ended",
		ProductSoldOutErr: "Product has been sold out",
		ProductNotFoundErr: "Product not found",
		InvalidRequestParamErr: "Invalid request parameter",
		UserValidationErr: "User validation error",
		NetworkBusyErr: "Network busy",
		IPBlockErr: "IP has been blocked",
		UserBlockErr: "User has been blocked",
		TimeoutErr: "Request timeout",
		ClientClosedErr: "Client has been closed",
	}
)
