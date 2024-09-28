package main

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name          string
		inputURL      string
		expected      string
		errorContains string
	}{
		{
			name:     "remove scheme",
			inputURL: "https://blog.boot.dev/path",
			expected: "blog.boot.dev/path",
		},
		{
			name:     "remove trailing slash",
			inputURL: "https://blog.boot.dev/path/",
			expected: "blog.boot.dev/path",
		},
		{
			name:     "lowercase capital letters",
			inputURL: "https://BLOG.boot.dev/PATH",
			expected: "blog.boot.dev/path",
		},
		{
			name:     "remove scheme and capitals and trailing slash",
			inputURL: "http://BLOG.boot.dev/path/",
			expected: "blog.boot.dev/path",
		},
		{
			name:          "handle invalid URL",
			inputURL:      `:\\invalidURL`,
			expected:      "",
			errorContains: "couldn't parse URL",
		},
		{
			name:     "multiple slashes",
			inputURL: `https://BLOG.boot.dev/PATH//`,
			expected: "blog.boot.dev/path",
		},
	}

	for i, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := normalizeURL(tc.inputURL)
			if err != nil && !strings.Contains(err.Error(), tc.errorContains) {
				t.Errorf("Test %v - '%s' FAIL: unexpected error: %v", i, tc.name, err)
				return
			} else if err != nil && tc.errorContains == "" {
				t.Errorf("Test %v - '%s' FAIL: unexpected error: %v", i, tc.name, err)
				return
			} else if err == nil && tc.errorContains != "" {
				t.Errorf("Test %v - '%s' FAIL: expected error containing '%v', got none.", i, tc.name, tc.errorContains)
				return
			}

			if actual != tc.expected {
				t.Errorf("Test %v - %s FAIL: expected URL: %v, actual: %v", i, tc.name, tc.expected, actual)
			}
		})
	}
}

func TestGetURLsFromHTML(t *testing.T) {
	cases := []struct {
		name          string
		inputURL      string
		inputBody     string
		expected      []string
		errorContains string
	}{
		{
			name:     "absolute URL",
			inputURL: "https://blog.boot.dev",
			inputBody: `
<html>
	<body>
		<a href="https://blog.boot.dev">
			<span>Boot.dev</span>
		</a>
	</body>
</html>
`,
			expected: []string{"https://blog.boot.dev"},
		},
		{
			name:     "relative URL",
			inputURL: "https://blog.boot.dev",
			inputBody: `
<html>
	<body>
		<a href="/path/one">
			<span>Boot.dev</span>
		</a>
	</body>
</html>
`,
			expected: []string{"https://blog.boot.dev/path/one"},
		},
		{
			name:     "absolute and relative URLs",
			inputURL: "https://blog.boot.dev",
			inputBody: `
<html>
	<body>
		<a href="/path/one">
			<span>Boot.dev</span>
		</a>
		<a href="https://other.com/path/one">
			<span>Boot.dev</span>
		</a>
	</body>
</html>
`,
			expected: []string{"https://blog.boot.dev/path/one", "https://other.com/path/one"},
		},
		{
			name:     "no href",
			inputURL: "https://blog.boot.dev",
			inputBody: `
<html>
	<body>
		<a>
			<span>Boot.dev></span>
		</a>
	</body>
</html>
`,
			expected: nil,
		},
		{
			name:     "bad HTML",
			inputURL: "https://blog.boot.dev",
			inputBody: `
<html body>
	<a href="path/one">
		<span>Boot.dev></span>
	</a>
</html body>
`,
			expected: []string{"https://blog.boot.dev/path/one"},
		},
		{
			name:     "invalid href URL",
			inputURL: "https://blog.boot.dev",
			inputBody: `
<html>
	<body>
		<a href=":\\invalidURL">
			<span>Boot.dev</span>
		</a>
	</body>
</html>
`,
			expected: nil,
		},
		{
			name:     "handle invalid base URL",
			inputURL: `:\\invalidBaseURL`,
			inputBody: `
<html>
	<body>
		<a href="/path">
			<span>Boot.dev</span>
		</a>
	</body>
</html>
`,
			expected:      nil,
			errorContains: "couldn't parse base URL",
		},
		{
			name:     "handle single /",
			inputURL: `https://blog.boot.dev`,
			inputBody: `
<html>
	<body>
		<a href="/">
			<span>Boot.dev</span>
		</a>
	</body>
</html>
`,
			expected: []string{"https://blog.boot.dev"},
		},
		{
			name:     "overlapping path",
			inputURL: `https://blog.boot.dev/tags`,
			inputBody: `
<html>
	<body>
		<a href="/tags/business">
			<span>Boot.dev</span>
		</a>
	</body>
</html>
`,
			expected: []string{"https://blog.boot.dev/tags/business"},
		},
		{
			name:     "deal with filename in url",
			inputURL: `https://blog.boot.dev/tags`,
			inputBody: `
<html>
	<body>
		<a href="/tags.xml">
			<span>Boot.dev</span>
		</a>
	</body>
</html>
`,
			expected: nil,
		},
		{
			name:     "duplicated urls",
			inputURL: "https://blog.boot.dev",
			inputBody: `
<html>
	<body>
		<a href="https://blog.boot.dev">
			<span>Boot.dev</span>
		</a>
		<a href="https://blog.boot.dev">
			<span>Boot.dev</span>
		</a>
	</body>
</html>
`,
			expected: []string{"https://blog.boot.dev"},
		},
	}

	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := getURLsFromHTML(tc.inputBody, tc.inputURL)
			if err != nil && !strings.Contains(err.Error(), tc.errorContains) {
				t.Errorf("Test %v - '%s' FAIL: unexpected error: %v", i, tc.name, err)
				return
			} else if err != nil && tc.errorContains == "" {
				t.Errorf("Test %v - '%s' FAIL: unexpected error: %v", i, tc.name, err)
				return
			} else if err == nil && tc.errorContains != "" {
				t.Errorf("Test %v - '%s' FAIL: expected error containing '%v', got none.", i, tc.name, tc.errorContains)
				return
			}

			if !reflect.DeepEqual(actual, tc.expected) {
				t.Errorf("Test %v - '%s' FAIL: expected URLs %v, got URLs %v", i, tc.name, tc.expected, actual)
				return
			}
		})
	}
}

// Test the sortMap method, should return a properly sorted slice of type kv
func TestSortMap(t *testing.T) {
	t.Run("test sort properly sorts by value first", func(t *testing.T) {
		test := make(map[string]int)
		test["google.com"] = 7
		test["cnn.com"] = 1
		test["abc.com"] = 2
		test["ftc.gov"] = 4
		test["eff.org"] = 6

		want := []kv{
			{key: "google.com", val: 7},
			{key: "eff.org", val: 6},
			{key: "ftc.gov", val: 4},
			{key: "abc.com", val: 2},
			{key: "cnn.com", val: 1},
		}

		got := sortMap(test)

		fail := false
		for i := range want {
			if got[i].key != want[i].key && got[i].val != want[i].val {
				fail = true
			}
		}

		if fail {
			t.Error("FAIL: expected same answer for both results")
			fmt.Println("results:", want)
			fmt.Println("expected:", got)
		}
	})

	t.Run("test multiple with same key", func(t *testing.T) {
		test := make(map[string]int)
		test["google.com"] = 7
		test["cnn.com"] = 1
		test["wired.com"] = 1
		test["headlights.com"] = 2
		test["arrival.com"] = 3
		test["abc.com"] = 2
		test["ftc.gov"] = 4
		test["wikipedia.org"] = 9
		test["eff.org"] = 6

		want := []kv{
			{key: "wikipedia.org", val: 9},
			{key: "google.com", val: 7},
			{key: "eff.org", val: 6},
			{key: "ftc.gov", val: 4},
			{key: "arrival.com", val: 3},
			{key: "abc.com", val: 2},
			{key: "headlights.com", val: 2},
			{key: "cnn.com", val: 1},
			{key: "wired.com", val: 1},
		}

		got := sortMap(test)

		fail := false
		for i := range want {
			if got[i].key != want[i].key && got[i].val != want[i].val {
				fail = true
			}
		}

		if fail {
			t.Error("FAIL: expected same answer for both results")
			fmt.Println("results:", want)
			fmt.Println("expected:", got)
		}

	})

}
