package junit

import (
	"encoding/xml"
	"time"
)

type Failure struct {
	Message  string `xml:"message,attr"`
	Type     string `xml:"type,attr"`
	Contents string `xml:",chardata"`
}

type Error struct {
	Message  string `xml:"message,attr"`
	Type     string `xml:"type,attr"`
	Contents string `xml:",chardata"`
}

type Property struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

type Skipped struct {
	Contents string `xml:",chardata"`
}

type SystemErr struct {
	Contents string `xml:",chardata"`
}

type SystemOut struct {
	Contents string `xml:",chardata"`
}

type TestCase struct {
	XMLName    xml.Name   `xml:"testcase"`
	Name       string     `xml:"name,attr"`
	Classname  string     `xml:"classname,attr"`
	Assertions string     `xml:"assertions,attr"`
	Status     string     `xml:"status,attr"`
	Time       string     `xml:"time,attr"`
	Skipped    *Skipped   `xml:"skipped,omitempty"`
	Failure    *Failure   `xml:"failure,omitempty"`
	Error      *Error     `xml:"error,omitempty"`
	SystemOut  *SystemOut `xml:"system-out,omitempty"`
	SystemErr  *SystemErr `xml:"system-err,omitempty"`
}

type TestSuite struct {
	XMLName    xml.Name    `xml:"testsuite"`
	Name       string      `xml:"name,attr"`
	Tests      int         `xml:"tests,attr"`
	Failures   int         `xml:"failures,attr"`
	Errors     int         `xml:"errors,attr"`
	Disabled   int         `xml:"disabled,attr"`
	Skipped    int         `xml:"skipped,attr"`
	Time       string      `xml:"time,attr"`
	Hostname   string      `xml:"hostname,attr"`
	ID         string      `xml:"id,attr"`
	Package    string      `xml:"package,attr"`
	Timestamp  time.Time   `xml:"timestamp,attr"`
	SystemOut  *SystemOut  `xml:"system-out,omitempty"`
	SystemErr  *SystemErr  `xml:"system-err,omitempty"`
	Properties []*Property `xml:"properties>property,omitempty"`
	TestCases  []*TestCase
}

func (ts *TestSuite) CurrentCase() *TestCase {
	return ts.TestCases[len(ts.TestCases)-1]
}

type TestSuites struct {
	XMLName    xml.Name `xml:"testsuites"`
	Name       string   `xml:"name,attr"`
	Tests      int      `xml:"tests,attr"`
	Failures   int      `xml:"failures,attr"`
	Errors     int      `xml:"errors,attr"`
	Disabled   int      `xml:"disabled,attr"`
	Time       string   `xml:"time,attr"`
	TestSuites []*TestSuite
}

func (ts *TestSuites) CurrentSuite() *TestSuite {
	return ts.TestSuites[len(ts.TestSuites)-1]
}

type JUnit []*TestSuites

func (j JUnit) CurrentSuites() *TestSuites {
	return j[len(j)-1]
}
