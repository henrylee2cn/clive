package cmd

import (
	"clive/u"
	"os"
	"testing"
)

func TestCmd(t *testing.T) {
	UnixIO()
	h := GetEnv("HOME")
	if h != u.Home {
		t.Fatalf("didn't get home")
	}
	c := AppCtx()
	if c.GetEnv("HOME") != u.Home {
		t.Fatalf("didn't get home")
	}
	t.Logf("ns is %s", NS())
	t.Logf("dot is %s", Dot())
	pwd, _ := os.Getwd()
	if Dot() != pwd {
		t.Fatalf("bad dot")
	}
	c.SetEnv("foo", "bar")
	if GetEnv("foo") != "bar" {
		t.Fatalf("didn't setenv")
	}
	out := Out("out")
	Eprintf("hi!\n")
	Eprintf("There!\n")
	Eprintf("There!\n")
	Warn("ho\n")
	Warn("ho")
	close(out)
}
