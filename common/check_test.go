package common

import "testing"

func TestConnectivityTest(t *testing.T) {
	type args struct {
		ipPorts string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "test a not exists ip and port",
			args:    args{"192.168.1.1:80"},
			wantErr: true,
		},
		{
			name:    "test a local ip",
			args:    args{"127.0.0.1:90"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ConnectivityTest(tt.args.ipPorts); (err != nil) != tt.wantErr {
				t.Errorf("ConnectivityTest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
