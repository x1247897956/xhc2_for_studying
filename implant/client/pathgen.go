package client

import (
	"fmt"
	"math/rand/v2"
	"net/url"

	"xhc2_for_studying/protocol"
)

// randomPath 从 C2Profile 的候选池中生成一次随机 URL 路径。
// ext 指定文件扩展名（如 "php"、"js"），由消息类型决定。
//
// 步骤:
//  1. 随机决定目录段数量 n ∈ [MinPathLength, MaxPathLength]
//  2. 从 PathSegments 中分离出目录段和文件段
//  3. 随机选 n 个目录段
//  4. 随机选 1 个文件名 + ext → 拼接成 filename.ext
//  5. 返回 [dir1, dir2, ..., filename.ext]
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
		minLen := p.MinPathLength
		maxLen := p.MaxPathLength
		n := minLen + rand.IntN(maxLen-minLen+1)
		for range n {
			segments = append(segments, dirs[rand.IntN(len(dirs))])
		}
	}

	if len(files) > 0 {
		filename := files[rand.IntN(len(files))]
		if ext != "" {
			filename = filename + ext // ext 本身带点，如 ".php"
		}
		segments = append(segments, filename)
	}

	return segments
}

// buildRandomURL 生成完整的随机 URL。
// baseURL 是 Server 根地址，p 是 C2 配置，ext 是消息类型对应的扩展名。
func buildRandomURL(baseURL string, p *protocol.C2Profile, ext string) (*url.URL, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse base url: %w", err)
	}

	segments := randomPath(p, ext)
	parsed = parsed.JoinPath(segments...)

	return parsed, nil
}

// embedNonce 按 NonceMode 将 nonce 嵌入 URL。
func embedNonce(u *url.URL, nonce int, nonceMode string) {
	switch nonceMode {
	case protocol.NonceModeURL:
		// 把 nonce 插在路径的倒数第二段（文件段之前）
		u.Path = insertNonceInPath(u.Path, nonce)
	case protocol.NonceModeURLParam:
		fallthrough
	default:
		q := u.Query()
		q.Set("_", fmt.Sprintf("%d", nonce))
		u.RawQuery = q.Encode()
	}
}

// insertNonceInPath 在路径的最后一段（文件段）之前插入 nonce。
// 例: /api/assets/chunk.js → /api/assets/4729183/chunk.js
func insertNonceInPath(path string, nonce int) string {
	parts := splitPath(path)
	if len(parts) < 2 {
		return path
	}
	// 在倒数第一段（文件名）之前插入 nonce
	last := parts[len(parts)-1]
	rest := parts[:len(parts)-1]
	rest = append(rest, fmt.Sprintf("%d", nonce))
	rest = append(rest, last)

	result := ""
	for _, p := range rest {
		result += "/" + p
	}
	return result
}

func splitPath(path string) []string {
	var parts []string
	for _, p := range splitSlash(path) {
		if p != "" {
			parts = append(parts, p)
		}
	}
	return parts
}

func splitSlash(s string) []string {
	var parts []string
	start := 0
	for i := range s {
		if s[i] == '/' {
			if i > start {
				parts = append(parts, s[start:i])
			}
			start = i + 1
		}
	}
	if start < len(s) {
		parts = append(parts, s[start:])
	}
	return parts
}
