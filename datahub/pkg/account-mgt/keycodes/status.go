package keycodes

const (
	KeycodeStatusUnknown      = 0
	KeycodeStatusNoKeycode    = 1
	KeycodeStatusInvalid      = 2
	KeycodeStatusExpired      = 3
	KeycodeStatusNotActivated = 4
	KeycodeStatusValid        = 5
)

var KeycodeStatusName = map[int]string{
	0: "Unknown",
	1: "No Keycode",
	2: "Invalid",
	3: "Expired",
	4: "Not Activated",
	5: "Valid",
}

var KeycodeStatusMessage = map[int]string{
	0: "Unknown keycode is detected",
	1: "No keycode is applied",
	2: "Invalid keycode is detected",
	3: "Keycode is expired",
	4: "Keycode is not activated",
	5: "A valid keycode is applied",
}

type KeycodeStatusObject struct {
}

func NewKeycodeStatusObject() *KeycodeStatusObject {
	keycodeStatusObject := KeycodeStatusObject{}
	return &keycodeStatusObject
}

func (c *KeycodeStatusObject) GetStatus() int {
	if c.isNoKeycode() {
		return KeycodeStatusNoKeycode
	}
	if c.isInvalid() {
		return KeycodeStatusInvalid
	}
	if c.isExpired() {
		return KeycodeStatusExpired
	}
	if c.isNotActivated() {
		return KeycodeStatusNotActivated
	}
	if c.isValid() {
		return KeycodeStatusValid
	}
	return KeycodeStatusUnknown
}

func (c *KeycodeStatusObject) isNoKeycode() bool {
	if KeycodeList == nil || len(KeycodeList) == 0 {
		return true
	}
	return false
}

func (c *KeycodeStatusObject) isInvalid() bool {
	if KeycodeSummary != nil {
		if KeycodeSummary.LicenseState == "Invalid" {
			return true
		}
	}
	return false
}

func (c *KeycodeStatusObject) isExpired() bool {
	if KeycodeSummary != nil {
		if KeycodeSummary.LicenseState == "Expired" {
			return true
		}
	}
	return false
}

func (c *KeycodeStatusObject) isNotActivated() bool {
	if KeycodeSummary != nil {
		if KeycodeSummary.KeycodeType == "Regular" && KeycodeSummary.LicenseState == "Valid" {
			if KeycodeSummary.Registered == false {
				return true
			}
		}
	}
	return false
}

func (c *KeycodeStatusObject) isValid() bool {
	if KeycodeSummary != nil {
		if KeycodeSummary.LicenseState == "Valid" {
			return true
		}
	}
	return false
}
