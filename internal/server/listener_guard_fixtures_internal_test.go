// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package server

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestListenerGuardFixtures builds every bypass in isolation before requiring rejection.
// Invalid source and dependency failures don't count as guard findings.
func TestListenerGuardFixtures(t *testing.T) {
	t.Parallel()
	fixtures := []struct {
		name, imports, source, finding string
	}{
		{"selector", `"net"`, `func bypass() { _, _ = net.Listen("tcp", "127.0.0.1:0") }`, "bypasses"},
		{"bare", `. "net"`, `func bypass() { _, _ = Listen("tcp", "127.0.0.1:0") }`, "func net.Listen"},
		{"package_alias", `network "net"`, `func bypass() { _, _ = network.Listen("tcp", "127.0.0.1:0") }`, "bypasses"},
		{"alias", `"net"`, `var bind = net.Listen; func bypass() { _, _ = bind("tcp", "127.0.0.1:0") }`, "bypasses"},
		{
			"alias_chain",
			`"net"`,
			`var first = net.Listen; var second = first; func bypass() { _, _ = second("tcp", "127.0.0.1:0") }`,
			"bypasses",
		},
		{
			"wrapper",
			`"net"`,
			`func bind() (net.Listener, error) { return net.Listen("tcp", "127.0.0.1:0") }; func bypass() { _, _ = bind() }`,
			"bypasses",
		},
		{
			"method_value",
			`"net"; "context"`,
			`func bypass() { var cfg net.ListenConfig; bind := cfg.Listen; _, _ = bind(context.Background(), "tcp", "127.0.0.1:0") }`,
			"func (*net.ListenConfig).Listen",
		},
		{
			"method_expression",
			`"net"; "context"`,
			`func bypass() { bind := (*net.ListenConfig).Listen; _, _ = bind(new(net.ListenConfig), context.Background(), "tcp", "127.0.0.1:0") }`,
			"bypasses",
		},
		{
			"renamed_receiver",
			`"net"; "context"`,
			`func bypass() { var other net.ListenConfig; _, _ = other.Listen(context.Background(), "tcp", "127.0.0.1:0") }`,
			"bypasses",
		},
		{
			"promoted_method",
			`"net"; "context"`,
			`type config struct { net.ListenConfig }; func bypass() { var cfg config; _, _ = cfg.Listen(context.Background(), "tcp", "127.0.0.1:0") }`,
			"bypasses",
		},
		{"tcp", `"net"`, `func bypass() { _, _ = net.ListenTCP("tcp", nil) }`, "bypasses"},
		{"unix", `"net"`, `func bypass() { _, _ = net.ListenUnix("unix", nil) }`, "bypasses"},
		{"packet", `"net"`, `func bypass() { _, _ = net.ListenPacket("udp", "127.0.0.1:0") }`, "bypasses"},
		{"udp", `"net"`, `func bypass() { _, _ = net.ListenUDP("udp", nil) }`, "bypasses"},
		{"unix_packet", `"net"`, `func bypass() { _, _ = net.ListenUnixgram("unixgram", nil) }`, "bypasses"},
		{"ip", `"net"`, `func bypass() { _, _ = net.ListenIP("ip:icmp", nil) }`, "bypasses"},
		{"file_listener", `"net"`, `func bypass() { _, _ = net.FileListener(nil) }`, "bypasses"},
		{"tls", `"crypto/tls"`, `func bypass() { _, _ = tls.Listen("tcp", "127.0.0.1:0", nil) }`, "bypasses"},
		{"tls_wrapper", `"crypto/tls"`, `func bypass() { _ = tls.NewListener(nil, nil) }`, "bypasses"},
		{"http", `"net/http"`, `func bypass() { _ = http.ListenAndServe("127.0.0.1:0", nil) }`, "bypasses"},
		{"https", `"net/http"`, `func bypass() { _ = http.ListenAndServeTLS("127.0.0.1:0", "", "", nil) }`, "bypasses"},
		{
			"http_method",
			`"net/http"`,
			`func bypass() { var srv http.Server; bind := srv.ListenAndServe; _ = bind() }`,
			"bypasses",
		},
		{"same_name", `"net"`, `func listenTCP() { _, _ = net.Listen("tcp", "127.0.0.1:0") }`, "bypasses"},
		{
			"unrelated_method",
			`"net"`,
			`type config struct{}; func (config) Listen() {}; func bypass() { config{}.Listen(); var _ net.Listener }`,
			"",
		},
	}
	for _, fixture := range fixtures {
		t.Run(fixture.name, func(t *testing.T) {
			t.Parallel()
			dir := buildListenerFixture(t, fixture.imports, fixture.source)
			findings := guardListenerDirectory(t, dir)
			if fixture.finding == "" {
				if len(findings) != 0 {
					t.Fatalf("valid source rejected: %v", findings)
				}
			} else if !strings.Contains(strings.Join(findings, "\n"), fixture.finding) {
				t.Fatalf("guard must reject %s with %q, got %v", fixture.name, fixture.finding, findings)
			}
		})
	}
}

// buildListenerFixture creates a temporary module and requires valid Go source.
func buildListenerFixture(t *testing.T, imports, source string) string {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		"go.mod":    "module listener/fixture\n\ngo 1.27\n",
		"server.go": "package fixture\nimport (" + imports + ")\n" + source + "\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	cmd := exec.CommandContext(t.Context(), "go", "build", "./...")
	cmd.Dir = dir
	var output bytes.Buffer
	cmd.Stdout, cmd.Stderr = &output, &output
	if err := cmd.Run(); err != nil {
		t.Fatalf("fixture must build: %v: %s", err, &output)
	}
	return dir
}
