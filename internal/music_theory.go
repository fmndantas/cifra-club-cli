package internal

import (
	"fmt"
	"slices"
)

type Note int

const (
	C      = Note(iota)
	CSharp = Note(iota)
	DFlat  = Note(iota)
	D      = Note(iota)
	DSharp = Note(iota)
	EFlat  = Note(iota)
	E      = Note(iota)
	F      = Note(iota)
	FSharp = Note(iota)
	GFlat  = Note(iota)
	G      = Note(iota)
	GSharp = Note(iota)
	AFlat  = Note(iota)
	A      = Note(iota)
	ASharp = Note(iota)
	BFlat  = Note(iota)
	B      = Note(iota)
)

var (
	Notes          = []Note{C, CSharp, DFlat, D, DSharp, EFlat, E, F, FSharp, GFlat, G, GSharp, AFlat, A, ASharp, BFlat, B}
	noteToIdx      = map[Note]int{C: 0, CSharp: 1, DFlat: 1, D: 2, DSharp: 3, EFlat: 3, E: 4, F: 5, FSharp: 6, GFlat: 6, G: 7, GSharp: 8, AFlat: 8, A: 9, ASharp: 10, BFlat: 10, B: 11}
	idxToSharpNote = map[int]Note{0: C, 1: CSharp, 2: D, 3: DSharp, 4: E, 5: F, 6: FSharp, 7: G, 8: GSharp, 9: A, 10: ASharp, 11: B}
	idxToFlatNote  = map[int]Note{0: C, 1: DFlat, 2: D, 3: EFlat, 4: E, 5: F, 6: GFlat, 7: G, 8: AFlat, 9: A, 10: BFlat, 11: B}
)

func (n *Note) String() string {
	switch *n {
	case C:
		return "C"
	case CSharp:
		return "C#"
	case DFlat:
		return "Db"
	case D:
		return "D"
	case DSharp:
		return "D#"
	case E:
		return "E"
	case F:
		return "F"
	case FSharp:
		return "F#"
	case GFlat:
		return "Gb"
	case G:
		return "G"
	case GSharp:
		return "G#"
	case AFlat:
		return "Ab"
	case A:
		return "A"
	case ASharp:
		return "A#"
	case BFlat:
		return "Bb"
	case B:
		return "B"
	default:
		return "?"
	}
}

func (n *Note) Same(another Note) bool {
	return int(*n) == int(another)
}

func (n *Note) ShouldTransposeToSharp() bool {
	yes := []Note{C, CSharp, D, DSharp, E, F, FSharp, G, GSharp, A, ASharp, B}
	return slices.ContainsFunc(yes, n.Same)
}

func (n *Note) Transpose(semitones int) Note {
	// TODO: this makes sense?
	if n == nil {
		return C
	}
	idx := noteToIdx[*n]
	switch {
	case semitones == 0:
		return *n
	case semitones > 0:
		interval := (idx + semitones) % 12
		if n.ShouldTransposeToSharp() {
			return idxToSharpNote[interval]
		} else {
			return idxToFlatNote[interval]
		}
	case semitones < 0:
		interval := idx + semitones
		if interval < 0 {
			interval += 12
		}
		if n.ShouldTransposeToSharp() {
			return idxToSharpNote[interval]
		} else {
			return idxToFlatNote[interval]
		}
	default:
		return *n
	}
}

type Chord struct {
	Root      Note
	Extension string
	Bass      *Note
}

func CreateChord(root Note, extension string) Chord {
	return Chord{Root: root, Extension: extension}
}

func CreateChordWithBass(root Note, extension string, bass Note) Chord {
	return Chord{Root: root, Extension: extension, Bass: &bass}
}

func (c *Chord) HasBass() bool {
	return c.Bass != nil
}

func (c *Chord) String() string {
	if c.HasBass() {
		return fmt.Sprintf("%s%s/%s", c.Root.String(), c.Extension, c.Bass)
	} else {
		return fmt.Sprintf("%s%s", c.Root.String(), c.Extension)
	}
}

func (c *Chord) Transpose(semitones int) Chord {
	var newBass *Note
	if c.HasBass() {
		transposed := c.Bass.Transpose(semitones)
		newBass = &transposed
	}
	return Chord{
		Root:      c.Root.Transpose(semitones),
		Extension: c.Extension,
		Bass:      newBass,
	}
}
