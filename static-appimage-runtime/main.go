package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/zipfs"
	"github.com/kardianos/osext"
	"github.com/orivej/e"
)

func main() {
	debug := false
	if len(os.Getenv("APPIMAGE_DEBUG"))>0 {
		debug = true
		fmt.Fprintf(os.Stderr, "[d] debug mode on\n")
	}
	executable, err := osext.Executable()
	e.Exit(err)
	if debug {
		fmt.Fprintf(os.Stderr, "[d] executable: %s\n", executable)
	}
	files, err := zipfs.NewZipTree(executable)
	e.Exit(err)

	mfs := zipfs.NewMemTreeFs(files)
	mfs.Name = fmt.Sprintf("fs(%s)", os.Args[0])

	if debug {
		fmt.Fprintf(os.Stderr, "[d] mfs.Name: %s\n", os.Args[0])
	}

	opts := &nodefs.Options{
		AttrTimeout:  10 * time.Second,
		EntryTimeout: 10 * time.Second,
		Debug:        debug,
	}

	mnt, err := ioutil.TempDir("", ".mount_")
	e.Exit(err)

	if debug {
		fmt.Fprintf(os.Stderr, "[d] mnt: %s\n", mnt)
	}

	server, _, err := nodefs.MountRoot(mnt, mfs.Root(), opts)
	e.Exit(err)

	go server.Serve()

	signals := make(chan os.Signal, 1)
	exitCode := 0
	go func() {
		signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
		<-signals

		if debug {
			fmt.Fprintf(os.Stderr, "[d] Received signal for mnt: %s\n", mnt)
		}
		err = server.Unmount()
		e.Exit(err)
		err = os.Remove(mnt)
		e.Exit(err)

		os.Exit(exitCode)
	}()

	if debug {
		fmt.Fprintf(os.Stderr, "[d] Waiting for mount\n")
	}

	err = server.WaitMount()
	e.Exit(err)

	cmd := exec.Cmd{
		Path:   path.Join(mnt, "AppRun"),
		Args:   os.Args,
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	argv0 := fmt.Sprintf("ARGV0=%s",os.Args[0])
	appdir := fmt.Sprintf("APPDIR=%s",mnt)
	evalsym, errs := filepath.EvalSymlinks(os.Args[0])
	if errs != nil {
		evalsym=os.Args[0]
	}
	abs, erra:= filepath.Abs(evalsym)
	if erra != nil {
		abs=""
	}
	appimage := fmt.Sprintf("APPIMAGE=%s", abs)
	currdir, errg := os.Getwd()
	if errg != nil {
		currdir=""
	}
	workdir := fmt.Sprintf("OWD=%s", currdir)
	cmd.Env = append(os.Environ(), argv0, appdir, appimage, workdir)
	if debug {
		fmt.Fprintf(os.Stderr, "[d] Running executable: %s with %s (wd: %s)", os.Args[0], abs, workdir)
	}
	err = cmd.Run()
	if cmd.ProcessState != nil {
		if waitStatus, ok := cmd.ProcessState.Sys().(syscall.WaitStatus); ok {
			exitCode = waitStatus.ExitStatus()
			err = nil
		}
	}
	e.Print(err)

	signals <- syscall.SIGTERM
	runtime.Goexit()
}
