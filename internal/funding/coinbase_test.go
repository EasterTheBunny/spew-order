package funding

import (
	"crypto/rsa"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCoinbaseVerifyRequest(t *testing.T) {

	src := strings.NewReader(`-----BEGIN PUBLIC KEY-----
MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEA9MsJBuXzFGIh/xkAA9Cy
QdZKRerV+apyOAWY7sEYV/AJg+AX/tW2SHeZj+3OilNYm5DlBi6ZzDboczmENrFn
mUXQsecsR5qjdDWb2qYqBkDkoZP02m9o9UmKObR8coKW4ZBw0hEf3fP9OEofG2s7
Z6PReWFyQffnnecwXJoN22qjjsUtNNKOOo7/l+IyGMVmdzJbMWQS4ybaU9r9Ax0J
4QUJSS/S4j4LP+3Z9i2DzIe4+PGa4Nf7fQWLwE45UUp5SmplxBfvEGwYNEsHvmRj
usIy2ZunSO2CjJ/xGGn9+/57W7/SNVzk/DlDWLaN27hUFLEINlWXeYLBPjw5GGWp
ieXGVcTaFSLBWX3JbOJ2o2L4MxinXjTtpiKjem9197QXSVZ/zF1DI8tRipsgZWT2
/UQMqsJoVRXHveY9q9VrCLe97FKAUiohLsskr0USrMCUYvLU9mMw15hwtzZlKY8T
dMH2Ugqv/CPBuYf1Bc7FAsKJwdC504e8kAUgomi4tKuUo25LPZJMTvMTs/9IsRJv
I7ibYmVR3xNsVEpupdFcTJYGzOQBo8orHKPFn1jj31DIIKociCwu6m8ICDgLuMHj
7bUHIlTzPPT7hRPyBQ1KdyvwxbguqpNhqp1hG2sghgMr0M6KMkUEz38JFElsVrpF
4z+EqsFcIZzjkSG16BjjjTkCAwEAAQ==
-----END PUBLIC KEY-----

date: 2014-07-09 13:37:00 UTC
version: 1`)

	s := &coinbaseSource{
		pubkeySrc: src,
		client: &http.Client{
			Timeout: time.Second * 3}}
	var pubkey *rsa.PublicKey
	var err error

	body := `{"order":{"id":null,"created_at":null,"status":"completed","event":null,"total_btc":` +
		`{"cents":100000000,"currency_iso":"BTC"},"total_native":{"cents":1000,"currency_iso":"USD"},` +
		`"total_payout":{"cents":1000,"currency_iso":"USD"},"custom":"123456789","receive_address":` +
		`"mzVoQenSY6RTBgBUcpSBTBAvUMNgGWxgJn","button":{"type":"buy_now","name":"Test Item",` +
		`"description":null,"id":null},"transaction":{"id":"53bdfe4d091c0d74a7000003",` +
		`"hash":"4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b","confirmations":0}}}`

	signature := `6yQRl17CNj5YSHSpF+tLjb0vVsNVEv021Tyy1bTVEQ69SWlmhwmJYuMc7jiDyeW9TLy4vRqSh4g4YEyN8eoQI` +
		`M57pMoNw6Lw6Oudubqwp+E3cKtLFxW0l18db3Z/vhxn5BScAutHWwT/XrmkCNaHyCsvOOGMekwrNO7mxX9QIx21FBaEejJ` +
		`eviSYrF8bG6MbmFEs2VGKSybf9YrElR8BxxNe/uNfCXN3P5tO8MgR5wlL3Kr4yq8e6i4WWJgD08IVTnrSnoZR6v8JkPA+f` +
		`n7I0M6cy0Xzw3BRMJAvdQB97wkobu97gFqJFKsOH2u/JR1S/UNP26vL0mzuAVuKAUwlRn0SUhWEAgcM3X0UCtWLYfCIb5Q` +
		`qrSHwlp7lwOkVnFt329Mrpjy+jAfYYSRqzIsw4ZsRRVauy/v3CvmjPI9sUKiJ5l1FSgkpK2lkjhFgKB3WaYZWy9ZfIAI9b` +
		`DyG8vSTT7IDurlUhyTweDqVNlYUsO6jaUa4KmSpg1o9eIeHxm0XBQ2c0Lv/T39KNc/VOAi1LBfPiQYMXD1e/8VuPPBTDGg` +
		`zOMD3i334ppSr36+8YtApAn3D36Hr9jqAfFrugM7uPecjCGuleWsHFyNnJErT0/amIt24Nh1GoiESEq42o7Co4wZieKZ+/` +
		`yeAlIUErJzK41ACVGmTnGoDUwEBXxADOdA=`

	t.Run("getSignaturePublicKey", func(t *testing.T) {
		pubkey, err = s.getSignaturePublicKey()
		assert.NoError(t, err)
	})

	t.Run("verifyRequest", func(t *testing.T) {
		err := s.verifyRequest(pubkey, signature, []byte(body))
		assert.NoError(t, err)
	})
}

func TestCoinbaseGetTime(t *testing.T) {
	s := &coinbaseSource{
		baseURL: "https://api.coinbase.com",
		client: &http.Client{
			Timeout: time.Second * 3}}

	_, err := s.getTime()
	assert.NoError(t, err)
}
