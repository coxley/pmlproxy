package server

import (
	"testing"
)

var t1 string = `@startuml
Bob -> Alice : hello
@enduml`

var s1 string = "SYWkIImgAStDuNBAJrBGjLDmpCbCJbMmKiX8pSd9vt98pKifpSq11000__y="

func BenchmarkShorten(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ToShort(t1)
	}
}

func BenchmarkExpand(b *testing.B) {
	for n := 0; n < b.N; n++ {
		FromShort(s1)
	}
}

func TestShort(t *testing.T) {
	type test struct {
		source string
		short  string
	}
	tests := []test{
		{t1, s1},
	}

	for _, test := range tests {
		enc, err := ToShort(test.source)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}

		if enc != test.short {
			t.Errorf("expected %s but got %s", test.short, enc)
		}

		dec, err := FromShort(test.short)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}

		if dec != test.source {
			t.Errorf("expected %s but got %s", test.source, dec)
		}
	}
}
func TestExpand(t *testing.T) {
	type test struct {
		short  string
		source string
	}
	tests := []test{
		// Upstream PlantUML compressions results are slightly different -- but
		// compatible. Make sure that we can decode them.
		{"SoWkIImgAStDuNBAJrBGjLDmpCbCJbMmKiX8pSd9vt98pKi1IW80", t1},
	}

	for _, test := range tests {
		dec, err := FromShort(test.short)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}

		if dec != test.source {
			t.Errorf("expected %s but got %s", test.source, dec)
		}
	}
}
