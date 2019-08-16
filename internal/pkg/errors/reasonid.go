package errors

// General
const (
	ReasonInvalidKeyReqPath         = 10000
	ReasonInvalidKeyReqParams       = 10001
	ReasonInvalidKeyReqBody         = 10002
	ReasonInvalidValueReqPath       = 10003
	ReasonInvalidValueReqParams     = 10004
	ReasonInvalidValueReqBody       = 10005
	ReasonInvalidReqFormat          = 10006
	ReasonInvalidJSONFormat         = 10007
	ReasonMissingFieldReq           = 10008
	ReasonIndexOutOfRange           = 10009
	ReasonFailedToConnectDB         = 10010
	ReasonFailedToReadDB            = 10011
	ReasonFailedToUpdateDB          = 10012
	ReasonFailedToExecCMD           = 10013
	ReasonInvalidParams             = 10014
	ReasonReservedChar              = 10015
	ReasonInvalidTimeSequence       = 10016
	ReasonInvalidTimeRange          = 10017
)

// Authentication
const (
	ReasonInvalidJWT                = 11000
	ReasonExpiredJWT                = 11001
	ReasonInvalidCredential         = 11002
	ReasonFailedToGenJWT            = 11003
)

// Authorization
const (
	ReasonNotAuthorized             = 12000
)

// Accounts
const (
	ReasonUserCreateFailed          = 13000
	ReasonUserReadFailed            = 13001
	ReasonUserUpdateFailed          = 13002
	ReasonUserDeleteFailed          = 13003
	ReasonUserNotFound              = 13004
	ReasonUserAlreadyExist          = 13005
	ReasonDomainNotFound            = 13006
)

// Events
const (
	ReasonSystemEventNotFound       = 14000
)

// Resources
const (
	ReasonAgentNotFound             = 15000
	ReasonHostNotFound              = 15100
	ReasonPhysicalDiskNotFound      = 15200
	ReasonVirtualMachineNotFound    = 15300
)

// Metrics
const (
	ReasonDiskMetricNotFount        = 16000
	ReasonHostMetricNotFount        = 16001
	ReasonInvalidMetricCategory     = 16002
	ReasonVmMetricNotFount          = 16003
)

// Predictions
const (
	ReasonDiskPredictionNotFount    = 17000
	ReasonHostPredictionNotFount    = 17001
	ReasonInvalidPredictionCategory = 17002
)

// Keycodes
const (
	ReasonKeycodeExpired                  = 18000
	ReasonKeycodeNotFound                 = 18001
	ReasonKeycodeNoLicenseFile            = 18100
	ReasonKeycodeIncorrectPadding         = 18101
	ReasonKeycodeAlreadyApplied           = 18102
	ReasonKeycodeInvalidSystemTime        = 18103
	ReasonKeycodeInvalidSignature         = 18200
	ReasonKeycodeSIDMismatch              = 18201
	ReasonKeycodeDomainMismatch           = 18202
	ReasonKeycodeInvalidTool              = 18203
	ReasonKeycodeInvalidState             = 18204
	ReasonKeycodeInvalidContent           = 18205
	ReasonKeycodeInvalidKeycode           = 18206
	ReasonKeycodeVersionMismatch          = 18207
	ReasonKeycodeOEMVendorMismatch        = 18208
)

// Licenses
const (
	ReasonLicenseInvalidLicense           = 19000
	ReasonLicenseNoLicense                = 19001
	ReasonLicenseFailedToConnectServer    = 19002
	ReasonLicenseFailedToReadServer       = 19003
	ReasonLicenseFailedToWriteServer      = 19004

	ReasonLicenseUserExpired              = 19100
	ReasonLicenseUserExceedUserMaximum    = 19101
	ReasonLicenseUserExceedDiskMaximum    = 19102
	ReasonLicenseUserExceedHostMaximum    = 19103
	ReasonLicenseUserHasLicense           = 19104
	ReasonLicenseUserNoLicense            = 19105

	ReasonLicenseDomainExpired            = 19200
	ReasonLicenseDomainExceedUserMaximum  = 19201
	ReasonLicenseDomainExceedDiskMaximum  = 19202
	ReasonLicenseDomainExceedHostMaximum  = 19203
	ReasonLicenseDomainHasLicense         = 19204
	ReasonLicenseDomainNoLicense          = 19205

	ReasonLicenseSystemExpired            = 19300
	ReasonLicenseSystemInvalid            = 19301
	ReasonLicenseSystemExceedUserMaximum  = 19302
	ReasonLicenseSystemExceedDiskMaximum  = 19303
	ReasonLicenseSystemExceedHostMaximum  = 19304
	ReasonLicenseSystemNoLicense          = 19305
)

//subscription
const (
	ReasonSubscriptionUserNotStarted      = 20000
	ReasonSubscriptionUserExpired         = 20001
	ReasonSubscriptionDomainNotStarted    = 20002
	ReasonSubscriptionDomainExpired       = 20003
)
