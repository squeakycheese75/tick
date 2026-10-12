package cli

import "testing"

func TestBriefIsAHiddenAliasForShortDaily(t *testing.T) {
	t.Parallel()

	brief := newBriefCmd(nil)
	daily := newDailyCmd(nil)

	if !brief.Hidden {
		t.Error("brief should be hidden from help")
	}
	if daily.Hidden {
		t.Error("daily should be visible")
	}
	if brief.Flags().Lookup("short") != nil {
		t.Error("brief is always short, so it has no --short flag")
	}
	if daily.Flags().Lookup("short") == nil {
		t.Error("daily should have --short")
	}
	for _, name := range []string{"portfolio", "news-count", "links", "truncate", "ai"} {
		if brief.Flags().Lookup(name) == nil || daily.Flags().Lookup(name) == nil {
			t.Errorf("both commands should take --%s", name)
		}
	}
}
