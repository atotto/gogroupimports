package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

var (
	write       = flag.Bool("w", false, "write result to (source) file instead of stdout")
	localPrefix = flag.String("local", "", "put imports beginning with this string after 3rd-party packages; comma-separated list")
)

func main() {
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	failed := false
	go func() {
		for _, file := range flag.Args() {
			if err := processFile(ctx, file); err != nil {
				if e, ok := err.(*exec.ExitError); ok {
					os.Exit(e.ExitCode())
				}
				fmt.Fprintf(os.Stderr, "FAIL: %s %s\n", file, err)
				failed = true
			}
		}
		cancel()
	}()
	<-ctx.Done()

	if failed {
		os.Exit(2)
	}
}

func processFile(ctx context.Context, file string) error {
	src, err := formatCode(ctx, nil, file)
	if err != nil {
		return err
	}

	s, err := ParseSourceFile(bytes.NewBuffer(src))
	if err != nil {
		return fmt.Errorf("parse source file: %s", err)
	}

	buf := bytes.NewBuffer(nil)
	buf.Write(s.PackageClause)
	for _, b := range s.ImportDecl {
		if bytes.Equal(b, []byte("\n")) {
			// ignore empty line
			continue
		}
		if bytes.HasPrefix(b, []byte("\t//")) {
			// ignore commented line
			continue
		}
		buf.Write(b)
	}
	buf.Write(s.TopLevelDecl)

	var args []string
	if *localPrefix != "" {
		args = append(args, "-local", *localPrefix)
	}
	dst, err := formatCode(ctx, buf, "", args...)
	if err != nil {
		return err
	}

	if *write {
		if err := os.WriteFile(file, dst, 0600); err != nil {
			return err
		}
	} else {
		if _, err := io.Copy(os.Stdout, bytes.NewReader(dst)); err != nil {
			return err
		}
	}
	return nil
}

func formatCode(ctx context.Context, r io.Reader, file string, arg ...string) (formattedCode []byte, err error) {
	cmd := exec.CommandContext(ctx, "goimports")
	if len(arg) != 0 {
		cmd.Args = append(cmd.Args, arg...)
	}
	if file != "" {
		cmd.Args = append(cmd.Args, file)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	if file != "" {
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stderr = &stderr
	}
	if r != nil {
		cmd.Stdin = r
	}

	if err := cmd.Run(); err != nil {
		if file != "" {
			return nil, err
		}
		r := bufio.NewReader(&stderr)
		for {
			b, err2 := r.ReadBytes('\n')
			if err2 == io.EOF {
				return nil, err
			}
			os.Stderr.Write(bytes.Replace(b, []byte("<standard input>"), []byte(file), 1))
		}
	}
	return stdout.Bytes(), nil
}
