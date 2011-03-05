package main

import (
	"os"
	"path"
	"fmt"
	"regexp"
	"bytes"
	goopt "github.com/droundy/goopt"
	"./highlight"
)

var byteNewLine []byte = []byte("\n")
// Used to prevent appear of sparse newline at the end of output
var prependNewLine = false

type StringList []string
var IgnoreDirs = StringList{"autom4te.cache", "blib", "_build", ".bzr", ".cdv",
	"cover_db", "CVS", "_darcs", "~.dep", "~.dot", ".git", ".hg", "~.nib",
    ".pc", "~.plst", "RCS", "SCCS", "_sgbak", ".svn"}

type RegexpList []*regexp.Regexp
var IgnoreFiles = regexpList([]string{`~$`, `#.+#$`, `[._].*\.swp$`, `core\.[0-9]+$`,
	`\.pyc$`, `\.o$`, `\.6$`})

var onlyName = goopt.Flag([]string{"-n", "--filename"}, []string{},
	"print only filenames", "")
var ignoreFiles = goopt.Strings([]string{"-x", "--exclude"}, "RE",
	"exclude files that match the regexp from search")

func main() {
	goopt.Description = func() string {
		return "Go search and replace in files"
	}
	goopt.Version = "0.1"
	goopt.Parse(nil)

	if len(goopt.Args) == 0 {
		println(goopt.Usage())
		return
	}

	IgnoreFiles = append(IgnoreFiles, regexpList(*ignoreFiles)...)

	pattern, err := regexp.Compile(goopt.Args[0])
	errhandle(err, "can't compile regexp %s", goopt.Args[0])

	searchFiles(pattern)
}

func errhandle(err os.Error, moreinfo string, a ...interface{}) {
	if err == nil {
		return
	}
	fmt.Fprintf(os.Stderr, "ERR %s\n%s\n", err,
		fmt.Sprintf(moreinfo, a...))
	os.Exit(1)
}

func regexpList(sa []string) RegexpList {
	ra := make(RegexpList, len(sa))
	for i, s := range sa {
		ra[i] = regexp.MustCompile(s)
	}
	return ra
}

func searchFiles(pattern *regexp.Regexp) {
	v := &GRVisitor{pattern}

	errors := make(chan os.Error, 64)

	path.Walk(".", v, errors)

	select {
	case err := <-errors:
		errhandle(err, "some error")
	default:
	}
}

type GRVisitor struct{
	pattern *regexp.Regexp
}

func (v *GRVisitor) VisitDir(fn string, fi *os.FileInfo) bool {
	if IgnoreDirs.Contains(fi.Name) {
		return false
	}
	return true
}

func (v *GRVisitor) VisitFile(fn string, fi *os.FileInfo) {
	if IgnoreFiles.Match(fn) {
		return
	}

	if fi.Size >= 1024*1024*10 {
		fmt.Fprintf(os.Stderr, "Skipping %s, too big: %d\n", fn, fi.Size)
		return
	}

	if fi.Size == 0 {
		return
	}

	f, err := os.Open(fn, os.O_RDONLY, 0666)
	errhandle(err, "can't open file %s", fn)

	content := make([]byte, fi.Size)
	n, err := f.Read(content)
	errhandle(err, "can't read file %s", fn)
	if int64(n) != fi.Size {
		panic(fmt.Sprintf("Not whole file was read, only %d from %d",
			n, fi.Size))
	}

	v.SearchFile(fn, content)

	f.Close()
}

func (v *GRVisitor) SearchFile(p string, content []byte) {
	linenum := 1
	last := 0
	hadOutput := false
	binary := false

	if bytes.IndexByte(content, 0) != -1 {
		binary = true
	}

	for _, bounds := range v.pattern.FindAllIndex(content, -1) {
		if prependNewLine {
			fmt.Println("")
			prependNewLine = false
		}

		if !hadOutput {
			hadOutput = true
			if binary && !*onlyName{
				fmt.Printf("Binary file %s matches\n", p)
				break
			} else {
				highlight.Printf("green", "%s\n", p)
			}
		}

		if *onlyName {
			return
		}

		linenum += bytes.Count(content[last:bounds[0]], byteNewLine)
		last = bounds[0]
		begin, end := beginend(content, bounds[0], bounds[1])

		if content[begin] == '\r' {
			begin += 1
		}

		highlight.Printf("bold yellow", "%d:", linenum)
		highlight.Reprintf("on_yellow", v.pattern, "%s\n", content[begin:end])
	}

	if hadOutput {
		prependNewLine = true
	}
}

// Given a []byte, start and finish of some inner slice, will find nearest
// newlines on both ends of this slice
func beginend(s []byte, start int, finish int) (begin int, end int) {
	begin = 0
	end = len(s)

	for i := start; i >= 0; i-- {
		if s[i] == byteNewLine[0] {
			// skip newline itself
			begin = i + 1
			break
		}
	}

	for i := finish; i < len(s); i++ {
		if s[i] == byteNewLine[0] {
			end = i
			break
		}
	}

	return
}

func (sl StringList) Contains(s string) bool {
	for _, x := range sl {
		if x == s {
			return true
		}
	}
	return false
}

func (rl RegexpList) Match(s string) bool {
	for _, x := range rl {
		if x.Match([]byte(s)) {
			return true
		}
	}
	return false
}
