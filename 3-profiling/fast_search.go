package main

import (
	"bufio"
	"fmt"
	"github.com/mailru/easyjson"
	"io"
	"os"
	"regexp"
	"stepikGoWebServices/model"
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

	var targetBrowsers = []string{
		"Android",
		"MSIE",
	}
	var seenBrowsers = make(map[string]interface{})

	var reg = regexp.MustCompile("@")

	fmt.Fprintln(out, "found users:")

	for i := 0; scanner.Scan(); i++ {
		var line = scanner.Bytes()
		var user = model.User{}
		var err = easyjson.Unmarshal(line, &user)
		if err != nil {
			panic(err)
		}

		var matchedBrowsers = make(map[string]interface{}, len(targetBrowsers))

		for _, browser := range user.Browsers {
			if len(browser) != 0 {
				for _, targetBrowser := range targetBrowsers {
					if strings.Contains(browser, targetBrowser) {
						matchedBrowsers[targetBrowser] = struct{}{}
						seenBrowsers[browser] = struct{}{}
					}
				}
			}
		}

		if len(matchedBrowsers) != len(targetBrowsers) {
			continue
		}

		//log.Println("Android and MSIE user:", user["name"], user["email"])
		var email = reg.ReplaceAllString(user.Email, " [at] ")

		fmt.Fprintf(out, "[%d] %s <%s>\n", i, user.Name, email)
	}

	fmt.Fprintln(out, "\nTotal unique browsers", len(seenBrowsers))
}
