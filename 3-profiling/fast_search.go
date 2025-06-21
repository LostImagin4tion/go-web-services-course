package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
)

func FastSearch(out io.Writer) {
	var file, err = os.Open(filePath)
	if err != nil {
		panic(err)
	}

	var scanner = bufio.NewScanner(file)

	const bufferSize = 100 * 1024 // 100 kb
	var scannerBuffer = make([]byte, bufferSize)
	scanner.Buffer(scannerBuffer, bufferSize)

	var targetBrowsers = [...]string{"Android", "MSIE"}
	var seenBrowsers = make(map[string]interface{})

	var reg = regexp.MustCompile("@")

	fmt.Fprintln(out, "found users:")

	for i := 0; scanner.Scan(); i++ {
		var line = scanner.Bytes()
		var user = make(map[string]interface{})
		//fmt.Printf("%v %v\n", err, line)
		var err = json.Unmarshal(line, &user)
		if err != nil {
			panic(err)
		}

		var browsers, ok = user["browsers"].([]interface{})
		if !ok {
			log.Println("Cant cast browsers")
			continue
		}

		var matchedBrowsers = make(map[string]interface{}, len(targetBrowsers))

		for _, browserRaw := range browsers {
			var browser, ok = browserRaw.(string)
			if !ok {
				//log.Println("Cant cast browser to string")
				continue
			}

			for _, targetBrowser := range targetBrowsers {
				if strings.Contains(browser, targetBrowser) {
					matchedBrowsers[targetBrowser] = struct{}{}
					seenBrowsers[browser] = struct{}{}
				}
			}
		}

		if len(matchedBrowsers) != len(targetBrowsers) {
			continue
		}

		//log.Println("Android and MSIE user:", user["name"], user["email"])
		var email = reg.ReplaceAllString(user["email"].(string), " [at] ")

		fmt.Fprintf(out, "[%d] %s <%s>\n", i, user["name"], email)
	}

	fmt.Fprintln(out, "\nTotal unique browsers", len(seenBrowsers))
}
