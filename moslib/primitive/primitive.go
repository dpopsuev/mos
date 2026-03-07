// Package primitive defines the lowest-level artifact types for Mos.
// It is consumed by testkit/world for integration tests and may be
// consumed by external projects (e.g. Asterisk). It deliberately has
// zero internal dependencies so that it remains a stable leaf package.
package primitive

import "time"

// Artifact is the fundamental unit of Mos. Every rule, contract,
// declaration, and bill is an artifact at its core: identity + lifecycle + spec.
// Governance commands are lexicon sugar on top of this primitive.
type Artifact struct {
	ID       string   `toml:"id"`
	Kind     string   `toml:"kind"`
	Title    string   `toml:"title"`
	Status   string   `toml:"status"`
	Scope    []string `toml:"scope,omitempty"`
	Spec     Spec     `toml:"spec"`
	Identity Identity `toml:"identity"`
}

// Spec holds the Gherkin feature text -- the definition of "done."
type Spec struct {
	Feature string `toml:"feature"`
}

// Identity tracks authorship and amendment state. This is a snapshot:
// who created it, current version, who last amended it.
// Full amendment history lives in the signed hash chain (.mos/history/).
type Identity struct {
	CreatedBy         string    `toml:"created_by"`
	CreatedAt         time.Time `toml:"created_at"`
	CreationSignature string    `toml:"creation_signature"`
	Version           int       `toml:"version"`
	LastAmendedBy     string    `toml:"last_amended_by"`
	LastAmendedAt     time.Time `toml:"last_amended_at"`
}

// Signer produces signatures for artifact operations.
type Signer interface {
	Fingerprint() string
	Sign(data []byte) (string, error)
}

// NewArtifact creates an artifact with the given parameters, signed by signer.
func NewArtifact(id, kind, title, spec string, signer Signer) (*Artifact, error) {
	now := time.Now().UTC()
	a := &Artifact{
		ID:     id,
		Kind:   kind,
		Title:  title,
		Status: "draft",
		Spec: Spec{
			Feature: spec,
		},
		Identity: Identity{
			CreatedBy: signer.Fingerprint(),
			CreatedAt: now,
			Version:   1,
		},
	}

	sig, err := signer.Sign([]byte(a.signingPayload()))
	if err != nil {
		return nil, err
	}
	a.Identity.CreationSignature = sig
	return a, nil
}

// Amend applies a mutation to the artifact and records the amendment.
func (a *Artifact) Amend(signer Signer, mutate func(a *Artifact)) error {
	mutate(a)
	now := time.Now().UTC()
	a.Identity.Version++
	a.Identity.LastAmendedBy = signer.Fingerprint()
	a.Identity.LastAmendedAt = now

	sig, err := signer.Sign([]byte(a.signingPayload()))
	if err != nil {
		return err
	}
	a.Identity.CreationSignature = sig
	return nil
}

func (a *Artifact) signingPayload() string {
	return a.ID + "|" + a.Kind + "|" + a.Title + "|" + a.Status + "|" + a.Spec.Feature
}
