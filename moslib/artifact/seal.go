package artifact

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/dpopsuev/mos/moslib/dsl"
)

// SealInfo holds metadata about a seal on a contract.
type SealInfo struct {
	Intent    string
	Operator  string
	Timestamp string
	Message   string
}

// ResolveOperator determines the current operator identity.
// Priority: $MOS_GPG_KEY > git config user.signingkey > "name <email>" from git config.
func ResolveOperator() (string, error) {
	if key := os.Getenv("MOS_GPG_KEY"); key != "" {
		return key, nil
	}

	out, err := exec.Command("git", "config", "user.signingkey").Output()
	if err == nil {
		key := strings.TrimSpace(string(out))
		if key != "" {
			return key, nil
		}
	}

	name, err1 := exec.Command("git", "config", "user.name").Output()
	email, err2 := exec.Command("git", "config", "user.email").Output()
	if err1 == nil && err2 == nil {
		n := strings.TrimSpace(string(name))
		e := strings.TrimSpace(string(email))
		if n != "" || e != "" {
			return fmt.Sprintf("%s <%s>", n, e), nil
		}
	}

	return "", fmt.Errorf("cannot resolve operator identity: set $MOS_GPG_KEY, or configure git user.signingkey / user.name+user.email")
}

// SealContract attaches a seal block to a contract.
func SealContract(root, id, intent, message string) error {
	contractPath, err := FindContractPath(root, id)
	if err != nil {
		return fmt.Errorf("SealContract: %w", err)
	}

	operator, err := ResolveOperator()
	if err != nil {
		return fmt.Errorf("SealContract: %w", err)
	}

	if err := dsl.WithArtifact(contractPath, func(ab *dsl.ArtifactBlock) error {
		if existing, _ := readSeal(ab); existing != nil {
			return fmt.Errorf("contract %s is already sealed by %s (intent: %s)", id, existing.Operator, existing.Intent)
		}
		sealBlock := &dsl.Block{
			Name:  "seal",
			Title: intent,
			Items: []dsl.Node{
				&dsl.Field{Key: "operator", Value: &dsl.StringVal{Text: operator}},
				&dsl.Field{Key: "timestamp", Value: &dsl.DateTimeVal{Raw: time.Now().UTC().Format(time.RFC3339)}},
			},
		}
		if message != "" {
			sealBlock.Items = append(sealBlock.Items, &dsl.Field{Key: "message", Value: &dsl.StringVal{Text: message}})
		}
		ab.Items = append(ab.Items, sealBlock)
		return nil
	}); err != nil {
		return fmt.Errorf("SealContract: %w", err)
	}

	mosDir := filepath.Join(root, MosDir)
	if ValidateContract != nil {
		return ValidateContract(contractPath, mosDir)
	}
	return nil
}

// UnsealContract removes the seal block from a contract.
func UnsealContract(root, id string, force bool) error {
	contractPath, err := FindContractPath(root, id)
	if err != nil {
		return fmt.Errorf("UnsealContract: %w", err)
	}

	var existing *SealInfo
	if err := dsl.WithArtifact(contractPath, func(ab *dsl.ArtifactBlock) error {
		existing, _ = readSeal(ab)
		if existing == nil {
			return nil
		}
		if !force {
			operator, err := ResolveOperator()
			if err != nil {
				return err
			}
			if operator != existing.Operator {
				myLevel, _ := ResolveAuthority(root, operator)
				theirLevel, _ := ResolveAuthority(root, existing.Operator)
				if myLevel <= theirLevel {
					return fmt.Errorf("contract %s is sealed by %s; your authority level (%d) is not higher than theirs (%d); use --force to override", id, existing.Operator, myLevel, theirLevel)
				}
			}
		}
		ab.Items = filterNodes(ab.Items, func(n dsl.Node) bool {
			blk, ok := n.(*dsl.Block)
			return ok && blk.Name == "seal"
		})
		return nil
	}); err != nil {
		return fmt.Errorf("UnsealContract: %w", err)
	}

	if existing == nil {
		return nil
	}

	mosDir := filepath.Join(root, MosDir)
	if ValidateContract != nil {
		return ValidateContract(contractPath, mosDir)
	}
	return nil
}

// CheckSeal reads the seal block from a contract if present.
func CheckSeal(root, id string) (*SealInfo, error) {
	contractPath, err := FindContractPath(root, id)
	if err != nil {
		return nil, fmt.Errorf("CheckSeal: %w", err)
	}
	ab, err := dsl.ReadArtifact(contractPath)
	if err != nil {
		return nil, fmt.Errorf("CheckSeal: %w", err)
	}
	return readSeal(ab)
}

// CheckSealForMutation verifies that the contract is not locked by another operator.
// Returns nil if the contract is not sealed, or if the current operator holds the lock.
func CheckSealForMutation(root, id string) error {
	seal, err := CheckSeal(root, id)
	if err != nil || seal == nil {
		return nil
	}

	operator, err := ResolveOperator()
	if err != nil {
		return fmt.Errorf("contract %s is sealed and cannot resolve operator: %w", id, err)
	}

	if operator == seal.Operator {
		return nil
	}

	return fmt.Errorf("contract %s is sealed by %s; unlock it first or use the lock owner's identity", id, seal.Operator)
}

// ResolveAuthority looks up the authority level for an operator from config.mos.
// Returns 0 if no authority block exists or operator is not listed.
func ResolveAuthority(root, operator string) (int, error) {
	configPath := filepath.Join(root, MosDir, ConfigFile)
	data, err := os.ReadFile(configPath)
	if err != nil {
		return 0, nil
	}
	f, err := dsl.Parse(string(data), nil) // config.mos (config file, not artifact): cannot migrate to WithArtifact
	if err != nil {
		return 0, nil
	}
	ab, ok := f.Artifact.(*dsl.ArtifactBlock)
	if !ok {
		return 0, nil
	}

	authBlock := dsl.FindBlock(ab.Items, "authority")
	if authBlock == nil {
		return 0, nil
	}

	roles := make(map[string]int)
	for _, item := range authBlock.Items {
		blk, ok := item.(*dsl.Block)
		if !ok || blk.Name != "role" || blk.Title == "" {
			continue
		}
		if level, ok := dsl.FieldInt(blk.Items, "level"); ok {
			roles[blk.Title] = int(level)
		}
	}

	for _, item := range authBlock.Items {
		blk, ok := item.(*dsl.Block)
		if !ok || blk.Name != "operator" || blk.Title != operator {
			continue
		}
		if role, ok := dsl.FieldString(blk.Items, "role"); ok {
			if level, ok := roles[role]; ok {
				return level, nil
			}
		}
	}
	return 0, nil
}

func readSeal(ab *dsl.ArtifactBlock) (*SealInfo, error) {
	sealBlk := dsl.FindBlock(ab.Items, "seal")
	if sealBlk == nil {
		return nil, nil
	}
	info := &SealInfo{Intent: sealBlk.Title}
	info.Operator, _ = dsl.FieldString(sealBlk.Items, "operator")
	info.Message, _ = dsl.FieldString(sealBlk.Items, "message")
	if f := dsl.FindField(sealBlk.Items, "timestamp"); f != nil {
		if dv, ok := f.Value.(*dsl.DateTimeVal); ok {
			info.Timestamp = dv.Raw
		}
	}
	return info, nil
}
