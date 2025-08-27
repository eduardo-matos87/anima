// tools/jwt_gen.go
package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"
)

func b64(v []byte) string { return base64.RawURLEncoding.EncodeToString(v) }

func main() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		fmt.Fprintln(os.Stderr, "JWT_SECRET não definido no ambiente")
		os.Exit(1)
	}

	var sub string
	var ttl int
	flag.StringVar(&sub, "sub", "user-123", "subject (user id)")
	flag.IntVar(&ttl, "ttl", 3600, "tempo de vida em segundos (exp)")
	flag.Parse()

	header := map[string]any{"alg": "HS256", "typ": "JWT"}
	payload := map[string]any{
		"sub": sub,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Duration(ttl) * time.Second).Unix(),
	}

	hb, _ := json.Marshal(header)
	pb, _ := json.Marshal(payload)

	seg := b64(hb) + "." + b64(pb)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(seg))
	sig := mac.Sum(nil)

	token := seg + "." + b64(sig)
	fmt.Println(token)
}
