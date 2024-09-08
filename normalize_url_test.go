package main

import (
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
