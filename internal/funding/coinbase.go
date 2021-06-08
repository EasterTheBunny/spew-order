package funding

import (
	"bytes"
	"crypto"
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/easterthebunny/spew-order/pkg/types"
	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
)

var (
	ErrRequestBodyNotVerified = errors.New("request body not verified")
)

// CoinbasePublicKeyURL = "https://www.coinbase.com/coinbase.pub"

// NewCoinbaseSource returns an implementation of the Source interface with a
// connection to Coinbase
func NewCoinbaseSource(config SourceConfig) Source {
	return &coinbaseSource{
		baseURL:   "https://api.coinbase.com",
		pubkeySrc: config.PublicKey,
		accounts:  make(map[string]string),
		client: &http.Client{
			Timeout: time.Second * 3},
		auditLog:  config.CallbackAudit,
		apikey:    config.APIKey,
		apisecret: config.APISecret}
}

type coinbaseSource struct {
	baseURL   string
	accounts  map[string]string
	apikey    string
	apisecret string
	client    *http.Client
	pubkeySrc io.Reader
	pubkey    *rsa.PublicKey
	auditLog  io.Writer
}

func (s *coinbaseSource) Name() string {
	return "COINBASE"
}

func (s *coinbaseSource) Supports(sym types.Symbol) bool {
	switch sym {
	case types.SymbolBitcoin, types.SymbolEthereum:
		return true
	default:
		return false
	}
}

// CreateAddress returns a new address for the given symbol
func (s *coinbaseSource) CreateAddress(sym types.Symbol) (address *Address, err error) {

	acct, ok := s.accounts[sym.String()]
	if !ok {
		err = s.getAccounts()
		if err != nil {
			return
		}

		acct, ok = s.accounts[sym.String()]
		if !ok {
			err = errors.New("account not found")
			return
		}
	}

	path := fmt.Sprintf("/v2/account/%s/addresses", acct)

	str := fmt.Sprintf(`{"name": "%s"}`, uuid.NewV4())
	resp, err := s.request("POST", path, strings.NewReader(str))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	data, err := s.extractResponsePayload(resp.Body)
	if err != nil {
		return
	}

	var obj coinbaseAddressResourceV2
	err = json.Unmarshal(data.Data, &obj)
	if err != nil {
		return
	}

	addr := Address{
		ID:   obj.ID,
		Hash: obj.Address}

	return &addr, nil
}

func (s *coinbaseSource) getSignaturePublicKey() (*rsa.PublicKey, error) {
	var err error

	b, err := readPubKey(s.pubkeySrc)
	if err != nil {
		return nil, errors.New("public key source not available")
	}

	// check the reader for new content; if there is no new content
	// and a public key exists, return the existing public key
	// if there is content in the reader, attempt to create a public
	// key from it
	if len(b) == 0 && s.pubkey != nil {
		return s.pubkey, nil
	}

	pubPem, _ := pem.Decode(b)
	if pubPem == nil {
		return nil, errors.New("pem decode error")
	}

	var parsedKey interface{}
	if parsedKey, err = x509.ParsePKIXPublicKey(pubPem.Bytes); err != nil {
		return nil, errors.New("unable to parse as PKCS1 public key")
	}

	var pubKey *rsa.PublicKey
	var ok bool
	if pubKey, ok = parsedKey.(*rsa.PublicKey); !ok {
		return nil, errors.New("unable to read public key")
	}

	s.pubkey = pubKey
	return pubKey, nil
}

func (s *coinbaseSource) Withdraw(*Transaction) error {
	return errors.New("not implemented")
}

func (s *coinbaseSource) OKResponse() int {
	return http.StatusOK
}

func (s *coinbaseSource) request(method string, path string, data interface{}) (*http.Response, error) {

	var body io.Reader
	var bodyBytes []byte
	var err error

	if data != nil {
		bodyBytes, err = json.Marshal(data)
		if err != nil {
			return nil, err
		}

		body = bytes.NewReader(bodyBytes)
	}

	tm, err := s.getTime()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s%s", s.baseURL, path)

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	hash := hmac.New(sha256.New, []byte(s.apisecret))
	_, err = io.WriteString(hash, fmt.Sprintf("%d%s%s%s", tm, method, path, string(bodyBytes)))
	if err != nil {
		return nil, err
	}
	encoded := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("CB-VERSION", "2021-05-27")
	req.Header.Add("CB-ACCESS-KEY", s.apikey)
	req.Header.Add("CB-ACCESS-SIGN", encoded)
	req.Header.Add("CB-ACCESS-TIMESTAMP", strconv.FormatInt(tm, 10))

	return s.client.Do(req)
}

func (s *coinbaseSource) verifyRequest(k *rsa.PublicKey, signature string, message []byte) error {

	sig, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return err
	}

	hashed := sha256.Sum256(message)
	err = rsa.VerifyPKCS1v15(k, crypto.SHA256, hashed[:], sig)
	if err != nil {
		return err
	}

	return nil
}

func (s *coinbaseSource) Callback() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ok := true

			for ok {
				var err error
				if r.Method != "POST" {
					err = errors.New("incorrect callback method")
					ctx = attachToContext(ctx, nil, &CallbackError{Status: http.StatusNotAcceptable, Err: err})
					break
				}

				signature := r.Header.Get("CB-SIGNATURE")
				body, err := ioutil.ReadAll(r.Body)
				if err != nil {
					err = fmt.Errorf("%w:read: %s", ErrRequestBodyParseError, err)
					ctx = attachToContext(ctx, nil, &CallbackError{Status: http.StatusBadRequest, Err: err})
					break
				}

				k, err := s.getSignaturePublicKey()
				if err != nil {
					err = fmt.Errorf("%w:pubkey: %s", ErrRequestBodyNotVerified, err)
					ctx = attachToContext(ctx, nil, &CallbackError{Status: http.StatusInternalServerError, Err: err})
					break
				}

				err = s.verifyRequest(k, signature, body)
				if err != nil {
					err = fmt.Errorf("%w:sign: %s", ErrRequestBodyNotVerified, err)
					ctx = attachToContext(ctx, nil, &CallbackError{Status: http.StatusForbidden, Err: err})
					break
				}

				_, err = s.auditLog.Write(body)
				if err != nil {
					err = fmt.Errorf("%w:auditlog: %s", ErrRequestBodyNotVerified, err)
					ctx = attachToContext(ctx, nil, &CallbackError{Status: http.StatusInternalServerError, Err: err})
					break
				}

				// find and add resource to context
				payload, err := s.extractNotificationPayload(r.Body)
				if err != nil {
					err = fmt.Errorf("%w:notification: %s", ErrRequestBodyParseError, err)
					ctx = attachToContext(ctx, nil, &CallbackError{Status: http.StatusBadRequest, Err: err})
					break
				}

				if payload.Attempts > 1 {
					log.Printf("multiple callback attempts '%d' found for notification id '%s'", payload.Attempts, payload.ID)
				}

				switch payload.Type {
				case cbNewPayment:
					var adr coinbaseAddressResourceV2
					var pmt coinbaseNewPaymentResourceV2

					err = json.Unmarshal(payload.Data, &adr)
					if err != nil {
						err = fmt.Errorf("%w:address: %s", ErrRequestBodyParseError, err)
						ctx = attachToContext(ctx, nil, &CallbackError{Status: http.StatusBadRequest, Err: err})
						break
					}

					err = json.Unmarshal(payload.AdditionalData, &pmt)
					if err != nil {
						err = fmt.Errorf("%w:payment: %s", ErrRequestBodyParseError, err)
						ctx = attachToContext(ctx, nil, &CallbackError{Status: http.StatusBadRequest, Err: err})
						break
					}

					tr := Transaction{
						Symbol:          coinbaseSymbol(string(pmt.Amount.Currency)),
						TransactionHash: pmt.Hash,
						Address:         adr.Address,
						Amount:          pmt.Amount.Amount,
					}

					ctx = attachToContext(ctx, tr, nil)
				}

				ok = false
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (s *coinbaseSource) extractNotificationPayload(r io.Reader) (*coinbaseNotificationPayloadV2, error) {

	callback := &coinbaseNotificationPayloadV2{}
	if err := json.NewDecoder(r).Decode(callback); err != nil {
		return nil, err
	}

	return callback, nil
}

func (s *coinbaseSource) extractResponsePayload(r io.Reader) (*coinbaseResponsePayloadV2, error) {

	data := &coinbaseResponsePayloadV2{}
	if err := json.NewDecoder(r).Decode(data); err != nil {
		return nil, err
	}

	return data, nil
}

func (s *coinbaseSource) getAccounts() error {
	path := "/v2/accounts"

	resp, err := s.request("GET", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	payload, err := s.extractResponsePayload(resp.Body)
	if err != nil {
		return err
	}

	var accts []coinbaseAccountResourceV2
	err = json.Unmarshal(payload.Data, &accts)
	if err != nil {
		return err
	}

	for _, acct := range accts {
		s.accounts[string(acct.Currency)] = acct.ID
	}

	return nil
}

func (s *coinbaseSource) getTime() (t int64, err error) {
	url := fmt.Sprintf("%s/v2/time", s.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	payload, err := s.extractResponsePayload(resp.Body)
	if err != nil {
		return
	}

	var e tm
	err = json.Unmarshal(payload.Data, &e)
	if err != nil {
		return
	}

	t = e.Epoch

	return
}

type coinbaseResourceType string

const (
	cbAddressResource     coinbaseResourceType = "address"
	cbAccountResource     coinbaseResourceType = "account"
	cbTransactionResource coinbaseResourceType = "transaction"
)

type coinbaseAddressResourceV2 struct {
	ID        string               `json:"id"`
	Address   string               `json:"address"`
	Name      string               `json:"name"`
	CreatedAt string               `json:"created_at"`
	UpdatedAt string               `json:"updated_at"`
	Network   string               `json:"network"`
	Resource  coinbaseResourceType `json:"resource"`
	Path      string               `json:"resource_path"`
}

type coinbaseNewPaymentResourceV2 struct {
	Hash        string                        `json:"hash"`
	Amount      coinbaseMoneyResourceV2       `json:"amount"`
	Transaction coinbaseTransactionResourceV2 `json:"transaction"`
}

type coinbaseTransactionResourceV2 struct {
	ID       string               `json:"id"`
	Resource coinbaseResourceType `json:"resource"`
	Path     string               `json:"resource_path"`
}

type coinbaseResponsePayloadV2 struct {
	Data json.RawMessage `json:"data"`
}

type coinbaseNotificationPayloadV2 struct {
	ID             string               `json:"id"`
	Type           coinbaseCallbackType `json:"type"`
	Data           json.RawMessage      `json:"data"`
	Attempts       int                  `json:"delivery_attempts"`
	AdditionalData json.RawMessage      `json:"additional_data"`
}

type coinbaseCallbackType string

const (
	cbNewPayment coinbaseCallbackType = "wallet:addresses:new-payment"
)

type coinbaseAccountResourceV2 struct {
	ID        string                  `json:"id"`
	Name      string                  `json:"name"`
	Primary   bool                    `json:"primary"`
	Type      coinbaseAccountType     `json:"type"`
	Currency  coinbaseCurrencyType    `json:"currency"`
	Balance   coinbaseMoneyResourceV2 `json:"balance"`
	CreatedAt string                  `json:"created_at"`
	UpdatedAt string                  `json:"updated_at"`
	Resource  coinbaseResourceType    `json:"resource"`
	Path      string                  `json:"resource_path"`
}

type coinbaseMoneyResourceV2 struct {
	Amount   decimal.Decimal      `json:"amount"`
	Currency coinbaseCurrencyType `json:"currency"`
}

type coinbaseAccountType string

const (
	cbWalletAccount coinbaseAccountType = "wallet"
)

type coinbaseCurrencyType string

const (
	cbBitcoinCurrency coinbaseCurrencyType = "BTC"
)

type tm struct {
	ISO   string `json:"iso"`
	Epoch int64  `json:"epoch"`
}

func coinbaseSymbol(s string) types.Symbol {
	switch s {
	case "BTC":
		return types.SymbolBitcoin
	case "ETH":
		return types.SymbolEthereum
	default:
		return types.SymbolBitcoin
	}
}

func readPubKey(r io.Reader) ([]byte, error) {
	output := []byte{}

	buf := make([]byte, 0, 1024)
	for {
		n, err := r.Read(buf[:cap(buf)])
		buf = buf[:n]
		if n == 0 {
			if err == nil {
				continue
			}
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}

		output = append(output, buf...)
		if err != nil && err != io.EOF {
			return output, err
		}
	}
	return output, nil
}
