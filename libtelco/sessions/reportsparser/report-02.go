package reportsparser

import (
	"io"
	"strconv"
	"strings"

	dt "github.com/masyagin1998/SchoolServer/libtelco/sessions/datatypes"

	"golang.org/x/net/html"
)

// AverageMarkReportParser возвращает средние баллы ученика.
// находится в inner-funcs, так как отчеты на всех серверах одинаковые.
func AverageMarkReportParser(r io.Reader) (*dt.AverageMarkReport, error) {
	parsedHTML, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	// Находит нод с табличкой
	var findAverageMarkTableNode func(*html.Node) *html.Node
	findAverageMarkTableNode = func(node *html.Node) *html.Node {
		if node.Type == html.ElementNode {
			for _, a := range node.Attr {
				if (a.Key == "class") && (a.Val == "table-print-num") {
					return node
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			n := findAverageMarkTableNode(c)
			if n != nil {
				return n
			}
		}
		return nil
	}

	// Формирует отчёт
	formAverageMarkReportTable := func(node *html.Node, studentAverageMarkReport map[string]string, classAverageMarkReport map[string]string) {
		if node != nil {
			subject := node.FirstChild.FirstChild
			studentAverageMark := subject.NextSibling
			classAverageMark := studentAverageMark.NextSibling

			subject = subject.FirstChild.NextSibling
			studentAverageMark = studentAverageMark.FirstChild.NextSibling
			classAverageMark = classAverageMark.FirstChild.NextSibling
			for subject != nil {
				if subject.FirstChild != nil && studentAverageMark.FirstChild != nil && classAverageMark.FirstChild != nil {
					studentAverageMarkReport[subject.FirstChild.Data] = studentAverageMark.FirstChild.Data
					classAverageMarkReport[subject.FirstChild.Data] = classAverageMark.FirstChild.Data
				}

				subject = subject.NextSibling
				studentAverageMark = studentAverageMark.NextSibling
				classAverageMark = classAverageMark.NextSibling
			}
		}
	}

	// Создаёт отчёт
	makeAverageMarkReportTable := func(node *html.Node) *dt.AverageMarkReport {
		report := dt.NewAverageMarkReport()
		tableNode := findAverageMarkTableNode(node)
		studentReport := make(map[string]string)
		classReport := make(map[string]string)
		formAverageMarkReportTable(tableNode, studentReport, classReport)
		for k, v := range studentReport {
			v1, err := strconv.ParseFloat(strings.Replace(v, ",", ".", 1), 32)
			if err != nil {
				v1 = -1.0
			}
			v2, err := strconv.ParseFloat(strings.Replace(classReport[k], ",", ".", 1), 32)
			if err != nil {
				v2 = -1.0
			}
			innerReport := dt.AverageMarkReportNote{
				Subject:     k,
				StudentMark: float32(v1),
				ClassMark:   float32(v2),
			}
			report.Table = append(report.Table, innerReport)
		}

		return report
	}

	return makeAverageMarkReportTable(parsedHTML), nil
}
