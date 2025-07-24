package dto

type CreateSignatureDeviceRequest struct {
	Algorithm string `json:"algorithm"`
	Label     string `json:"label,omitempty"`
}

type CreateSignatureDeviceResponse struct {
	ID string `json:"id"`
}

type SignTransactionRequest struct {
	Data string `json:"data_to_be_signed"`
}

type SignTransactionResponse struct {
	Signature  string `json:"signature"`
	SignedData string `json:"signed_data"`
}

type GetDeviceResponse struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}
