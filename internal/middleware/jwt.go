package middleware

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"net/http"
	"strings"

	"github.com/easterthebunny/spew-order/internal/auth"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"
)

var (
	// ErrUnauthorized ...
	ErrUnauthorized = errors.New("token is unauthorized")
	// ErrExpired ...
	ErrExpired = errors.New("token is expired")
	// ErrNBFInvalid ...
	ErrNBFInvalid = errors.New("token nbf validation failed")
	// ErrIATInvalid ...
	ErrIATInvalid = errors.New("token iat validation failed")
	// ErrNoTokenFound ...
	ErrNoTokenFound = errors.New("no token found")
	// ErrAlgoInvalid ...
	ErrAlgoInvalid = errors.New("algorithm mismatch")
	// ErrParamsMissing ...
	ErrParamsMissing = errors.New("token parameters missing")
	// ErrKeyMustBePEMEncoded ...
	ErrKeyMustBePEMEncoded = errors.New("certificate not PEM encoded")
	// ErrNotRSAPublicKey ...
	ErrNotRSAPublicKey = errors.New("rsa key not a public key")
)

// NewJWTAuth ...
func NewJWTAuth(url string) (*JWT, error) {
	return &JWT{domain: url}, nil
}

// JWT ...
type JWT struct {
	alg       jwa.SignatureAlgorithm
	verifyKey interface{}
	verifier  jwt.ParseOption
	domain    string
	token     jwt.Token
}

// Jwks ...
type Jwks struct {
	Keys []JSONWebKeys `json:"keys"`
}

// JSONWebKeys ...
type JSONWebKeys struct {
	Alg string   `json:"alg"`
	Kty string   `json:"kty"`
	Kid string   `json:"kid"`
	Use string   `json:"use"`
	N   string   `json:"n"`
	E   string   `json:"e"`
	X5c []string `json:"x5c"`
}

// Verifier ...
func (j *JWT) Verifier() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			token, err := tokenFromRequest(j, r, tokenFromHeader)
			if err != nil || token == nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			j.token = token

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Subject ...
func (j *JWT) Subject() string {
	return j.token.Subject()
}

// UpdateAuthz ...
func (j *JWT) UpdateAuthz(a *auth.Authorization) {
	a.ID = j.token.Subject()
	email, ok := j.token.Get("name")
	if ok {
		a.Email = email.(string)
	}

	pic, ok := j.token.Get("picture")
	if ok {
		a.Avatar = pic.(string)
	}

	uname, ok := j.token.Get("nickname")
	if ok {
		a.Name = uname.(string)
		a.Username = uname.(string)
	}
}

func tokenFromRequest(j *JWT, r *http.Request, findTokenFns ...func(r *http.Request) string) (jwt.Token, error) {
	var tokenString string

	// Extract token string from the request by calling token find functions in
	// the order they where provided. Further extraction stops if a function
	// returns a non-empty string.
	for _, fn := range findTokenFns {
		tokenString = fn(r)
		if tokenString != "" {
			break
		}
	}
	if tokenString == "" {
		return nil, ErrNoTokenFound
	}

	return verifyToken(j, tokenString)
}

func verifyToken(j *JWT, tokenString string) (jwt.Token, error) {
	keyPath := "https://" + j.domain + "/.well-known/jwks.json"
	keySet, err := jwk.Fetch(keyPath)
	if err != nil {
		return nil, err
	}

	j.verifier = jwt.WithKeySet(keySet)

	token, err := jwt.Parse(bytes.NewReader([]byte(tokenString)), j.verifier)
	if err != nil {
		return token, err
	}

	if err := jwt.Validate(token); err != nil {
		return token, err
	}

	return token, nil
}

// SetVerifier ...
func (j *JWT) SetVerifier(kid string) error {
	cert := ""
	resp, err := http.Get("https://" + j.domain + "/.well-known/jwks.json")

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var jwks = Jwks{}
	err = json.NewDecoder(resp.Body).Decode(&jwks)

	if err != nil {
		return err
	}

	for k := range jwks.Keys {
		if kid == jwks.Keys[k].Kid {
			j.alg = jwa.SignatureAlgorithm(jwks.Keys[k].Alg)
			cert = "-----BEGIN CERTIFICATE-----\n" + jwks.Keys[k].X5c[0] + "\n-----END CERTIFICATE-----"

			pubKey, err := parseRSAPublicKeyFromPem([]byte(cert))
			if err != nil {
				return err
			}

			j.verifyKey = pubKey
			j.verifier = jwt.WithVerify(j.alg, j.verifyKey)

			return nil
		}
	}

	return errors.New("unable to find appropriate key")
}

func parseRSAPublicKeyFromPem(key []byte) (*rsa.PublicKey, error) {
	var err error

	// Parse PEM block
	var block *pem.Block
	if block, _ = pem.Decode(key); block == nil {
		return nil, ErrKeyMustBePEMEncoded
	}

	// Parse the key
	var parsedKey interface{}
	if parsedKey, err = x509.ParsePKIXPublicKey(block.Bytes); err != nil {
		if cert, err := x509.ParseCertificate(block.Bytes); err == nil {
			parsedKey = cert.PublicKey
		} else {
			return nil, err
		}
	}

	var pkey *rsa.PublicKey
	var ok bool
	if pkey, ok = parsedKey.(*rsa.PublicKey); !ok {
		return nil, ErrNotRSAPublicKey
	}

	return pkey, nil
}

// Decode ...
func (j *JWT) Decode(tokenString string) (jwt.Token, error) {
	return j.parse([]byte(tokenString))
}

func (j *JWT) parse(payload []byte) (jwt.Token, error) {
	return jwt.Parse(bytes.NewReader([]byte(payload)))
}

func tokenFromHeader(r *http.Request) string {
	// Get token from authorization header.
	bearer := r.Header.Get("Authorization")
	if len(bearer) > 7 && strings.ToUpper(bearer[0:6]) == "BEARER" {
		return bearer[7:]
	}
	return ""
}
