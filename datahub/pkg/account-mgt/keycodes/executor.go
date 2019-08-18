package keycodes

import (
	"encoding/base64"
	"encoding/json"
	Errors "github.com/containers-ai/alameda/internal/pkg/errors"
	"os/exec"
	"strings"
)

type Result struct {
	Status int    `json:"status"`
	Reason string `json:"reason"`
}

type AllKeycodesResult struct {
	Result
	Data    []*Keycode `json:"data"`
	Summary *Keycode   `json:"summary"`
}

type KeycodeResult struct {
	Result
	Data *Keycode `json:"data"`
}

type RegistrationDataResult struct {
	Result
	Data string `json:"data"`
}

type KeycodeExecutor struct {
	Cli      string
	LdapArgs []string
}

func NewKeycodeExecutor() *KeycodeExecutor {
	keycodeCli := KeycodeExecutor{}
	keycodeCli.Cli = KeycodeCliPath
	keycodeCli.LdapArgs = make([]string, 0)

	if LdapConfig != nil {
		keycodeCli.LdapArgs = append(keycodeCli.LdapArgs,
			"--ldap-portal", LdapConfig.Address,
			"--ldap-base-dn", LdapConfig.BaseDN,
			"--ldap-user", LdapConfig.AdminID,
			"--ldap-password", LdapConfig.AdminPW,
			"--encode-key", base64.StdEncoding.EncodeToString(KeycodeAesKey),
		)
	}

	return &keycodeCli
}

func (c *KeycodeExecutor) AddKeycode(keycode string) error {
	result := KeycodeResult{}

	// Prepend args
	args := make([]string, 0)
	args = append([]string{"--add-keycode", keycode}, c.LdapArgs...)

	cmd := exec.Command(c.Cli, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		scope.Error(string(out))
		return Errors.NewError(Errors.ReasonFailedToExecCMD, "keycode")
	}

	stripped := strings.TrimSuffix(string(out), "\n")
	stringList := strings.Split(string(stripped), "\n")

	err = json.Unmarshal([]byte(stringList[len(stringList)-1]), &result)
	if err != nil {
		scope.Error(err.Error())
		return Errors.NewError(Errors.ReasonInvalidJSONFormat)
	}

	if result.Status != 0 {
		scope.Error(result.Reason)
		err := c.translateErrorId(result.Status)
		if err.(*Errors.InternalError).ErrorID == Errors.ReasonKeycodeAlreadyApplied {
			err.(*Errors.InternalError).Append(strings.Fields(result.Reason)[2])
		}
		return err
	}

	return nil
}

func (c *KeycodeExecutor) DeleteKeycode(keycode string) error {
	result := Result{}

	// Prepend args
	args := make([]string, 0)
	args = append([]string{"--delete-keycode", keycode}, c.LdapArgs...)

	cmd := exec.Command(c.Cli, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		scope.Error(string(out))
		return err
	}

	stripped := strings.TrimSuffix(string(out), "\n")
	stringList := strings.Split(string(stripped), "\n")

	err = json.Unmarshal([]byte(stringList[len(stringList)-1]), &result)
	if err != nil {
		scope.Error(err.Error())
		return Errors.NewError(Errors.ReasonInvalidJSONFormat)
	}

	if result.Status != 0 {
		scope.Error(result.Reason)
		return c.translateErrorId(result.Status)
	}

	return nil
}

func (c *KeycodeExecutor) GetKeycode(keycode string) (*Keycode, error) {
	result := KeycodeResult{}

	cmd := exec.Command(c.Cli, "--get-keycode", keycode)
	out, err := cmd.CombinedOutput()
	if err != nil {
		scope.Error(string(out))
		return result.Data, Errors.NewError(Errors.ReasonFailedToExecCMD, "keycode")
	}

	stripped := strings.TrimSuffix(string(out), "\n")
	stringList := strings.Split(string(stripped), "\n")

	err = json.Unmarshal([]byte(stringList[len(stringList)-1]), &result)
	if err != nil {
		scope.Error(err.Error())
		return result.Data, Errors.NewError(Errors.ReasonInvalidJSONFormat)
	}

	if result.Status != 0 {
		scope.Error(result.Reason)
		return nil, c.translateErrorId(result.Status)
	}

	return result.Data, nil
}

func (c *KeycodeExecutor) GetKeycodeSummary() (*Keycode, error) {
	result := KeycodeResult{}

	cmd := exec.Command(c.Cli, "--get-keycode-summary")
	out, err := cmd.CombinedOutput()
	if err != nil {
		scope.Error(string(out))
		return result.Data, Errors.NewError(Errors.ReasonFailedToExecCMD, "keycode")
	}

	stripped := strings.TrimSuffix(string(out), "\n")
	stringList := strings.Split(string(stripped), "\n")

	err = json.Unmarshal([]byte(stringList[len(stringList)-1]), &result)
	if err != nil {
		scope.Error(err.Error())
		return result.Data, Errors.NewError(Errors.ReasonInvalidJSONFormat)
	}

	if result.Status != 0 {
		if result.Status == 2 {
			return nil, nil
		} else {
			scope.Error(result.Reason)
			return nil, c.translateErrorId(result.Status)
		}
	}

	return result.Data, nil
}

func (c *KeycodeExecutor) GetAllKeycodes() ([]*Keycode, *Keycode, error) {
	result := AllKeycodesResult{}

	cmd := exec.Command(c.Cli, "--get-all-keycodes")
	out, err := cmd.CombinedOutput()
	if err != nil {
		scope.Error(string(out))
		return make([]*Keycode, 0), nil, err
	}

	stripped := strings.TrimSuffix(string(out), "\n")
	stringList := strings.Split(string(stripped), "\n")

	err = json.Unmarshal([]byte(stringList[len(stringList)-1]), &result)
	if err != nil {
		scope.Error(err.Error())
		return make([]*Keycode, 0), nil, Errors.NewError(Errors.ReasonInvalidJSONFormat)
	}

	if result.Status != 0 {
		if result.Status == 2 {
			return make([]*Keycode, 0), &Keycode{LicenseState: "Invalid"}, nil
		} else {
			scope.Error(result.Reason)
			return make([]*Keycode, 0), nil, c.translateErrorId(result.Status)
		}
	}

	return result.Data, result.Summary, nil
}

func (c *KeycodeExecutor) GetRegistrationData() (string, error) {
	result := RegistrationDataResult{}

	cmd := exec.Command(c.Cli, "--get-registration-data")
	out, err := cmd.CombinedOutput()
	if err != nil {
		scope.Error(string(out))
		return "", err
	}

	stripped := strings.TrimSuffix(string(out), "\n")
	stringList := strings.Split(string(stripped), "\n")

	err = json.Unmarshal([]byte(stringList[len(stringList)-1]), &result)
	if err != nil {
		scope.Error(err.Error())
		return "", Errors.NewError(Errors.ReasonInvalidJSONFormat)
	}

	if result.Status != 0 {
		scope.Error(result.Reason)
		return "", c.translateErrorId(result.Status)
	}

	return result.Data, nil
}

func (c *KeycodeExecutor) PutSignatureData(signatureData string) error {
	result := Result{}

	// Prepend args
	args := make([]string, 0)
	args = append([]string{"--put-signature-data", signatureData}, c.LdapArgs...)

	cmd := exec.Command(c.Cli, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		scope.Error(string(out))
		return Errors.NewError(Errors.ReasonFailedToExecCMD, "keycode")
	}

	stripped := strings.TrimSuffix(string(out), "\n")
	stringList := strings.Split(string(stripped), "\n")

	err = json.Unmarshal([]byte(stringList[len(stringList)-1]), &result)
	if err != nil {
		scope.Error(err.Error())
		return Errors.NewError(Errors.ReasonInvalidJSONFormat)
	}

	if result.Status != 0 {
		scope.Error(result.Reason)
		return c.translateErrorId(result.Status)
	}

	return nil
}

func (c *KeycodeExecutor) PutSignatureDataFile(filePath string) error {
	result := Result{}

	// Prepend args
	args := make([]string, 0)
	args = append([]string{"--put-signature-data-file", filePath}, c.LdapArgs...)

	cmd := exec.Command(c.Cli, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		scope.Error(string(out))
		return Errors.NewError(Errors.ReasonFailedToExecCMD, "keycode")
	}

	stripped := strings.TrimSuffix(string(out), "\n")
	stringList := strings.Split(string(stripped), "\n")

	err = json.Unmarshal([]byte(stringList[len(stringList)-1]), &result)
	if err != nil {
		scope.Error(err.Error())
		return Errors.NewError(Errors.ReasonInvalidJSONFormat)
	}

	if result.Status != 0 {
		scope.Error(result.Reason)
		return c.translateErrorId(result.Status)
	}

	return nil
}

func (c *KeycodeExecutor) translateErrorId(errorId int) error {
	if errorId == 2 {
		return Errors.NewError(Errors.ReasonKeycodeNoLicenseFile)
	}
	if errorId == 14 {
		return Errors.NewError(Errors.ReasonKeycodeIncorrectPadding)
	}
	if errorId == 17 {
		return Errors.NewError(Errors.ReasonKeycodeAlreadyApplied)
	}
	if errorId == 62 {
		return Errors.NewError(Errors.ReasonKeycodeInvalidSystemTime)
	}
	if errorId == 10001 {
		return Errors.NewError(Errors.ReasonKeycodeInvalidSignature)
	}
	if errorId == 10002 {
		return Errors.NewError(Errors.ReasonKeycodeSIDMismatch)
	}
	if errorId == 10003 {
		return Errors.NewError(Errors.ReasonKeycodeDomainMismatch)
	}
	if errorId == 10004 {
		return Errors.NewError(Errors.ReasonKeycodeInvalidTool)
	}
	if errorId == 10005 {
		return Errors.NewError(Errors.ReasonKeycodeInvalidState)
	}
	if errorId == 10006 {
		return Errors.NewError(Errors.ReasonKeycodeInvalidContent)
	}
	if errorId == 10007 {
		return Errors.NewError(Errors.ReasonKeycodeInvalidKeycode)
	}
	if errorId == 10008 {
		return Errors.NewError(Errors.ReasonKeycodeVersionMismatch)
	}
	if errorId == 10009 {
		return Errors.NewError(Errors.ReasonKeycodeOEMVendorMismatch)
	}

	return Errors.NewError(errorId)
}
