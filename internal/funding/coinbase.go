package funding

import (
	"crypto"
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

func NewCoinbaseSource(audit io.Writer, src io.Reader) Source {
	return &coinbaseSource{
		baseURL:   "https://api.coinbase.com/v2",
		pubkeySrc: src,
		accounts:  make(map[string]string),
		client: &http.Client{
			Timeout: time.Second * 3},
		auditLog: audit}
}

type coinbaseSource struct {
	baseURL   string
	accounts  map[string]string
	bearer    string
	client    *http.Client
	pubkeySrc io.Reader
	pubkey    *rsa.PublicKey
	auditLog  io.Writer
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

func (s *coinbaseSource) request(method string, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.bearer))

	return s.client.Do(req)
}

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

	url := fmt.Sprintf("%s/account/%s/addresses", s.baseURL, acct)

	str := fmt.Sprintf(`{"name": "%s"}`, uuid.NewV4())
	resp, err := s.request("POST", url, strings.NewReader(str))
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

	if s.pubkey != nil {
		return s.pubkey, nil
	}

	var err error

	b, err := ioutil.ReadAll(s.pubkeySrc)
	if err != nil {
		return nil, errors.New("public key source not available")
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

func (s *coinbaseSource) Withdraw(*Transaction) error {
	return errors.New("not implemented")
}

func (s *coinbaseSource) OKResponse() int {
	return http.StatusOK
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

func (s *coinbaseSource) getAccounts() error {
	url := fmt.Sprintf("%s/accounts", s.baseURL)

	str := fmt.Sprintf(`{"name": "%s"}`, uuid.NewV4())
	resp, err := s.request("POST", url, strings.NewReader(str))
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
