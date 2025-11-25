# go-erriter

[![Go Reference](https://pkg.go.dev/badge/github.com/burdiyan/go-erriter.svg)](https://pkg.go.dev/github.com/burdiyan/go-erriter)
[![Go Report Card](https://goreportcard.com/badge/github.com/burdiyan/go-erriter)](https://goreportcard.com/report/github.com/burdiyan/go-erriter)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A Go package that provides iterators capable of returning errors while properly handling cleanup errors within the iterator itself and after iteration finishes.

## Overview

Package `erriter` extends Go 1.23's range-function iterators with error handling capabilities. It allows you to create iterators that can fail during iteration and ensures that cleanup errors are properly propagated even if you need to return early from the iteration loop.

This is particularly useful for I/O operations like database queries or network streams, where errors can occur both during iteration (e.g., reading rows from a database) and during cleanup (e.g., closing connections or rolling back transactions). By using `Seq`, you can safely handle both types of errors without losing information.

Ideally, this functionality should exist in the Go standard library, but since it doesn't, `erriter` provides a way to establish consistent patterns and improve composability across projects that need error-aware iterators.

## Installation

```bash
go get github.com/burdiyan/go-erriter
```

## Usage

### Consuming a Seq

`Seq[T]` is similar to `iter.Seq` but allows returning an error. To iterate over a `Seq`, use the `All()` method which returns three values: an iterator, a discard function, and a check function.

```go
func someFunc(it erriter.Seq[string]) (err error) {
    items, discard, check := it.All()
    // Defer discard to ensure cleanup errors are handled,
    // even if we return early from the loop.
    defer discard(&err)

    for item := range items {
        if err := handleItem(item); err != nil {
            // Safe to return here. The deferred discard will
            // propagate this error along with any cleanup errors.
            return err
        }
    }

    // Call check() at the end of iteration to handle any errors.
    if err := check(); err != nil {
        return err
    }

    return nil
}
```

### Producing a Seq

To create a `Seq`, use the `Seq()` function with a callback that yields values and returns an error. The callback is responsible for cleanup and should return any error that occurred during iteration.

```go
// Example: creating a Seq that reads from a fake database connection
func queryDatabase(db *Database, query string) erriter.Seq[string] {
    return erriter.Seq(func(yield func(string) bool) error {
        // Open a connection (or cursor)
        cursor, err := db.Query(query)
        if err != nil {
            return err
        }
        // Ensure cleanup happens, even if yield returns false
        defer func() {
            if closeErr := cursor.Close(); closeErr != nil {
                err = errors.Join(err, closeErr)
            }
        }()

        // Iterate and yield results
        for cursor.Next() {
            value := cursor.Value()
            // If yield returns false, the caller stopped iterating
            if !yield(value) {
                break
            }
        }

        // Return any errors from iteration
        return cursor.Err()
    })
}
```

## License

MIT
