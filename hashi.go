package main

import (
	"bytes"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

const (
	HASHIDIR = ".hashi"
	PUBDIR   = ".pub"
)

type Vars map[string]string

// renameExt renames extension (if any) from oldext to newext
// If oldext is an empty string - extension is extracted automatically.
// If path has no extension - new extension is appended
func renameExt(path, oldext, newext string) string {
	if oldext == "" {
		oldext = filepath.Ext(path)
	}
	if oldext == "" || strings.HasSuffix(path, oldext) {
		return strings.TrimSuffix(path, oldext) + newext
	} else {
		return path
	}
}

// globals returns list of global OS environment variables that start
// with HASHI prefix as Vars, so the values can be used inside templates
func globals() Vars {
	vars := Vars{}
	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")
		if strings.HasPrefix(pair[0], "HASHI_") {
			vars[strings.ToLower(pair[0][3:])] = pair[1]
		}
	}
	return vars
}

// run executes a command or a script. Vars define the command environment,
// each hashi var is converted into OS environemnt variable with HASHI_ prefix
// prepended.  Additional variable $HASHI contains path to the hashi binary. Command
// stderr is printed to hashi stderr, command output is returned as a string.
func run(vars Vars, cmd string, args ...string) (string, error) {
	// First check if partial exists (.html)
	if b, err := ioutil.ReadFile(filepath.Join(HASHIDIR, cmd+".html")); err == nil {
		return string(b), nil
	}

	var errbuf, outbuf bytes.Buffer
	c := exec.Command(cmd, args...)
	env := []string{"HASHI=" + os.Args[0], "HASHI_OUTDIR=" + PUBDIR}
	env = append(env, os.Environ()...)
	for k, v := range vars {
		env = append(env, "HASHI_"+strings.ToUpper(k)+"="+v)
	}
	c.Env = env
	c.Stdout = &outbuf
	c.Stderr = &errbuf

	err := c.Run()

	if errbuf.Len() > 0 {
		log.Println("ERROR:", errbuf.String())
	}
	if err != nil {
		return "", err
	}
	return string(outbuf.Bytes()), nil
}

// getVars returns list of variables defined in a text file and actual file
// content following the variables declaration. Header is separated from
// content by an empty line. Header can be either YAML or JSON.
// If no empty newline is found - file is treated as content-only.
func getVars(path string, globals Vars) (Vars, string, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	s := string(b)

	// Pick some default values for content-dependent variables
	v := Vars{}
	title := strings.Replace(strings.Replace(path, "_", " ", -1), "-", " ", -1)
	v["title"] = strings.ToTitle(title)
	v["description"] = ""
	v["file"] = path
	v["url"] = path[:len(path)-len(filepath.Ext(path))] + ".html"
	v["output"] = filepath.Join(PUBDIR, v["url"])

	// Override default values with globals
	for name, value := range globals {
		v[name] = value
	}

	// Add layout if none is specified
	if _, ok := v["layout"]; !ok {
		v["layout"] = "layout.html"
	}

	delim := "\n---\n"
	if sep := strings.Index(s, delim); sep == -1 {
		return v, s, nil
	} else {
		header := s[:sep]
		body := s[sep+len(delim):]
		// here we can seek for the title

		vars := Vars{}
		if err := yaml.Unmarshal([]byte(header), &vars); err != nil {
			fmt.Println("ERROR: failed to parse header", err)
			return nil, "", err
		} else {
			// Override default values + globals with the ones defines in the file
			for key, value := range vars {
				v[key] = value
			}
		}
		if strings.HasPrefix(v["url"], "./") {
			v["url"] = v["url"][2:]
		}
		return v, body, nil
	}
}

// Render expanding hashi plugins and variables
func render(s string, vars Vars) (string, error) {
	delim_open := "{{"
	delim_close := "}}"
	out := &bytes.Buffer{}

	//do anchor code here
	//err := pitchTable(&s)
	//	//THIS is added before conversition to HTML
	//	if err != nil {
	//		fmt.Println(err)
	//	}
	for {
		if from := strings.Index(s, delim_open); from == -1 {
			out.WriteString(s)
			return out.String(), nil
		} else {
			if to := strings.Index(s, delim_close); to == -1 {
				return "", fmt.Errorf("Close delim not found")
			} else {
				out.WriteString(s[:from])
				cmd := s[from+len(delim_open) : to]
				s = s[to+len(delim_close):]
				m := strings.Fields(cmd)
				if len(m) == 1 {
					if v, ok := vars[m[0]]; ok {
						out.WriteString(v)
						continue
					}

					if res, err := run(vars, m[0], m[1:]...); err == nil {
						out.WriteString(res)
					} else {
						fmt.Println(err)
					}
				}
			}
		}
		return s, nil
	}
}

// Renders markdown with the given layout into html expanding all the macros
func buildMarkdown(path string, w io.Writer, vars Vars) error {
	v, body, err := getVars(path, vars)
	if err != nil {
		return err
	}
	content, err := render(body, v)
	if err != nil {
		return err
	}
	//	fridayString := string(blackfriday.Run([]byte(content),blackfriday.WithExtensions(blackfriday.CommonExtensions|blackfriday.AutoHeadingIDs)))
	made := Markdown([]byte(content))
	fridayString := string(made[:])
	pitchWriter(&fridayString)
	//this is done right after converting to HTML
	v["content"] = fridayString

	if w == nil {
		out, err := os.Create(filepath.Join(PUBDIR, renameExt(path, "", ".html")))
		if err != nil {
			return err
		}
		defer out.Close()
		w = out
	}
	//	fmt.Println(v["content"])
	//this is to check how your current html looks before building

	return buildHTML(filepath.Join(HASHIDIR, v["layout"]), w, v)
}

// Renders text file expanding all variable macros inside it
func buildHTML(path string, w io.Writer, vars Vars) error {
	v, body, err := getVars(path, vars)
	if err != nil {
		return err
	}
	if body, err = render(body, v); err != nil {
		return err
	}
	tmpl, err := template.New("").Delims("<%", "%>").Parse(body)
	if err != nil {
		return err
	}
	if w == nil {
		f, err := os.Create(filepath.Join(PUBDIR, path))
		if err != nil {
			return err
		}
		defer f.Close()
		w = f
	}
	return tmpl.Execute(w, vars)
}

// Copies file as is from path to writer
func buildRaw(path string, w io.Writer) error {
	in, err := os.Open(path)
	if err != nil {
		return err
	}
	defer in.Close()
	if w == nil {
		if out, err := os.Create(filepath.Join(PUBDIR, path)); err != nil {
			return err
		} else {
			defer out.Close()
			w = out
		}
	}
	_, err = io.Copy(w, in)
	return err
}

func build(path string, w io.Writer, vars Vars) error {
	ext := filepath.Ext(path)
	if ext == ".md" || ext == ".mkd" {
		return buildMarkdown(path, w, vars)
	} else if ext == ".html" || ext == ".xml" {
		return buildHTML(path, w, vars)
	} else {
		return buildRaw(path, w)
	}
}

func buildAll(watch bool) {
	lastModified := time.Unix(0, 0)
	modified := false

	vars := globals()
	for {
		os.Mkdir(PUBDIR, 0755)
		filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
			// ignore hidden files and directories
			if filepath.Base(path)[0] == '.' || strings.HasPrefix(path, ".") {
				return nil
			}
			// inform user about fs walk errors, but continue iteration
			if err != nil {
				fmt.Println("error:", err)
				return nil
			}

			if info.IsDir() {
				os.Mkdir(filepath.Join(PUBDIR, path), 0755)
				return nil
			} else if info.ModTime().After(lastModified) {
				if !modified {
					// First file in this build cycle is about to be modified
					run(vars, "prehook")
					modified = true
				}
				log.Println("build:", path)
				return build(path, nil, vars)
			}
			return nil
		})
		if modified {
			// At least one file in this build cycle has been modified
			run(vars, "posthook")
			modified = false
		}
		if !watch {
			break
		}
		lastModified = time.Now()
		time.Sleep(1 * time.Second)
	}
}

func init() {
	// prepend .hashi to $PATH, so plugins will be found before OS commands
	p := os.Getenv("PATH")
	p = HASHIDIR + ":" + p
	os.Setenv("PATH", p)
}

func main() {
	if len(os.Args) == 1 {
		fmt.Println(os.Args[0], "<command> [args]")
		return
	}
	cmd := os.Args[1]
	args := os.Args[2:]
	switch cmd {
	case "build":
		if len(args) == 0 {
			buildAll(false)
		} else if len(args) == 1 {
			if err := build(args[0], os.Stdout, globals()); err != nil {
				fmt.Println("ERROR: " + err.Error())
			}
		} else {
			fmt.Println("ERROR: too many arguments")
		}
	case "watch":
		buildAll(true)
	case "var":
		if len(args) == 0 {
			fmt.Println("var: filename expected")
		} else {
			s := ""
			if vars, _, err := getVars(args[0], Vars{}); err != nil {
				fmt.Println("var: " + err.Error())
			} else {
				if len(args) > 1 {
					for _, a := range args[1:] {
						s = s + vars[a] + "\n"
					}
				} else {
					for k, v := range vars {
						s = s + k + ":" + v + "\n"
					}
				}
			}
			fmt.Println(strings.TrimSpace(s))
		}
	default:
		if s, err := run(globals(), cmd, args...); err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(s)
		}
	}
}
