package main

import (
	"strconv"
	"strings"
)

const EMPTY_SPACE = "　"

func pitchWriter(s *string) {

	var body string = ""
	if s != nil {
		//instead just return
		//	return
		body = *s
	}
	var cutUp []string
	// Get rid of all of these indents invert ifs
	if strings.ContainsAny(body, "{}") {
		// replace windows new lines with normal new lines and split the string
		cutUp = strings.Split(strings.ReplaceAll(body, "\r\n", "\n"), "\n")
		for i, _ := range cutUp {
			if strings.ContainsAny(cutUp[i], "{}") {
				cutUp[i] = setPitchNum(cutUp[i])
			}
		}
		*s = strings.Join(cutUp, "\n")
	}
}
func setPitchNum(s string) string {
	parsed := findAllMatch(s)
	var combined string
	for i, e := range parsed {
		if strings.ContainsAny(parsed[i], "{}") {
			insideBracket := match(parsed[i])
			startDel := strings.Index(e, "{")
			endDel := strings.Index(e, "}")
			after := e[endDel+1:]
			word := e[:startDel]
			if len(word) == 0 {
				combined = combined + parsed[i]
				continue
			}
			if insideBracket == "-" {
				var build = []string{EMPTY_SPACE, word}
				var pattern = []uint8{2, 0}
				parsed[i] = buildPitch(build, pattern) + after
			}
			if isNumber(insideBracket) != true {
				combined = combined + parsed[i]
				continue
			}
			takeNum, _ := strconv.Atoi(match(e))
			num := uint8(takeNum)

			moraLength := getMoraLength(word)
			if moraLength >= num {
				pitch, pattern := toPitch(word, num, moraLength)
				parsed[i] = buildPitch(pitch, pattern) + after
			}

		}
		combined = combined + parsed[i]
	}
	return combined
}

func toPitch(s string, pitchNum uint8, moraLength uint8) ([]string, []uint8) {
	// 0 is low, 1 is overline and 2 is a drop
	runes := []rune(s)
	//Heiban
	if pitchNum == 0 {
		return []string{s, EMPTY_SPACE}, []uint8{0, 1}
	}
	//Atamadaka
	if pitchNum == 1 {
		if moraLength == 1 {
			return []string{s}, []uint8{2}
		}
		var pattern = []uint8{2, 0}
		if isYoon(runes[1]) {
			highPart := string(runes[0:2])
			lowPart := string(runes[2:])
			return []string{highPart, lowPart}, pattern
		}
		return []string{string(runes[0]), string(runes[1:])}, pattern
	}
	//Odaka
	var finalIndex uint8 = uint8(len(runes)) - 1
	if pitchNum == moraLength {
		var pattern = []uint8{0, 2}
		if isYoon(runes[finalIndex]) {
			lowPart := string(runes[:finalIndex-1])
			highPart := string(runes[finalIndex-1:])
			return []string{lowPart, highPart}, pattern
		}
		lowPart := string(runes[:finalIndex])
		highPart := string(runes[finalIndex])
		return []string{lowPart, highPart}, pattern
	}
	//Nakadaka
	realNum := getPitchNum(s, pitchNum)
	if isYoon(runes[realNum]) {
		var pattern = []uint8{0, 1, 2, 0}
		start := string(runes[0 : realNum-1])
		overline := string(runes[realNum-1])
		drop := string(runes[realNum])
		after := string(runes[realNum+1:])
		return []string{start, overline, drop, after}, pattern
	}
	var pattern = []uint8{0, 2, 0}
	start := string(runes[0:realNum])
	drop := string(runes[realNum])
	after := string(runes[realNum+1:])
	return []string{start, drop, after}, pattern
}

func getPitchNum(s string, pitchNum uint8) uint8 {
	runes := []rune(s)
	num := pitchNum - 1

	for i := uint8(0); i < num; i++ {
		if isYoon(runes[i]) {
			num++
		}
	}
	if isYoon(runes[num]) {
		num++
	}
	if isYoon(runes[num+1]) {
		num++
	}
	return num
}

func getMoraLength(s string) uint8 {
	runes := []rune(s)
	var count uint8
	for _, element := range runes {
		if !(isYoon(element)) {
			count++
		}
	}
	return count

}

func isDelimeter(r rune) bool {
	//if unicode.IsSpace(r) {
	//	return true
	//	}
	switch r {
	case '<', '>', ',', '\t', '\v', '\f', ' ', 0x85, 0xA0, '・', '　', '、', '(', ')', '（', '[', ']', '［', '］':
		return true
	}

	return false

}

func findAllMatch(s string) []string {
	//splits the string, if its a normal character we keep building up
	// our string, and for delimeters we split them apart, this is a bit
	//faster than regex and also allows multiple delimeters in a row
	// without the pitch being off, which has the added benefit of
	// multiple spaces
	var b strings.Builder
	b.Grow(len(s))
	var build []string
	for _, e := range s {
		if !isDelimeter(e) {
			b.WriteRune(e)
			continue
		}

		if isDelimeter(e) {
			build = append(build, b.String())
			b.Reset()
			build = append(build, string(e))
			continue
		}
	}
	return build

}

func buildPitch(s []string, p []uint8) string {
	builder := strings.Builder{}

	builder.WriteString(`<span class="pitch-accent">`)
	for i, e := range p {
		switch e {
		case 0:
			builder.WriteString(`<span>` + s[i] + `</span>`)
		case 1:
			builder.WriteString(`<span class="overline">` + s[i] + `</span>`)
		case 2:
			builder.WriteString(`<span class="drop">` + s[i] + `</span>`)
		}
	}
	builder.WriteString(`</span>`)

	return builder.String()
}

// This function is not utilized here but left for possible future use, turns a string of japanese text and splits it into groups based on mora
// For example きゃ is grouped together and きや is grouped as two different mora.
func toMora(s string) []string {
	var mora []string
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		if i != len(runes)-1 && isYoon(runes[i]) != true && isYoon(runes[i+1]) {
			mora = append(mora, string(runes[i])+string(runes[i+1]))
			i++
			continue

		}
		mora = append(mora, string(runes[i]))

	}
	return mora
}

func isYoon(r rune) bool {
	switch r {
	case 'ぁ', 'ぃ', 'ぅ', 'ぇ', 'ぉ', 'ゃ', 'ゅ', 'ょ', 'ゎ', 'ァ', 'ィ', 'ゥ', 'ェ', 'ォ', 'ャ', 'ュ', 'ョ', 'ヮ':
		return true
	}

	return false

}

func match(s string) string {
	i := strings.Index(s, "{")
	if i >= 0 {
		j := strings.Index(s, "}")
		if j >= 0 {
			return s[i+1 : j]
		}
	}
	return ""
}

func isNumber(s string) bool {
	if s == "" {
		return false
	}
	_, err := strconv.Atoi(string(s))
	// replace with return err
	if err != nil {
		return false
	}
	return true
}
