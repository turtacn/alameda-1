package errors

import (
	"fmt"
)

var reasonMap = map[int]string{}

func init() {
	// General
	reasonMap[ReasonInvalidKeyReqPath]     = "Invalid key(%s) in request path."
	reasonMap[ReasonInvalidKeyReqParams]   = "Invalid key(%s) in request parameters."
	reasonMap[ReasonInvalidKeyReqBody]     = "Invalid key(%s) in request body."
	reasonMap[ReasonInvalidValueReqPath]   = "Invalid value of key(%s) in request path."
	reasonMap[ReasonInvalidValueReqParams] = "Invalid value of key(%s) in request parameters."
	reasonMap[ReasonInvalidValueReqBody]   = "Invalid value of key(%s) in request body."
	reasonMap[ReasonInvalidReqFormat]      = "Invalid format of request %s."
	reasonMap[ReasonInvalidJSONFormat]     = "Invalid format of JSON string."
	reasonMap[ReasonMissingFieldReq]       = "Field(%s) is not given in request."
	reasonMap[ReasonIndexOutOfRange]       = "Index out of range."
	reasonMap[ReasonFailedToConnectDB]     = "Failed to connect to database(%s)."
	reasonMap[ReasonFailedToReadDB]        = "Failed to read data from database(%s)."
	reasonMap[ReasonFailedToUpdateDB]      = "Failed to update data to database(%s)."
	reasonMap[ReasonFailedToExecCMD]       = "Failed to execute command(%s)."
	reasonMap[ReasonInvalidParams]         = "Invalid parameters."
	reasonMap[ReasonReservedChar]          = "Character(%s) is reserved in (%s)"
	reasonMap[ReasonInvalidTimeSequence]   = "Key(%s) must be before key(%s)."
	reasonMap[ReasonInvalidTimeRange]      = "Key(%s) and key(%s) must be within %s."

	// Authentication
	reasonMap[ReasonInvalidJWT]        = "Invalid JWT."
	reasonMap[ReasonExpiredJWT]        = "Expired JWT."
	reasonMap[ReasonFailedToGenJWT]    = "Failed to generate JWT."
	reasonMap[ReasonInvalidCredential] = "Invalid credential."

	// Authorization
	reasonMap[ReasonNotAuthorized] = "Not authorized to access this resource/api."

	// Accounts
	reasonMap[ReasonUserCreateFailed] = "Failed to create user(%s)."
	reasonMap[ReasonUserReadFailed]   = "Failed to get user(%s) info."
	reasonMap[ReasonUserUpdateFailed] = "Failed to update user(%s) info."
	reasonMap[ReasonUserDeleteFailed] = "Failed to delete user(%s)."
	reasonMap[ReasonUserNotFound]     = "User(%s) is not found."
	reasonMap[ReasonUserAlreadyExist] = "User(%s) already exists."
	reasonMap[ReasonDomainNotFound]   = "Domain(%s) is not found."

	// Events
	reasonMap[ReasonSystemEventNotFound] = "System event(%s-%s) of host(%s) is not found."

	// Resources
	reasonMap[ReasonAgentNotFound]          = "Agent(%s) is not found."
	reasonMap[ReasonHostNotFound]           = "Host(%s) is not found."
	reasonMap[ReasonPhysicalDiskNotFound]   = "Physical disk(%s) is not found."
	reasonMap[ReasonVirtualMachineNotFound] = "Virtual machine(%s) is not found."

	// Metrics
	reasonMap[ReasonDiskMetricNotFount]    = "Metrics of disk(%s) are not found."
	reasonMap[ReasonHostMetricNotFount]    = "Metrics of host(%s) are not found."
	reasonMap[ReasonInvalidMetricCategory] = "Invalid metric category(%s) for this host."
	reasonMap[ReasonVmMetricNotFount]      = "Metrics of virtual machine(%s) are not found."

	// Predictions
	reasonMap[ReasonDiskPredictionNotFount]    = "Predictions of disk(%s) are not found."
	reasonMap[ReasonHostPredictionNotFount]    = "Predictions of host(%s) are not found."
	reasonMap[ReasonInvalidPredictionCategory] = "Invalid prediction category(%s) for this host."

	// Keycodes
	reasonMap[ReasonKeycodeExpired]                  = "Keycode expired."
	reasonMap[ReasonKeycodeNotFound]                 = "Keycode(%s) is not found."
	reasonMap[ReasonKeycodeNoLicenseFile]            = "No license file."
	reasonMap[ReasonKeycodeIncorrectPadding]         = "Incorrect padding."
	reasonMap[ReasonKeycodeAlreadyApplied]           = "Keycode(%s) is already applied."
	reasonMap[ReasonKeycodeInvalidSystemTime]        = "Invalid system time."
	reasonMap[ReasonKeycodeInvalidSignature]         = "Invalid signature."
	reasonMap[ReasonKeycodeSIDMismatch]              = "SID mismatch."
	reasonMap[ReasonKeycodeDomainMismatch]           = "Domain mismatch."
	reasonMap[ReasonKeycodeInvalidTool]              = "Invalid tool."
	reasonMap[ReasonKeycodeInvalidState]             = "Invalid state."
	reasonMap[ReasonKeycodeInvalidContent]           = "Invalid license file."
	reasonMap[ReasonKeycodeInvalidKeycode]           = "Invalid keycode."
	reasonMap[ReasonKeycodeVersionMismatch]          = "Version mismatch."
	reasonMap[ReasonKeycodeOEMVendorMismatch]        = "OEM vendor mismatch."

	// Licenses
	reasonMap[ReasonLicenseInvalidLicense]           = "Invalid license."
	reasonMap[ReasonLicenseNoLicense]                = "No license."

	reasonMap[ReasonLicenseFailedToConnectServer]    = "Failed to connect to license server."
	reasonMap[ReasonLicenseFailedToReadServer]       = "Failed to read data from license server."
	reasonMap[ReasonLicenseFailedToWriteServer]      = "Failed to write data to license server."

	reasonMap[ReasonLicenseUserExpired]              = "User(%s) license expired."
	reasonMap[ReasonLicenseUserExceedUserMaximum]    = "User(%s) license exceeds the maximum number of users."
	reasonMap[ReasonLicenseUserExceedDiskMaximum]    = "User(%s) license exceeds the maximum number of disks."
	reasonMap[ReasonLicenseUserExceedHostMaximum]    = "User(%s) license exceeds the maximum number of hosts."
	reasonMap[ReasonLicenseUserHasLicense]           = "User(%s) already has a license."
	reasonMap[ReasonLicenseUserNoLicense]            = "User(%s) does not have a license."

	reasonMap[ReasonLicenseDomainExpired]            = "Domain(%s) license expired."
	reasonMap[ReasonLicenseDomainExceedUserMaximum]  = "Domain(%s) license exceeds the maximum number of users."
	reasonMap[ReasonLicenseDomainExceedDiskMaximum]  = "Domain(%s) license exceeds the maximum number of disks."
	reasonMap[ReasonLicenseDomainExceedHostMaximum]  = "Domain(%s) license exceeds the maximum number of hosts."
	reasonMap[ReasonLicenseDomainHasLicense]         = "Domain(%s) already has a license."
	reasonMap[ReasonLicenseDomainNoLicense]          = "Domain(%s) does not have a license."

	reasonMap[ReasonLicenseSystemExpired]            = "System license expired."
	reasonMap[ReasonLicenseSystemInvalid]            = "System license is invalid."
	reasonMap[ReasonLicenseSystemExceedUserMaximum]  = "System license exceeds the maximum number of users."
	reasonMap[ReasonLicenseSystemExceedDiskMaximum]  = "System license exceeds the maximum number of disks."
	reasonMap[ReasonLicenseSystemExceedHostMaximum]  = "System license exceeds the maximum number of hosts."
	reasonMap[ReasonLicenseSystemNoLicense]          = "System dose not have a license."

	//Subscription
	reasonMap[ReasonSubscriptionUserNotStarted]      = "User(%s) subscription is not started."
	reasonMap[ReasonSubscriptionUserExpired]         = "User(%s) subscription expired."
	reasonMap[ReasonSubscriptionDomainNotStarted]    = "Domain(%s) subscription is not started."
	reasonMap[ReasonSubscriptionDomainExpired]       = "Domain(%s) subscription expired."
}

func GetReason(reasonId int, args ...interface{}) string {
	return fmt.Sprintf(reasonMap[reasonId], args...)
}
