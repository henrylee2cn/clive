package cmd

import (
	"clive/zx"
	fpath "path"
	"strings"
	"bytes"
	"unicode/utf8"
	"errors"
)

func Stat(path string) (zx.Dir, error) {
	path = AbsPath(path)
	rc := NS().Stat(path)
	d := <-rc
	return d, cerror(rc)
}

func Get(path string, off, count int64) <-chan []byte {
	path = AbsPath(path)
	return NS().Get(path, off, count)
}

func Put(path string, ud zx.Dir, off int64, dc <-chan []byte) <-chan zx.Dir {
	path = AbsPath(path)
	return NS().Put(path, ud, off, dc)
}

func Wstat(path string, ud zx.Dir) (zx.Dir, error) {
	path = AbsPath(path)
	rc := NS().Wstat(path, ud)
	d := <-rc
	return d, cerror(rc)
}

func Remove(path string) error {
	path = AbsPath(path)
	return <-NS().Remove(path)
}

func RemoveAll(path string) error {
	path = AbsPath(path)
	return <-NS().RemoveAll(path)
}

// Issue a find for these names ("filename,predicate")
// If no predicate is given, then "depth<1" is used.
// eg. /a -> just /a; /a, -> subtree at /a
// Errors are reported by sending an error.
// The chan error status is not nil if there's an error.
// The Upath attribute in the dir entries returned mimics the paths given
// in the names.
// The Rpath attribute in the dir entries provide a path relative to the one
// specified by the user.
func Dirs(names ...string) chan interface{} {
	ns := NS()
	rc := make(chan interface{})
	go func() {
		var err error
		for _, name := range names {
			if len(name) > 0 && name[0] == '#' {
				d := zx.Dir{"path": name, "name": name,
					"Upath": name, "type": "c"}
				rc <- d
				continue
			}
			toks := strings.SplitN(name, ",", 2)
			if toks[0] == "" {
				toks[0] = "."
			}
			if len(toks) == 1 {
				toks = append(toks, "0")
			}
			toks[0] = fpath.Clean(toks[0])
			name = AbsPath(toks[0])
			Dprintf("getdirs: find %s %s\n", name, toks[1])
			dc := ns.Find(name, toks[1], "/", "/", 0)
			for d := range dc {
				if d == nil {
					break
				}
				d["Upath"] = d["path"]
				if toks[0] != name && zx.HasPrefix(d["path"], name) {
					u := fpath.Join(toks[0], zx.Suffix(d["path"], name))
					d["Upath"] = u
				}
				d["Rpath"] = zx.Suffix(d["path"], name)
				if d["err"] != "" {
					if d["err"] != "pruned" {
						err = errors.New(d["err"])
						rc <- err
					}
					continue
				}
				if ok := rc <- d; !ok {
					close(dc, cerror(rc))
					return
				}
			}
			if derr := cerror(dc); derr != nil {
				err = derr
				if ok := rc <- err; !ok {
					return
				}
			} else {
				close(dc) // in case a null was sent but no error
			}
		}
		close(rc, err)
	}()
	return rc
}

// Like Dirs(), but sends also file contents
func Files(names ...string) chan interface{} {
	ns := NS()
	rc := make(chan interface{})
	go func() {
		var err error
		for _, name := range names {
			if len(name) > 0 && name[0] == '#' {
				d := zx.Dir{"path": name, "name": name,
					"Upath": name, "type": "c"}
				rc <- d
				continue
			}
			toks := strings.SplitN(name, ",", 2)
			if toks[0] == "" {
				toks[0] = "."
			}
			if len(toks) == 1 {
				toks = append(toks, "0")
			}
			toks[0] = fpath.Clean(toks[0])
			name = AbsPath(toks[0])
			Dprintf("getfiles: findget %s %s\n", name, toks[1])
			dc := ns.FindGet(name, toks[1], "/", "/", 0)
			for m := range dc {
				d, ok := m.(zx.Dir)
				if !ok {
					if ok := rc <- m; !ok {
						close(dc, cerror(rc))
						return
					}
					continue
				}
				d["Upath"] = d["path"]
				if toks[0] != name && zx.HasPrefix(d["path"], name) {
					u := fpath.Join(toks[0], zx.Suffix(d["path"], name))
					d["Upath"] = u
				}
				d["Rpath"] = zx.Suffix(d["path"], name)
				if d["err"] != "" {
					if d["err"] != "pruned" {
						err = errors.New(d["err"])
						rc <- err
					}
					continue
				}
				if ok := rc <- d; !ok {
					close(dc, cerror(rc))
					return
				}
			}
			if derr := cerror(dc); derr != nil {
				err = derr
				if ok := rc <- err; !ok {
					return
				}
			} else {
				close(dc) // in case a null was sent but no error
			}
		}
		close(rc, err)
	}()
	return rc
}

// Process a stream of input []byte data and send one line at a time
func ByteLines(c <-chan []byte) <-chan []byte {
	sep := '\n'
	rc := make(chan []byte)
	go func() {
		var buf bytes.Buffer
		saved := []byte{}
		for d := range c {
			if len(saved) > 0 {
				nb := []byte{}
				nb = append(nb, saved...)
				nb = append(nb, d...)
				d = nb
				saved = nil
			}
			for len(d) > 0 && utf8.FullRune(d) {
				r, n := utf8.DecodeRune(d)
				d = d[n:]
				buf.WriteRune(r)
				if r == sep {
					nb := make([]byte, buf.Len())
					copy(nb, buf.Bytes())
					if ok := rc <- nb; !ok {
						close(c, cerror(rc))
						return
					}
					buf.Reset()
				}
			}
			saved = d
		}
		if len(saved) > 0 {
			buf.Write(saved)
		}
		if buf.Len() > 0 {
			rc <- buf.Bytes()
		}
		close(rc, cerror(c))
	}()
	return rc
}

// Process a stream of input file data and send one line at a time.
func Lines(c <-chan interface{}) <-chan interface{} {
	sep := '\n'
	rc := make(chan interface{})
	go func() {
		var buf bytes.Buffer
		saved := []byte{}
		for m := range c {
			d, ok := m.([]byte)
			if !ok {
				if len(saved) > 0 {
					rc <- saved
					saved = nil
				}
				if ok := rc <- m; !ok {
					close(c, cerror(rc))
					return
				}
				continue
			}
			if len(saved) > 0 {
				nb := []byte{}
				nb = append(nb, saved...)
				nb = append(nb, d...)
				d = nb
				saved = nil
			}
			for len(d) > 0 && utf8.FullRune(d) {
				r, n := utf8.DecodeRune(d)
				d = d[n:]
				buf.WriteRune(r)
				if r == sep {
					nb := make([]byte, buf.Len())
					copy(nb, buf.Bytes())
					if ok := rc <- nb; !ok {
						close(c, cerror(rc))
						return
					}
					buf.Reset()
				}
			}
			saved = d
		}
		if len(saved) > 0 {
			buf.Write(saved)
		}
		if buf.Len() > 0 {
			rc <- buf.Bytes()
		}
		close(rc, cerror(c))
	}()
	return rc
}

// pipe an input chan and make sure the output
// issues one message per file in the input containing all data.
// non []byte messages forwarded as-is.
func FullFiles(c <-chan interface{}) <-chan interface{} {
	rc := make(chan interface{})
	go func() {
		var b *bytes.Buffer
		for m := range c {
			switch d := m.(type) {
			case []byte:
				if b == nil {
					b = &bytes.Buffer{}
				}
				b.Write(d)
			default:
				if b != nil {
					if ok := rc <- b.Bytes(); !ok {
						close(c, cerror(rc))
						break
					}
					b = nil
				}
				if ok := rc <- m; !ok {
					close(c, cerror(rc))
					break
				}
			}
		}
		if b != nil {
			if ok := rc <- b.Bytes(); !ok {
				close(c, cerror(rc))
			}
		}
		close(rc, cerror(c))
	}()
	return rc
}
