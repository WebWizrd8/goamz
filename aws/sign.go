package aws

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"sort"
	"strings"
)

type V2Signer struct {
	auth    Auth
	service ServiceInfo
}

var b64 = base64.StdEncoding

func (s *V2Signer) Sign(method, path string, params map[string]string) {
	params["AWSAccessKeyId"] = s.auth.AccessKey
	params["SignatureVersion"] = "2"
	params["SignatureMethod"] = "HmacSHA256"

	// AWS specifies that the parameters in a signed request must
	// be provided in the natural order of the keys. This is distinct
	// from the natural order of the encoded value of key=value.
	// Percent and equals affect the sorting order.
	var keys, sarray []string
	for k, _ := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		sarray = append(sarray, Encode(k)+"="+Encode(params[k]))
	}
	joined := strings.Join(sarray, "&")
	payload := method + "\n" + s.service.Endpoint + "\n" + path + "\n" + joined
	hash := hmac.New(sha256.New, []byte(s.auth.SecretKey))
	hash.Write([]byte(payload))
	signature := make([]byte, b64.EncodedLen(hash.Size()))
	b64.Encode(signature, hash.Sum(nil))

	params["Signature"] = string(signature)
}
