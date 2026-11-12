package web

import (
	"strings"
	"testing"
)

func TestIconEscapesClassAttribute(t *testing.T) {
	got := string(icon("wallet", `x" onmouseover="alert(1)`))
	if !strings.Contains(got, `class="x&#34; onmouseover=&#34;alert(1)"`) {
		t.Fatalf("class attribute was not escaped: %s", got)
	}
}
