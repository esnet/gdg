package encode

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeAndEscapeSpecialChars(t *testing.T) {
	in := "Stardust perfSONAR"
	expected := "Stardust\\+perfSONAR"
	result := EncodeEscapeSpecialChars(in)
	assert.Equal(t, result, expected)
	result = DecodeEscapeSpecialChars(result)
	assert.Equal(t, result, in)
}

func TestEncode(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
		skip bool
	}{
		{
			name: "basic test1",
			args: args{
				s: "k&r",
			},
			want: "k%26r",
		},
		{
			name: "basic test 2",
			args: args{
				s: "t / n",
			},
			want: "t+%2F+n",
		},
		{
			name: "stardust test",
			args: args{
				s: "Stardust perfSONAR",
			},
			want: "Stardust+perfSONAR",
		},
	}
	for _, tt := range tests {
		if tt.skip {
			t.Log("Skipping test", "name", tt.name)
			continue
		}

		res := Encode(tt.args.s)
		assert.Equal(t, tt.want, res)
		assert.Equal(t, Decode(res), tt.args.s)
	}
}

func TestEncodePath(t *testing.T) {
	in := []string{"t", "n / t", "booh", "k&r"}
	out := EncodePath(nil, in...)
	assert.Equal(t, "t/n+%2F+t/booh/k%26r", out)
}
