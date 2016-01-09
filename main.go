package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

var idle = flag.Duration("idle", time.Duration(10*time.Minute), "how long disk needs to be idle")

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s [OPTS] DISK [DISK..]\n", os.Args[0])
	flag.PrintDefaults()
}

type DevID uint64

type Device struct {
	Name string

	Ops struct {
		Reads  uint64
		Writes uint64
	}
}

// TODO no portable go access to the defines in linux/kdev_t.h
func mkdev(maj uint64, min uint64) uint64 {
	return maj<<8 | min
}

type Ignore struct{}

func (Ignore) Scan(state fmt.ScanState, verb rune) error {
	_, err := state.Token(true, nil)
	return err
}

func main() {
	prog := filepath.Base(os.Args[0])
	log.SetFlags(0)
	log.SetPrefix(prog + ": ")

	flag.Usage = usage
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(2)
	}

	devs := make(map[DevID]*Device, flag.NArg())
	for _, name := range flag.Args() {
		fi, err := os.Stat(name)
		if err != nil {
			log.Fatalf("no such device: %v", err)
		}
		st, ok := fi.Sys().(*syscall.Stat_t)
		if !ok {
			log.Fatalf("don't know how to handle stat: %#v", fi.Sys())
		}
		majmin := DevID(st.Rdev)
		devs[majmin] = &Device{Name: name}
	}

	diskstats, err := os.Open("/proc/diskstats")
	if err != nil {
		log.Fatalf("cannot open diskstats: %v", err)
	}

	for {
		_, err := diskstats.Seek(0, 0)
		if err != nil {
			log.Fatalf("cannot rewind diskstats: %v", err)
		}
		r := bufio.NewReader(diskstats)

		// https://www.kernel.org/doc/Documentation/iostats.txt
		//   3    0   hda 446216 784926 9550688 4382310 424847 312726 5922052 19310380 0 3376340 23705160
		// maj min devname 1reads . . . 5writes . . . 9inprogress . .
		var ignore Ignore
		var maj, min, reads, writes, inprogress uint64
		for {
			_, err = fmt.Fscanln(r, &maj, &min, ignore,
				&reads, ignore, ignore, ignore,
				&writes, ignore, ignore, ignore,
				&inprogress, ignore, ignore,
			)
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("cannot parse diskstats: %v", err)
			}

			majmin := DevID(mkdev(maj, min))
			dev, ok := devs[majmin]
			if !ok {
				continue
			}

			prev := dev.Ops
			dev.Ops.Reads = reads
			dev.Ops.Writes = writes
			if inprogress == 0 &&
				prev.Reads == dev.Ops.Reads &&
				prev.Writes == dev.Ops.Writes {
				log.Printf("spinning down %s", dev.Name)
				err = spindown(dev.Name)
				if err != nil {
					log.Fatalf("cannot spin down %s: %v", dev.Name, err)
				}
			}
		}

		// we measure just once per idle period, which means the disk
		// may need to be idle almost 2*idle to be noticed as idle.
		// good enough.
		time.Sleep(*idle)
	}
}
