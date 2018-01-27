package weatherline

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNotifyError_Error(t *testing.T) {
	err := notifyError{
		Status:  10,
		Message: "TEST Message",
	}

	expected := "10: TEST Message"

	actual := err.Error()
	if actual != expected {
		t.Errorf("Expected to get [%s], but got [%s]", expected, actual)
	}
}

func TestNewLineNotify(t *testing.T) {
	token := "abcde"

	notify := NewLineNotify(token)
	if notify == nil {
		t.Fatal("function returns nil")
	}

	n, ok := notify.(*lineNotify)
	if !ok {
		t.Fatal("Expected lineNotify instance, but not.")
	}

	if n.token != token {
		t.Fatalf("token is not same: %s != %s", n.token, token)
	}

	if n.url != lineNotifyAPIBase {
		t.Fatalf("Expected url is %s, but it's %s.", lineNotifyAPIBase, n.url)
	}

	if n.httpClient == nil {
		t.Fatal("httpClient is nil")
	}
}

func writeLineNotifyResponse(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	if _, err := fmt.Fprintf(w, `{"status":%d,"message":"%s"}`, status, message); err != nil {
		panic(err)
	}
}

func lineNotifyFunc(token, msg string, resStatus int, resMessage string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeLineNotifyResponse(w, http.StatusBadRequest, fmt.Sprintf("Unexpected request: method = %s", r.Method))
			return
		}

		if r.URL.Path != "/api/notify" {
			writeLineNotifyResponse(w, http.StatusBadRequest, fmt.Sprintf("Unexpected request: path = %s", r.URL.Path))
			return
		}

		if contentType := r.Header.Get("Content-Type"); contentType != "application/x-www-form-urlencoded" {
			writeLineNotifyResponse(w, http.StatusBadRequest, fmt.Sprintf("Unexpected request: `Content-Type` header = %s", contentType))
			return
		}

		if authorization := r.Header.Get("Authorization"); authorization != "Bearer "+token {
			writeLineNotifyResponse(w, http.StatusBadRequest, fmt.Sprintf("Unexpected request: `Authorization` header = %s", authorization))
			return
		}

		err := r.ParseForm()
		if err != nil {
			writeLineNotifyResponse(w, http.StatusBadRequest, fmt.Sprintf("Failed to parse form: %v", err))
			return
		}

		if postMsg := r.PostFormValue("message"); postMsg != msg {
			writeLineNotifyResponse(w, http.StatusBadRequest, fmt.Sprintf("Unexpected request: `message` data = %s", postMsg))
			return
		}

		writeLineNotifyResponse(w, resStatus, resMessage)
	}
}

func TestLineNotify_Send(t *testing.T) {
	tests := []struct {
		token      string
		resStatus  int
		resMessage string

		n   lineNotify
		msg string

		expected error
	}{
		// TEST0 {{{
		{
			token:      "XXXXX",
			resStatus:  http.StatusOK,
			resMessage: "OK",

			n:   lineNotify{token: "XXXXX"},
			msg: "TEST0",

			expected: nil,
		},
		// }}}
		// TEST1 {{{
		{
			token: "XXXXX",

			n:   lineNotify{token: "YYYYY"},
			msg: "TEST1",

			expected: fmt.Errorf("%d: Unexpected request: `Authorization` header = Bearer YYYYY", http.StatusBadRequest),
		},
		// }}}
		// TEST2 {{{
		{
			token:      "XXXXX",
			resStatus:  http.StatusInternalServerError,
			resMessage: "InternalServerError",

			n:   lineNotify{token: "XXXXX"},
			msg: "TEST2",

			expected: notifyError{
				Status:  http.StatusInternalServerError,
				Message: "InternalServerError",
			},
		},
		// }}}
	}

	for i, tt := range tests {
		tt := tt // capture
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			t.Parallel()

			server := httptest.NewTLSServer(http.HandlerFunc(lineNotifyFunc(tt.token, tt.msg, tt.resStatus, tt.resMessage)))
			defer server.Close()

			tt.n.httpClient = server.Client()
			tt.n.url = server.URL

			err := tt.n.Send(tt.msg)
			if err != nil {
				if tt.expected == nil {
					t.Errorf("Expected no error occurred, but it occurred (%v)", err)
				} else if tt.expected != err && err.Error() != tt.expected.Error() {
					t.Errorf("Expected to get [%v], but got [%v]", tt.expected, err)
				}
			} else {
				if tt.expected != nil {
					t.Errorf("It was expected that an error occurred, but it did not occur")
				}
			}
		})
	}
}
