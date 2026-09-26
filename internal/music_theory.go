package internal

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
	Notes = []Note{C, CSharp, DFlat, D, DSharp, EFlat, E, F, FSharp, GFlat, G, GSharp, AFlat, A, ASharp, BFlat, B}
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

type Chord struct {
	Root      Note
	Extension string
	Bass      *Note
}

func (c *Chord) HasBass() bool {
	return c.Bass != nil
}

func CreateChord(root Note, extension string) Chord {
	return Chord{Root: root, Extension: extension}
}

func CreateChordWithBass(root Note, extension string, bass Note) Chord {
	return Chord{Root: root, Extension: extension, Bass: &bass}
}
