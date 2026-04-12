package sanitize

import "testing"

func TestStripHTML_RemovesTags(t *testing.T) {
	input := "<script>alert(1)</script>hello"
	got := StripHTML(input)
	want := "hello"

	if got != want {
		t.Fatalf("unexpected stripped value: got %q want %q", got, want)
	}
}

func TestStripHTML_PreservesText(t *testing.T) {
	input := "plain text"
	got := StripHTML(input)

	if got != input {
		t.Fatalf("plain text should be unchanged: got %q want %q", got, input)
	}
}

func TestTruncateString_Truncates(t *testing.T) {
	input := "truncate-me"
	got := TruncateString(input, 8)
	want := "truncate"

	if got != want {
		t.Fatalf("unexpected truncation: got %q want %q", got, want)
	}
}

func TestTruncateString_ShortPassesThrough(t *testing.T) {
	input := "short"
	got := TruncateString(input, 10)

	if got != input {
		t.Fatalf("short string should be unchanged: got %q want %q", got, input)
	}
}

func TestTruncateString_HandlesUnicode(t *testing.T) {
	input := "こんにちは世界"
	got := TruncateString(input, 4)
	want := "こんにち"

	if got != want {
		t.Fatalf("unicode truncation should be rune-aware: got %q want %q", got, want)
	}
}
