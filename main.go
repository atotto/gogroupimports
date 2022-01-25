package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

var (
	localPrefix = flag.String("local", "", "put imports beginning with this string after 3rd-party packages; comma-separated list")
)

func main() {
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()
	go func() {
		for _, file := range flag.Args() {
			if err := fixImports(ctx, file); err != nil {
				log.Printf("FAIL: %s %s", file, err)
			}
		}
		cancel()
	}()
	<-ctx.Done()
}

func fixImports(ctx context.Context, file string) error {
	var perm os.FileMode = 0600
	f, err := os.OpenFile(file, os.O_RDONLY, perm)
	if err != nil {
		return err
	}
	defer f.Close()

	src, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	hash := sha256.New()
	hash.Write(src)
	h1 := hash.Sum(nil)

	s := ParseSourceFile(src)

	buf := bytes.NewBuffer(nil)
	buf.Write(s.PackageClause)
	for _, b := range s.ImportDecl {
		buf.Write(b)
	}
	buf.Write(s.TopLevelDecl)

	cmd := exec.CommandContext(ctx, "goimports")
	if *localPrefix != "" {
		cmd.Args = append(cmd.Args, "-local", *localPrefix)
	}
	var b1 bytes.Buffer
	var b2 bytes.Buffer
	cmd.Stdout = &b1
	cmd.Stderr = &b2
	cmd.Stdin = buf

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %s", string(bytes.ReplaceAll(b2.Bytes(), []byte("<standard input>"), []byte(file))), err)
	}

	hash = sha256.New()
	hash.Write(b1.Bytes())
	h2 := hash.Sum(nil)

	if bytes.Equal(h1, h2) {
		// No change
		return nil
	}

	if err := writeFile(file, &b1, perm); err != nil {
		return err
	}
	return nil
}

func writeFile(name string, r io.Reader, perm os.FileMode) error {
	f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, r)
	if err1 := f.Close(); err1 != nil && err == nil {
		err = err1
	}
	return err
}
