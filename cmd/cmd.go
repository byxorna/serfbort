package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"text/template"
)

var sanitizeTagRegexp = regexp.MustCompile(`[^A-Z0-9_]`)

// Runner will handle parameterizing a command with a given argument
// and run it, returning the result and exit code
type Runner struct {
	target         string
	scriptTemplate string
	argument       string
}

// TemplateData is the data available to the command template
type TemplateData struct {
	Argument string
}

// New creates a new command Runner, given a target, script template, and an argument
func New(target, scriptTemplate, arg string) Runner {
	return Runner{
		scriptTemplate: scriptTemplate,
		argument:       arg,
	}
}

// Run will parameterize the script, run it, and return a reader over its output
func (r Runner) Run() (io.Reader, error) {
	tmpl, err := template.New(r.target).Parse(r.scriptTemplate)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Running template %v\n", tmpl)
	buf := new(bytes.Buffer)
	if err := tmpl.Execute(buf, TemplateData{r.argument}); err != nil {
		return nil, err
	}
	script := buf.String()
	fmt.Printf("Command: %q\n", script)

	cmd := exec.Command("/bin/sh", "-c", script)
	cmd.Env = append(os.Environ(),
		"SERFBORT_TARGET="+r.target,
		"SERFBORT_ARGUMENT="+r.argument,
	)

	outputBuf := new(bytes.Buffer)
	cmd.Stderr = outputBuf
	cmd.Stdout = outputBuf

	if err := cmd.Start(); err != nil {
		return outputBuf, err
	}

	err = cmd.Wait()
	if err != nil {
		return outputBuf, err
	}

	return outputBuf, nil
}
