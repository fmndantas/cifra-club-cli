package internal_test

import (
	"fmt"
	"testing"

	"github.com/fmndantas/cifraclubcli/internal"
	"github.com/stretchr/testify/assert"
)

func TestNoteSame(t *testing.T) {
	tests := []struct {
		id   int
		a    internal.Note
		b    internal.Note
		want bool
	}{
		{1, internal.C, internal.C, true},
		{2, internal.C, internal.CSharp, false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", tt.id), func(t *testing.T) {
			assert.Equal(t, tt.want, tt.a.Same(tt.b))
		})
	}
}

func TestNoteTransposition(t *testing.T) {
	tests := []struct {
		id             int
		note           internal.Note
		semitones      int
		expectedResult internal.Note
	}{
		{1, internal.C, 0, internal.C},
		{2, internal.C, 1, internal.CSharp},
		{3, internal.C, -1, internal.B},
		{4, internal.C, 12, internal.C},
		{5, internal.C, 13, internal.CSharp},
		{6, internal.AFlat, 2, internal.BFlat},
		{7, internal.GSharp, 2, internal.ASharp},
		{8, internal.AFlat, -2, internal.GFlat},
		{9, internal.GSharp, -2, internal.FSharp},
		{10, internal.GSharp, -1, internal.G},
		{11, internal.FSharp, 6, internal.C},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", tt.id), func(t *testing.T) {
			result := tt.note.Transpose(tt.semitones)
			assert.Equalf(t, tt.expectedResult, result, "expected: %s, got: %s", tt.expectedResult.String(), result.String())
		})
	}
}
