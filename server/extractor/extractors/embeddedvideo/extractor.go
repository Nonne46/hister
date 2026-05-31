// SPDX-License-Identifier: AGPL-3.0-or-later

// Package embeddedvideo extracts embedded video URLs from HTML documents.
package embeddedvideo

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"golang.org/x/net/html"

	"github.com/asciimoo/hister/config"
	"github.com/asciimoo/hister/server/document"
	"github.com/asciimoo/hister/server/types"
)

// knownVideoHosts is the set of hostname substrings that identify an iframe
// src as a video embed. Only the relevant path fragment is checked so that
// both www.youtube.com and youtube.com match.
var knownVideoHosts = []string{
	"https://youtube.com/embed/",
	"https://youtube.com/v/",
	"https://youtu.be/",
	"https://player.vimeo.com/video/",
	"https://vimeo.com/video/",
	"https://dailymotion.com/embed/",
	"https://bitchute.com/embed/",
	"https://rumble.com/embed/",
	"https://twitch.tv/embed/",
	"https://facebook.com/plugins/video",
	"https://instagram.com/p/",
	"https://tiktok.com/embed/",
	"https://ok.ru/videoembed/",
	"https://rutube.ru/play/embed/",
	"https://ted.com/talks/",
	"https://wistia.com/medias/",
	"https://jwplatform.com/players/",
	"https://cdn.jwplayer.com/players/",
	"https://brightcove.net/services/mobile/streaming/index/",
	"https://metacafe.com/embed/",
	"https://streamable.com/e/",
	"https://odysee.com/$/embed/",
	"https://peertube.",
}

// htmlQuickCheck holds the byte strings used for a cheap pre-scan. If none
// of these appear in the raw HTML there is nothing for this extractor to do.
var htmlQuickCheck = []string{"<iframe", "<video", "<embed", "<object"}

// EmbeddedVideoExtractor scans HTML for embedded video elements and stores
// their URLs in d.Metadata["videos"]. It always returns ExtractorContinue so
// the rest of the extractor chain runs normally.
type EmbeddedVideoExtractor struct {
	cfg *config.Extractor
}

func (e *EmbeddedVideoExtractor) Name() string {
	return "EmbeddedVideo"
}

func (e *EmbeddedVideoExtractor) Description() string {
	return "Scans HTML for embedded video tags (iframe, video, embed, object) and stores discovered video URLs in document metadata."
}

func (e *EmbeddedVideoExtractor) GetConfig() *config.Extractor {
	if e.cfg == nil {
		return &config.Extractor{Enable: true, Options: map[string]any{}}
	}
	return e.cfg
}

func (e *EmbeddedVideoExtractor) SetConfig(c *config.Extractor) error {
	for k := range c.Options {
		return fmt.Errorf("unknown option %q", k)
	}
	e.cfg = c
	return nil
}

// Match returns true only when the raw HTML plausibly contains a video
// element, avoiding the tokenizer overhead on pages without any embeds.
func (e *EmbeddedVideoExtractor) Match(d *document.Document) bool {
	if len(d.HTML) == 0 {
		return false
	}
	lower := strings.ToLower(d.HTML)
	for _, needle := range htmlQuickCheck {
		if strings.Contains(lower, needle) {
			return true
		}
	}
	return false
}

// Extract scans d.HTML for video embedding elements and appends any
// discovered URLs to d.Metadata["videos"]. It always returns
// ExtractorContinue so that text-extraction extractors still run.
func (e *EmbeddedVideoExtractor) Extract(d *document.Document) (types.ExtractorState, error) {
	videos := extractVideoURLs(d.HTML)
	if len(videos) > 0 {
		d.AddMetadata("videos", videos)
	}
	return types.ExtractorContinue, nil
}

// Preview does not provide a custom rendering; let the chain continue.
func (e *EmbeddedVideoExtractor) Preview(d *document.Document) (types.PreviewResponse, types.ExtractorState, error) {
	return types.PreviewResponse{}, types.ExtractorContinue, nil
}

// extractVideoURLs tokenizes raw HTML and returns all video embed URLs found
// in iframe/video/source/embed/object elements.
func extractVideoURLs(rawHTML string) []string {
	var videos []string
	seen := make(map[string]struct{})

	add := func(u string) {
		u = strings.TrimSpace(u)
		if u == "" {
			return
		}
		if _, dup := seen[u]; dup {
			return
		}
		seen[u] = struct{}{}
		videos = append(videos, u)
	}

	z := html.NewTokenizer(bytes.NewReader([]byte(rawHTML)))
	inVideo := false // true while inside a <video> element
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			if errors.Is(z.Err(), io.EOF) {
				return videos
			}
			return videos
		case html.StartTagToken, html.SelfClosingTagToken:
			name, hasAttr := z.TagName()
			tag := string(name)
			switch tag {
			case "video":
				inVideo = true
				if src := attrVal(z, hasAttr, "src"); src != "" {
					add(src)
				}
			case "source":
				if inVideo {
					if src := attrVal(z, hasAttr, "src"); src != "" {
						add(src)
					}
				}
			case "iframe":
				if src := attrVal(z, hasAttr, "src"); src != "" && isVideoEmbedURL(src) {
					add(src)
				}
			case "embed":
				if src := attrVal(z, hasAttr, "src"); src != "" && isVideoEmbedURL(src) {
					add(src)
				}
			case "object":
				if data := attrVal(z, hasAttr, "data"); data != "" && isVideoEmbedURL(data) {
					add(data)
				}
			}
		case html.EndTagToken:
			name, _ := z.TagName()
			if string(name) == "video" {
				inVideo = false
			}
		}
	}
}

// attrVal reads all attributes from the current token and returns the value
// of the requested key (case-insensitive), or "" if not present.
func attrVal(z *html.Tokenizer, hasAttr bool, wantKey string) string {
	for hasAttr {
		var k, v []byte
		k, v, hasAttr = z.TagAttr()
		if strings.EqualFold(string(k), wantKey) {
			return string(v)
		}
	}
	return ""
}

// isVideoEmbedURL returns true when the URL belongs to a known video
// hosting / embed service.
func isVideoEmbedURL(u string) bool {
	lower := strings.ToLower(u)
	for _, host := range knownVideoHosts {
		if strings.HasPrefix(lower, host) {
			return true
		}
	}
	return false
}
