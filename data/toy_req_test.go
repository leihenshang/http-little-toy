package data

import (
	"net/http"
	"testing"
)

func TestToyReq_Check(t *testing.T) {
	tests := []struct {
		name    string
		toyReq  *ToyReq
		wantErr bool
	}{
		{
			name: "URL is empty",
			toyReq: &ToyReq{
				Url: "",
				Method: http.MethodGet,
			},
			wantErr: true,
		},
		{
			name: "Invalid HTTP method",
			toyReq: &ToyReq{
				Url: "http://example.com",
				Method: "INVALID",
			},
			wantErr: true,
		},
		{
			name: "Valid request",
			toyReq: &ToyReq{
				Url: "http://example.com",
				Method: http.MethodGet,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.toyReq.Check()
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
			name: "Valid GET method",
			method: http.MethodGet,
			wantErr: false,
		},
		{
			name: "Valid POST method",
			method: http.MethodPost,
			wantErr: false,
		},
		{
			name: "Invalid method",
			method: "INVALID",
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