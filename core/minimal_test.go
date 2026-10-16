//go:build m1ss_minimal && with_utls

package core

import (
	"context"
	"crypto/ecdh"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	M "github.com/sagernet/sing/common/metadata"
)

// Exercise the real TLS handshake and VLESS authentication, rather than only
// checking that a tag or constructor is present. No public network is used.
func TestMinimalRealityVisionRoundTrip(t *testing.T) {
	target := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "m1ss-reality-vision")
	}))
	target.TLS = &tls.Config{MinVersion: tls.VersionTLS13, CurvePreferences: []tls.CurveID{tls.X25519}}
	target.StartTLS()
	defer target.Close()
	_, targetPortText, _ := net.SplitHostPort(target.Listener.Addr().String())
	targetPort, _ := strconv.Atoi(targetPortText)
	key, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	privateKey := base64.RawURLEncoding.EncodeToString(key.Bytes())
	publicKey := base64.RawURLEncoding.EncodeToString(key.PublicKey().Bytes())
	port := freePort(t)
	const uuid = "11111111-1111-1111-1111-111111111111"
	server := NewCore()
	defer server.Stop()
	serverJSON := fmt.Sprintf(`{"log":{"level":"error"},"inbounds":[{"type":"vless","tag":"vless-in","listen":"127.0.0.1","listen_port":%d,"users":[{"name":"alice","uuid":%q,"flow":"xtls-rprx-vision"}],"tls":{"enabled":true,"server_name":"example.com","reality":{"enabled":true,"private_key":%q,"short_id":["0123456789abcdef"],"handshake":{"server":"127.0.0.1","server_port":%d}}}}],"outbounds":[{"type":"direct","tag":"direct"}]}`, port, uuid, privateKey, targetPort)
	if err = server.Start([]byte(serverJSON)); err != nil {
		t.Fatal(err)
	}
	for _, fingerprint := range []string{"chrome", "firefox"} {
		t.Run(fingerprint, func(t *testing.T) {
			client := NewCore()
			defer client.Stop()
			clientJSON := fmt.Sprintf(`{"log":{"level":"error"},"outbounds":[{"type":"vless","tag":"proxy","server":"127.0.0.1","server_port":%d,"uuid":%q,"flow":"xtls-rprx-vision","tls":{"enabled":true,"server_name":"example.com","utls":{"enabled":true,"fingerprint":%q},"reality":{"enabled":true,"public_key":%q,"short_id":"0123456789abcdef"}}}]}`, port, uuid, fingerprint, publicKey)
			if err := client.Start([]byte(clientJSON)); err != nil {
				t.Fatal(err)
			}
			out, found := client.GetInstance().outbound.Outbound("proxy")
			if !found {
				t.Fatal("VLESS outbound missing")
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			conn, err := out.DialContext(ctx, "tcp", M.ParseSocksaddr(target.Listener.Addr().String()))
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close()
			conn.SetDeadline(time.Now().Add(10 * time.Second))
			// Send TLS traffic through Vision as well as the outer Reality TLS.
			inner := tls.Client(conn, &tls.Config{InsecureSkipVerify: true, ServerName: "example.com"})
			if _, err = io.WriteString(inner, "GET / HTTP/1.1\r\nHost: example.com\r\nConnection: close\r\n\r\n"); err != nil {
				t.Fatal(err)
			}
			body, err := io.ReadAll(inner)
			if err != nil || !strings.Contains(string(body), "m1ss-reality-vision") {
				t.Fatalf("proxy round trip: %v; response %s", err, body)
			}
		})
	}
	// The panel's live user edits must still reach the unchanged VLESS service.
	handled, err := server.UpdateInboundUsers([]byte(`{"type":"vless","tag":"vless-in","users":[{"name":"bob","uuid":"22222222-2222-2222-2222-222222222222","flow":"xtls-rprx-vision"}]}`))
	if err != nil || !handled {
		t.Fatalf("live VLESS user update: handled=%v, err=%v", handled, err)
	}
}

func TestMinimalRejectsUnavailableListeners(t *testing.T) {
	if !slices.Equal(InboundRegistry().OptionTypes(), []string{"vless"}) {
		t.Fatal("shipping build includes an unintended inbound")
	}
	for _, transport := range []string{"ws", "http", "grpc", "quic", "httpupgrade"} {
		t.Run(transport, func(t *testing.T) {
			c := NewCore()
			defer c.Stop()
			config, _ := json.Marshal(map[string]any{"inbounds": []any{map[string]any{
				"type": "vless", "tag": "unsupported", "transport": map[string]string{"type": transport},
			}}})
			if err := c.Start(config); err == nil || !strings.Contains(err.Error(), "native TCP only") {
				t.Fatalf("expected a clear native-TCP error, got %v", err)
			}
		})
	}
}
