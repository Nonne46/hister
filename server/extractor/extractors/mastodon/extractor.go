// SPDX-License-Identifier: AGPL-3.0-or-later

package mastodon

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/asciimoo/hister/config"
	"github.com/asciimoo/hister/server/document"
	"github.com/asciimoo/hister/server/extractor/urlutil"
	"github.com/asciimoo/hister/server/sanitizer"
	"github.com/asciimoo/hister/server/types"

	"github.com/PuerkitoBio/goquery"
	"github.com/rs/zerolog/log"
)

type MastodonExtractor struct {
	cfg *config.Extractor
}

func (e *MastodonExtractor) Name() string {
	return "Mastodon"
}

func (e *MastodonExtractor) Description() string {
	return "Extracts toots as individual documents from Mastodon websites."
}

func (e *MastodonExtractor) GetConfig() *config.Extractor {
	if e.cfg == nil {
		return &config.Extractor{
			Enable:  true,
			Options: map[string]any{},
		}
	}
	return e.cfg
}

func (e *MastodonExtractor) SetConfig(c *config.Extractor) error {
	e.cfg = c
	return nil
}

func (e *MastodonExtractor) Match(d *document.Document) bool {
	if strings.Contains(d.HTML, `"repository":"mastodon/mastodon"`) {
		return true
	}
	if d.Metadata != nil && d.Metadata["type"] == "toot" {
		return true
	}
	return false
}

func (e *MastodonExtractor) Extract(d *document.Document) (types.ExtractorState, error) {
	if d.Metadata != nil && d.Metadata["type"] == "toot" {
		return types.ExtractorStop, nil
	}
	d.SkipIndexing = true
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(d.HTML))
	if err != nil {
		return types.ExtractorContinue, nil
	}

	pu, err := url.Parse(d.URL)
	if err != nil {
		return types.ExtractorContinue, nil
	}

	doc.Find(".status").Each(func(_ int, s *goquery.Selection) {
		c := s.Find(".status__content")
		urlutil.RewriteURLs(c, pu)
		u, exists := s.Find(".status__relative-time").Attr("href")
		if !exists {
			log.Debug().Msg("Failed to find URL for toot")
			return
		}
		h, err := c.Html()
		if err != nil {
			log.Debug().Msg("Failed to extract HTML for toot")
			return
		}
		td := &document.Document{
			URL:    urlutil.ResolveURL(pu, u),
			Title:  "Mastodon toot: " + s.Find(".display-name").Text(),
			Text:   c.Text(),
			HTML:   h,
			UserID: d.UserID,
			Metadata: map[string]any{
				"type": "toot",
			},
		}
		d.ExtraDocuments = append(d.ExtraDocuments, td)
	})

	return types.ExtractorStop, nil
}

func (e *MastodonExtractor) Preview(d *document.Document) (types.PreviewResponse, types.ExtractorState, error) {
	// TODO enhance the toot preview
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(d.HTML))
	if err != nil {
		return types.PreviewResponse{}, types.ExtractorContinue, err
	}

	var b strings.Builder

	// --- Title -----------------------------------------------------------
	if title := strings.TrimSpace(doc.Find("h1").First().Text()); title != "" {
		fmt.Fprintf(&b, "<h2>%s</h2>\n", title)
	}
	b.WriteString(d.HTML)

	// Always sanitize HTML before returning it to strip scripts, event
	// handlers, and other potentially unsafe markup.
	return types.PreviewResponse{
		Content: sanitizer.SanitizeHTML(b.String()),
	}, types.ExtractorStop, nil
}
