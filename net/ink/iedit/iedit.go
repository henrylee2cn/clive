/*
	Simple editor mostly to test the ink UI framework
	Creates a text control to edit the text and then prints the text
	when it exits.
*/
package main

import (
	"clive/cmd"
	"clive/zx"
	"clive/cmd/opt"
	"clive/net/ink"
)

func edits(t *ink.Txt) {
	t.GetText()
	defer t.PutText()
	// and we'd edit here....
	// XXX: It would be good to have t.Ins and t.Del to do an edit that's just
	// ins or del
	// XXX: We might implement marks in the js and debug the current ones
	// in txt, and then add MarkIns to keep on adding text to the mark without
	// holding the text.
}

func edit(t *ink.Txt) {
	in := t.Events()
	for ev := range in {
		cmd.Warn("got text: %v", ev.Args)
		switch ev.Args[0] {
		case "start":
			continue
			// Example: keep only a single view
			vs := t.Views()
			for _, v := range vs {
				if v != ev.Src {
					t.CloseView(v)
				}
			}
			// Example: do some edits from the program.
			go edits(t)
		case "tag":
			if len(ev.Args) == 1 || ev.Args[1] != "Del" {
				continue
			}
			t.Close()
		case "end":
			// Example: delete the text when all views are gone
			vs := t.Views()
			cmd.Dprintf("views %v\n", t.Views())
			if len(vs) == 0 {
				t.Close()
				return
			}
		}
	}
}

func buttons(bs *ink.ButtonSet, rs *ink.RadioSet, t *ink.Txt) {
	in := bs.Events()
	rs.SendEventsTo(in)
	for ev := range in {
		cmd.Warn("buttons: %v", ev.Args)
		if ev.Args[0] == "Set" {
			s := style
			if bold {
				s += "b"
			}
			if italic {
				s += "i"
			}
			t.SetFont(s);
		}
	}
}

var (
	bold, italic bool
	style = "t"
)

func main() {
	opts := opt.New("[file]")
	c := cmd.AppCtx()
	opts.NewFlag("D", "debug", &c.Debug)
	rdonly := false
	opts.NewFlag("r", "read only", &rdonly)
	cmd.UnixIO()
	args := opts.Parse()
	var t *ink.Txt
	if len(args) == 0 {
		t = ink.NewTxt("1234", "abc")
	} else {
		dat, err := zx.GetAll(cmd.NS(), cmd.AbsPath(args[0]))
		if err != nil {
			cmd.Fatal(err)
		}
		t = ink.NewTaggedTxt(args[0] + " Del", string(dat))
	}
	go edit(t)
	if rdonly {
		t.NoEdits()
	}

	bs := ink.NewButtonSet(&ink.Button{Tag: "One", Name: "one"},
		&ink.Button{Tag: "Two", Name: "two"},
		&ink.Button{Tag: "B", Name: "b", Value: &bold},
		&ink.Button{Tag: "I", Name: "i", Value: &italic})
	rs := ink.NewRadioSet(&style, &ink.Button{Tag: "R", Name: "r"},
		&ink.Button{Tag: "T", Name: "t"})
	go buttons(bs, rs, t)

	pg := ink.NewPg("/", "Example text editing:", bs, rs, t)
	pg.Tag = "Clive's iedit"
	ink.ServeLoginFor("/")

	go ink.Serve(":8181")
	t.Wait()
	for rs := range t.Get(0, -1) {
		cmd.Printf("%s", string(rs))
	}
}
