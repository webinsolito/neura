package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

const (
	capSystemObserve  = "system.observe"
	capFSList         = "fs.list"
	capFSRead         = "fs.read"
	capWindowsObserve = "windows.observe"

	maxPassportTTL = 5 * time.Minute
	clockSkew      = 5 * time.Second
)

type CapabilityPassport struct {
	ID           string    `json:"id"`
	Subject      string    `json:"subject"`
	PurposeHash  string    `json:"purpose_hash"`
	Capabilities []string  `json:"capabilities"`
	IssuedAt     time.Time `json:"issued_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	Signature    string    `json:"signature"`
}

type CapabilityGate struct {
	key []byte
	now func() time.Time
}

func NewCapabilityGate() (*CapabilityGate, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("capability gate key generation failed: %w", err)
	}
	return &CapabilityGate{key: key, now: func() time.Time { return time.Now().UTC() }}, nil
}

func newCapabilityGateForTest(key []byte, now func() time.Time) (*CapabilityGate, error) {
	if len(key) < 32 {
		return nil, errors.New("capability gate key too short")
	}
	cp := append([]byte(nil), key...)
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &CapabilityGate{key: cp, now: now}, nil
}

func normalizeCapabilities(in []string) ([]string, error) {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, c := range in {
		c = strings.TrimSpace(strings.ToLower(c))
		if c == "" {
			continue
		}
		if c == "*" || strings.Contains(c, "*") {
			return nil, errors.New("wildcard capabilities are forbidden")
		}
		if _, ok := seen[c]; ok {
			continue
		}
		seen[c] = struct{}{}
		out = append(out, c)
	}
	if len(out) == 0 {
		return nil, errors.New("at least one capability required")
	}
	sort.Strings(out)
	return out, nil
}

func purposeDigest(purpose string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(purpose)))
	return hex.EncodeToString(sum[:16])
}

func randomPassportID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (g *CapabilityGate) Issue(subject, purpose string, capabilities []string, ttl time.Duration) (CapabilityPassport, error) {
	if g == nil || len(g.key) < 32 || g.now == nil {
		return CapabilityPassport{}, errors.New("capability gate unavailable")
	}
	subject = strings.TrimSpace(subject)
	if subject == "" {
		return CapabilityPassport{}, errors.New("capability subject required")
	}
	caps, err := normalizeCapabilities(capabilities)
	if err != nil {
		return CapabilityPassport{}, err
	}
	if ttl <= 0 || ttl > maxPassportTTL {
		return CapabilityPassport{}, errors.New("invalid capability passport ttl")
	}
	id, err := randomPassportID()
	if err != nil {
		return CapabilityPassport{}, fmt.Errorf("passport id generation failed: %w", err)
	}
	now := g.now().UTC()
	p := CapabilityPassport{
		ID:           id,
		Subject:      subject,
		PurposeHash:  purposeDigest(purpose),
		Capabilities: caps,
		IssuedAt:     now,
		ExpiresAt:    now.Add(ttl),
	}
	p.Signature = g.sign(p)
	return p, nil
}

func passportCanonical(p CapabilityPassport) string {
	return strings.Join([]string{
		p.ID,
		p.Subject,
		p.PurposeHash,
		strings.Join(p.Capabilities, ","),
		p.IssuedAt.UTC().Format(time.RFC3339Nano),
		p.ExpiresAt.UTC().Format(time.RFC3339Nano),
	}, "\n")
}

func (g *CapabilityGate) sign(p CapabilityPassport) string {
	mac := hmac.New(sha256.New, g.key)
	_, _ = mac.Write([]byte(passportCanonical(p)))
	return hex.EncodeToString(mac.Sum(nil))
}

func containsCapability(caps []string, wanted string) bool {
	i := sort.SearchStrings(caps, wanted)
	return i < len(caps) && caps[i] == wanted
}

func (g *CapabilityGate) Authorize(p CapabilityPassport, required, purpose string) error {
	if g == nil || len(g.key) < 32 || g.now == nil {
		return errors.New("capability gate unavailable")
	}
	if p.ID == "" || p.Subject == "" || p.Signature == "" {
		return errors.New("capability passport incomplete")
	}
	caps, err := normalizeCapabilities(p.Capabilities)
	if err != nil {
		return fmt.Errorf("invalid capability passport: %w", err)
	}
	if strings.Join(caps, "\x00") != strings.Join(p.Capabilities, "\x00") {
		return errors.New("capability passport not canonical")
	}
	expected, err := hex.DecodeString(g.sign(p))
	if err != nil {
		return errors.New("capability signature generation failed")
	}
	actual, err := hex.DecodeString(p.Signature)
	if err != nil || !hmac.Equal(expected, actual) {
		return errors.New("capability passport signature invalid")
	}
	now := g.now().UTC()
	if p.IssuedAt.After(now.Add(clockSkew)) {
		return errors.New("capability passport issued in the future")
	}
	if !p.ExpiresAt.After(now) {
		return errors.New("capability passport expired")
	}
	if p.ExpiresAt.Sub(p.IssuedAt) <= 0 || p.ExpiresAt.Sub(p.IssuedAt) > maxPassportTTL {
		return errors.New("capability passport lifetime invalid")
	}
	if p.PurposeHash != purposeDigest(purpose) {
		return errors.New("capability passport purpose mismatch")
	}
	required = strings.TrimSpace(strings.ToLower(required))
	if required == "" || !containsCapability(p.Capabilities, required) {
		return fmt.Errorf("capability %q denied", required)
	}
	return nil
}

func capabilityForTool(name string) (string, bool) {
	switch name {
	case "system.info":
		return capSystemObserve, true
	case "fs.list":
		return capFSList, true
	case "fs.read":
		return capFSRead, true
	case "windows.processes":
		return capWindowsObserve, true
	default:
		return "", false
	}
}

func capabilityCatalog() map[string]string {
	return map[string]string{
		"system.info":       capSystemObserve,
		"fs.list":           capFSList,
		"fs.read":           capFSRead,
		"windows.processes": capWindowsObserve,
	}
}
