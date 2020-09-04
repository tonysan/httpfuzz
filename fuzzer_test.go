package httpfuzz

import (
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
)

func testLogger(t *testing.T) *log.Logger {
	return log.New(testWriter{t}, "test", log.LstdFlags)
}

type testWriter struct {
	t *testing.T
}

func (tw testWriter) Write(p []byte) (n int, err error) {
	tw.t.Log(string(p))
	return len(p), nil
}

func TestFuzzerCalculatesCorrectNumberOfRequests(t *testing.T) {
	wordlist, err := os.Open("testdata/useragents.txt")
	if err != nil {
		t.Fatal(err)
	}
	request, _ := http.NewRequest("GET", "", nil)
	client := &Client{
		Client: &http.Client{},
	}
	config := &Config{
		TargetHeaders:   []string{"Host", "Pragma", "User-Agent"},
		TargetParams:    []string{"fuzz"},
		TargetPathArgs:  []string{"user"},
		FuzzDirectory:   true,
		Wordlist:        wordlist,
		Seed:            &Request{request},
		Client:          client,
		Logger:          testLogger(t),
		TargetDelimiter: '*',
		URLScheme:       "http",
	}
	fuzzer := &Fuzzer{config}
	count, err := fuzzer.RequestCount()
	if err != nil {
		t.Fatal(err)
	}

	const expectedCount = 30
	if count != expectedCount {
		t.Fatalf("Exepected %d requests, got %d", expectedCount, count)
	}
}

func TestFuzzerGeneratesExpectedNumberOfRequests(t *testing.T) {
	wordlist, err := os.Open("testdata/useragents.txt")
	if err != nil {
		t.Fatal(err)
	}
	request, _ := http.NewRequest("GET", "", nil)
	client := &Client{
		Client: &http.Client{},
	}
	config := &Config{
		TargetHeaders:   []string{"Host", "Pragma", "User-Agent"},
		TargetParams:    []string{"fuzz"},
		TargetPathArgs:  []string{"user"},
		FuzzDirectory:   true,
		Wordlist:        wordlist,
		Seed:            &Request{request},
		Client:          client,
		TargetDelimiter: '*',
		Logger:          testLogger(t),
		URLScheme:       "http",
	}
	fuzzer := &Fuzzer{config}
	expectedCount, err := fuzzer.RequestCount()
	if err != nil {
		t.Fatal(err)
	}

	requests := fuzzer.GenerateRequests()
	count := 0
	for job := range requests {
		// A nil request represents the end of stream.
		if job == nil {
			break
		}

		if job.Request == nil {
			t.Fatalf("Nil request received for %+v", *job)
		}

		count++
		// Prevent it from running forever if too many requests come back.
		if count > expectedCount {
			t.Fatalf("Too many requests are being sent, expected %d, got %d", expectedCount, count)
		}
	}

	if count != expectedCount {
		t.Fatalf("Too few requests are being sent, expected %d, got %d", expectedCount, count)
	}
}

func TestFuzzerGeneratesCorrectRequestsRequestBody(t *testing.T) {
	wordlist, err := os.Open("testdata/useragents.txt")
	if err != nil {
		t.Fatal(err)
	}
	request, _ := http.NewRequest("POST", "", strings.NewReader("{\"type\": \"*body*\", \"second\": \"*value*\"}"))
	client := &Client{
		Client: &http.Client{},
	}
	config := &Config{
		TargetHeaders:   []string{"Host", "Pragma", "User-Agent"},
		TargetParams:    []string{"fuzz"},
		FuzzDirectory:   true,
		Wordlist:        wordlist,
		Seed:            &Request{request},
		TargetDelimiter: '*',
		Client:          client,
		Logger:          testLogger(t),
		URLScheme:       "http",
	}
	fuzzer := &Fuzzer{config}
	expectedCount, err := fuzzer.RequestCount()
	if err != nil {
		t.Fatal(err)
	}

	sanityCount := 35
	if expectedCount != sanityCount {
		t.Fatalf("Wrong count, expected %d, got %d", sanityCount, expectedCount)
	}

	requests := fuzzer.GenerateRequests()
	count := 0
	for job := range requests {
		if job.Request == nil {
			t.Fatalf("Nil request received for %+v", *job)
		}

		count++
		// Prevent it from running forever if too many requests come back.
		if count > expectedCount {
			t.Fatalf("Too many requests are being sent, expected %d, got %d", expectedCount, count)
		}
	}

	if count != expectedCount {
		t.Fatalf("Too few requests are being sent, expected %d, got %d", expectedCount, count)
	}
}
