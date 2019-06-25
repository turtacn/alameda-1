package errors

// General
const (
	ReasonInvalidKeyReqPath     = 10000
	ReasonInvalidKeyReqParams   = 10001
	ReasonInvalidKeyReqBody     = 10002
	ReasonInvalidValueReqPath   = 10003
	ReasonInvalidValueReqParams = 10004
	ReasonInvalidValueReqBody   = 10005
	ReasonInvalidReqFormat      = 10006
	ReasonInvalidJSONFormat     = 10007
	ReasonMissingKeyReqBody     = 10008
	ReasonIndexOutOfRange       = 10009
	ReasonFailedToConnectDB     = 10010
	ReasonFailedToReadDB        = 10011
	ReasonFailedToUpdateDB      = 10012
	ReasonFailedToExecCMD       = 10013
	ReasonInvalidParams         = 10014
	ReasonReservedChar          = 10015
	ReasonInvalidTimeSequence   = 10016
	ReasonInvalidTimeRange      = 10017
	ReasonInvalidRequest        = 10018
)

// Authentication
const (
	ReasonInvalidJWT        = 11000
	ReasonExpiredJWT        = 11001
	ReasonInvalidCredential = 11002
	ReasonFailedToGenJWT    = 11003
)

// Authorization
const (
	ReasonNotAuthorized = 12000
)

// Accounts
const (
	ReasonUserCreateFailed = 13000
	ReasonUserReadFailed   = 13001
	ReasonUserUpdateFailed = 13002
	ReasonUserDeleteFailed = 13003
	ReasonUserNotFound     = 13004
	ReasonUserAlreadyExist = 13005
	ReasonDomainNotFound   = 13006
)

// Events
const (
	ReasonSystemEventNotFound = 14000
)

// Resources
const (
	ReasonAgentNotFound          = 15000
	ReasonHostNotFound           = 15100
	ReasonPhysicalDiskNotFound   = 15200
	ReasonVirtualMachineNotFound = 15300
)

// Metrics
const (
	ReasonDiskMetricNotFount    = 16000
	ReasonHostMetricNotFount    = 16001
	ReasonInvalidMetricCategory = 16002
	ReasonVmMetricNotFount      = 16003
)

// Predictions
const (
	ReasonDiskPredictionNotFount    = 17000
	ReasonHostPredictionNotFount    = 17001
	ReasonInvalidPredictionCategory = 17002
)

// Licenses
const (
	ReasonLicenseInvalidLicense        = 18000
	ReasonLicenseInvalidKeycode        = 18001
	ReasonLicenseInvalidContent        = 18002
	ReasonLicenseInvalidSignature      = 18003
	ReasonLicenseInvalidTool           = 18004
	ReasonLicenseInvalidState          = 18005
	ReasonLicenseSIDMismatch           = 18006
	ReasonLicenseDomainMismatch        = 18007
	ReasonLicenseVersionMismatch       = 18008
	ReasonLicenseOEMVendorMismatch     = 18009
	ReasonLicenseNoLicenseFile         = 18010
	ReasonLicenseKeycodeNotFound       = 18011
	ReasonLicenseKeycodeAlreadyApplied = 18012
	ReasonLicenseIncorrectPadding      = 18013
	ReasonLicenseFailedToConnectServer = 18014
	ReasonLicenseFailedToReadServer    = 18015
	ReasonLicenseFailedToWriteServer   = 18016

	ReasonLicenseUserExpired           = 18017
	ReasonLicenseUserExceedUserMaximum = 18018
	ReasonLicenseUserExceedDiskMaximum = 18019
	ReasonLicenseUserExceedHostMaximum = 18020
	ReasonLicenseUserHasLicense        = 18021
	ReasonLicenseUserNoLicense         = 18022

	ReasonLicenseDomainExpired           = 18023
	ReasonLicenseDomainExceedUserMaximum = 18024
	ReasonLicenseDomainExceedDiskMaximum = 18025
	ReasonLicenseDomainExceedHostMaximum = 18026
	ReasonLicenseDomainHasLicense        = 18027
	ReasonLicenseDomainNoLicense         = 18028

	ReasonLicenseSystemExpired           = 18029
	ReasonLicenseSystemInvalid           = 18030
	ReasonLicenseSystemExceedUserMaximum = 18031
	ReasonLicenseSystemExceedDiskMaximum = 18032
	ReasonLicenseSystemExceedHostMaximum = 18033
	ReasonLicenseSystemNoLicense         = 18034

	ReasonLicenseNoLicense = 18035

	ReasonLicenseInvalidSystemTime = 18036
)

//subscription
const (
	ReasonSubscriptionUserNotStarted   = 19000
	ReasonSubscriptionUserExpired      = 19001
	ReasonSubscriptionDomainNotStarted = 19002
	ReasonSubscriptionDomainExpired    = 19003
)
