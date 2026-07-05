package calculator

import "testing"

func TestAdd(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int
		expected int
	}{
		{"positive numbers", 2, 3, 5},
		{"with zero", 5, 0, 5},
		{"negative numbers", -2, -3, -5},
		{"mixed signs", -2, 5, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Add(tt.a, tt.b)
			if got != tt.expected {
				t.Errorf("Add(%d, %d) = %d; want %d", tt.a, tt.b, got, tt.expected)
			}
		})
	}
}

func TestSubtract(t *testing.T) {
	got := Subtract(10, 4)
	want := 6
	if got != want {
		t.Errorf("Subtract(10, 4) = %d; want %d", got, want)
	}
}

func TestMultiply(t *testing.T) {
	got := Multiply(3, 4)
	want := 12
	if got != want {
		t.Errorf("Multiply(3, 4) = %d; want %d", got, want)
	}
}

func TestDivide(t *testing.T) {
	t.Run("valid division", func(t *testing.T) {
		got, err := Divide(10, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 5 {
			t.Errorf("Divide(10, 2) = %d; want 5", got)
		}
	})

	t.Run("divide by zero", func(t *testing.T) {
		_, err := Divide(10, 0)
		if err == nil {
			t.Error("expected an error when dividing by zero, got nil")
		}
	})
}
