package client

import (
	"fmt"
	"math/rand/v2"
	"net/url"
	"strconv"
	"strings"

	"xhc2_for_studying/protocol"
)

// randomPath generates a single random URL path from the C2Profile candidate pool.
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
		if n > len(dirs) {
			n = len(dirs)
		}
		dirPool := append([]string(nil), dirs...)
		for range n {
			idx := rand.IntN(len(dirPool))
			segments = append(segments, dirPool[idx])
			dirPool = append(dirPool[:idx], dirPool[idx+1:]...)
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

// buildRandomURL constructs a complete random URL from the base URL, fixed
// prefix, and randomized profile path.
func buildRandomURL(baseURL, pathPrefix string, p *protocol.C2Profile, ext string) (*url.URL, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse base url: %w", err)
	}
	segments := append(prefixSegments(pathPrefix), randomPath(p, ext)...)
	parsed = parsed.JoinPath(segments...)
	if parsed.Host != "" && parsed.Path != "" && !strings.HasPrefix(parsed.Path, "/") {
		parsed.Path = "/" + parsed.Path
	}
	return parsed, nil
}

func prefixSegments(pathPrefix string) []string {
	pathPrefix = strings.Trim(pathPrefix, "/")
	if pathPrefix == "" {
		return nil
	}
	return strings.Split(pathPrefix, "/")
}

// embedEncoderNonce embeds the encoder negotiation nonce into the URL query
// parameter "?_=<integer>". The server extracts the encoder ID via
// nonce % EncoderModulus.
func embedEncoderNonce(u *url.URL, nonce int) {
	q := u.Query()
	q.Set("_", strconv.Itoa(nonce))
	u.RawQuery = q.Encode()
}
