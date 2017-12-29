package store

import (
	"testing"
)

func TestNew(t *testing.T) {
	store := New()

	// At the point of writing the test, New() can only fail if it's
	// not initialized propertly. Either the RWMutex is created wrongly
	// or the map is not initialized - in either case, the below should
	// result in a panic.
	store.Store("testKey", "testValue")
}

func TestCStore_Store(t *testing.T) {
	type args struct {
		key   string
		value string
	}
	tests := []struct {
		name string
		args args
		want args
	}{
		{"key_value_ok", args{"testKey", "testValue"}, args{"testKey", "testValue"}},
		{"value_empty", args{"testKey", ""}, args{"testKey", ""}},
		{"key_empty", args{"", "testValue"}, args{"", "testValue"}},
		{"both_empty", args{"", ""}, args{"", ""}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := New()

			cs.Store(tt.args.key, tt.args.value)

			val, ok := cs.store[tt.want.key]
			if !ok {
				t.Errorf("failed to store key: %q", tt.args.key)
			}

			if val != tt.want.value {
				t.Errorf("stored key (%q) is not what expected: %q", val, tt.args.key)
			}
		})
	}
}

func TestCStore_Load(t *testing.T) {
	type args struct {
		key   string
		value string
	}
	tests := []struct {
		name    string
		args    args
		want    args
		present bool
	}{
		{"key_value_ok", args{"testKey", "testValue"}, args{"testKey", "testValue"}, true},
		{"value_empty", args{"testKey", ""}, args{"testKey", ""}, true},
		{"key_empty", args{"", "testValue"}, args{"", "testValue"}, true},
		{"both_empty", args{"", ""}, args{"", ""}, true},
		{"nothing_to_load", args{"", ""}, args{"testKey", ""}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := New()
			cs.store[tt.args.key] = tt.args.value

			val, ok := cs.Load(tt.want.key)
			if ok != tt.present {
				t.Errorf("expected %t for ok, is %t instead.", tt.present, ok)
			}
			if !ok {
				return
			}

			if val != tt.want.value {
				t.Errorf("loaded key %q while expecting %q", val, tt.want.value)
			}

		})
	}
}

func TestCStore_Exists(t *testing.T) {
	type args struct {
		key   string
		value string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"key_value_ok", args{"testKey", "testValue"}, true},
		{"value_empty", args{"testKey", ""}, true},
		{"key_empty", args{"", "testValue"}, true},
		{"both_empty", args{"", ""}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := New()
			cs.store[tt.args.key] = tt.args.value

			if cs.Exists(tt.args.key) != tt.want {
				t.Errorf("expected existence to be %t, is %t.", tt.want, !tt.want)
			}
		})
	}
}

func TestCStore_StoreKey(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		args args
		want args
	}{
		{"key_ok", args{"testKey"}, args{"testKey"}},
		{"key_empty", args{"testKey"}, args{"testKey"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := New()

			cs.StoreKey(tt.args.key)

			_, ok := cs.store[tt.want.key]
			if !ok {
				t.Errorf("failed to store key: %q", tt.args.key)
			}
		})
	}
}

func TestCStore_ToMap(t *testing.T) {
	type args struct {
		key   string
		value string
	}
	tests := []struct {
		name string
		args args
		want args
	}{
		{"key_value_ok", args{"testKey", "testValue"}, args{"testKey", "testValue"}},
		{"value_empty", args{"testKey", ""}, args{"testKey", ""}},
		{"key_empty", args{"", "testValue"}, args{"", "testValue"}},
		{"both_empty", args{"", ""}, args{"", ""}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := New()

			cs.Store(tt.args.key, tt.args.value)

			cc := cs.ToMap()

			val, ok := cc[tt.args.key]
			if !ok {
				t.Errorf("key-value pair %q:%q not present in copied map", tt.args.key, tt.args.value)
			}

			if val != tt.args.value {
				t.Errorf("copied map contains %q instead of %q for key %q", val, tt.args.value, tt.args.key)
			}

		})
	}
}
