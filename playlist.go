/* -.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.

* File Name : playlist.go

* Purpose :

* Creation Date : 03-16-2014

* Last Modified : Tue 08 Apr 2014 09:36:27 PM UTC

* Created By : Kiyor

_._._._._._._._._._._._._._._._._._._._._.*/

package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type Media struct {
	Name string
	Url  string
	T    string
}

type Medias struct {
	Title string
	Ms    []Media
}

var (
	host       *string = flag.String("host", "hostname", "host")
	fullurl    *bool   = flag.Bool("fullurl", false, "set up full url")
	dir        *string = flag.String("dir", "/home/nginx/html", "rootdir")
	supportFmt         = []string{"mp3", "wav", "mp4"}
	mediaType          = map[string]string{
		"mp3": "audio",
		"wav": "audio",
		"mp4": "video",
	}
	Host, Dir string
)

func init() {
	Host = "http://" + *host
	Dir = *dir
}

func main() {
	files, _ := find(Dir, 0)
	var dirs []string
	for _, v := range files {
		dirs = append(dirs, file2dir(v))
	}
	dirs = removeDuplicates(dirs)
	for _, dir := range dirs {
		// 		fmt.Println(dir)
		list, _ := find(dir, 1)
		// 		fmt.Println(dir)
		ms := list2media(dir, list)
		// 		fmt.Println(mkPlaylist(dir, ms))
		write2file(dir, mkPlaylist(dir, ms))
	}
}

func write2file(dir string, player string) {
	err := ioutil.WriteFile(dir+"player", []byte(player), 0644)
	if err != nil {
		panic(err)
	}
}

func mkPlaylist(dir string, m []Media) string {
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
	// 	fmt.Println(buf.String())
	return buf.String()
}

func dir2title(dir string) string {
	token := strings.Split(dir, "/")
	return token[len(token)-2 : len(token)-1][0]
}

func list2media(dir string, list []string) []Media {
	var ms []Media
	for _, f := range list {
		var m Media
		m.Name = f[len(dir):]
		ft := getFileType(f)
		m.T = mediaType[ft] + "/" + ft
		if *fullurl {
			m.Url = Host + file2url(f)
		} else {
			m.Url = file2url(f)
		}
		ms = append(ms, m)
	}
	return ms
}

func file2url(file string) string {
	var Url *url.URL
	Url, err := url.Parse(Host)
	if err != nil {
		panic("host not correct")
	}
	file = file[len(Dir):]
	Url.Path += file
	return Url.String()[len(Host):]
}

func file2dir(file string) string {
	token := strings.Split(file, "/")
	token = token[:len(token)-1]
	var dir string
	for _, v := range token {
		dir += v + "/"
	}
	return dir
}

func find(location string, depth int) ([]string, error) {
	var files []string
	locationToken := strings.Split(location, "/")
	err := filepath.Walk(location, func(path string, _ os.FileInfo, _ error) error {
		if needPlaylist(path) {
			if depth == 0 {
				files = append(files, path)
			} else {
				pathToken := strings.Split(path, "/")
				if len(locationToken)+depth > len(pathToken) {
					files = append(files, path)
					// 					fmt.Println(path, len(locationToken)+depth, len(pathToken))
				} else {
					// 					fmt.Println("skip path", path)
					// 					fmt.Println(path, len(locationToken)+depth, len(pathToken))
				}
			}
		}
		return nil
	})
	return files, err
}

func needPlaylist(file string) bool {
	token := strings.Split(file, ".")
	for _, v := range supportFmt {
		if v == token[len(token)-1:][0] {
			return true
		}
	}
	return false
}

func getFileType(file string) string {
	token := strings.Split(file, ".")
	return token[len(token)-1:][0]
}

func removeDuplicates(a []string) []string {
	result := []string{}
	seen := map[string]string{}
	for _, val := range a {
		if _, ok := seen[val]; !ok {
			result = append(result, val)
			seen[val] = val
		}
	}
	return result
}
