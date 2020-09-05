package entity

import "testing"

func TestNewID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		count int
	}{
		{name: "Valid", count: 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ids := make(map[ID]struct{})

			for i := 0; i <= tt.count; i++ {
				id := NewID()
				if _, ok := ids[id]; ok {
					t.Error("Created ID already exists")
					return
				}
				ids[id] = struct{}{}
			}

		})
	}
}
