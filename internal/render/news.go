package render

import (
	"io"

	"github.com/squeakycheese75/tick/internal/domain"
)

func RenderNewsItem(w io.Writer, r domain.NewsSummary, opts NewsOptions) error {
	out := &writer{w: w}

	out.printf("News for %s\n\n", r.Ticker)

	if len(r.Headlines) == 0 {
		out.println("No recent headlines")
		return out.err
	}

	limit := opts.MaxHeadlines
	if limit <= 0 || limit > len(r.Headlines) {
		limit = len(r.Headlines)
	}

	for i := 0; i < limit; i++ {
		h := r.Headlines[i]
		title := h.Title
		if opts.TruncateTitles {
			title = truncate(title, opts.HeadlineMaxLen)
		}

		out.printf("- %s\n", title)
		if opts.ShowLinks && h.URL != "" {
			out.printf("  🔗 %s\n", h.URL)
		}
		if i < limit-1 {
			out.println("")
		}
	}

	return out.err
}
