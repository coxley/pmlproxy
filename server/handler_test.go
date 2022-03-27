package server

import (
	"context"
	"strings"
	"testing"

	"github.com/coxley/pmlproxy/pb"
)

func TestPages(t *testing.T) {
	h := DefaultHandler
	ctx := context.Background()
	go h.ManageWorkers(ctx)
	defer ctx.Done()

	table := []struct {
		text        string
		expectedCnt int
	}{
		{"@startuml\nrectangle Foo\n@enduml", 1},
		{"@startuml\nrectangle foo\n@enduml\n@startumlrectangle Bar\n@enduml", 2},
		{"@startuml\nrectangle Foo\n@enduml\n@startumlrectangle Bar\n@enduml\n@startuml\nrectangle Baz\n@enduml", 3},
		// mis-matching startXYZ
		{"@startuml\nrectangle Foo\n@enduml\n@startumlrectangle Bar", 0},
	}
	for _, tc := range table {
		req := pb.RenderRequest{
			Diagram: &pb.Diagram{Source: tc.text},
			Format:  pb.Format_PNG,
		}
		res, err := h.Render(ctx, &req)
		if err != nil && tc.expectedCnt != 0 {
			t.Errorf("found error: %v\ndiagram: %v", err, tc.text)
		}
		if len(res.Data) != tc.expectedCnt {
			t.Errorf("expected %v pages, got %v\ndiagram: %v", tc.expectedCnt, len(res.Data), tc.text)
		}
	}
}

var e1 = `@startuml
!include <tupadr3/common>
!include <tupadr3/devicons/java>
@enduml`

var e2 = `@startuml


hide stereotype
sprite $java [48x48/16] {
000000000000000000000000000000000000000000000000
000000000000000000000000000000000000000000000000
000000000000000000000000000000000000000000000000
000000000000000000000000000000000000000000000000
000000000000000000000000000000000000000000000000
000000000000000000000000000120000000000000000000
000000000000000000000000000080000000000000000000
0000000000000000000000000001B0000000000000000000
0000000000000000000000000005A0000000000000000000
000000000000000000000000000C70000000000000000000
000000000000000000000000006F10000000000000000000
00000000000000000000000005F700000000000000000000
0000000000000000000000006FB000000000000000000000
000000000000000000000009FB0001650000000000000000
0000000000000000000001CFA0018C200000000000000000
000000000000000000002DF9005E90000000000000000000
00000000000000000001DF8008F700000000000000000000
0000000000000000000BFB005FA000000000000000000000
0000000000000000001FF200CF3000000000000000000000
0000000000000000001FD000FF4000000000000000000000
0000000000000000000DE000EFB000000000000000000000
00000000000000000005F2008FF600000000000000000000
00000000000000000000AA001EFE00000000000000000000
000000000000000000000B4004FF40000000000000000000
0000000000000000000000A100BF20000000000000000000
00000000000000000000000400A900000011000000000000
00000000000000016730000001800000046CB10000000000
00000000000005BD30000000010000120000DC0000000000
0000000000001DFEA7765567889BCB7000008F1000000000
000000000000001468899988753200000000BE0000000000
000000000000000001000000000000000004F50000000000
0000000000000000C800000000024200004E500000000000
0000000000000000CFFDCBBBCEFFFC201881000000000000
00000000000000000268AAAAA86400003000000000000000
000000000000000000000000000000000000000000000000
000000000000000009A10000002410000000000000000000
00000000000000000CFFFEDDEFFFF6000000000000000000
00000000000003520049CEEFFCA720000000000000000000
000000000019B50000000000000000000002400000000000
0000000002FF820000000000000000000378000000000000
00000000018DFFDB9766555566789ACCA610160000000000
00000000000001357899AABAA9875310026AA10000000000
0000000000000003542211111233579BCB72000000000000
000000000000000002457778888765310000000000000000
000000000000000000000000000000000000000000000000
000000000000000000000000000000000000000000000000
000000000000000000000000000000000000000000000000
000000000000000000000000000000000000000000000000
}


skinparam folderBackgroundColor<<DEV JAVA>> White
@enduml`

func TestExtract(t *testing.T) {
	h := DefaultHandler
	ctx := context.Background()
	go h.ManageWorkers(ctx)
	defer ctx.Done()

	table := []struct {
		text         string
		expandMacros bool
		expected     string
	}{
		{e1, false, e1},
		{e1, true, e2},
	}
	for _, tc := range table {
		res, err := h.Render(ctx, &pb.RenderRequest{
			Diagram: &pb.Diagram{Source: tc.text},
			Format:  pb.Format_PNG,
		})
		if err != nil {
			t.Errorf("unexpected failure in test setup: %v", err)
		}
		ex, err := h.Extract(ctx, &pb.ExtractRequest{Data: res.Data[0], ExpandMacros: tc.expandMacros})

		got := ex.Diagram.Source
		exp := tc.expected
		if got != exp {
			t.Errorf("diagram doesn't match expected\nExpected: %v\nGot: %v", exp, got)
		}
	}
}

func removeBlankLines(s string) string {
	for strings.Contains(s, "\n\n") {
		s = strings.ReplaceAll(s, "\n\n", "\n")
	}
	return s
}

func TestDivideMetadata(t *testing.T) {
	s := `@startuml
...
...
@enduml
@startuml
...
@enduml`
	e1 := "@startuml\n...\n...\n@enduml\n"
	e2 := "@startuml\n...\n@enduml"
	g1, g2 := divideMetadata(s)
	if e1 != g1 {
		t.Errorf("part 1 is not equal\nwanted: %v\ngot: %v", []byte(e1), []byte(g1))
	}
	if e2 != g2 {
		t.Errorf("part 2 is not equal\nwanted: %v\ngot: %v", e2, g2)
	}
}

// Randomly pulled from Github — ~245ns/op as of 2022-03-21
var benchSource = `
@startuml

actor Utilisateur as user
participant "formSign.js" as form <<Controleur formulaire>>
participant "Sign.java" as controler <<(C,#ADD1B2) Controleur formulaire>>
participant "Secure.java" as secure <<(C,#ADD1B2) authentification>>
participant "Security.java" as security <<(C,#ADD1B2) securite>>

box "Application Web" #LightBlue
	participant form
end box

box "Serveur Play" #LightGreen
	participant controler
	participant secure
	participant security
end box

user -> form : submitSignIn()
form -> form : getParameters()
form -> form : result = checkFields()

alt result

    form -> controler : formSignIn(email,pwd)
    controler -> controler : result = checkFields()

    alt result
    	controler -> secure : Secure.authenticate(email, pwd, true);
    	secure -> security : onAuthenticated()
    	security --> form : renderJSON(0);
    	form --> user : display main page
    else !result
    	controler --> form : renderJSON(1)
    	form --> user : display error
    end

else !result
	form --> user : display error
end

@enduml
`

func benchmarkExtract(b *testing.B, format pb.Format) {
	h := DefaultHandler
	ctx := context.Background()
	go h.ManageWorkers(ctx)

	req := pb.RenderRequest{
		Diagram: &pb.Diagram{Source: benchSource},
		Format:  pb.Format_PNG,
	}
	res, err := h.Render(ctx, &req)
	if err != nil {
		b.Errorf("unexpected error in benchmark setup: %v", err)
	}

	for n := 0; n < b.N; n++ {
		x, err := h.Extract(ctx, &pb.ExtractRequest{Data: res.Data[0]})
		if err != nil {
			b.Errorf("unexpected error in benchmark execution: %v", err)
		}

		pre := normalizeText(benchSource)
		post := x.Diagram.Source

		if !strings.EqualFold(pre, post) {
			b.Errorf("Pre and post extract don't match:\nPre: %v\nPost: %v", pre, post)
		}
	}
}

func BenchmarkPNG(b *testing.B) {
	benchmarkExtract(b, pb.Format_PNG)
}

func BenchmarkSVG(b *testing.B) {
	benchmarkExtract(b, pb.Format_SVG)
}
