package cli

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/coxley/pmlproxy/pb"
)

var rootCmd = &cobra.Command{
	Use:   "pml",
	Short: "Manipulate diagrams with the pml.us toolchain built on-top of PlantUML",
	Long: `PlantUML is a markup language to create diagrams from text — and reverse.

Diagrams are represented by their source text or "short code" format. The
syntax is easy enough to read, flexible enough to be powerful. We pair that
engine with additional features on our side to make a really cool suite.

Our goal is to reduce the friction of creating diagrams that persist.
`,
	Example: `
cat diagram.puml | pml shorten
pml render SyfFKj2rKt3CoKnELR1Io4ZDoSa70000
pml expand SyfFKj2rKt3CoKnELR1Io4ZDoSa70000

cat diagram.puml | pml render > diagram.png
cat diagram.puml | pml render | upload_somewhere
download_somewhere | pml extract > original.puml

cat <<EOF | pml render --format SVG > diagram.svg
@startuml
Bob -> Alice
@enduml
EOF
`,
	CompletionOptions: cobra.CompletionOptions{HiddenDefaultCmd: true},
}

func Execute() error {
	return rootCmd.Execute()
}

var (
	addr     string
	insecure bool
	errorC   = color.New(color.FgHiRed)
	warningC = color.New(color.FgYellow)
)

func init() {
	// TODO: Set this via config, env, or flag. (maybe viper)
	rootCmd.PersistentFlags().StringVar(&addr, "addr", "", "which proxy server to talk to? (eg: localhost:6969")
	rootCmd.PersistentFlags().BoolVarP(&insecure, "insecure", "i", false, "disable TLS instead of verifying against system cert pool")
	rootCmd.SetGlobalNormalizationFunc(normalizeFlags)
	flag.Set("logtostderr", "true")
}

func normalizeFlags(f *pflag.FlagSet, name string) pflag.NormalizedName {
	// allow Go-style flags: thisShit -> this-shit
	name = camelToSnake(name, true)

	// any _ should be replaced with '-'
	name = strings.Replace(name, "_", "-", -1)

	return pflag.NormalizedName(name)
}

func getClient() (pb.PlantUMLClient, error) {
	var conn grpc.ClientConnInterface
	var err error
	if !insecure {
		certPool, err := x509.SystemCertPool()
		if err != nil {
			return nil, err
		}
		config := &tls.Config{RootCAs: certPool}
		conn, err = grpc.Dial(addr, grpc.WithTransportCredentials(credentials.NewTLS(config)))
	} else {
		conn, err = grpc.Dial(addr, grpc.WithInsecure())
	}
	if err != nil {
		return nil, err
	}
	return pb.NewPlantUMLClient(conn), nil
}

func fatalf(f string, v ...interface{}) {
	errorC.Fprintf(os.Stderr, f, v...)
	os.Exit(1)
}

func fatalfUsage(cmd *cobra.Command, f string, v ...interface{}) {
	errorC.Fprintf(os.Stderr, f, v...)
	fmt.Fprintln(os.Stderr)
	cmd.Usage()
	os.Exit(1)
}

func errorf(f string, v ...interface{}) {
	errorC.Fprintf(os.Stderr, f, v...)
}

func warningf(f string, v ...interface{}) {
	warningC.Fprintf(os.Stderr, f, v...)
}

// camelToSnake converts ThisShit to this_shit
func camelToSnake(s string, hyphen bool) string {
	s = strings.TrimSpace(s)
	var buf bytes.Buffer
	var prev rune
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			// Make exceptions for ID
			if prev == 'I' && r == 'D' {
				buf.WriteRune(r)
				continue
			}
			if !hyphen {
				buf.WriteRune('_')
			} else {
				buf.WriteRune('-')
			}
		}
		buf.WriteRune(unicode.ToUpper(r))
		prev = r
	}
	return strings.ToLower(buf.String())
}
