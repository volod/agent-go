package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha256"
	"debug/buildinfo"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime/debug"
	"slices"
	"strings"
	"time"
)

// sumsFile lists the archive checksums in `sha256sum -c` format.
const sumsFile = "SHA256SUMS"

// zipEpoch is the entry time when a binary has no commit time; ZIP cannot store earlier times.
var zipEpoch = time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC)

// Config describes one packaging run.
type Config struct {
	App       string   // command name; binaries are BinDir/App-GOOS-GOARCH[.exe]
	BinDir    string   // directory of the cross-built binaries
	OutDir    string   // directory for the archives and SHA256SUMS
	Version   string   // when set, every binary must report exactly this version
	Files     []string // extra files placed next to the binary in every archive
	Platforms []string // GOOS/GOARCH pairs, one archive each
}

// stamp is what packaging needs from a binary's build information.
type stamp struct {
	version string
	time    time.Time // commit time, used for every archive entry
}

// readStamp reads the build information of the Go binary at path.
func readStamp(path string) (stamp, error) {
	bi, err := buildinfo.ReadFile(path)
	if err != nil {
		return stamp{}, fmt.Errorf("read build info: %w", err)
	}
	return stampOf(bi)
}

func stampOf(bi *debug.BuildInfo) (stamp, error) {
	v := bi.Main.Version
	if v == "" || v == "(devel)" {
		return stamp{}, errors.New("no module version; build it with go build in a Git checkout")
	}
	s := stamp{version: v, time: zipEpoch}
	for _, kv := range bi.Settings {
		if kv.Key != "vcs.time" {
			continue
		}
		if t, err := time.Parse(time.RFC3339, kv.Value); err == nil && t.After(zipEpoch) {
			s.time = t.UTC()
		}
	}
	return s, nil
}

// entry is one member of an archive.
type entry struct {
	name string // slash-separated path inside the archive
	src  string // file on disk; "" for a directory
	mode fs.FileMode
}

// Package writes one archive per platform and SHA256SUMS into cfg.OutDir and
// returns the archive names in platform order. read returns a binary's stamp;
// tests replace it.
func Package(cfg Config, read func(path string) (stamp, error)) ([]string, error) {
	type target struct{ goos, goarch, bin string }
	var (
		targets []target
		st      stamp
	)
	for i, p := range cfg.Platforms {
		goos, goarch, ok := strings.Cut(p, "/")
		if !ok || goos == "" || goarch == "" || strings.Contains(goarch, "/") {
			return nil, fmt.Errorf("platform %q: want GOOS/GOARCH", p)
		}
		bin := filepath.Join(cfg.BinDir, cfg.App+"-"+goos+"-"+goarch+exeSuffix(goos))
		s, err := read(bin)
		if err != nil {
			return nil, fmt.Errorf("%s: %w (run make build-all)", bin, err)
		}
		switch {
		case cfg.Version != "" && s.version != cfg.Version:
			return nil, fmt.Errorf("%s reports version %s, want %s; build a clean checkout of that tag", bin, s.version, cfg.Version)
		case i > 0 && s.version != st.version:
			return nil, fmt.Errorf("%s reports version %s, other binaries %s; rebuild them together", bin, s.version, st.version)
		}
		st = s
		targets = append(targets, target{goos, goarch, bin})
	}

	if err := os.MkdirAll(cfg.OutDir, 0o755); err != nil {
		return nil, err
	}
	var names, sums []string
	for _, t := range targets {
		stem := strings.Join([]string{cfg.App, st.version, t.goos, t.goarch}, "-")
		entries := []entry{
			{name: stem + "/", mode: fs.ModeDir | 0o755},
			{name: stem + "/" + cfg.App + exeSuffix(t.goos), src: t.bin, mode: 0o755},
		}
		for _, f := range cfg.Files {
			entries = append(entries, entry{name: stem + "/" + filepath.Base(f), src: f, mode: 0o644})
		}
		name, write := stem+".tar.gz", writeTarGz
		if t.goos == "windows" {
			name, write = stem+".zip", writeZip
		}
		sum, err := writeArchive(filepath.Join(cfg.OutDir, name), entries, st.time, write)
		if err != nil {
			return nil, err
		}
		names = append(names, name)
		sums = append(sums, fmt.Sprintf("%x  %s\n", sum, name))
	}
	slices.Sort(sums)
	return names, os.WriteFile(filepath.Join(cfg.OutDir, sumsFile), []byte(strings.Join(sums, "")), 0o644)
}

func exeSuffix(goos string) string {
	if goos == "windows" {
		return ".exe"
	}
	return ""
}

type archiveWriter func(w io.Writer, entries []entry, mtime time.Time) error

// writeArchive creates path with write and returns the SHA-256 of its bytes.
func writeArchive(path string, entries []entry, mtime time.Time, write archiveWriter) ([]byte, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	h := sha256.New()
	err = write(io.MultiWriter(f, h), entries, mtime)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return nil, fmt.Errorf("write %s: %w", path, err)
	}
	return h.Sum(nil), nil
}

func writeTarGz(w io.Writer, entries []entry, mtime time.Time) error {
	gz := gzip.NewWriter(w) // the zero header carries no name or time
	tw := tar.NewWriter(gz)
	for _, e := range entries {
		hdr := &tar.Header{Name: e.name, Mode: int64(e.mode.Perm()), ModTime: mtime, Typeflag: tar.TypeDir}
		if e.src == "" {
			if err := tw.WriteHeader(hdr); err != nil {
				return err
			}
			continue
		}
		hdr.Typeflag = tar.TypeReg
		err := copyFile(e.src, func(size int64) (io.Writer, error) {
			hdr.Size = size
			return tw, tw.WriteHeader(hdr)
		})
		if err != nil {
			return err
		}
	}
	if err := tw.Close(); err != nil {
		return err
	}
	return gz.Close()
}

func writeZip(w io.Writer, entries []entry, mtime time.Time) error {
	zw := zip.NewWriter(w)
	for _, e := range entries {
		hdr := &zip.FileHeader{Name: e.name, Modified: mtime}
		hdr.SetMode(e.mode)
		if e.src == "" {
			if _, err := zw.CreateHeader(hdr); err != nil {
				return err
			}
			continue
		}
		hdr.Method = zip.Deflate
		if err := copyFile(e.src, func(int64) (io.Writer, error) { return zw.CreateHeader(hdr) }); err != nil {
			return err
		}
	}
	return zw.Close()
}

// copyFile copies src into the writer that open returns for its size.
func copyFile(src string, open func(size int64) (io.Writer, error)) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return err
	}
	w, err := open(info.Size())
	if err != nil {
		return err
	}
	_, err = io.Copy(w, f)
	return err
}
