package cli

import (
	"os"
	"testing"
)

func TestGetValidInput(t *testing.T) {
	// Save original stdin
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	tests := []struct {
		name      string
		inputs    []string
		validator func(string) error
		want      string
	}{
		{
			name:      "Valid input first try",
			inputs:    []string{"test\n"},
			validator: validateFirstName,
			want:      "test",
		},
		{
			name:      "Valid input after invalid",
			inputs:    []string{"123\n", "test\n"},
			validator: validateFirstName,
			want:      "test",
		},
		{
			name:      "Valid email",
			inputs:    []string{"test@example.com\n"},
			validator: validateEmail,
			want:      "test@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a pipe and make it stdin
			r, w, _ := os.Pipe()
			os.Stdin = r

			// Write all inputs
			go func() {
				for _, input := range tt.inputs {
					w.Write([]byte(input))
				}
				w.Close()
			}()

			if got := getValidInput("Test: ", tt.validator); got != tt.want {
				t.Errorf("getValidInput() = %v, want %v", got, tt.want)
			}
		})
	}
}
