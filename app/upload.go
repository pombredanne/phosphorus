package app

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"willstclair.com/phosphorus/db"
)

var Token = &Resource{
	Path: "/_upload",
	Get:  UploadToken}

type _uploadForm struct {
	Key            string `json:"key"`
	AWSAccessKeyId string `json:"AWSAccessKeyId"`
	Policy         string `json:"policy"`
	Signature      string `json:"signature"`
}

func (f *_uploadForm) Flush(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(f)
}

var s3policy = `{"conditions":[["starts-with","$key","%s/"],{"bucket":"phosphorus-upload"},{"acl":"private"},["starts-with","$Content-Type","text/csv"],{"success_action_status":"200"}],"expiration":"%s"}`

func UploadToken(r *http.Request, e *Env, m map[string]interface{}) (Response, error) {
	account := &Account{Username: m["session"].(*Session).Username}
	err := db.GetItem(e.Accounts, account)
	if err != nil {
		return nil, fmt.Errorf("Bad authentication")
	}

	expires := time.Now().UTC().Add(1 * time.Minute)
	policy := fmt.Sprintf(s3policy, account.Username,
		expires.Format("2006-01-02T15:04:05.000Z"))
	encodedPolicy := base64.StdEncoding.EncodeToString([]byte(policy))

	mac := hmac.New(sha1.New, []byte(e.Auth.SecretKey))
	mac.Write([]byte(encodedPolicy))
	sig := mac.Sum(nil)

	return &_uploadForm{
		Key:            fmt.Sprintf("%s/%d", account.Username, e.IdGen.SafeId()),
		AWSAccessKeyId: e.Auth.AccessKey,
		Policy:         encodedPolicy,
		Signature:      base64.StdEncoding.EncodeToString(sig)}, nil

}
