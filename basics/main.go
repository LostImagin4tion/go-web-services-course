package main

import (
	"fmt"
	"io"
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"
)

type FileMetadata struct {
	fmt.Stringer

	Path            string
	Depth           int
	UseVerticalLine []bool
}

func (f *FileMetadata) String() string {
	return fmt.Sprintf("{%s}", f.Path)
}

func getTreePrefix(
	depth int,
	lineStartSymbol string,
	useVerticalLine []bool,
) string {
	var prefix = fmt.Sprintf("%s───", lineStartSymbol)
	var resultPrefix = ""

	var tabs = make([]string, depth-1)
	for i := range depth - 1 {
		if useVerticalLine[i] {
			tabs = append(tabs, "│\t")
		} else {
			tabs = append(tabs, "\t")
		}
	}

	resultPrefix = fmt.Sprintf(
		"%s%s",
		strings.Join(tabs, ""),
		prefix,
	)

	return resultPrefix
}

func handleDir(
	isRootFile bool,
	printFiles bool,
	lineStartSymbol string,
	fileMetadata FileMetadata,
	fileInfo os.FileInfo,
	stack *[]FileMetadata,
) (string, error) {
	var file, err = os.Open(fileMetadata.Path)
	if err != nil {
		return "", fmt.Errorf("failed to open file '%s', error %s", fileMetadata.Path, err)
	}

	var children, _ = file.ReadDir(0)
	var childrenFiltered = filter(&children, func(item os.DirEntry) bool {
		return item.IsDir() || printFiles
	})

	sort.Slice(
		childrenFiltered,
		func(i, j int) bool {
			return childrenFiltered[i].Name() < childrenFiltered[j].Name()
		},
	)

	var childrenMetadata = mapped(&childrenFiltered, func(i int, child os.DirEntry) FileMetadata {
		var useVerticalLineForThisChild = i != len(childrenFiltered)-1

		var newUseVerticalLine = append(
			copied(&fileMetadata.UseVerticalLine),
			useVerticalLineForThisChild,
		)

		return FileMetadata{
			Path:            fmt.Sprintf("%s/%s", fileMetadata.Path, child.Name()),
			Depth:           fileMetadata.Depth + 1,
			UseVerticalLine: newUseVerticalLine,
		}
	})

	*stack = append(childrenMetadata, *stack...)

	var printString string
	if !isRootFile {
		printString = fmt.Sprintf(
			"%s%s\n",
			getTreePrefix(
				fileMetadata.Depth,
				lineStartSymbol,
				fileMetadata.UseVerticalLine,
			),
			fileInfo.Name(),
		)
	}
	return printString, nil
}

func handleRegularFile(
	isRootFile bool,
	lineStartSymbol string,
	fileInfo os.FileInfo,
	fileMetadata FileMetadata,
) string {
	var size = fileInfo.Size()
	var sizeStr string
	if size == 0 {
		sizeStr = "empty"
	} else {
		sizeStr = fmt.Sprintf("%sb", strconv.Itoa(int(size)))
	}

	var printString string
	if !isRootFile {
		printString = fmt.Sprintf(
			"%s%s (%s)\n",
			getTreePrefix(
				fileMetadata.Depth,
				lineStartSymbol,
				fileMetadata.UseVerticalLine,
			),
			fileInfo.Name(),
			sizeStr,
		)
	}
	return printString
}

func dirTree(out io.Writer, root string, printFiles bool) error {
	var stack = []FileMetadata{
		{
			Path:            root,
			Depth:           0,
			UseVerticalLine: []bool{},
		},
	}
	var isRootFile = true

	for len(stack) != 0 {
		var fileMetadata = stack[0]

		stack = slices.Delete(stack, 0, 1)

		var fileInfo, err = os.Lstat(fileMetadata.Path)

		if err != nil {
			return fmt.Errorf("cannot open file '%s': error %s", fileMetadata.Path, err)
		}

		var printString string

		var lineStartSymbol = "├"
		if !isRootFile && (len(stack) == 0 || stack[0].Depth != fileMetadata.Depth) {
			lineStartSymbol = "└"
		}

		switch mode := fileInfo.Mode(); {
		case mode.IsDir():
			printString, err = handleDir(
				isRootFile,
				printFiles,
				lineStartSymbol,
				fileMetadata,
				fileInfo,
				&stack,
			)
			if err != nil {
				return err
			}

		case mode.IsRegular() && printFiles:
			printString = handleRegularFile(
				isRootFile,
				lineStartSymbol,
				fileInfo,
				fileMetadata,
			)
		}

		if len(printString) != 0 {
			var _, err = out.Write([]byte(printString))
			if err != nil {
				return fmt.Errorf("failed to write '%s' to output stream", printString)
			}
		}

		isRootFile = false
	}

	return nil
}

func main() {
	out := os.Stdout

	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}

	path := os.Args[1]
	shouldPrintFiles := len(os.Args) == 3 && os.Args[2] == "-f"

	err := dirTree(out, path, shouldPrintFiles)

	if err != nil {
		panic(err.Error())
	}
}
