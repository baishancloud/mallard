package httptoken

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/baishancloud/mallard/corelib/zaplog"
	"golang.org/x/time/rate"
)

const (
	// DefaultRateLimit is default rate limit for one user in one second
	DefaultRateLimit = 10
)

var (
	tokenMap        map[string]*VerifierUser
	rateMap         map[string]*rate.Limiter
	tokenLock       sync.RWMutex
	tokeFileModTime int64

	log = zaplog.Zap("httptoken")
)

type (
	// VerifierUser is setting for one verifier user
	VerifierUser struct {
		User      string `json:"user"`
		Token     string `json:"token"`
		RateLimit int    `json:"rate_limit"`
	}
)

// SyncVerifier reloads token file in time loop
func SyncVerifier(file string, interval time.Duration) {
	if file == "" {
		log.Info("no-verify-file")
		return
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		refreshVerifyFile(file)
		<-ticker.C
	}
}

// VerifyRequest checks user and token from http request
func VerifyRequest(r *http.Request) (string, string, bool) {
	r.ParseForm()
	user, token := r.Header.Get("Transfer-User"), r.Header.Get("Transfer-Token")
	if user == "" {
		user = r.Form.Get("user")
	}
	if token == "" {
		token = r.Form.Get("token")
	}
	return user, token, VerifyUserToken(user, token)
}

// VerifyUserToken checks user and token
func VerifyUserToken(user string, token string) bool {
	if token == "" || user == "" {
		return false
	}
	tokenLock.RLock()
	defer tokenLock.RUnlock()
	return tokenMap[user].Token == token
}

func refreshVerifyFile(file string) {
	info, err := os.Stat(file)
	if err != nil {
		log.Warn("refresh-error", "error", err)
		return
	}
	modTime := info.ModTime().Unix()
	if modTime == tokeFileModTime {
		log.Debug("refresh-no-change")
		return
	}
	b, err := ioutil.ReadFile(file)
	if err != nil {
		log.Warn("refresh-error", "error", err)
		return
	}
	tokens := make(map[string]*VerifierUser)
	if err = json.Unmarshal(b, &tokens); err != nil {
		log.Warn("refresh-error", "error", err)
		return
	}
	tokenLock.Lock()
	tokenMap = tokens
	rateMap = make(map[string]*rate.Limiter)
	for user, tk := range tokens {
		if tk.RateLimit == 0 {
			tk.RateLimit = DefaultRateLimit
		}
		rateMap[user] = rate.NewLimiter(rate.Every(time.Second), tk.RateLimit)
	}
	log.Debug("refresh-ok", "tokens", tokens)
	tokeFileModTime = modTime
	tokenLock.Unlock()
}

// VerifyAllowLimit checks rate limit of user
func VerifyAllowLimit(user string) bool {
	tokenLock.RLock()
	defer tokenLock.RUnlock()
	rt := rateMap[user]
	if rt == nil {
		return false
	}
	return rt.Allow()
}

var (
	// ErrorTokenInvalid is error of token is wrong
	ErrorTokenInvalid = errors.New("token-invalid")
	// ErrorLimitExceeded is error of requests frequency over limit
	ErrorLimitExceeded = errors.New("limit-exceeded")
)

// VerifyAndAllow checks token and rate limit together,
// if token is wrong, return ErrorTokenInvalid
// if limit exceeded, return ErrorLimitExceeded
func VerifyAndAllow(r *http.Request) (string, string, error) {
	user, token, ok := VerifyRequest(r)
	if !ok {
		return "", "", ErrorTokenInvalid
	}
	if !VerifyAllowLimit(user) {
		return "", "", ErrorLimitExceeded
	}
	return user, token, nil
}

// GetUserVerifier gets user token info by user name
func GetUserVerifier(user string) *VerifierUser {
	tokenLock.RLock()
	defer tokenLock.RUnlock()
	return tokenMap[user]
}
