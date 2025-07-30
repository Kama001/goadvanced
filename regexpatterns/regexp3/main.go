// Print top N IP addresses causing HTTP 500/502/503/504 errors with count.

package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

type ipc struct {
	ip  string
	cnt int
}

func main() {
	ptm := regexp.MustCompile(`(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}) - - \[.*\] \".*\" 50[0-9] .*`)
	cwd, _ := os.Getwd()
	file, _ := os.Open(cwd + "/server.log")
	defer file.Close()
	scanner := bufio.NewScanner(file)
	ips := make(map[string]int)
	ipsSli := []ipc{}
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		res := ptm.FindStringSubmatch(line)
		if len(res) > 0 {
			ips[res[1]]++
		}
	}
	for k, v := range ips {
		ipsSli = append(ipsSli, ipc{k, v})
	}
	sort.Slice(ipsSli, func(i, j int) bool {
		return ipsSli[i].cnt > ipsSli[j].cnt
	})
	fmt.Println(ipsSli)
}
