package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"newsaggregator/internal/db"
	"time"
)

const HOUR = 3600

type AccessToken struct {
	token      []byte
	timeToLive time.Time
}

func NewAccessToken(db *db.DB) *AccessToken {
	a := &AccessToken{}
	tk := a.generateToken()

	hash := sha256.Sum256(tk)
	dbKey := append([]byte("token:"), hash[:]...)

	ttlTimestamp := time.Now().Add(24 * time.Hour).Unix()

	tv := make([]byte, 8)
	binary.BigEndian.PutUint64(tv, uint64(ttlTimestamp))

	db.Insert(dbKey, tv)
	a.token = tk
	a.timeToLive = time.Unix(ttlTimestamp, 0)
	return a
}

func (a *AccessToken) generateToken() (t []byte) {
	t = make([]byte, 32)
	rand.Read(t)

	return t
}
