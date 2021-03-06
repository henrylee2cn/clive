package txt

import (
	"clive/dbg"
	"fmt"
	"testing"
)

struct test {
	op    Tedit
	p0    int
	n     int
	txt   string
	out   string
	fails bool
	m     string
}

var (
	debug  bool
	printf = dbg.FlagPrintf(&debug)

	tests = []test{
		test{
			op: Eins, p0: 0, n: 0, txt: "12345",
			out: `12345`,
		},
		test{
			op: Eins, p0: 0, n: 0, txt: "xyz",
			out: `xyz12345`,
		},
		test{
			op: Eins, p0: 2, n: 0, txt: "qwerty",
			out: `xyqwertyz12345`,
		},
		test{
			op: Eins, p0: 20, n: 0, txt: "xx",
			out:   `xyqwertyz12345`,
			fails: true,
		},
		test{
			op: Edel, p0: 20, n: 0, txt: "",
			out: `xyqwertyz12345`,
		},
		test{
			op: Edel, p0: 0, n: 2, txt: "xy",
			out: `qwertyz12345`,
		},
		test{
			op: Edel, p0: 6, n: 3, txt: "z12",
			out: `qwerty345`,
		},
	}
)

struct pair {
	a, b int
}

func TestLineAt(t *testing.T) {
	debug = testing.Verbose()
	tx := NewEditing([]rune("abc\ndef\n\nghi\n"))
	nbs := [...]int{1, 1, 1, 1, 2, 2, 2, 2, 3, 4, 4, 4, 4, 5}
	printf("=>\n%s\n", tx.Sprint())
	for o := 0; o <= tx.Len(); o++ {
		n := tx.LineAt(o)
		printf("\t%d -> %d\n", o, n)
		if n != nbs[o] {
			t.Fatalf("bad line number")
		}
	}
	lns := [...]pair{{1, 1}, {0, 0}, {0, 3}, {0, 4}, {1, 5}, {4, 4}}
	bs := [...]pair{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 2}, {2, 2}}
	for i, p := range lns {
		a, b := tx.LinesAt(p.a, p.b)
		printf("\t%v-> %d %d\n", p, a, b)
		if a != bs[i].a || b != bs[i].b {
			t.Fatalf("bad line number")
		}
	}
}

func TestLineOff(t *testing.T) {
	debug = testing.Verbose()
	tx := NewEditing([]rune("abc\ndef\n\nghi\n"))
	offs := [...]int{0, 0, 4, 8, 9, 13, 13}
	lns := [...]int{1, 1, 2, 3, 4, 5, 5}
	printf("=>\n%s\n", tx.Sprint())
	for n := 0; n <= 6; n++ {
		o := tx.LineOff(n)
		ln := tx.LineAt(o)
		printf("\t%d -> %d\t%d\n", n, o, ln)
		if o != offs[n] {
			t.Fatalf("bad off")
		}
		if ln != lns[n] {
			t.Fatalf("bad ln")
		}
	}
	ps := [...]pair{{1, 1}, {1, 2}, {2, 2}, {2, 3}, {3, 3}, {4, 4}, {4, 5}, {6, 6}}
	os := [...]pair{{0, 0}, {0, 8}, {4, 8}, {4, 9}, {8, 9}, {9, 13}, {9, 13}, {13, 13}}
	for i, p := range ps {
		a, b := tx.LinesOffs(p.a, p.b)
		printf("\t%v-> %d %d\n", p, a, b)
		if a != os[i].a || b != os[i].b {
			t.Fatalf("bad line number")
		}
	}
}

func TestInsDel(t *testing.T) {
	debug = testing.Verbose()

	tx := NewEditing(nil)
	for _, tst := range tests {
		var err error
		if tst.op == Eins {
			printf("ins %s %d\n", tst.txt, tst.p0)
			err = tx.Ins([]rune(tst.txt), tst.p0)
			printf("-> %v\n", err)
			if tst.fails && err == nil {
				t.Fatalf("didn't fail")
			}
			if !tst.fails && err != nil {
				t.Fatalf("did fail: %s", err)
			}
		} else {
			printf("del %d %d\n", tst.p0, tst.n)
			rs := tx.Del(tst.p0, tst.n)
			printf("-> %s\n", string(rs))
			if tst.txt != string(rs) {
				t.Fatalf("bad delete")
			}
		}
		out := tx.String()
		printf("\t\t\tout: `%s`,\n", out)
		printf("=>\n%s\n", tx.Sprint())
		if tst.out != "" && out != tst.out {
			t.Fatalf("bad text")
		}
	}
}

var tests2 = []test{
	test{
		op: Eins, p0: 0, n: 0, txt: "ABC",
		out: `[[m0 0] [m1 7] [m2 11] [m3 17]]`,
	},
	test{
		op: Eins, p0: 5, n: 0, txt: "DE",
		out: `[[m0 0] [m1 9] [m2 13] [m3 19]]`,
	},
	test{
		op: Eins, p0: 13, n: 0, txt: "FG",
		out: `[[m0 0] [m1 9] [m2 13] [m3 21]]`,
	},
	test{
		op: Eins, p0: 21, n: 0, txt: "HI",
		out: `[[m0 0] [m1 9] [m2 13] [m3 21]]`,
	},
	test{
		op: Edel, p0: 22, n: 2,
		out: `[[m0 0] [m1 9] [m2 13] [m3 21]]`,
	},
	test{
		op: Edel, p0: 16, n: 2,
		out: `[[m0 0] [m1 9] [m2 13] [m3 19]]`,
	},
	test{
		op: Edel, p0: 18, n: 2,
		out: `[[m0 0] [m1 9] [m2 13] [m3 18]]`,
	},
	test{
		op: Edel, p0: 7, n: 4,
		out: `[[m0 0] [m1 7] [m2 9] [m3 14]]`,
	},
	test{
		op: Eins, m: "m0", txt: "12",
		out: `[[m0 2] [m1 9] [m2 11] [m3 16]]`,
	},
	test{
		op: Eins, m: "m0", txt: "34",
		out: `[[m0 4] [m1 11] [m2 13] [m3 18]]`,
	},
	test{
		op: Eins, m: "m2", txt: "12",
		out: `[[m0 4] [m1 11] [m2 15] [m3 20]]`,
	},
	test{
		op: Eins, m: "m3", txt: "ab",
		out: `[[m0 4] [m1 11] [m2 15] [m3 22]]`,
	},
	test{
		op: Edel, m: "m3", n: 2,
		out: `[[m0 4] [m1 11] [m2 15] [m3 20]]`,
	},
	test{
		op: Edel, m: "m0", n: 6,
		out: `[[m0 0] [m1 7] [m2 11] [m3 16]]`,
	},
	test{
		op: Edel, m: "m2", n: 6,
		out: `[[m0 0] [m1 5] [m2 5] [m3 10]]`,
	},
}

func TestMark(t *testing.T) {
	debug = testing.Verbose()

	tx := NewEditing(nil)
	for _, tst := range tests {
		if tst.op != Eins {
			break
		}
		tx.Ins([]rune(tst.txt), tst.p0)
	}
	out := tx.String()
	printf("txt:`%s`,\n", out)
	printf("=>\n%s\n", tx.Sprint())
	m0 := tx.SetMark("m0", 0)
	m1 := tx.SetMark("m1", 4)
	m2 := tx.SetMark("m2", 8)
	m3 := tx.SetMark("m3", tx.Len())
	marks := []*Mark{m0, m1, m2, m3}
	printf("%s\n", marks)
	printf("=>\n%s\n", tx.SprintMarks())

	for _, tst := range tests2 {
		if tst.m != "" && tst.op == Eins {
			printf("markins %s %s\n", tst.m, tst.txt)
			err := tx.MarkIns(tst.m, []rune(tst.txt))
			printf("-> %v\n", err)
		} else if tst.m != "" && tst.op == Edel {
			printf("markdel %s %d\n", tst.m, tst.n)
			rs := tx.MarkDel(tst.m, tst.n)
			printf("-> %s\n", string(rs))
		} else if tst.op == Eins {
			printf("ins %s %d\n", tst.txt, tst.p0)
			err := tx.Ins([]rune(tst.txt), tst.p0)
			printf("-> %v\n", err)
		} else {
			printf("del %d %d\n", tst.p0, tst.n)
			rs := tx.Del(tst.p0, tst.n)
			printf("-> %s\n", string(rs))
		}
		ms := fmt.Sprintf("%s", marks)
		printf("\t\t\tout: `%s`,\n", ms)
		printf("=>\n%s\n", tx.SprintMarks())
		if tst.out != "" && tst.out != ms {
			t.Fatalf("bad out")
		}
	}
}
