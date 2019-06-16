package live

import "testing"

func Test_getUserId(t *testing.T) {
	type args struct {
		fs string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
		{
			name: "",
			args: args{fs: "567399844883"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i, err := getUserID(tt.args.fs)
			if err != nil {
				t.Error(err)
				return
			}
			t.Log(i)
		})
	}
}
