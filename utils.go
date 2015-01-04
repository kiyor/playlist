/* -.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.

* File Name : utils.go

* Purpose :

* Creation Date : 01-04-2015

* Last Modified : Sun 04 Jan 2015 05:55:24 AM UTC

* Created By : Kiyor

_._._._._._._._._._._._._._._._._._._._._.*/

package main

import (
	"strings"
)

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

func list2media(fs []file) []Media {
	var ms []Media
	for _, f := range fs {
		var m Media
		m.file = f
		m.updateT()
		m.updateUrl()
		m.updateSubtitle()
		if *fullurl {
			m.Url = Host + m.Url
		}
		ms = append(ms, m)
	}
	return ms
}

func dir2title(dir string) string {
	token := strings.Split(dir, "/")
	return token[len(token)-2 : len(token)-1][0]
}
