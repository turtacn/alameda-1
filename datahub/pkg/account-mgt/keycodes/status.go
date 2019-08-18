package keycodes

const (
	KeycodeStatusNoKeycode    = "No Keycode"
	KeycodeStatusInvalid      = "Invalid"
	KeycodeStatusExpired      = "Expired"
	KeycodeStatusNotActivated = "Not Activated"
	KeycodeStatusValid        = "Valid"
	KeycodeStatusUnknown      = "Unknown"
)

type KeycodeStatus struct {
}

func NewKeycodeStatus() *KeycodeStatus {
	keycodeStatus := KeycodeStatus{}
	return &keycodeStatus
}

func (c *KeycodeStatus) GetStatus() string {
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

func (c *KeycodeStatus) isNoKeycode() bool {
	if KeycodeList == nil || len(KeycodeList) == 0 {
		return true
	}
	return false
}

func (c *KeycodeStatus) isInvalid() bool {
	if KeycodeSummary != nil {
		if KeycodeSummary.LicenseState == "Invalid" {
			return true
		}
	}
	return false
}

func (c *KeycodeStatus) isExpired() bool {
	if KeycodeSummary != nil {
		if KeycodeSummary.LicenseState == "Expired" {
			return true
		}
	}
	return false
}

func (c *KeycodeStatus) isNotActivated() bool {
	if KeycodeSummary != nil {
		if KeycodeSummary.KeycodeType == "Regular" && KeycodeSummary.LicenseState == "Valid" {
			if KeycodeSummary.Registered == false {
				return true
			}
		}
	}
	return false
}

func (c *KeycodeStatus) isValid() bool {
	if KeycodeSummary != nil {
		if KeycodeSummary.LicenseState == "Valid" {
			return true
		}
	}
	return false
}
