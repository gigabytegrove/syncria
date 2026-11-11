package agentsecurity

import (
 "crypto/rand"
 "crypto/sha256"
 "crypto/subtle"
 "encoding/hex"
 "errors"
 "sync"
 "time"
)

var (
 ErrInvalidCode = errors.New("invalid or expired pairing code")
 ErrNotApproved = errors.New("device enrollment has not been approved")
 ErrRevoked = errors.New("device identity is revoked")
)

type Enrollment struct {
 ID string
 DeviceName string
 ExpiresAt time.Time
 Approved bool
 Redeemed bool
}

type Registry struct {
 mu sync.Mutex
 pending map[string]Enrollment
 devices map[string][32]byte
 revoked map[string]bool
 now func() time.Time
}

// NewRegistry is an in-memory reference implementation for secure enrollment.
// Persistent device identity storage is required before production deployment.
func NewRegistry() *Registry {
 return &Registry{pending:make(map[string]Enrollment),devices:make(map[string][32]byte),revoked:make(map[string]bool),now:time.Now}
}

func randomHex(n int) (string,error) {
 b:=make([]byte,n)
 if _,err:=rand.Read(b);err!=nil{return "",err}
 return hex.EncodeToString(b),nil
}

// Request creates a one-time opaque pairing code without enrolling the device.
func (r *Registry) Request(name string,ttl time.Duration)(Enrollment,string,error){
 if len(name)==0||len(name)>128||ttl<=0||ttl>24*time.Hour{return Enrollment{},"",errors.New("invalid enrollment request")}
 id,err:=randomHex(16);if err!=nil{return Enrollment{},"",err}
 code,err:=randomHex(32);if err!=nil{return Enrollment{},"",err}
 e:=Enrollment{ID:id,DeviceName:name,ExpiresAt:r.now().Add(ttl)}
 r.mu.Lock();defer r.mu.Unlock()
 r.pending[code]=e
 return e,code,nil
}

// Approve must only be called after a separate authenticated administrator action.
func (r *Registry) Approve(code string) error {
 r.mu.Lock();defer r.mu.Unlock()
 e,ok:=r.pending[code]
 if !ok||!r.now().Before(e.ExpiresAt)||e.Redeemed{return ErrInvalidCode}
 e.Approved=true;r.pending[code]=e
 return nil
}

// Redeem issues a per-device secret once. The raw secret is never retained.
func (r *Registry) Redeem(code string)(string,string,error){
 r.mu.Lock();defer r.mu.Unlock()
 e,ok:=r.pending[code]
 if !ok||!r.now().Before(e.ExpiresAt)||e.Redeemed{return "","",ErrInvalidCode}
 if !e.Approved{return "","",ErrNotApproved}
 secret,err:=randomHex(32);if err!=nil{return "","",err}
 r.devices[e.ID]=sha256.Sum256([]byte(secret))
 e.Redeemed=true
 delete(r.pending,code)
 return e.ID,secret,nil
}

func (r *Registry) Authenticate(id,secret string)error{
 r.mu.Lock();defer r.mu.Unlock()
 if r.revoked[id]{return ErrRevoked}
 digest,ok:=r.devices[id];if !ok{return errors.New("unknown device")}
 supplied:=sha256.Sum256([]byte(secret))
 if subtle.ConstantTimeCompare(digest[:],supplied[:])!=1{return errors.New("invalid device credential")}
 return nil
}

func (r *Registry) Revoke(id string) error {
 r.mu.Lock();defer r.mu.Unlock()
 if _,ok:=r.devices[id];!ok{return errors.New("unknown device")}
 delete(r.devices,id)
 r.revoked[id]=true
 return nil
}
