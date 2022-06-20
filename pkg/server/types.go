package server

type HandshakeRequestBody struct {
	ReportBase64        string `json:"report_base64"`
	NodePublicKeyBase64 string `json:"node_public_key_base64"`
}

type HandshakeResponseBody struct {
	EncryptedOraclePrivateKeyBase64 string `json:"encrypted_oracle_private_key_base64"`
}
