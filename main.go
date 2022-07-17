package main

import (
	"bufio"
	"flag"
	"io"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

// ParseFile parse a proto file and generate the result file
func ParseFile(inPath string) {

	//log.WithField("path", inPath).Debug("parse file begin")

	inf, err := os.Open(inPath)
	if err != nil {
		log.WithFields(logrus.Fields{
			"path":  inPath,
			"error": err},
		).Fatal("open path error")
	}
	defer inf.Close()

	outf, err := os.Create(path.Base(inPath) + ".parse")
	if err != nil {
		log.WithField("error", err).Fatal("create result file error")
	}
	defer outf.Close()

	reader := bufio.NewReader(inf)
	writer := bufio.NewWriter(outf)

	//log.WithField("path", inPath).Debug("read begin")

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				log.WithField("error", err).Error("read error")
			} else {
				parseLine(line, writer)
				log.Debug("read over")
			}
			break
		}

		parseLine(line, writer)
	}

	writer.Flush()
	//log.WithField("path", inPath).Info("parse file over!")
}

type ParseType int

const (
	ParseMsg ParseType = iota
	ParseEnum
)

// regexp
var (
	messageReg   = regexp.MustCompile(`message[ \t]+([_0-9a-zA-Z]+)`)                                                                // (name)
	enumReg      = regexp.MustCompile(`enum[ \t]+([_0-9a-zA-Z]+)`)                                                                   // (name)
	fieldReg     = regexp.MustCompile(`(required|optional|repeated)?[ \t]+([._0-9a-zA-Z]+)[ \t]+([_0-9a-zA-Z]+)[ \t]*=[ \t]*[0-9]+`) // (option) (type) (name)
	indent       = '\t'
	indentCount  = 0
	parseType    = make([]ParseType, 0, 0) // wether we parse a message? could be a enum
	parseTypeIdx = -1
)

// parseLine parse a line read from proto file
func parseLine(line string, writer io.Writer) {
	//log.Debug("parse line", line)

	var builder strings.Builder
	for i := 0; i < indentCount; i++ {
		builder.WriteByte(byte(indent))
	}

	var needWrite = false
	if match := messageReg.FindStringSubmatch(line); match != nil {

		log.WithField("message", match[1]).Debug("message found")

		builder.WriteString("message")
		builder.WriteByte(byte(indent))
		builder.WriteString(match[1])
		indentCount++
		needWrite = true

		parseTypeIdx++
		if parseTypeIdx < len(parseType) {
			parseType[parseTypeIdx] = ParseMsg
		} else {
			parseType = append(parseType, ParseMsg)
		}

	} else if match = enumReg.FindStringSubmatch(line); match != nil {

		log.WithField("enum", match[1]).Debug("enum found")

		builder.WriteString("enum")
		builder.WriteByte(byte(indent))
		builder.WriteString(match[1])
		indentCount++
		needWrite = true

		parseTypeIdx++
		if parseTypeIdx < len(parseType) {
			parseType[parseTypeIdx] = ParseEnum
		} else {
			parseType = append(parseType, ParseEnum)
		}

	} else if match = fieldReg.FindStringSubmatch(line); match != nil {

		log.WithFields(logrus.Fields{
			"option": match[1],
			"type":   match[2],
			"name":   match[3]},
		).Debug("field found")

		option := match[1]
		if len(option) == 0 {
			option = "optional"
		}

		builder.WriteString("field")
		builder.WriteByte(byte(indent))
		builder.WriteString(option)
		builder.WriteByte(byte(indent))
		builder.WriteString(match[2])
		builder.WriteByte(byte(indent))
		builder.WriteString(match[3])
		needWrite = true

	} else if strings.Index(line, "}") != -1 {

		log.WithField("parseType", parseType[parseTypeIdx]).Debug("find end")
		log.Debug(parseType, parseTypeIdx)

		switch parseType[parseTypeIdx] {
		case ParseMsg:
			builder.WriteString("msgend")
			needWrite = true
		}

		parseTypeIdx--
		indentCount--

	}

	if needWrite {
		builder.WriteByte(byte('\n'))
		writer.Write([]byte(builder.String()))
	}
}

/////////////////////////////////////////
var log = logrus.New()

func init() {
	log.SetLevel(logrus.DebugLevel)
}

func main() {
	indentArg := flag.String("indent", "-", "indent used in parse file")
	pathArg := flag.String("path", "", "input proto file path")
	flag.Parse()

	indent = rune((*indentArg)[0])

	// x.proto
	if len(*pathArg) < 7 {
		log.WithField("path", *pathArg).Fatal("path error")
	}

	ParseFile(*pathArg)
}
