package data

import (
	"net/http"
	"testing"
)

func TestToyReq_Check(t *testing.T) {
	validBase := func() *ToyReq {
		return &ToyReq{
			Url:      "http://example.com",
			Method:   http.MethodGet,
			Thread:   10,
			Duration: 10,
			Timeout:  10,
		}
	}

	tests := []struct {
		name    string
		toyReq  *ToyReq
		wantErr bool
	}{
		{
			name: "URL is empty",
			toyReq: &ToyReq{
				Url:    "",
				Method: http.MethodGet,
			},
			wantErr: true,
		},
		{
			name: "Invalid HTTP method",
			toyReq: &ToyReq{
				Url:    "http://example.com",
				Method: "INVALID",
			},
			wantErr: true,
		},
		{
			name:    "Valid request",
			toyReq:  validBase(),
			wantErr: false,
		},
		{
			name:    "Negative thread rejected", // L3
			toyReq:  func() *ToyReq { r := validBase(); r.Thread = -1; return r }(),
			wantErr: true,
		},
		{
			name:    "Zero thread rejected", // L3
			toyReq:  func() *ToyReq { r := validBase(); r.Thread = 0; return r }(),
			wantErr: true,
		},
		{
			name:    "Thread beyond MaxThread rejected", // L3
			toyReq:  func() *ToyReq { r := validBase(); r.Thread = MaxThread + 1; return r }(),
			wantErr: true,
		},
		{
			name:    "Zero duration rejected", // L3
			toyReq:  func() *ToyReq { r := validBase(); r.Duration = 0; return r }(),
			wantErr: true,
		},
		{
			name:    "Duration beyond MaxDuration rejected", // L3
			toyReq:  func() *ToyReq { r := validBase(); r.Duration = MaxDuration + 1; return r }(),
			wantErr: true,
		},
		{
			name:    "Zero timeout rejected", // L3
			toyReq:  func() *ToyReq { r := validBase(); r.Timeout = 0; return r }(),
			wantErr: true,
		},
		{
			name:    "Header without colon rejected", // S2
			toyReq:  func() *ToyReq { r := validBase(); r.Header = MyStrSlice{"badheader"}; return r }(),
			wantErr: true,
		},
		{
			name:    "Header with CRLF rejected", // S2
			toyReq:  func() *ToyReq { r := validBase(); r.Header = MyStrSlice{"X-A: v\r\nX-Injected: 1"}; return r }(),
			wantErr: true,
		},
		{
			name:    "Header with newline in name rejected", // S2
			toyReq:  func() *ToyReq { r := validBase(); r.Header = MyStrSlice{"X-B\nN: v"}; return r }(),
			wantErr: true,
		},
		{
			name: "Valid headers pass", // S2
			toyReq: func() *ToyReq {
				r := validBase()
				r.Header = MyStrSlice{"Content-Type: application/json", "X-Toy-Test: abc"}
				return r
			}(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.toyReq.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ToyReq.Check() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheckHttpMethod(t *testing.T) {
	tests := []struct {
		name    string
		method  string
		wantErr bool
	}{
		{
			name:    "Valid GET method",
			method:  http.MethodGet,
			wantErr: false,
		},
		{
			name:    "Valid POST method",
			method:  http.MethodPost,
			wantErr: false,
		},
		{
			name:    "Invalid method",
			method:  "INVALID",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkHttpMethod(tt.method)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkHttpMethod() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseHeaders(t *testing.T) {
	r := &ToyReq{Header: MyStrSlice{"Content-Type:  application/json ", "X-Toy: v1"}}
	parsed, err := r.ParseHeaders()
	if err != nil {
		t.Fatalf("ParseHeaders() unexpected error: %v", err)
	}
	if len(parsed) != 2 {
		t.Fatalf("ParseHeaders() len = %d, want 2", len(parsed))
	}
	if parsed[0][0] != "Content-Type" || parsed[0][1] != "application/json" {
		t.Errorf("ParseHeaders() trims failed: %+v", parsed[0])
	}
	if parsed[1] != [2]string{"X-Toy", "v1"} {
		t.Errorf("ParseHeaders() = %+v, want [X-Toy v1]", parsed[1])
	}
}

func TestMyStrSlice_Set(t *testing.T) {
	s := &MyStrSlice{}
	value := "test: value"
	err := s.Set(value)
	if err != nil {
		t.Errorf("MyStrSlice.Set() error = %v, wantErr nil", err)
	}
	if len(*s) != 1 || (*s)[0] != value {
		t.Errorf("MyStrSlice.Set() = %v, want %v", *s, []string{value})
	}
}

func TestMyStrSlice_String(t *testing.T) {
	s := &MyStrSlice{"a", "b", "c"}
	result := s.String()
	expected := "[a b c]"
	if result != expected {
		t.Errorf("MyStrSlice.String() = %v, want %v", result, expected)
	}
}
