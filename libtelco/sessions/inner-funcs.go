// Copyright (C) 2018 Mikhail Masyagin & Andrey Koshelev

/*
Package sessions - данный файл содержит в себе внутренние функции обработки сайта.
*/
package sessions

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// incDate увеличивает текущую дату на один день.
func incDate(date string) (string, error) {
	// Объявляем необходимые функции

	// Принимает дату, записанную в виде строки, и возвращает в виде int'ов день, месяц и год этой даты
	getStringDates := func(str string) (int, int, int, error) {
		s := strings.Split(str, ".")
		if len(s) != 3 {
			// Дата записана неправильно
			return 0, 0, 0, fmt.Errorf("Wrond date: %s", date)
		}
		var err error
		day := 0
		month := 0
		year := 0
		day, err = strconv.Atoi(s[0])
		if err != nil {
			return 0, 0, 0, err
		}
		month, err = strconv.Atoi(s[1])
		if err != nil {
			return 0, 0, 0, err
		}
		year, err = strconv.Atoi(s[2])
		if err != nil {
			return 0, 0, 0, err
		}
		return day, month, year, nil
	}

	// Преобразует три числа (day, month, year) в строку-дату "day.month.year"
	makeDate := func(day int, month int, year int) string {
		var str string
		if day < 10 {
			str += "0"
		}
		str += strconv.Itoa(day) + "."
		if month < 10 {
			str += "0"
		}
		str += strconv.Itoa(month) + "." + strconv.Itoa(year)
		return str
	}

	// Преобразуем дату в три int'а
	day, month, year, err := getStringDates(date)
	if err != nil {
		return "", err
	}
	if (day == 0) && (month == 0) && (year == 0) {
		return "", fmt.Errorf("Wrond date: %s", date)
	}

	// Создаём из полученных int'ов структуру time.Date и получаем следующий день
	currentDay := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	oneDayLater := currentDay.AddDate(0, 0, 1)
	year2, month2, day2 := oneDayLater.Date()

	// Преобразуем полученные числа обратно в строку
	nextDay := makeDate(day2, int(month2), year2)
	return nextDay, nil
}

// totalMarkReportParser возвращает успеваемость ученика.
// находится в inner-funcs, так как отчеты на всех серверах одинаковые.
func totalMarkReportParser(r io.Reader) (*TotalMarkReport, error) {
	parsedHTML, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	// Находит нод с табличкой
	var findNode func(*html.Node) *html.Node
	findNode = func(node *html.Node) *html.Node {
		if node.Type == html.ElementNode {
			for _, a := range node.Attr {
				if (a.Key == "class") && (a.Val == "table-print-num") {
					return node
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			n := findNode(c)
			if n != nil {
				return n
			}
		}
		return nil
	}

	// Формирует отчёт
	var formTotalMarkReport func(*html.Node, map[string][]int)
	formTotalMarkReport = func(node *html.Node, report map[string][]int) {
		for _, a := range node.Attr {
			if (a.Key == "class") && (a.Val == "cell-text") {
				// Нашли урок
				report[node.FirstChild.Data] = make([]int, 7)
				i := 0
				flag := false
				for c := node.NextSibling; c != nil; c = c.NextSibling {
					if c.FirstChild != nil {
						mark, err := strconv.Atoi(c.FirstChild.Data)
						if flag {
							i++
						}
						if (err != nil) || (i > 6) {
							continue
						}
						flag = true
						report[node.FirstChild.Data][i] = mark
					}
				}
			}
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			formTotalMarkReport(c, report)
		}
	}

	// Создаёт отчёт
	makeTotalMarkReport := func(node *html.Node) (*TotalMarkReport, error) {
		var report TotalMarkReport
		tableNode := findNode(node)
		data := make(map[string][]int)
		formTotalMarkReport(tableNode, data)
		for k, v := range data {
			innerReport := TotalMarkReportNote{k, v[0], v[1], v[2], v[3], v[4], v[5], v[6]}
			report.Table = append(report.Table, innerReport)
		}
		return &report, nil
	}

	return makeTotalMarkReport(parsedHTML)
}

// averageMarkReportParser возвращает средние баллы ученика.
// находится в inner-funcs, так как отчеты на всех серверах одинаковые.
func averageMarkReportParser(r io.Reader) (*AverageMarkReport, error) {
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
				studentAverageMarkReport[subject.FirstChild.Data] = studentAverageMark.FirstChild.Data
				classAverageMarkReport[subject.FirstChild.Data] = classAverageMark.FirstChild.Data

				subject = subject.NextSibling
				studentAverageMark = studentAverageMark.NextSibling
				classAverageMark = classAverageMark.NextSibling
			}
		}
	}

	// Создаёт отчёт
	makeAverageMarkReportTable := func(node *html.Node) *AverageMarkReport {
		var report AverageMarkReport
		tableNode := findAverageMarkTableNode(node)
		studentReport := make(map[string]string)
		classReport := make(map[string]string)
		formAverageMarkReportTable(tableNode, studentReport, classReport)
		for k, v := range studentReport {
			v1, err := strconv.ParseFloat(v, 32)
			if err != nil {
				v1 = -1.0
			}
			v2, err := strconv.ParseFloat(classReport[k], 32)
			if err != nil {
				v2 = -1.0
			}
			innerReport := AverageMarkReportNote{k, float32(v1), float32(v2)}
			report.Table = append(report.Table, innerReport)
		}

		return &report
	}

	return makeAverageMarkReportTable(parsedHTML), nil
}

// averageMarkDynReportParser возвращает динамику среднего балла ученика.
// находится в inner-funcs, так как отчеты на всех серверах одинаковые.
func averageMarkDynReportParser(r io.Reader) (*AverageMarkDynReport, error) {
	parsedHTML, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	// Находит нод с табличкой
	var findAverageMarkDynTableNode func(*html.Node) *html.Node
	findAverageMarkDynTableNode = func(node *html.Node) *html.Node {
		if node.Type == html.ElementNode {
			for _, a := range node.Attr {
				if (a.Key == "class") && (a.Val == "table-print-num") {
					return node
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			n := findAverageMarkDynTableNode(c)
			if n != nil {
				return n
			}
		}
		return nil
	}

	// Формирует отчёт
	formAverageMarkDynReportTable := func(node *html.Node, data *[]AverageMarkDynReportNote) error {
		if node != nil {
			stage := node.FirstChild.FirstChild
			var studentWorksAmount, studentAverageMark, classWorksAmount, classAverageMark *html.Node
			hasWorks := false

			// Проверяем вид отчёта
			if stage.NextSibling.NextSibling.NextSibling != nil {
				hasWorks = true
			}
			if hasWorks {
				studentWorksAmount = stage.NextSibling
				studentAverageMark = studentWorksAmount.NextSibling
				classWorksAmount = studentAverageMark.NextSibling
				classAverageMark = classWorksAmount.NextSibling

				studentWorksAmount = studentWorksAmount.FirstChild.NextSibling
				studentAverageMark = studentAverageMark.FirstChild.NextSibling
				classWorksAmount = classWorksAmount.FirstChild.NextSibling
				classAverageMark = classAverageMark.FirstChild.NextSibling
			} else {
				studentAverageMark = stage.NextSibling
				classAverageMark = studentAverageMark.NextSibling

				studentAverageMark = studentAverageMark.FirstChild.NextSibling
				classAverageMark = classAverageMark.FirstChild.NextSibling
			}
			stage = stage.FirstChild.NextSibling
			for stage != nil {
				var note AverageMarkDynReportNote
				note.Date = stage.FirstChild.Data
				note.StudentAverageMark = studentAverageMark.FirstChild.Data
				note.ClassAverageMark = classAverageMark.FirstChild.Data

				if hasWorks {
					note.StudentWorksAmount, err = strconv.Atoi(studentWorksAmount.FirstChild.Data)
					if err != nil {
						return err
					}
					note.ClassWorksAmount, err = strconv.Atoi(classWorksAmount.FirstChild.Data)
					if err != nil {
						return err
					}
					studentWorksAmount = studentWorksAmount.NextSibling
					classWorksAmount = classWorksAmount.NextSibling
				}
				(*data) = append((*data), note)
				stage = stage.NextSibling
				studentAverageMark = studentAverageMark.NextSibling
				classAverageMark = classAverageMark.NextSibling
			}

			return nil
		}
		return errors.New("Node is nil in func formAverageMarkDynReportTable")
	}

	// Создаёт отчёт
	makeAverageMarkDynReportTable := func(node *html.Node) (*AverageMarkDynReport, error) {
		var report AverageMarkDynReport
		tableNode := findAverageMarkDynTableNode(node)
		report.Data = make([]AverageMarkDynReportNote, 0, 10)
		err := formAverageMarkDynReportTable(tableNode, &report.Data)
		return &report, err
	}
	return makeAverageMarkDynReportTable(parsedHTML)
}

// studentGradeReportParser возвращает отчет об успеваемости ученика по предмету.
// находится в inner-funcs, так как отчеты на всех серверах одинаковые.
func studentGradeReportParser(r io.Reader) (*StudentGradeReport, error) {
	parsedHTML, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	// Находит нод с табличкой
	var findPerformanceTableNode func(*html.Node) *html.Node
	findPerformanceTableNode = func(node *html.Node) *html.Node {
		if node.Type == html.ElementNode {
			for _, a := range node.Attr {
				if (a.Key == "class") && (a.Val == "table-print") {
					return node
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			n := findPerformanceTableNode(c)
			if n != nil {
				return n
			}
		}
		return nil
	}

	// Формирует отчёт
	formStudentGradeReportTable := func(node *html.Node) ([]StudentGradeReportNote, error) {
		notes := make([]StudentGradeReportNote, 0, 10)
		if node != nil {
			noteNode := node.FirstChild.FirstChild
			for noteNode = noteNode.NextSibling; noteNode != nil && len(noteNode.Attr) == 0; noteNode = noteNode.NextSibling {
				// Добавляем запись
				perfNote := *new(StudentGradeReportNote)
				c := noteNode.FirstChild
				perfNote.Type = c.FirstChild.Data
				c = c.NextSibling
				perfNote.Theme = c.FirstChild.Data
				c = c.NextSibling
				perfNote.DateOfCompletion = c.FirstChild.Data
				c = c.NextSibling
				perfNote.Mark, err = strconv.Atoi(c.FirstChild.Data)
				if err != nil {
					return notes, err
				}
				notes = append(notes, perfNote)
			}
		}
		return notes, nil
	}

	// Создаёт отчёт
	makeStudentGradeReportTable := func(node *html.Node) (*StudentGradeReport, error) {
		var report StudentGradeReport
		tableNode := findPerformanceTableNode(node)
		report.Data, err = formStudentGradeReportTable(tableNode)
		return &report, err
	}

	return makeStudentGradeReportTable(parsedHTML)
}

// studentTotalReportParser возвращает отчет о посещениях ученика.
// находится в inner-funcs, так как отчеты на всех серверах одинаковые.
func studentTotalReportParser(r io.Reader) (*StudentTotalReport, error) {
	parsedHTML, err := html.Parse(r)
	if err != nil {
		return nil, err
	}
	// Находит нод с табличкой
	var findStudentTotalTableNode func(*html.Node) *html.Node
	findStudentTotalTableNode = func(node *html.Node) *html.Node {
		if node.Type == html.ElementNode {
			for _, a := range node.Attr {
				if a.Key == "class" && a.Val == "table-print" {
					return node
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			n := findStudentTotalTableNode(c)
			if n != nil {
				return n
			}
		}
		return nil
	}
	// Разделяет строку по пробелу
	splitBySpace := func(str string) []string {
		strings := make([]string, 0, 3)
		start := 0
		for i := 0; i < len(str); i++ {
			if str[i] == 194 || str[i] == 160 {
				if i < len(str)-1 && str[i+1] == 160 {
					strings = append(strings, str[start:i])
					start = i + 2
				}
			}
		}
		if start < len(str) {
			strings = append(strings, str[start:])
		}
		return strings
	}
	// Формирует отчёт
	formStudentTotalReportTable := func(node *html.Node) ([]Month, map[string]string, error) {
		months := make([]Month, 0, 3)
		averageMarks := make(map[string]string)
		if node != nil {
			// Добавляем месяцы
			monthNode := node.FirstChild.FirstChild.FirstChild
			for monthNode = monthNode.NextSibling; len(monthNode.Attr) == 1 && monthNode.Attr[0].Key == "colspan"; monthNode = monthNode.NextSibling {
				month := *new(Month)
				month.Name = monthNode.FirstChild.Data
				numberOfDaysInMonth, err := strconv.Atoi(monthNode.Attr[0].Val)
				if err != nil {
					return months, averageMarks, err
				}
				month.Days = make([]Day, numberOfDaysInMonth)
				months = append(months, month)
			}
			// Добавляем дни
			dayNode := node.FirstChild.FirstChild.NextSibling
			// Текущий месяц в months
			currentMonth := 0
			// Сколько дней добавили для текущего месяца
			dayNumberInMonth := 0
			// Всего дней в отчёте
			overallNumberOfDays := 0
			for dayNode = dayNode.FirstChild; dayNode != nil; dayNode = dayNode.NextSibling {
				if dayNumberInMonth == len(months[currentMonth].Days) {
					currentMonth++
					dayNumberInMonth = 0
				}
				day := *new(Day)
				day.Number, err = strconv.Atoi(dayNode.FirstChild.Data)
				day.Marks = make(map[string][]string)
				if err != nil {
					return months, averageMarks, err
				}
				months[currentMonth].Days[dayNumberInMonth] = day

				dayNumberInMonth++
				overallNumberOfDays++
			}
			// Идём по остальной части таблицы
			noteNode := node.FirstChild.FirstChild.NextSibling
			for noteNode = noteNode.NextSibling; noteNode != nil; noteNode = noteNode.NextSibling {
				currentMonth = 0
				dayNumberInMonth = 0
				c := noteNode.FirstChild
				subjectName := c.FirstChild.Data
				for i := 0; i < overallNumberOfDays; i++ {
					if dayNumberInMonth == len(months[currentMonth].Days) {
						currentMonth++
						dayNumberInMonth = 0
					}
					c = c.NextSibling
					var marks []string
					if c.FirstChild.FirstChild != nil {
						for c2 := c.FirstChild; c2 != nil; c2 = c2.NextSibling {
							var s []string
							if c2.FirstChild != nil {
								s = splitBySpace(c2.FirstChild.Data)
							} else {
								s = splitBySpace(c2.Data)
							}
							for k := 0; k < len(s); k++ {
								marks = append(marks, s[k])
							}
						}
					} else {
						marks = splitBySpace(c.FirstChild.Data)
					}
					// Избавляемся от строк из непечатаемых символом
					finalMarks := make([]string, 0, 1)
					for el := range marks {
						if len([]byte(marks[el])) != 0 {
							finalMarks = append(finalMarks, marks[el])
						}
					}
					if len(finalMarks) != 0 {
						months[currentMonth].Days[dayNumberInMonth].Marks[subjectName] = finalMarks
					}

					dayNumberInMonth++
				}
				averageMarks[subjectName] = c.NextSibling.FirstChild.Data
			}
		} else {
			return months, averageMarks, errors.New("Node is nil in func formStudentTotalReportTable")
		}

		return months, averageMarks, nil
	}
	// Создаёт отчёт
	makeStudentTotalReportTable := func(node *html.Node) (*StudentTotalReport, error) {
		var report StudentTotalReport
		tableNode := findStudentTotalTableNode(node)
		report.MainTable, report.AverageMarks, err = formStudentTotalReportTable(tableNode)
		return &report, err
	}
	return makeStudentTotalReportTable(parsedHTML)
}
