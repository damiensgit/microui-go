package atlas

// Atlas dimensions
const (
	AtlasWidth  = 128
	AtlasHeight = 128
)

// Icon indices (matching microui commands.go)
const (
	IconClose     = 1
	IconCheck     = 2
	IconCollapsed = 3
	IconExpanded  = 4
	// IconResize = 5 exists in microui but has no atlas graphic

	AtlasFont = 100 // Base index for font characters (+ ASCII code)
)

// Rect represents a rectangle in the atlas
type Rect struct {
	X, Y, W, H int
}

// AtlasRects maps icon/character codes to atlas positions
var AtlasRects = map[int]Rect{
	IconClose:     {88, 68, 16, 16},
	IconCheck:     {0, 0, 18, 18},
	IconExpanded:  {118, 68, 7, 5},
	IconCollapsed: {113, 68, 5, 7},
	// ASCII printable characters (32-127)
	AtlasFont + 32:  {84, 68, 2, 17},  // space
	AtlasFont + 33:  {39, 68, 3, 17},  // !
	AtlasFont + 34:  {114, 51, 5, 17}, // "
	AtlasFont + 35:  {34, 17, 7, 17},  // #
	AtlasFont + 36:  {28, 34, 6, 17},  // $
	AtlasFont + 37:  {58, 0, 9, 17},   // %
	AtlasFont + 38:  {103, 0, 8, 17},  // &
	AtlasFont + 39:  {86, 68, 2, 17},  // '
	AtlasFont + 40:  {42, 68, 3, 17},  // (
	AtlasFont + 41:  {45, 68, 3, 17},  // )
	AtlasFont + 42:  {34, 34, 6, 17},  // *
	AtlasFont + 43:  {40, 34, 6, 17},  // +
	AtlasFont + 44:  {48, 68, 3, 17},  // ,
	AtlasFont + 45:  {51, 68, 3, 17},  // -
	AtlasFont + 46:  {54, 68, 3, 17},  // .
	AtlasFont + 47:  {124, 34, 4, 17}, // /
	AtlasFont + 48:  {46, 34, 6, 17},  // 0
	AtlasFont + 49:  {52, 34, 6, 17},  // 1
	AtlasFont + 50:  {58, 34, 6, 17},  // 2
	AtlasFont + 51:  {64, 34, 6, 17},  // 3
	AtlasFont + 52:  {70, 34, 6, 17},  // 4
	AtlasFont + 53:  {76, 34, 6, 17},  // 5
	AtlasFont + 54:  {82, 34, 6, 17},  // 6
	AtlasFont + 55:  {88, 34, 6, 17},  // 7
	AtlasFont + 56:  {94, 34, 6, 17},  // 8
	AtlasFont + 57:  {100, 34, 6, 17}, // 9
	AtlasFont + 58:  {57, 68, 3, 17},  // :
	AtlasFont + 59:  {60, 68, 3, 17},  // ;
	AtlasFont + 60:  {106, 34, 6, 17}, // <
	AtlasFont + 61:  {112, 34, 6, 17}, // =
	AtlasFont + 62:  {118, 34, 6, 17}, // >
	AtlasFont + 63:  {119, 51, 5, 17}, // ?
	AtlasFont + 64:  {18, 0, 10, 17},  // @
	AtlasFont + 65:  {41, 17, 7, 17},  // A
	AtlasFont + 66:  {48, 17, 7, 17},  // B
	AtlasFont + 67:  {55, 17, 7, 17},  // C
	AtlasFont + 68:  {111, 0, 8, 17},  // D
	AtlasFont + 69:  {0, 35, 6, 17},   // E
	AtlasFont + 70:  {6, 35, 6, 17},   // F
	AtlasFont + 71:  {119, 0, 8, 17},  // G
	AtlasFont + 72:  {18, 17, 8, 17},  // H
	AtlasFont + 73:  {63, 68, 3, 17},  // I
	AtlasFont + 74:  {66, 68, 3, 17},  // J
	AtlasFont + 75:  {62, 17, 7, 17},  // K
	AtlasFont + 76:  {12, 51, 6, 17},  // L
	AtlasFont + 77:  {28, 0, 10, 17},  // M
	AtlasFont + 78:  {67, 0, 9, 17},   // N
	AtlasFont + 79:  {76, 0, 9, 17},   // O
	AtlasFont + 80:  {69, 17, 7, 17},  // P
	AtlasFont + 81:  {85, 0, 9, 17},   // Q
	AtlasFont + 82:  {76, 17, 7, 17},  // R
	AtlasFont + 83:  {18, 51, 6, 17},  // S
	AtlasFont + 84:  {24, 51, 6, 17},  // T
	AtlasFont + 85:  {26, 17, 8, 17},  // U
	AtlasFont + 86:  {83, 17, 7, 17},  // V
	AtlasFont + 87:  {38, 0, 10, 17},  // W
	AtlasFont + 88:  {90, 17, 7, 17},  // X
	AtlasFont + 89:  {30, 51, 6, 17},  // Y
	AtlasFont + 90:  {36, 51, 6, 17},  // Z
	AtlasFont + 91:  {69, 68, 3, 17},  // [
	AtlasFont + 92:  {124, 51, 4, 17}, // \
	AtlasFont + 93:  {72, 68, 3, 17},  // ]
	AtlasFont + 94:  {42, 51, 6, 17},  // ^
	AtlasFont + 95:  {15, 68, 4, 17},  // _
	AtlasFont + 96:  {48, 51, 6, 17},  // ` (backtick)
	AtlasFont + 97:  {54, 51, 6, 17},  // a
	AtlasFont + 98:  {97, 17, 7, 17},  // b
	AtlasFont + 99:  {0, 52, 5, 17},   // c
	AtlasFont + 100: {104, 17, 7, 17}, // d
	AtlasFont + 101: {60, 51, 6, 17},  // e
	AtlasFont + 102: {19, 68, 4, 17},  // f
	AtlasFont + 103: {66, 51, 6, 17},  // g
	AtlasFont + 104: {111, 17, 7, 17}, // h
	AtlasFont + 105: {75, 68, 3, 17},  // i
	AtlasFont + 106: {78, 68, 3, 17},  // j
	AtlasFont + 107: {72, 51, 6, 17},  // k
	AtlasFont + 108: {81, 68, 3, 17},  // l
	AtlasFont + 109: {48, 0, 10, 17},  // m
	AtlasFont + 110: {118, 17, 7, 17}, // n
	AtlasFont + 111: {0, 18, 7, 17},   // o
	AtlasFont + 112: {7, 18, 7, 17},   // p
	AtlasFont + 113: {14, 34, 7, 17},  // q
	AtlasFont + 114: {23, 68, 4, 17},  // r
	AtlasFont + 115: {5, 52, 5, 17},   // s
	AtlasFont + 116: {27, 68, 4, 17},  // t
	AtlasFont + 117: {21, 34, 7, 17},  // u
	AtlasFont + 118: {78, 51, 6, 17},  // v
	AtlasFont + 119: {94, 0, 9, 17},   // w
	AtlasFont + 120: {84, 51, 6, 17},  // x
	AtlasFont + 121: {90, 51, 6, 17},  // y
	AtlasFont + 122: {10, 68, 5, 17},  // z
	AtlasFont + 123: {31, 68, 4, 17},  // {
	AtlasFont + 124: {96, 51, 6, 17},  // |
	AtlasFont + 125: {35, 68, 4, 17},  // }
	AtlasFont + 126: {102, 51, 6, 17}, // ~
	AtlasFont + 127: {108, 51, 6, 17}, // DEL (placeholder)
}
