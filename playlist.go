/* -.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.

* File Name : playlist.go

* Purpose :

* Creation Date : 03-16-2014

* Last Modified : Sun 04 Jan 2015 07:41:08 AM UTC

* Created By : Kiyor

_._._._._._._._._._._._._._._._._._._._._.*/

package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/kiyor/gfind/lib"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"text/template"
)

type Media struct {
	file
	T   string //type like video/mp4
	Sub []Subtitle
}

type Subtitle struct {
	file
	Lang string
}

type file struct {
	os.FileInfo
	Name         string
	path         string
	Url          string
	ext          string
	Episode      int
	isConverting bool
}

type Medias struct {
	Title string
	Ms    []Media
}

var (
	host      *string = flag.String("host", "hostname", "host")
	fullurl   *bool   = flag.Bool("fullurl", false, "set up full url")
	dir       *string = flag.String("dir", "/home/nginx/html", "rootdir")
	verbose   *bool   = flag.Bool("v", false, "output verbose")
	MEDIATYPE         = map[string]string{
		"mp3": "audio",
		"wav": "audio",
		"mp4": "video",
	}
	CONVERTCMD = map[string]string{
		"mkv": "/usr/local/bin/ffmpeg -i \"{@}.mkv\" -vcodec copy -acodec copy \"{@}.mp4\"",
		"wmv": "/usr/local/bin/ffmpeg -i \"{@}.wmv\" -c:v libx264 -crf 23 -c:a libfaac -q:a 100 \"{@}.mp4\"",
		"avi": "/usr/local/bin/ffmpeg -i \"{@}.avi\" -c:v libx264 -crf 23 -c:a libfaac -q:a 100 \"{@}.mp4\"",
		"ass": "/usr/local/bin/ass2srt.pl -f `file -bi \"{@}.ass\"|cut -d= -f2` -t utf8 \"{@}.ass\" \"{@}.srt\"",
	}
	CONVFMT = map[string]string{
		"mkv": "mp4",
		"avi": "mp4",
		"wmv": "mp4",
		"ass": "srt",
	}
	LANGMAP = map[string]string{
		"zh-cn": "sc,GB",
		"zh-tw": "tc,BIG5",
	}
	Host, Dir       string
	LOCKFILE        = "/var/run/playlist/playlist.lock"
	SUBTITLEFMT     = "srt"
	convertingQueue = make(chan *file)
	wg              sync.WaitGroup
)

func init() {
	if _, err := os.Stat(LOCKFILE); err == nil {
		os.Exit(1)
	}
	if err := ioutil.WriteFile(LOCKFILE, []byte(""), 0644); err != nil {
		fmt.Println("not able to write lock file", LOCKFILE)
		os.Exit(1)
	}
	flag.Parse()
	Host = "http://" + *host
	Dir = *dir
}

func main() {
	defer func() {
		if err := os.Remove(LOCKFILE); err != nil {
			fmt.Println("cannot remove lock file", LOCKFILE)
			os.Exit(1)
		}
		os.Exit(0)
	}()
	go converting()
	files, err := find(Dir, 0)
	if err != nil {
		log.Println(err.Error())
	}
	var dirs []string
	for _, v := range files {
		dirs = append(dirs, v.getDir())
	}
	dirs = removeDuplicates(dirs)
	for _, dir := range dirs {
		list, _ := find(dir, 1)
		ms := list2media(list)
		write2file(dir, mkPlaylist(dir, ms))
	}
	wg.Wait()
}

func write2file(dir string, player string) error {
	err := ioutil.WriteFile(dir+"player", []byte(player), 0644)
	return err
}

func mkPlaylist(dir string, m []Media) string {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in mkPlaylist", r)
		}
	}()
	token := strings.Split(m[0].T, "/")
	var t *template.Template
	if token[0] == "audio" {
		// 		fmt.Println(dir, "is audio")
		t = template.New("audio.tmpl")
	} else if token[0] == "video" {
		// 		fmt.Println(dir, "is mp4")
		t = template.New("video.tmpl")
	}
	t = template.Must(t.ParseGlob(Dir + "/templates/*.tmpl"))
	var buf bytes.Buffer
	var ms Medias
	ms.Title = dir2title(dir)
	ms.Ms = m
	err := t.Execute(&buf, ms)
	if err != nil {
		fmt.Println(err)
	}
	return buf.String()
}

// get file name with out ext
func (f *file) getPrefix() string {
	token := strings.Split(f.Name, ".")
	// 	fmt.Println("token", token)
	if len(token) < 2 {
		return f.Name
	}
	var prefix string
	for _, v := range token[:len(token)-1] {
		prefix += v + "."
	}
	return prefix[:len(prefix)-1]
}

func (m *Media) updateSubtitle() {
	prefix := m.getPrefix()
	var ss []Subtitle
	var conf gfind.FindConf
	r := strings.NewReplacer("(", "\\(", ")", "\\)", "[", "\\[", "]", "\\]", ".", "\\.", "\\", "\\\\", " ", "\\s", "'", "\\'")
	if *verbose {
		fmt.Println("REPLACE", r.Replace(prefix), m.ext)
	}
	conf.Name = ".*" + r.Replace(prefix) + ".*"
	conf.Ext = SUBTITLEFMT
	conf.Dir = m.getDir()
	conf.Ftype = "f"
	if *verbose {
		fmt.Println("conf", conf)
	}
	fs := gfind.Find(conf)
	if len(fs) != 0 {
		for _, v := range fs {
			var s Subtitle
			s.update(v.Path)
			s.guessLang()
			s.getEpisode()
			ss = append(ss, s)
			if *verbose {
				fmt.Println("file", m.Name, "has Subtitle", s.Name, s.Lang)
			}
		}
	}
	m.Sub = ss
}

func (f *file) getEpisode() {
	re, err := regexp.Compile(`(\[|\s)(\d\d)(\]|\s)`)
	if err != nil {
		// 		panic(err)
		return
	}
	if !re.MatchString(f.Name) {
		return
	}
	s := re.FindStringSubmatch(f.Name)
	if len(s) < 4 {
		return
	}
	e, err := strconv.Atoi(s[2])
	if err != nil {
		return
	}
	f.Episode = e
}

func (s *Subtitle) guessLang() {
	for keyLang, v := range LANGMAP {
		keyStrings := strings.Split(v, ",")
		for _, keyString := range keyStrings {
			r, err := regexp.Compile(`.*` + keyString + `.*`)
			if err != nil {
				fmt.Println(err.Error())
			} else if r.MatchString(s.Name) {
				s.Lang = keyLang
				return
			}
		}
	}
	s.Lang = "en"
}

func (f *file) updateUrl() {
	var Url *url.URL
	Url, err := url.Parse(Host)
	if err != nil {
		panic("host not correct")
	}
	// remove root path for generate url
	if f.IsDir() {
		return
	}
	file := f.path[len(Dir):]
	Url.Path += file
	f.Url = Url.String()[len(Host):]
}

func (f *file) getDir() string {
	token := strings.Split(f.path, "/")
	token = token[:len(token)-1]
	var dir string
	for _, v := range token {
		dir += v + "/"
	}
	return dir
}

func find(location string, depth int) ([]file, error) {
	var files []file
	locationToken := strings.Split(location, "/")
	err := filepath.Walk(location, func(path string, f os.FileInfo, _ error) error {
		if *verbose {
			fmt.Println("found file", path)
		}
		var myfile file
		myfile.FileInfo = f
		myfile.update(path)
		if myfile.needConv() {
			wg.Add(1)
			convertingQueue <- &myfile
		}
		if myfile.needPlaylist() {
			myfile.getEpisode()
			if depth == 0 {
				files = append(files, myfile)
			} else {
				pathToken := strings.Split(path, "/")
				if len(locationToken)+depth > len(pathToken) {
					files = append(files, myfile)
				} else {
				}
			}
		}
		return nil
	})
	return files, err
}

func converting() {
	for {
		select {
		case myfile := <-convertingQueue:
			myfile.convert()
			wg.Done()
		}
	}
}

func (f *file) update(path string) {
	f.path = path
	f.Name = f.FileInfo.Name()
	f.updateFileExt()
	f.updateUrl()
}

func (f *file) needPlaylist() bool {
	for k, _ := range MEDIATYPE {
		if k == f.ext {
			return true
		}
	}
	return false
}

func (f *file) needConv() bool {
	for k, _ := range CONVFMT {
		if k == f.ext {
			if *verbose {
				fmt.Println("need convert", f.Name)
			}
			return true
		}
	}
	return false
}

func (f *file) convert() {
	if f.isConverted() || f.IsDir() || f.isConverting {
		return
	}
	f.isConverting = true
	prefix := f.getDir() + f.getPrefix()
	convCmd := strings.Replace(CONVERTCMD[f.ext], "{@}", prefix, -1)
	log.Println(convCmd)
	cmd := exec.Command("/bin/bash", "-c", convCmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println("not able to convert file", f.Name, err)
	}
	f.isConverting = false
}

func (f *file) isConverted() bool {
	prefix := f.getDir() + f.getPrefix()
	chkF := prefix + "." + CONVFMT[f.ext]
	if _, err := os.Stat(chkF); err == nil {
		return true
	}
	return false
}

func (f *file) updateFileExt() {
	token := strings.Split(f.FileInfo.Name(), ".")
	f.ext = token[len(token)-1:][0]
}

func (m *Media) updateT() {
	m.T = MEDIATYPE[m.ext] + "/" + m.ext
}
