package action

import (
	"encoding/json"
	"reflect"
	"sort"
	"testing"
	"time"
)

func Test_actionStore_AddAction(t *testing.T) {

	shutdown := make(chan struct{}, 1)
	t.Cleanup(func() {
		shutdown <- struct{}{}
	})

	store := NewActionStore(shutdown)

	t.Run("Test provided use case", func(t *testing.T) {

		expected := []*output{
			{
				Name:    "jump",
				Average: 150,
			},
			{
				Name:    "run",
				Average: 75,
			},
		}

		param1 := `{"action":"jump", "time":100}`
		param2 := `{"action":"run", "time":75}`
		param3 := `{"action":"jump", "time":200}`
		var err error
		err = store.AddAction(param1)
		if err != nil {
			t.Fatalf("unexpected error adding param1: %v", err)
		}
		err = store.AddAction(param2)
		if err != nil {
			t.Fatalf("unexpected error adding param2: %v", err)
		}
		err = store.AddAction(param3)
		if err != nil {
			t.Fatalf("unexpected error adding param3: %v", err)
		}
		// allow work queue to process, a wait group could handle this better
		time.Sleep(1 * time.Second)

		actualStr := store.GetStats()
		var actual []*output
		err = json.Unmarshal([]byte(actualStr), &actual)
		if err != nil {
			t.Fatalf("failed to marshal actual: %v", err)
		}

		sort.Slice(actual, func(i, j int) bool {
			return actual[i].Name < actual[j].Name
		})
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("actionStore.GetStats() expected = %v, got %v", expected, actual)
		}

	})
}

func BenchmarkAddAction(b *testing.B) {
	shutdown := make(chan struct{}, 1)
	b.Cleanup(func() {
		shutdown <- struct{}{}
	})

	param1 := `{"action":"jump", "time":100}`
	expected := []*output{
		{
			Name:    "jump",
			Average: 100,
		},
	}

	store := NewActionStore(shutdown)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		store.AddAction(param1)
	}
	b.StopTimer()

	// allow work queue to process, a wait group could handle this better
	time.Sleep(1 * time.Second)

	actualStr := store.GetStats()
	var actual []*output
	err := json.Unmarshal([]byte(actualStr), &actual)
	if err != nil {
		b.Fatalf("failed to marshal actual: %v", err)
	}

	if !reflect.DeepEqual(actual, expected) {
		b.Errorf("actionStore.GetStats() expected = %v, got %v", expected, actual)
	}
}
