package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type event struct {
	counts float64
	value  float64
}

type entry struct {
	time   float64
	events map[string]*event
}

func main() {
	var path string

	flag.StringVar(&path, "path", "tensor.txt", "")
	flag.Parse()

	file, err := os.Open(path)
	if err != nil {
		log.Printf("open file: %s, error: %s", path, err)
	}

	indices := []string{
		"task-clock",
		"context-switches",
		"cpu-migrations",
		"page-faults",
		"cycles",
		"stalled-cycles-frontend",
		"instructions",
		"stalled-cycles-per-insn",
		"branches",
		"branch-misses",
		"L1-dcache-loads",
		"L1-dcache-load-misses",
		"LLC-loads",
	}

	currTime := float64(0)

	records := []*entry(nil)
	idx := -1

	reader := bufio.NewReader(file)
	for {
		bytes, _, err := reader.ReadLine()
		if err != nil {
			if err.Error() != "EOF" {
				log.Printf("read line error: %s", err)
			}

			break
		}

		line := strings.TrimSpace(string(bytes))

		splits := regSplit(line, "[ ]+")

		if len(splits) < 2 {
			continue
		} else if splits[0] == "#" {
			continue
		} else if splits[1] == "<not" {
			continue
		}

		time, _ := strconv.ParseFloat(splits[0], 64)

		var record *entry

		if time == currTime {
			record = records[idx]
		} else {
			idx++

			currTime = time
			record = &entry{
				time:   time,
				events: map[string]*event{},
			}
			records = append(records, record)
		}

		var name string
		var e *event
		if splits[1] == "#" {
			name, e = type1(splits)
		} else if splits[2] == "msec" {
			name, e = type2(splits)
		} else {
			name, e = type3(splits)
		}

		record.events[name] = e

		//fmt.Println(splits)
	}

	title := "id,time"
	for _, name := range indices {
		title += "," + name + ",count"
	}

	fmt.Println(title)

	for i, record := range records {
		line := fmt.Sprintf("%d,%f", i+1, record.time)
		for _, name := range indices {
			if e, ok := record.events[name]; ok {
				line += fmt.Sprintf(",%f,%f", e.value, e.counts)
			} else {
				line += ",n/a,n/a"
				//fmt.Println(record.time, name)
			}
		}
		fmt.Println(line)
	}
}

func regSplit(text string, delimeter string) []string {
	reg := regexp.MustCompile(delimeter)
	indexes := reg.FindAllStringIndex(text, -1)
	laststart := 0
	result := make([]string, len(indexes)+1)
	for i, element := range indexes {
		result[i] = text[laststart:element[0]]
		laststart = element[1]
	}
	result[len(indexes)] = text[laststart:len(text)]
	return result
}

// [1.000106418 # 1.02 stalled cycles per insn (61.50%)]
func type1(splits []string) (string, *event) {
	e := &event{}
	var err error
	e.value, err = strconv.ParseFloat(strings.Replace(splits[2], ",", "", -1), 64)
	if err != nil {
		fmt.Println("wa1 stalled-cycles-per-insn")
	}
	return "stalled-cycles-per-insn", e
}

// [1.000106418 1,510.48 msec task-clock # 1.510 CPUs utilized]
func type2(splits []string) (string, *event) {
	e := &event{}
	var err error
	e.counts, _ = strconv.ParseFloat(strings.Replace(splits[1], ",", "", -1), 64)
	e.value, err = strconv.ParseFloat(strings.Replace(splits[5], ",", "", -1), 64)
	if err != nil {
		fmt.Println("wa2 task-clock")
	}
	return "task-clock", e
}

// [1.000106418 217 context-switches # 0.144 K/sec]
func type3(splits []string) (string, *event) {
	e := &event{}
	var err error
	e.counts, _ = strconv.ParseFloat(strings.Replace(splits[1], ",", "", -1), 64)
	e.value, err = strconv.ParseFloat(strings.Replace(strings.Replace(splits[4], ",", "", -1), "%", "", -1), 64)
	if err != nil {
		fmt.Println("wa3", splits[2])
	}
	return splits[2], e
}
