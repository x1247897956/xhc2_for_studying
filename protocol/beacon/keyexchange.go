package beacon

// KeyExchangeRequest 是 implant 在加密握手阶段发送的请求。
// Body 包含用服务端 Age 公钥加密后的 32 字节对称密钥(sKey)。
// 除此之外不包含其他信息，字段预留以便后续扩展(如 implant 指纹)。
type KeyExchangeRequest struct {
	EncryptedKey []byte `json:"encrypted_key"`
}

// KeyExchangeResponse 是服务端对握手请求的响应。
// 服务端用响应体确认握手成功，后续通信均使用 sKey 加密。
type KeyExchangeResponse struct {
	OK       bool   `json:"ok"`
	BeaconID string `json:"beacon_id,omitempty"`
	Error    string `json:"error,omitempty"`
}
