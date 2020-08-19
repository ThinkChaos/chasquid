package expvarom

import (
	"expvar"
	"io/ioutil"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var (
	testI1 = NewInt("testI1", "int test var")
	testI2 = expvar.NewInt("testI2")

	testF = NewFloat("testF", "float test var")

	testMI  = NewMap("testMI", "label", "int map test var")
	testMF  = NewMap("testMF", "label", "float map test var")
	testMXI = expvar.NewMap("testMXI")
	testMXF = expvar.NewMap("testMXF")

	testMEmpty = expvar.NewMap("testMEmpty") //nolint // Unused.

	testMOther = expvar.NewMap("testMOther")

	testS = expvar.NewString("testS")

	// Naming test cases.
	testN1 = expvar.NewInt("name/1z")
	testN2 = NewInt("name$2", "name with $")
	testN3 = expvar.NewInt("3name")
	testN4 = expvar.NewInt("nAme_4Z")
	testN5 = expvar.NewInt("ñame_5")
)

const expected string = `_ame_5 5

i3name 3

nAme_4Z 4

name_1z 1

# HELP name_2 name with $
name_2 2

# HELP testF float test var
testF 3.43434

# HELP testI1 int test var
testI1 1

testI2 2

# HELP testMF float map test var
testMF{label="key2.0"} 6.6
testMF{label="key2.1"} 6.61
testMF{label="key2.2-ñaca"} 6.62
testMF{label="key2.3-a\\b"} 6.63
testMF{label="key2.4-	"} 6.64
testMF{label="key2.5-a\nb"} 6.65
testMF{label="key2.6-a\"b"} 6.66
testMF{label="key2.7-\\u00f1aca-A\\t\\xff\\xfe\\xfdB"} 6.67

# HELP testMI int map test var
testMI{label="key1"} 5

testMXF{key="key4"} 8e-08

testMXI{key="key3"} 7

# Generated by expvarom
# EXPERIMENTAL - Format is not fully standard yet
# Ignored variables: ["cmdline" "memstats" "testS"]
`

func TestHandler(t *testing.T) {
	testI1.Add(1)
	testI2.Add(2)
	testF.Add(3.43434)
	testMI.Add("key1", 5)

	// Test some strange keys in this map to check they're escaped properly.
	testMF.AddFloat("key2.0", 6.60)
	testMF.AddFloat("key2.1", 6.61)
	testMF.AddFloat("key2.2-ñaca", 6.62)
	testMF.AddFloat(`key2.3-a\b`, 6.63)
	testMF.AddFloat("key2.4-\t", 6.64)
	testMF.AddFloat("key2.5-a\nb", 6.65)
	testMF.AddFloat(`key2.6-a"b`, 6.66)
	testMF.AddFloat("key2.7-ñaca-A\t\xff\xfe\xfdB", 6.67) // Invalid utf8.

	testMXI.Add("key3", 7)
	testMXF.AddFloat("key4", 8e-8)
	testS.Set("lalala")

	testN1.Add(1)
	testN2.Add(2)
	testN3.Add(3)
	testN4.Add(4)
	testN5.Add(5)

	// Map with an unsupported type.
	testMOther.Set("keyX", &expvar.String{})

	req := httptest.NewRequest("get", "/metrics", nil)
	w := httptest.NewRecorder()
	MetricsHandler(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	if diff := cmp.Diff(expected, string(body)); diff != "" {
		t.Errorf("MetricsHandler() mismatch (-want +got):\n%s", diff)
	}
}

func TestMapLabelAccident(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("NewMap did not panic as expected")
		}
	}()

	NewMap("name", "label with spaces", "description")
}
