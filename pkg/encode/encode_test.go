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

func TestEncodePath_CustomEncoder(t *testing.T) {
	// A custom encoder that upper-cases each segment instead of URL-encoding.
	upper := func(s string) string {
		result := make([]byte, len(s))
		for i := range s {
			c := s[i]
			if c >= 'a' && c <= 'z' {
				c -= 32
			}
			result[i] = c
		}
		return string(result)
	}

	in := []string{"foo", "bar", "baz"}
	out := EncodePath(upper, in...)
	assert.Equal(t, "FOO/BAR/BAZ", out)
}

func TestEncodePath_SingleSegmentStringSplit(t *testing.T) {
	// When a single slash-delimited string is passed it is split on "/" and each
	// resulting segment is encoded individually. Spaces within a segment are
	// encoded too, but the "/" delimiter is consumed by the split — not encoded.
	got := EncodePath(nil, "k&r/hello world/baz")
	// Split produces ["k&r", "hello world", "baz"]; each is Encode()'d.
	assert.Equal(t, "k%26r/hello+world/baz", got)
}

func TestEncodePath_CustomEncoder_NilFallsBackToEncode(t *testing.T) {
	// Passing nil encoder must fall back to the default Encode function.
	in := []string{"hello world", "k&r"}
	out := EncodePath(nil, in...)
	assert.Equal(t, "hello+world/k%26r", out)
}
