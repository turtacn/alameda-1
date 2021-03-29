package keycodes

import (
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
}

func NewKeycodeExecutor() *KeycodeExecutor {
	keycodeCli := KeycodeExecutor{}
	return &keycodeCli
}

func (c *KeycodeExecutor) AddKeycode(keycode string) error {
	return nil
}

func (c *KeycodeExecutor) DeleteKeycode(keycode string) error {
	return nil
}

func (c *KeycodeExecutor) GetKeycode(keycode string) (*Keycode, error) {
	result := KeycodeResult{}


	return result.Data, nil
}

func (c *KeycodeExecutor) GetKeycodeSummary() (*Keycode, error) {
	result := KeycodeResult{}
	return result.Data, nil
}

func (c *KeycodeExecutor) GetAllKeycodes() ([]*Keycode, *Keycode, error) {
	result := AllKeycodesResult{}
	return result.Data, result.Summary, nil
}

func (c *KeycodeExecutor) GetRegistrationData() (string, error) {
	result := RegistrationDataResult{}
	return result.Data, nil
}

func (c *KeycodeExecutor) PutSignatureData(signatureData string) error {
	return nil
}

func (c *KeycodeExecutor) PutSignatureDataFile(filePath string) error {
	return nil
}