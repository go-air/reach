package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var timeRe = regexp.MustCompile(`Time\s+=\s+(\d+\.\d+)`)

func getTime(line string) (float64, error) {
	matches := timeRe.FindStringSubmatch(line)
	if len(matches) == 0 {
		return 0.0, fmt.Errorf("didn't match: %s\n", line)
	}
	secString := matches[1]
	return strconv.ParseFloat(secString, 64)
}

func main() {
	for _, fn := range os.Args[1:] {
		f, e := os.Open(fn)
		if e != nil {
			log.Fatalf("error opening %s: %s\n", fn, e.Error())
		}
		scanner := bufio.NewScanner(f)
		result := 2
		var endLine string
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "of miter") {
				result = 1
				endLine = line
			} else if strings.Contains(line, "UNDECIDED") {
				result = 0
				endLine = line
			} else if strings.Contains(line, "proved") {
				result = -1
				endLine = line
			}
		}
		if err := scanner.Err(); err != nil {
			log.Fatalf("error scanning: %s\n", err)
		}
		f.Close()
		t, e := getTime(endLine)
		if e != nil {
			log.Fatalf("couldn't get time from '%s': %s\n", endLine, e.Error())
		}
		fmt.Printf("%s %d %f\n", fn, result, t)
	}
}
