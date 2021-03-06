// Copyright (C) 2018 Mikhail Masyagin & Andrey Koshelev

package reports

import (
	"bytes"
	"fmt"

	dt "github.com/masyagin1998/SchoolServer/libtelco/sessions/datatypes"
	"github.com/masyagin1998/SchoolServer/libtelco/sessions/reportsparser"
	"github.com/masyagin1998/SchoolServer/libtelco/sessions/servers/01/check"

	"github.com/pkg/errors"

	gr "github.com/levigross/grequests"
)

/*
01 тип.
*/

// GetTotalMarkReport возвращает успеваемость ученика с сервера первого типа.
func GetTotalMarkReport(s *dt.Session, studentID string) (*dt.TotalMarkReport, error) {
	p := "http://"

	// 0-ой Post-запрос.
	r0 := func() (bool, error) {
		ro := &gr.RequestOptions{
			Data: map[string]string{
				"AT":        s.AT,
				"LoginType": "0",
				"RPTID":     "0",
				"ThmID":     "1",
				"VER":       s.VER,
			},
			Headers: map[string]string{
				"Origin":                    p + s.Serv.Link,
				"Upgrade-Insecure-Requests": "1",
				"Referer":                   p + s.Serv.Link + "/asp/Reports/Reports.asp",
			},
		}
		r, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/ReportStudentTotalMarks.asp", ro)
		if err != nil {
			return false, err
		}
		defer func() {
			_ = r.Close()
		}()
		return check.CheckResponse(s, r)
	}
	flag, err := r0()
	if err != nil {
		return nil, errors.Wrap(err, "0 POST")
	}
	if !flag {
		flag, err = r0()
		if err != nil {
			return nil, errors.Wrap(err, "retrying 0 POST")
		}
		if !flag {
			return nil, fmt.Errorf("retry didn't work for 0 POST")
		}
	}

	// 1-ый Post-запрос.
	r1 := func() (bool, error) {
		ro := &gr.RequestOptions{
			Data: map[string]string{
				"A":         "",
				"AT":        s.AT,
				"BACK":      "/asp/Reports/ReportStudentTotalMarks.asp",
				"LoginType": "0",
				"NA":        "",
				"PCLID":     "",
				"PP":        "/asp/Reports/ReportStudentTotalMarks.asp",
				"RP":        "",
				"RPTID":     "0",
				"RT":        "",
				"SID":       studentID,
				"TA":        "",
				"ThmID":     "1",
				"VER":       s.VER,
			},
			Headers: map[string]string{
				"Origin":                    p + s.Serv.Link,
				"Upgrade-Insecure-Requests": "1",
				"Referer":                   p + s.Serv.Link + "/asp/Reports/ReportStudentTotalMarks.asp",
			},
		}
		r, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/ReportStudentTotalMarks.asp", ro)
		if err != nil {
			return false, err
		}
		defer func() {
			_ = r.Close()
		}()

		flag, err := check.CheckResponse(s, r)
		return flag, err
	}
	flag, err = r1()
	if err != nil {
		return nil, errors.Wrap(err, "1 POST")
	}
	if !flag {
		flag, err = r1()
		if err != nil {
			return nil, errors.Wrap(err, "retrying 1 POST")
		}
		if !flag {
			return nil, fmt.Errorf("retry didn't work for 1 POST")
		}
	}

	// 2-ой Post-запрос.
	r2 := func() ([]byte, bool, error) {
		ro := &gr.RequestOptions{
			Data: map[string]string{
				"A":         "",
				"AT":        s.AT,
				"BACK":      "/asp/Reports/ReportStudentTotalMarks.asp",
				"ISTF":      "0",
				"LoginType": "0",
				"NA":        "",
				"PCLID":     "",
				"PP":        "/asp/Reports/ReportStudentTotalMarks.asp",
				"RP":        "0",
				"RPTID":     "0",
				"RT":        "0",
				"SID":       studentID,
				"TA":        "",
				"ThmID":     "1",
				"VER":       s.VER,
			},
			Headers: map[string]string{
				"Origin":           p + s.Serv.Link,
				"X-Requested-With": "XMLHttpRequest",
				"at":               s.AT,
				"Referer":          p + s.Serv.Link + "/asp/Reports/ReportStudentTotalMarks.asp",
			},
		}
		r, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/StudentTotalMarks.asp", ro)
		if err != nil {
			return nil, false, err
		}
		defer func() {
			_ = r.Close()
		}()
		flag, err := check.CheckResponse(s, r)
		return r.Bytes(), flag, err
	}
	b, flag, err := r2()
	if err != nil {
		return nil, errors.Wrap(err, "2 POST")
	}

	if !flag {
		b, flag, err = r2()
		if err != nil {
			return nil, errors.Wrap(err, "retrying 2 POST")
		}
		if !flag {
			return nil, fmt.Errorf("retry didn't work for 2 POST")
		}
	}

	// Если мы дошли до этого места, то можно распарсить HTML-страницу,
	// находящуюся в теле ответа, и найти в ней отчет об успеваемости ученика.
	return reportsparser.TotalMarkReportParser(bytes.NewReader(b))
}
