package core

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	. "github.com/Comcast/sheens/util/testutil"
)

type MatchTest struct {
	Pattern       interface{}              `json:"p"`
	Message       interface{}              `json:"m"`
	Expected      []map[string]interface{} `json:"w,omitempty"`
	Error         bool                     `json:"err,omitempty"`
	Title         string                   `json:"title,omitempty"`
	Doc           string                   `json:"doc,omitempty"`
	Verbose       bool                     `json:"verbose,omitempty"`
	BenchmarkOnly bool                     `json:"benchmarkOnly,omitempty"`
}

func (t MatchTest) Name(i int) string {
	if t.Title == "" {
		return fmt.Sprintf("%d", i)
	} else {
		return fmt.Sprintf("%03d %s", i, t.Title)
	}
}

func (t MatchTest) Fprintf(w io.Writer, i int) {
	title := t.Title
	if title == "" {
		title = "Anonymous example"
	}
	fmt.Fprintf(w, "\n## %d. %s\n\n", i, title)
	if t.Doc != "" {
		fmt.Fprintf(w, "\n%s\n", t.Doc)
	}
	fmt.Fprintf(w, "The pattern\n```JSON\n%s\n```\n\n", JS(t.Pattern))
	fmt.Fprintf(w, "matched against\n```JSON\n%s\n```\n\n", JS(t.Message))
	if t.Error {
		fmt.Fprintf(w, "should return an error.\n")
	} else {
		fmt.Fprintf(w, "should return\n```JSON\n%s\n```\n", JS(t.Expected))
	}
}

func compareMatchResult(bss []Bindings, expected []map[string]interface{}) bool {
	if len(bss) != len(expected) {
		return false
	}

	m := make(map[int]map[string]interface{})
	for i, got := range bss {
		m[i] = map[string]interface{}(got)
	}

	for _, e := range expected {
		found := false
		for k, v := range m {
			if reflect.DeepEqual(e, v) {
				delete(m, k)
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return 0 == len(m)
}

func (mt *MatchTest) Run(t *testing.T, check bool) {
	bss, err := Matches(nil, mt.Pattern, mt.Message)
	if !check {
		return
	}
	if err != nil {
		if !mt.Error {
			t.Fatal(err)
		}
	} else {
		if mt.Error {
			t.Fatal("expected an error")
		}
	}

	if !compareMatchResult(bss, mt.Expected) {
		t.Fatalf("match test failed: pattern: %s message: %s got: %s expected: %s\n",
			JS(mt.Pattern), JS(mt.Message), JS(bss), JS(mt.Expected))
	}
}

func getMatchTests() ([]MatchTest, error) {
	js, err := ioutil.ReadFile("match_test.js")
	if err != nil {
		return nil, err
	}
	var tests []MatchTest
	if err = json.Unmarshal(js, &tests); err != nil {
		return nil, err
	}
	return tests, nil
}

func TestMatch(t *testing.T) {
	tests, err := getMatchTests()
	if err != nil {
		t.Fatal(err)
	}
	md, err := os.Create("match.md")
	if err != nil {
		t.Fatal(err)
	}
	defer md.Close()

	fmt.Fprintf(md, `# Pattern matching examples

Generated from test cases.

`)

	for i, test := range tests {
		if test.BenchmarkOnly {
			continue
		}
		test.Fprintf(md, i)
		t.Run(test.Name(i), func(t *testing.T) {
			test.Run(t, true)
		})
	}
}

func BenchmarkMatch(b *testing.B) {
	tests, err := getMatchTests()
	if err != nil {
		b.Fatal(err)
	}
	for i, test := range tests {
		b.Run(test.Name(i), func(b *testing.B) {
			for j := 0; j < b.N; j++ {
				test.Run(nil, false)
			}
		})
	}
}
