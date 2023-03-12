package main

import "strings"
import "strconv"

// REFACTOR toPitch disgusting unreadable code.

func pitchWriter(s *string) error {

	var body string = ""
	if s != nil {
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
	} else {
		return nil

	}
	*s = strings.Join(cutUp, "\n")
	return nil
}
func setPitchNum(s string) string {
	parsed := findAllMatch(s)
	var combined string
	for i, e := range parsed {
		if strings.ContainsAny(parsed[i], "{}") {
			insideBracket := match(parsed[i])
			if isNumber(insideBracket) != true {
				combined = combined + parsed[i]
				continue
			}
			takeNum, _ := strconv.Atoi(match(e))
			num := uint8(takeNum)
			startDel := strings.Index(e, "{")
			endDel := strings.Index(e, "}")
			after := e[endDel+1:]
			word := e[:startDel]
			if len(word) == 0 {
				combined = combined + parsed[i]
				continue

			}

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
	build := make([]string, 0, 4)
	pattern := make([]uint8, 0, 4)
	runes := []rune(s)
	//Heiban and atamadaka
	if pitchNum == 0 || pitchNum == 1 {
		var start, end uint8
		if pitchNum == 0 {
			start = 0
			end = 1
		} else {
			start = 2
			end = 0
		}
		if moraLength == 1 {
			pattern = append(pattern, start)
			build = append(build, s)
			return build, pattern
		}
		pattern = append(pattern, start)
		pattern = append(pattern, end)
		if isYoon(runes[1]) {
			build = append(build, string(runes[0:2]))
			build = append(build, string(runes[2:]))
			return build, pattern
		}
		build = append(build, string(runes[0]))
		build = append(build, string(runes[1:]))
		return build, pattern
	}
	//odaka
	if pitchNum == moraLength {
		if moraLength == 2 {
			pattern = append(pattern, 0)
			pattern = append(pattern, 2)
			if isYoon(runes[1]) {
				build = append(build, string(runes[0:2]))
				build = append(build, string(runes[2:]))
				return build, pattern
			}
			build = append(build, string(runes[0]))
			build = append(build, string(runes[1:]))
			return build, pattern

		}
		end1 := uint8(1)
		start := uint8(len(runes) - 1)
		pattern = append(pattern, 0)
		pattern = append(pattern, 1)
		pattern = append(pattern, 2)
		if isYoon(runes[1]) {
			end1 = end1 + 1
		}
		if isYoon(runes[len(runes)-1]) {
			start = start - 1
		}
		build = append(build, string(runes[0:end1]))
		build = append(build, string(runes[end1:start]))
		build = append(build, string(runes[start:]))
		return build, pattern

	}
	//nakadaka case for words of length 3, they can only have pitch on 2 ifthey are nakadaka or else they fall on another previous pattern
	if pitchNum == 2 {
		pattern = append(pattern, 0)
		pattern = append(pattern, 2)
		pattern = append(pattern, 0)
		end1 := uint8(1)
		if isYoon(runes[1]) {
			end1++
			build = append(build, string(runes[0:end1]))
		} else {
			build = append(build, string(runes[0]))
		}
		if isYoon(runes[end1+1]) {
			build = append(build, string(runes[end1:end1+2]))
			build = append(build, string(runes[end1+2:]))
		} else {
			build = append(build, string(runes[end1]))
			build = append(build, string(runes[end1+1:]))
		}

		return build, pattern

	}
	// nakadaka case others
	pattern = append(pattern, 0)
	pattern = append(pattern, 1)
	pattern = append(pattern, 2)
	pattern = append(pattern, 0)
	end1 := uint8(1)
	if isYoon(runes[1]) {
		end1++
	}
	build = append(build, string(runes[0:end1]))
	realnum := getPitchNum(s, pitchNum)
	build = append(build, string(runes[end1:realnum]))
	build = append(build, string(runes[realnum]))
	build = append(build, string(runes[realnum+1:]))

	return build, pattern
}

func getPitchNum(s string, pitchNum uint8) uint8 {
	runes := []rune(s)
	num := pitchNum - 1

	for i := uint8(0); i < num; i++ {
		if isYoon(runes[i]) {
			num++
		}
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
	case '<', '>', ',', '\t', '\v', '\f', ' ', 0x85, 0xA0, '・', '　', '、', '(', '（':
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
//			builder.WriteString(`<span class="drop">` + s[i] + `</span>`)
			builder.WriteString(`<span class="drop">` + s[i] + `</span><span class="drop-line"></span>`)
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
	if err != nil {
		return false
	}
	return true
}
