package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"time"
	"html/template"

	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"
)

const (
	defaultTemplate = `<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta http-equiv="X-UA-Compatible" content="IE=edge">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<link rel="preconnect" href="https://fonts.googleapis.com">
	<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
	<link href="https://fonts.googleapis.com/css2?family=Poppins&display=swap" rel="stylesheet">
		<title>{{ .Title }}</title>
		<style>
			h1 {
				color: blue
			}
			body {
				font-family: 'Poppins', sans-serif;
			}
		</style>
	</head>
	<body>
		{{ .Body }}
	</body>
	</html>`
)

type content struct {
	Title string
	Body template.HTML
}

func main() {
	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "Mardown Preview Tool, Developed by Nelson Isioma")
		fmt.Fprintln(flag.CommandLine.Output(), "Copyright " + strconv.Itoa(time.Now().Local().Year()) + "\n")
		fmt.Fprintln(flag.CommandLine.Output(), "Usage Information:")
		flag.PrintDefaults()
	}

	fileName := flag.String("file", "", "Markdown file to preview")
	skipPreview := flag.Bool("skip", false, "Skip auto-preview")
	tFName := flag.String("t", "", "Alternate template name")
	flag.Parse()

	if *fileName == "" {
		flag.Usage()
		os.Exit(1)
	}

	if err := run(*fileName, *tFName, os.Stdout, *skipPreview); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(fileName string, tFName string, out io.Writer, skipPreview bool) error {
	input, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}

	htmlData, err := parseContent(input, tFName)
	if err != nil {
		return err
	}

	tempFile, err := os.CreateTemp("", "mdp*.html")
	if err != nil {
		return err
	}

	if err := tempFile.Close(); err != nil {
		return err
	}

	outName := tempFile.Name()
	fmt.Fprintln(out, outName)

	if err := saveHTML(outName, htmlData); err != nil {
		return err
	}

	if skipPreview {
		return nil
	}
	//defer os.Remove(outName)
	return preview(outName)
}

func parseContent(input []byte, tFname string) ([]byte, error) {
	output := blackfriday.Run(input)
	body := bluemonday.UGCPolicy().SanitizeBytes(output)

	t, err := template.New("mdp").Parse(defaultTemplate)

	if err != nil {
		return nil, err
	}

	if tFname != "" {
		t, err = template.ParseFiles(tFname)
		if err != nil {
			return nil, err
		}
	}

	c := content{
		Title: "Markdown Preview Tool",
		Body: template.HTML(body),
	}

	var buffer bytes.Buffer

	if err := t.Execute(&buffer, c); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func saveHTML(outName string, data []byte) error {
	return os.WriteFile(outName, data, 0644)
}

func preview(fileName string) error {
	cName := ""
	cParams := []string {}

	switch runtime.GOOS {
	case "linux":
		cName = "xdg-open"
	case "windows":
		cName = "cmd.exe"
		cParams = []string{"/C", "start"}
	case "darwin":
		cName = "open"
	default: 
		return fmt.Errorf("OS not supported")
	}

	cParams = append(cParams, fileName)
	cPath, err := exec.LookPath(cName)

	if err != nil {
		return err
	}

	err = exec.Command(cPath, cParams...).Run()

	// TODO: Refactor this when you learn how to handle signals
	time.Sleep(2 * time.Second)
	return err
}