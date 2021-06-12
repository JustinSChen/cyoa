package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/template"
)

type Chapter struct {
	Title   string
	Story   []string
	Options []Option
}

type Option struct {
	Text string
	Arc  string
}

// Handler for web version
type webHandler struct {
	chapters map[string]Chapter
	template  *template.Template
}

// Handler for Web version
func (wh webHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var path string
	if req.URL.Path == "" || req.URL.Path == "/" {
		wh.template.Execute(w, wh.chapters["intro"])
	} else {
		path = strings.TrimLeft(req.URL.Path, "/")
		wh.template.Execute(w, wh.chapters[path])
	}
}

// Reader for cli version
type cliReader struct {
	chapters 	map[string]Chapter
	reader    	*bufio.Reader
}

func (cr cliReader) getOption(options []Option) Option {
	input, err := cr.reader.ReadString('\n')
	if err != nil {
		fmt.Println("Didn't catch that. Please try again...")
		return cr.getOption(options)
	}

	// Check if option is legal
	selected, err := strconv.Atoi(strings.TrimRight(input, "\n"));
	if selected <= 0 || selected > len(options) || err != nil {
		fmt.Println("That wasn't one of the choices! Please try again...")
		return cr.getOption(options)
	}

	return options[selected-1]
}

func (cr cliReader) showChapter(chapterName string) {
	chapter := cr.chapters[chapterName]

	// End story if no options available
	if len(chapter.Options) == 0 {
		fmt.Printf("***** The End *****\n")
		os.Exit(0)
	}

	fmt.Printf("\n***** %v *****\n\n", chapter.Title)

	for _, p := range chapter.Story {
		fmt.Printf("%v\n", p)
	}

	fmt.Printf("\nYou decide to: \n")
	for i, opt := range chapter.Options {
		fmt.Printf("%d: %s\n", i+1, opt.Text)
	}

	option := cr.getOption(chapter.Options)
	cr.showChapter(option.Arc)
}

func main() {
	// Create optional flags
	cli := flag.Bool("cli", false, "To use this web app via CLI or via web")
	flag.Parse()

	// Open the json file and parse the story within in
	jsonFile, err := os.Open("gopher.json")
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()

	jsonBytes, _ := ioutil.ReadAll(jsonFile)
	var parsedStory map[string]Chapter
	err = json.Unmarshal(jsonBytes, &parsedStory)
	if err != nil {
		panic(err)
	}

	// CLI version or Web version
	if *cli {
		s := cliReader{
			chapters: parsedStory,
			reader: bufio.NewReader(os.Stdin),
		}
		s.showChapter("intro")
	} else {
		tpl := template.Must(template.ParseFiles("cyoa.html"))
		fmt.Println("Open browser to localhost:8080 to play.")
		http.ListenAndServe(":8080", webHandler{
			chapters: parsedStory,
			template: tpl,
		})
	}
}