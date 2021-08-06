package utils

import (
	"bytes"
	"fmt"
	"html/template"
	"regexp"
	"strconv"
	"strings"
)

var (
	temperatureRegexp = regexp.MustCompile(`\s*(?P<num>\d+)\s*C?\s*`)
	sizeRegexp        = regexp.MustCompile(`\s*(?P<size>\d+)\s*(?P<unit>\w+)\s*`)
)

func TemperatureStrToInt(str string) (num int, err error) {
	err = fmt.Errorf("%s is not a valid temperature str", str)
	match := temperatureRegexp.FindStringSubmatch(str)
	if len(match) == 0 {
		return
	}

	for i, name := range temperatureRegexp.SubexpNames() {
		if name == "num" {
			num, err = strconv.Atoi(match[i])
			break
		}
	}

	return
}

func SizeStrToMiBInt(sizeStr string) (int, error) {
	bytes, err := SizeStrToByteInt(sizeStr)
	if err != nil {
		return 0, err
	}
	return bytes / 1024 / 1024, nil
}

func SizeStrToByteInt(sizeStr string) (num int, err error) {
	err = fmt.Errorf("%s is not a valid size str", sizeStr)
	match := sizeRegexp.FindStringSubmatch(sizeStr)
	if len(match) == 0 {
		return
	}

	r := make(map[string]string)
	for i, name := range sizeRegexp.SubexpNames() {
		if name != "" {
			r[name] = match[i]
		}
	}

	if len(r) != 2 {
		return
	}

	size, err := strconv.Atoi(r["size"])
	if err != nil {
		return
	}

	unit := r["unit"]

	switch strings.ToLower(unit) {
	case "byte":
		return size, nil
	case "kib":
	case "ki":
		return size * 1024, nil
	case "mib":
	case "mi":
		return size * 1024 * 1024, nil
	case "gib":
	case "gi":
		return size * 1024 * 1024 * 1024, nil
	case "tib":
	case "ti":
		return size * 1024 * 1024 * 1024 * 1024, nil
	case "pib":
	case "pi":
		return size * 1024 * 1024 * 1024 * 1024 * 1024, nil
	case "eib":
	case "ei":
		return size * 1024 * 1024 * 1024 * 1024 * 1024 * 1024, nil
	case "mb":
	case "m":
		return size * 1000 * 1000, nil
	case "gb":
	case "g":
		return size * 1000 * 1000 * 1000, nil
	case "tb":
	case "t":
		return size * 1000 * 1000 * 1000 * 1000, nil
	}

	return
}

func SplitToIntList(str string) ([]int, error) {
	pieces := strings.Split(strings.TrimSpace(str), ",")

	res := make([]int, 0, len(pieces))

	for _, p := range pieces {
		n, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			return nil, err
		}

		res = append(res, n)
	}

	return res, nil
}

func Partition(s string, sep string) (head string, retSep string, tail string) {
	// Partition(s, sep) -> (head, sep, tail)
	index := strings.Index(s, sep)
	if index == -1 {
		head = s
		retSep = ""
		tail = ""
	} else {
		head = s[:index]
		retSep = sep
		tail = s[len(head)+len(sep):]
	}
	return
}

func FormatCommitId(commitId string) string {
	commitId = strings.ToLower(commitId)
	if len(commitId) < 7 {
		return commitId
	}
	return commitId[:7]
}

func StringPtr(s string) *string {
	return &s
}

func RenderTemplate(dicts map[string]string, toRenderTmpl string) (string, error) {
	var output bytes.Buffer
	t, err := template.New("renderTemplate").Parse(toRenderTmpl)
	if err != nil {
		return "", err
	}
	if err := t.Execute(&output, dicts); err != nil {
		return "", err
	}
	return output.String(), nil
}
