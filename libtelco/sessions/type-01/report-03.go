// Copyright (C) 2018 Mikhail Masyagin & Andrey Koshelev

package type01

import (
	"bytes"
	"fmt"

	dt "github.com/masyagin1998/SchoolServer/libtelco/sessions/data-types"
	"github.com/masyagin1998/SchoolServer/libtelco/sessions/inner"
	ss "github.com/masyagin1998/SchoolServer/libtelco/sessions/session"

	"github.com/pkg/errors"

	gr "github.com/levigross/grequests"
)

/*
03 тип.
*/

// GetAverageMarkDynReport возвращает динамику среднего балла ученика с сервера первого типа.
func GetAverageMarkDynReport(s *ss.Session, dateBegin, dateEnd, Type, studentID string) (*dt.AverageMarkDynReport, error) {
	p := "http://"

	r0 := func() (bool, error) {
		// 0-ой Post-запрос.
		ro := &gr.RequestOptions{
			Data: map[string]string{
				"AT":        s.AT,
				"LoginType": "0",
				"RPTID":     "2",
				"ThmID":     "1",
				"VER":       s.VER,
			},
			Headers: map[string]string{
				"Origin":                    p + s.Serv.Link,
				"Upgrade-Insecure-Requests": "1",
				"Referer":                   p + s.Serv.Link + "/asp/Reports/Reports.asp",
			},
		}
		r, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/ReportStudentAverageMarkDyn.asp", ro)
		if err != nil {
			return false, err
		}
		defer func() {
			_ = r.Close()
		}()
		return checkResponse(s, r)
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
				"ADT":       dateBegin,
				"AT":        s.AT,
				"BACK":      "/asp/Reports/ReportStudentAverageMarkDyn.asp",
				"DDT":       dateEnd,
				"LoginType": "0",
				"MT":        Type,
				"NA":        "",
				"PCLID":     "",
				"PP":        "/asp/Reports/ReportStudentAverageMarkDyn.asp",
				"RP":        "",
				"RPTID":     "2",
				"RT":        "",
				"SID":       studentID,
				"TA":        "",
				"ThmID":     "1",
				"VER":       s.VER,
			},
			Headers: map[string]string{
				"Origin":                    p + s.Serv.Link,
				"Upgrade-Insecure-Requests": "1",
				"Referer":                   p + s.Serv.Link + "/asp/Reports/ReportStudentAverageMarkDyn.asp",
			},
		}
		r1, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/ReportStudentAverageMarkDyn.asp", ro)
		if err != nil {
			return false, err
		}
		defer func() {
			_ = r1.Close()
		}()
		return checkResponse(s, r1)
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
			return nil, fmt.Errorf("retry didn't work 1 POST")
		}
	}

	// 2-ой Post-запрос.
	r2 := func() ([]byte, bool, error) {
		ro := &gr.RequestOptions{
			Data: map[string]string{
				"A":         "",
				"ADT":       dateBegin,
				"AT":        s.AT,
				"BACK":      "/asp/Reports/ReportStudentAverageMarkDyn.asp",
				"DDT":       dateEnd,
				"LoginType": "0",
				"MT":        Type,
				"NA":        "",
				"PCLID":     "",
				"PP":        "/asp/Reports/ReportStudentAverageMarkDyn.asp",
				"RP":        "",
				"RPTID":     "2",
				"RT":        "",
				"SID":       studentID,
				"TA":        "",
				"ThmID":     "1",
				"VER":       s.VER,
			},
			Headers: map[string]string{
				"Origin":           p + s.Serv.Link,
				"X-Requested-With": "XMLHttpRequest",
				"at":               s.AT,
				"Referer":          p + s.Serv.Link + "/asp/Reports/ReportStudentAverageMarkDyn.asp",
			},
		}
		r, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/StudentAverageMarkDyn.asp", ro)
		if err != nil {
			return nil, false, err
		}
		defer func() {
			_ = r.Close()
		}()
		flag, err := checkResponse(s, r)
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
	// находящуюся в теле ответа, и найти в ней отчет о динамике среднего балла.
	return inner.AverageMarkDynReportParser(bytes.NewReader(b))
}
