package main

import "testing"

func TestUnpacking(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "base1", input: "a4bc2d5e", want: "aaaabccddddde", wantErr: false},
		{name: "base2", input: "abcd", want: "abcd", wantErr: false},
		{name: "invalid string", input: "45", want: "", wantErr: true},
		{name: "empty string", input: "", want: "", wantErr: false},
		{name: "slash", input: "qwe\\4\\5", want: "qwe45", wantErr: false},
		{name: "slash with digit", input: "qwe\\45", want: "qwe44444", wantErr: false},
		{name: "big digit 1", input: "a10b", want: "aaaaaaaaaab", wantErr: false},
		{name: "big digit 2", input: "a10b11", want: "aaaaaaaaaabbbbbbbbbbb", wantErr: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := Unpacking(test.input)
			if test.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("expected nil, got error: %v", err)
				}
			}
			if got != test.want {
				t.Errorf("expected %s, got %s", test.want, got)
			}
		})
	}
}
