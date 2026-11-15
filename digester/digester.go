package main

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sync/errgroup"
)

type result struct {
	path string
	sum  [md5.Size]byte
	err  error
}

func main() {
	fmt.Printf("Digesting %s\n", os.Args[1])
	r, err := MD5All(os.Args[1])
	if err != nil {
		panic(err)
	}
	fmt.Println(r)
}

func MD5All(root string) (res map[string][md5.Size]byte, err error) {
	paths := make(chan string)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		return walkFiles(egCtx, root, paths)
	})

	c := make(chan result)
	const numDigesters = 20
	for range numDigesters {
		eg.Go(func() error {
			digester(egCtx, paths, c)
			return nil
		})
	}

	go func() {
		_ = eg.Wait()
		close(c)
	}()

	defer func() {
		if err != nil {
			cancel()
			_ = eg.Wait()
		} else {
			err = eg.Wait()
		}
	}()

	m := make(map[string][md5.Size]byte)
	for r := range c {
		if r.err != nil {
			return nil, r.err
		}
		m[r.path] = r.sum
	}

	return m, nil
}

func walkFiles(ctx context.Context, root string, paths chan<- string) error {
	defer close(paths)
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		select {
		case paths <- path:
		case <-ctx.Done():
			return errors.New("walk canceled")
		}
		return nil
	})
}

func digester(ctx context.Context, paths <-chan string, c chan<- result) {
	for path := range paths {
		data, err := os.ReadFile(path)
		select {
		case c <- result{path, md5.Sum(data), err}:
		case <-ctx.Done():
			return
		}
	}
}
