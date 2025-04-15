package glog

type Validator interface {
	Validate(keys *Keys) bool
}

type SimpleValidator struct {
	apiKey    string
	apiSecret string

	Validator
}

func (sp *SimpleValidator) Validate(keys *Keys) bool {
	return sp.apiKey == keys.apiKey && sp.apiSecret == keys.apiSecret
	//return keys.apiKey == "231a9c79-a227-469e-b6f4-78235569f557" && keys.apiSecret == "92aa03df-e789-46a7-9485-facf38e0c8296843bf7a-3189-45c7-bee8-062791eb2b3a"
}
