package main_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/google/goterm/term"
	"github.com/matryer/is"
)

var binpath string

func TestMain(m *testing.M) {
	tmpdir, err := ioutil.TempDir("", "ixargstest")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpdir)

	binpath = filepath.Join(tmpdir, "ixargs")

	buildcmd := exec.Command("go", "build", "-o", binpath)
	if err := buildcmd.Run(); err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}

func Test(t *testing.T) {
	is := is.New(t)

	cases := map[string]struct {
		cmd []string
		in  string
		out string
	}{
		"run command": {
			cmd: []string{binpath, "echo"},
			in:  "a\nb\nc\n\x04",
			out: "a\r\nb\r\nc\r\na b c\r\n",
		},
		"prompt": {
			cmd: []string{"sh", "-c", fmt.Sprintf("echo hoge | %q sh -c '{ read a; echo $a; echo $@; }' dummy", binpath)},
			in:  "acm\n\x04",
			out: "acm\r\nacm\r\nhoge\r\n",
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			main := exec.Command(c.cmd[0], c.cmd[1:]...)
			main.SysProcAttr = &syscall.SysProcAttr{
				Setsid:  true,
				Setctty: true,
			}

			pty, err := term.OpenPTY()
			is.NoErr(err)
			defer pty.Close()

			main.Stdin, main.Stdout, main.Stderr = pty.Slave, pty.Slave, pty.Slave
			is.NoErr(main.Start())

			go func() {
				_, err := io.WriteString(pty.Master, c.in)
				is.NoErr(err)
			}()

			master := newReaderChan(t, pty.Master)
			defer master.Cancel()

			sigcld := bufferedsigcld()

			var out string
		L:
			for {
				select {
				case bs := <-master.C:
					out += string(bs)

				case <-sigcld:
					break L
				}
			}

			is.Equal(out, c.out)
		})
	}
}

type readerChan struct {
	C       <-chan []byte
	cancelC chan<- struct{}
}

func newReaderChan(t *testing.T, r io.Reader) *readerChan {
	is := is.New(t)

	errC := make(chan error)
	cc := make(chan []byte)
	go func() {
		for {
			p := make([]byte, 1024)
			i, err := r.Read(p)
			if i > 0 {
				cc <- p[:i]
			}
			if err != nil {
				errC <- err
				break
			}
		}
	}()

	c := make(chan []byte)
	cancelC := make(chan struct{})
	go func() {
		for {
			select {
			case bs := <-cc:
				c <- bs
			case err := <-errC:
				is.NoErr(err)
			case <-cancelC:
				return
			}
		}
	}()

	return &readerChan{
		C:       c,
		cancelC: cancelC,
	}
}

func (c *readerChan) Cancel() {
	close(c.cancelC)
}

func bufferedsigcld() <-chan struct{} {
	c := make(chan struct{})
	go func() {
		sig := make(chan os.Signal)
		signal.Notify(sig, syscall.SIGCLD)

		<-sig
		<-time.After(100 * time.Millisecond)
		close(c)
	}()

	return c
}
