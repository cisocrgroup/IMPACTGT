package main

import (
	"fmt"
	"strings"
	"unicode"
)

// glyphs maps custom code points (mostly from the Unicode Private
// Area) to more appropriate unicode strings.
//
// See https://ocr-d.de/de/gt-guidelines/trans/ocr_d_koordinationsgremium_codierung.html#ocr_d_koordinationsgremium_codierung
// or https://folk.uib.no/hnooh/mufi/specs/MUFI-CodeChart-2-0.pdf
// for more information.
var glyphs = map[rune]string{
	0xEBA3: "",             // LATIN SMALL LIGATURE LONG S L
	0xF219: "e",             // LATIN SMALL LETTER E EXTENDED BAR FORM
	0xE781: "",             // LATIN SMALL LETTER Y WITH LATIN SMALL LETTER E ABOVE
	0xEEC4: "ck",            // LATIN SMALL LIGATURE CK
	0xEEC6: "pp",            // LATIN SMALL LIGATURE PP
	0xEEDC: "tz",            // LATIN SMALL LIGATURE TZ
	0xF50D: "q\u0301",       // LATIN SMALL LETTER Q LIGATED WITH FINAL ET AND ACUTE ACCENT
	0xE8BF: "q&",            // LATIN SMALL LETTER Q LIGATED WITH FINAL ET
	0xEBA7: "\u017f\u017fi", // LATIN SMALL LIGATURE LONG S LONG S I ()
	0xF50E: "q\u0301",       // LATIN SMALL LETTER Q WITH ACUTE ACCENT
	0xE5DC: "n\u0304",       // LATIN SMALL LETTER N WITH MEDIUM HIGH MACRON ABOVE
	0xF519: "m\u0303",       // LATIN SMALL LETTER M WITH TILDE
	0xF526: "p\u0309",       // LATIN SMALL LETTER P WITH HOOK ABOVE
	0xF509: "q\u0305",       // LATIN SMALL LETTER Q LIGATED WITH FINAL ET WITH OVERLINE
	0xE6E2: "t\u0301",       // LATIN SMALL LETTER T WITH ACUTE
	0xE682: "q\u0307",       // LATIN SMALL LETTER Q WITH DOT ABOVE ()
	0xE665: "p\u0304",       // LATIN SMALL LETTER P WITH MACRON ABOVE
	0xE681: "q\u0304",       // LATIN SMALL LETTER Q WITH MACRON
	0xEBAC: "sv",            // LATIN SMALL LIGATURE LONG S INSULAR V
	0xEED6: "pp",            // LATIN SMALL LIGATURE PP
	0xEED7: "pp",            // LATIN SMALL LIGATURE PP FLOURISH
	0xF510: "r\u0304",       // LATIN SMALL LETTER R WITH MACRON ABOVE
	0xF50B: "l'",            // LATIN SMALL LETTER L WITH APOSTROPHE
	0xF506: "h\u030a",       // LATIN SMALL LETTER H WITH RING ABOVE
	0xF527: "v\u0309",       // LATIN SMALL LETTER V WITH HOOK ABOVE
	0xF508: "q\u030a",       // LATIN SMALL LETTER Q WITH RING ABOVE
	0xF525: "h\u0309",       // LATIN SMALL LETTER H WITH HOOK ABOVE
	0xF524: "b\u0309",       // LATIN SMALL LETTER B WITH HOOK ABOVE
	0xF504: "g\u030a",       // LATIN SMALL LETTER G WITH RING ABOVE
	0xF511: "s\u0304",       // LATIN SMALL LETTER S WITH MACRON ABOVE
	0xF530: "s\u0308",       // LATIN SMALL LETTER S WITH DIAERESIS
	0xF521: "h",             // LATIN SMALL LETTER H WITH RIGHT DESCENDER AND CURL
	0xF505: "g\u0304",       // LATIN SMALL LETTER G WITH MACRON ABOVE
	0xF501: "c\u0304",       // LATIN SMALL LETTER C WITH MACRON ABOVE
	0xF50F: "q\u0303",       // LATIN SMALL LETTER Q WITH TILDE
	0xF512: "t\u0303",       // LATIN SMALL LETTER T WITH TILDE
	0xF513: "v\u0306",       // LATIN SMALL LETTER V WITH BREVE
	0xF517: "c\u0303",       // LATIN SMALL LETTER C WITH TILDE
	0xF518: "r\u0303",       // LATIN SMALL LETTER R WITH TILDE
	0xF51A: "q",             // LATIN SMALL LETTER Q WITH DIAGONAL STROKE AND DIAERESIS
	0xF51F: "p\u0308",       // LATIN SMALL LETTER P WITH DIAERESIS
	0xF520: "c\u0308",       // LATIN SMALL ABBREVIATION SIGN CON WITH DIAERESIS
	0xF522: "c\u0308",       // LATIN SMALL LETTER C WITH DIAERESIS
	0xF523: "q\u0308",       // LATIN SMALL LETTER Q WITH DIAERESIS
	0xF52F: "q\u0308&",      // LATIN SMALL LETTER Q LIGATED WITH FINAL ET WITH DIAERESIS
	0xF1AC: ";",             // LATIN ABBREVIATION SIGN SEMICOLON
	0xF51B: "d.",            // ABBREVIATION SIGN DER
	0xEEC7: "pp",
	0xE5B8: "m\u0304",
	0xEFA1: "ae",
	0xEADA: "\ufb05",
	0xF4F9: "ll",
	0x00D1: "N\u0303",
	0xEEC5: "ct",
	0xF502: "ch",
	0xE42C: "a\u0364",
	0xE644: "o\u0364",
	0xEBA2: "\u017fi",
	0xE72B: "u\u0364",
	0xEBA6: "\u017f\u017f",
}

// normalize transposes custom unicode points in the ground-truth
// files to appropriate unicode sequences.  If a non-printable code
// cannot be transformed, its hexadecimal representation is used
// instead.
func normalize(str string) string {
	var b strings.Builder
	for _, c := range str {
		if unicode.IsPrint(c) {
			b.WriteRune(c)
			continue
		}
		if repl, ok := glyphs[c]; ok {
			b.WriteString(repl)
			continue
		}
		b.WriteString(fmt.Sprintf("[%4X]", c))
	}
	return b.String()
}
