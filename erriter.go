// Package erriter provides iterators that can return errors.
// It allows iterating over sequences that may fail while properly handling
// cleanup errors within the iterator itself, and after the iteration finishes.
package erriter

import (
	"errors"
	"iter"
)

// Seq is like [iter.Seq] but is allowed to return an error.
// To actually range over Seq users must call [Seq.All],
// which returns a standard iterator over the values,
// plus the discard and check functions which let the user
// handle any possible cleanup errors, even if they need to return
// from the iteration loop early.
//
// ## Consumer Pattern
//
// The usage pattern for consuming a Seq is as follows:
//
//	func someFunc(it erriter.Seq[string]) (err error) {
//	    items, discard, check := it.All()
//	    // Deferring discard here makes sure that we handle the cleanup errors
//	    // even if we have to return from within the loop.
//	    defer discard(&err)
//
//	    for item := range items {
//	        if err := handleItem(item); err != nil {
//	            // We can safely return here.
//	            // This error will be bubbled up by discard,
//	            // and any possible cleanup errors will be joined with it.
//	            return err
//	        }
//	    }
//
//	    // If we finished the iteration, we are warned to call check() by the compiler.
//	    // After check() is called, the deferred discard will do nothing, because the cleanup had already happened.
//	    // Calling check more than once is a bug and will panic.
//	    if err := check(); err != nil {
//	        return err
//	    }
//
//	    // Maybe more code here...
//
//	    return nil
//	}
//
// ## Producer Pattern
//
// To create a Seq, use the [Make] function with a callback that yields values and returns an error:
//
//	func queryRows(db *Database) erriter.Seq[string] {
//	    return erriter.Make(func(yield func(string) bool) error {
//	        // Open a database connection/cursor
//	        cursor, err := db.OpenCursor()
//	        if err != nil {
//	            return err
//	        }
//	        // Ensure cleanup happens even if yield returns false (caller stopped iterating)
//	        defer func() {
//	            if closeErr := cursor.Close(); closeErr != nil {
//	                err = errors.Join(err, closeErr)
//	            }
//	        }()
//
//	        // Iterate and yield results
//	        for cursor.Next() {
//	            row := cursor.Row()
//	            // If yield returns false, the caller stopped iterating
//	            if !yield(row) {
//	                break
//	            }
//	        }
//
//	        // Return any errors from iteration or from the query itself
//	        return cursor.Err()
//	    })
//	}
type Seq[T any] func(yield func(T) bool) error

// Make is a convenience function for creating a [Seq].
// As Go can't infer generic type parameters on function types,
// but can infer them on function values, using this function lets you wrap the literal
// iteration function without having to specify the type parameter explicitly.
func Make[T any](fn func(yield func(T) bool) error) Seq[T] {
	return Seq[T](fn)
}

// All returns an iterator, a discard function, and a check function.
// The iterator yields the items from the sequence.
// The discard function should be deferred and will handle any cleanup errors.
// The check function should be called at the end of the iteration.
// Calling check more than once is a bug and will panic.
func (iter Seq[T]) All() (it iter.Seq[T], discard func(*error), check func() error) {
	var err error
	var checkCalled bool

	discard = func(errp *error) {
		if checkCalled {
			return
		}
		*errp = errors.Join(*errp, err)
	}

	check = func() error {
		if checkCalled {
			panic("BUG: check called twice")
		}
		checkCalled = true
		return err
	}

	it = func(yield func(T) bool) {
		err = iter(yield)
	}

	return it, discard, check
}
