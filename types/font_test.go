package types

import "testing"

func TestMockFont(t *testing.T) {
	f := &MockFont{
		Widths: map[rune]int{'a': 5, 'b': 6},
		H:      12,
	}

	if f.Height() != 12 {
		t.Errorf("Height() = %d, want 12", f.Height())
	}

	if w := f.Width("ab"); w != 11 {
		t.Errorf("Width(\"ab\") = %d, want 11", w)
	}

	if w := f.Width("xyz"); w != 24 { // 3 * default 8
		t.Errorf("Width(\"xyz\") = %d, want 24", w)
	}

	if w := f.Width(""); w != 0 {
		t.Errorf("Width(\"\") = %d, want 0", w)
	}
}

func TestMockFont_DefaultHeight(t *testing.T) {
	f := &MockFont{}

	if f.Height() != 16 {
		t.Errorf("Height() = %d, want 16 (default)", f.Height())
	}
}
