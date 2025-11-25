package erriter

import (
	"fmt"
	"slices"
	"testing"
)

func TestSeq(t *testing.T) {
	out, err := doTest(1)
	if err != nil {
		t.Errorf("doTest(1) returned unexpected error: %v", err)
	}
	if !slices.Equal(out, []int{1}) {
		t.Errorf("doTest(1) returned %v, expected [1]", out)
	}

	out, err = doTest(2)
	if err == nil || err.Error() != "error midway" {
		t.Errorf("doTest(2) returned error %v, expected 'error midway'", err)
	}
	if !slices.Equal(out, []int{1, 2}) {
		t.Errorf("doTest(2) returned %v, expected [1, 2]", out)
	}

	out, err = doTest(3)
	if err != nil {
		t.Errorf("doTest(3) returned unexpected error: %v", err)
	}
	if !slices.Equal(out, []int{1, 2, 3}) {
		t.Errorf("doTest(3) returned %v, expected [1, 2, 3]", out)
	}

	out, err = doTest(4)
	if err == nil || err.Error() != "cleanup error" {
		t.Errorf("doTest(4) returned error %v, expected 'cleanup error'", err)
	}
	if out != nil {
		t.Errorf("doTest(4) returned %v, expected nil", out)
	}
}

func doTest(n int) (out []int, err error) {
	it := Seq(func(yield func(int) bool) (err error) {
		if !yield(1) {
			return err
		}
		if !yield(2) {
			return err
		}
		err = fmt.Errorf("error midway")
		if !yield(3) {
			return err
		}
		err = nil
		if !yield(4) {
			return err
		}

		return fmt.Errorf("cleanup error")
	})

	items, discard, check := it.All()
	defer discard(&err)

	var i int
	for x := range items {
		if i == n {
			return out, err
		}
		out = append(out, x)
		i++
	}

	if err := check(); err != nil {
		return nil, err
	}

	return out, nil
}
