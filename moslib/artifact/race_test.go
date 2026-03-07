package artifact

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/dpopsuev/mos/moslib/dsl"
)

func TestRace_ConcurrentLexiconAdd(t *testing.T) {
	root := setupScaffold(t)

	const goroutines = 10
	var wg sync.WaitGroup
	errors := make([]error, goroutines)

	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			key := fmt.Sprintf("term_%d", idx)
			desc := fmt.Sprintf("Description for term %d", idx)
			errors[idx] = AddTerm(root, key, desc)
		}(i)
	}
	wg.Wait()

	for i, err := range errors {
		if err != nil {
			t.Errorf("goroutine %d: AddTerm failed: %v", i, err)
		}
	}

	terms, err := ListTerms(root)
	if err != nil {
		t.Fatalf("ListTerms after concurrent add: %v", err)
	}
	if len(terms) != goroutines {
		t.Errorf("expected %d terms after concurrent add, got %d (lost %d writes)",
			goroutines, len(terms), goroutines-len(terms))
	}
}

func TestRace_ConcurrentContractCreate(t *testing.T) {
	root := setupScaffold(t)

	const goroutines = 10
	var wg sync.WaitGroup
	errors := make([]error, goroutines)

	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			id := fmt.Sprintf("CON-RACE-%03d", idx)
			_, errors[idx] = CreateContract(root, id, ContractOpts{
				Title:  fmt.Sprintf("Race contract %d", idx),
				Status: "draft",
				Goal:   fmt.Sprintf("Test concurrent create %d", idx),
			})
		}(i)
	}
	wg.Wait()

	for i, err := range errors {
		if err != nil {
			t.Errorf("goroutine %d: CreateContract failed: %v", i, err)
		}
	}

	contracts, err := ListContracts(root, ListOpts{})
	if err != nil {
		t.Fatalf("ListContracts after concurrent create: %v", err)
	}
	if len(contracts) != goroutines {
		t.Errorf("expected %d contracts after concurrent create, got %d",
			goroutines, len(contracts))
	}
}

func TestRace_ConcurrentLexiconAddRemove(t *testing.T) {
	root := setupScaffold(t)

	for i := 0; i < 5; i++ {
		AddTerm(root, fmt.Sprintf("seed_%d", i), fmt.Sprintf("Seed term %d", i))
	}

	const goroutines = 10
	var wg sync.WaitGroup
	errors := make([]error, goroutines)

	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			if idx%2 == 0 {
				errors[idx] = AddTerm(root, fmt.Sprintf("new_%d", idx), fmt.Sprintf("New term %d", idx))
			} else {
				errors[idx] = RemoveTerm(root, fmt.Sprintf("seed_%d", idx%5))
			}
		}(i)
	}
	wg.Wait()

	writeErrors := 0
	for _, err := range errors {
		if err != nil {
			writeErrors++
		}
	}

	terms, err := ListTerms(root)
	if err != nil {
		t.Fatalf("ListTerms after concurrent add/remove: %v", err)
	}

	vocabPath := filepath.Join(root, ".mos", "lexicon", "default.mos")
	data, _ := os.ReadFile(vocabPath)
	f, parseErr := dsl.Parse(string(data), nil)
	if parseErr != nil {
		t.Fatalf("lexicon file corrupted after concurrent operations: %v", parseErr)
	}
	_ = f

	t.Logf("Concurrent add/remove: %d terms remaining, %d operation errors, %d total ops",
		len(terms), writeErrors, goroutines)
}

func TestRace_ConcurrentContractStatusUpdate(t *testing.T) {
	root := setupScaffold(t)

	CreateContract(root, "CON-FLIP", ContractOpts{
		Title:  "Status Flip Target",
		Status: "draft",
		Goal:   "Test concurrent status updates",
	})

	const goroutines = 5
	var wg sync.WaitGroup
	errors := make([]error, goroutines)
	statuses := []string{"active", "complete", "draft", "active", "complete"}

	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			errors[idx] = UpdateContractStatus(root, "CON-FLIP", statuses[idx])
		}(i)
	}
	wg.Wait()

	moveErrors := 0
	for i, err := range errors {
		if err != nil {
			t.Logf("goroutine %d (status=%s): %v", i, statuses[i], err)
			moveErrors++
		}
	}

	contracts, err := ListContracts(root, ListOpts{})
	if err != nil {
		t.Fatalf("ListContracts after concurrent status update: %v", err)
	}

	found := false
	for _, c := range contracts {
		if c.ID == "CON-FLIP" {
			found = true
			t.Logf("Final status: %s, path: %s", c.Status, c.Path)
		}
	}
	if !found {
		t.Error("CON-FLIP lost after concurrent status updates -- contract disappeared")
	}

	if moveErrors > 0 {
		t.Logf("WARNING: %d/%d concurrent status updates failed -- race condition confirmed", moveErrors, goroutines)
	}
}

// --- helpers ---
