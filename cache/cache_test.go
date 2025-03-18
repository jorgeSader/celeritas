package cache

import (
	"testing"
	"time"
)

func TestRedisCache_Has(t *testing.T) {
	// Reset cache for isolation
	if err := resetCache(); err != nil {
		t.Fatalf("Failed to reset cache: %v", err)
	}

	tests := []struct {
		name    string
		key     string
		setup   func() error
		want    bool
		wantErr bool
		errMsg  string
	}{
		{
			name:    "non-existent key",
			key:     "test",
			want:    false,
			wantErr: false,
		},
		{
			name: "existent key",
			key:  "test",
			setup: func() error {
				return testRedisCache.Set("test", "hello")
			},
			want:    true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				if err := tt.setup(); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}
			got, err := testRedisCache.Has(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Has() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("Has() error = %v, want %v", err, tt.errMsg)
			}
			if got != tt.want {
				t.Errorf("Has() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRedisCache_Get(t *testing.T) {
	if err := resetCache(); err != nil {
		t.Fatalf("Failed to reset cache: %v", err)
	}

	tests := []struct {
		name    string
		key     string
		setup   func() error
		want    interface{}
		wantErr bool
	}{
		{
			name:    "non-existent key",
			key:     "test",
			want:    nil,
			wantErr: false,
		},
		{
			name: "existent key",
			key:  "test",
			setup: func() error {
				return testRedisCache.Set("test", "hello world")
			},
			want:    "hello world",
			wantErr: false,
		},
		{
			name: "expired key",
			key:  "test",
			setup: func() error {
				return testRedisCache.Set("test", "temp", 1) // 1-second TTL
			},
			want:    nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				if err := tt.setup(); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
				if tt.name == "expired key" {
					testRedisServer.FastForward(2 * time.Second) // Use global server
				}
			}
			got, err := testRedisCache.Get(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRedisCache_Set(t *testing.T) {
	if err := resetCache(); err != nil {
		t.Fatalf("Failed to reset cache: %v", err)
	}

	tests := []struct {
		name    string
		key     string
		value   interface{}
		expires []int
		wantErr bool
	}{
		{
			name:    "set without expiration",
			key:     "test",
			value:   "hello",
			wantErr: false,
		},
		{
			name:    "set with expiration",
			key:     "test2",
			value:   42,
			expires: []int{1},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := testRedisCache.Set(tt.key, tt.value, tt.expires...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				got, err := testRedisCache.Get(tt.key)
				if err != nil {
					t.Errorf("Get() after Set() failed: %v", err)
					return
				}
				if got != tt.value {
					t.Errorf("Set() stored %v, want %v", got, tt.value)
				}
			}
		})
	}
}

func TestRedisCache_Forget(t *testing.T) {
	if err := resetCache(); err != nil {
		t.Fatalf("Failed to reset cache: %v", err)
	}

	tests := []struct {
		name    string
		key     string
		setup   func() error
		wantErr bool
	}{
		{
			name:    "delete existent key",
			key:     "test",
			setup:   func() error { return testRedisCache.Set("test", "data") },
			wantErr: false,
		},
		{
			name:    "delete non-existent key",
			key:     "test2",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				if err := tt.setup(); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}
			err := testRedisCache.Forget(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Forget() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			exists, err := testRedisCache.Has(tt.key)
			if err != nil {
				t.Errorf("Has() after Forget() failed: %v", err)
				return
			}
			if exists {
				t.Errorf("Key %s still exists after Forget()", tt.key)
			}
		})
	}
}

func TestRedisCache_Empty(t *testing.T) {
	if err := resetCache(); err != nil {
		t.Fatalf("Failed to reset cache: %v", err)
	}

	// Setup multiple keys
	keys := []string{"test1", "test2", "test3"}
	for _, key := range keys {
		if err := testRedisCache.Set(key, "data"); err != nil {
			t.Fatalf("Setup failed: %v", err)
		}
	}

	err := testRedisCache.Empty()
	if err != nil {
		t.Errorf("Empty() error = %v", err)
		return
	}

	for _, key := range keys {
		exists, err := testRedisCache.Has(key)
		if err != nil {
			t.Errorf("Has() after Empty() failed for %s: %v", key, err)
			return
		}
		if exists {
			t.Errorf("Key %s still exists after Empty()", key)
		}
	}
}

func TestRedisCache_EmptyByMatch(t *testing.T) {
	if err := resetCache(); err != nil {
		t.Fatalf("Failed to reset cache: %v", err)
	}

	// Setup keys with a pattern
	keys := map[string]string{
		"user:1": "data1",
		"user:2": "data2",
		"other":  "data3",
	}
	for key, value := range keys {
		if err := testRedisCache.Set(key, value); err != nil {
			t.Fatalf("Setup failed: %v", err)
		}
	}

	err := testRedisCache.EmptyByMatch("user*")
	if err != nil {
		t.Errorf("EmptyByMatch() error = %v", err)
		return
	}

	tests := []struct {
		key      string
		shouldBe bool // true if key should still exist
	}{
		{"user:1", false},
		{"user:2", false},
		{"other", true},
	}

	for _, tt := range tests {
		exists, err := testRedisCache.Has(tt.key)
		if err != nil {
			t.Errorf("Has() after EmptyByMatch() failed for %s: %v", tt.key, err)
			return
		}
		if exists != tt.shouldBe {
			t.Errorf("Key %s existence = %v, want %v after EmptyByMatch()", tt.key, exists, tt.shouldBe)
		}
	}
}
