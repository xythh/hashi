package main

import (
	"bytes"
	"encoding/csv"
	"errors"
	"log"
	"strings"
	"text/template"
)

type Line struct {
	Name, Category, Reference, Description string
	Number                                 int
}

const tableFormat = `<tr id="row{{.Number}}" class="rows"><th class="{{.Category}}"><a href="{{.Reference}}" id="{{.Number}}" class="links">{{.Name}}</a></th><td><div class="container">{{.Description}}</div><div id="entry{{.Number}}"></div></tr>`

// takes a string pointer, checks if pitchTable is there if it is parse if not dont alter string and return error if delims dont match
// error with using other {{}}
func pitchTable(s *string) error {
	var body string = ""
	if s != nil {
		body = *s
	}

	//	var body string = *s
	const delimOpen = "{{[pitchTable}}"
	const delimClose = "{{]pitchTable}}"
	var missingClose, missingOpen bool
	from := strings.Index(body, delimOpen)
	to := strings.Index(body, delimClose)

	if from == -1 {
		missingOpen = true
	}
	if to == -1 {
		missingClose = true
	}
	if missingOpen && missingClose {
		return nil

	}
	if missingOpen || missingClose {
		return errors.New("Mismatch, did not find both opening and closing tag")

	}

	before := body[:from]
	after := body[to+len(delimClose):]

	inside := body[from+len(delimOpen) : to]

	r := csv.NewReader(strings.NewReader(inside))
	r.Comment = '#'
	r.TrimLeadingSpace = false
	records, err := r.ReadAll()
	if err != nil {
		return err
	}
	var fixed []Line
	t := template.Must(template.New("table").Parse(tableFormat))
	for i, record := range records {
		e := Line{record[0], record[1], record[2], record[3], i + 1}
		fixed = append(fixed, e)
	}
	var b bytes.Buffer
	for _, r := range fixed {
		err := t.Execute(&b, &r)
		if err != nil {
			log.Println("executing template:", err)
		}
	}
	var table string = `
<table id="toptable"><tbody><tr><th id="concept">Concept <span id="results"></span></th><th>Notes</th></tr></tbody><tbody id="tablebody">
` + b.String() + `</tbody></table>`
	*s = before + table + after

	return nil
}
