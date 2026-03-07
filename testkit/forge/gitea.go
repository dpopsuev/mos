package forge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func init() {
	// t.Cleanup handles container teardown; Ryuk is redundant and fails on
	// some Docker setups (rootless, cgroup v2, etc.).
	if os.Getenv("TESTCONTAINERS_RYUK_DISABLED") == "" {
		os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
	}
}

const (
	giteaImage    = "gitea/gitea:latest"
	giteaUser     = "mosadmin"
	giteaPassword = "mosadmin1234"
	giteaEmail    = "admin@mos.dev"
)

type giteaForge struct {
	t       testing.TB
	baseURL string
	token   string
	repos   map[string]string
}

// Gitea creates a Forge backed by a real Gitea container.
// Requires Docker. Skips the test if Docker is unavailable.
func Gitea(t testing.TB) Forge {
	t.Helper()

	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        giteaImage,
		ExposedPorts: []string{"3000/tcp"},
		Env: map[string]string{
			"GITEA__security__INSTALL_LOCK": "true",
			"GITEA__server__DISABLE_SSH":    "true",
			"USER_UID":                      "1000",
			"USER_GID":                      "1000",
		},
		WaitingFor: wait.ForHTTP("/api/v1/version").
			WithPort("3000/tcp").
			WithStartupTimeout(90 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Skipf("could not start Gitea container (Docker unavailable?): %v", err)
	}
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Logf("terminate gitea container: %v", err)
		}
	})

	mappedPort, err := container.MappedPort(ctx, "3000/tcp")
	if err != nil {
		t.Fatalf("mapped port: %v", err)
	}
	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("container host: %v", err)
	}

	baseURL := fmt.Sprintf("http://%s:%s", host, mappedPort.Port())

	createAdmin(t, container, ctx)

	token := createAPIToken(t, baseURL)

	f := &giteaForge{
		t:       t,
		baseURL: baseURL,
		token:   token,
		repos:   make(map[string]string),
	}
	return f
}

func createAdmin(t testing.TB, container testcontainers.Container, ctx context.Context) {
	t.Helper()
	code, out, err := container.Exec(ctx, []string{
		"su", "git", "-c",
		fmt.Sprintf("gitea admin user create --admin --username %s --password %s --email %s",
			giteaUser, giteaPassword, giteaEmail),
	})
	if err != nil {
		t.Fatalf("exec gitea admin create: %v", err)
	}
	if code != 0 {
		body, _ := io.ReadAll(out)
		t.Fatalf("gitea admin create exited %d: %s", code, body)
	}
}

func createAPIToken(t testing.TB, baseURL string) string {
	t.Helper()
	payload := map[string]interface{}{
		"name":   "test-token",
		"scopes": []string{"all"},
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", baseURL+"/api/v1/users/"+giteaUser+"/tokens", bytes.NewReader(body))
	req.SetBasicAuth(giteaUser, giteaPassword)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("create token request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("create token: status %d: %s", resp.StatusCode, respBody)
	}

	var result struct {
		Sha1 string `json:"sha1"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode token response: %v", err)
	}
	return result.Sha1
}

func (f *giteaForge) CreateRepo(name string) (string, error) {
	payload := map[string]interface{}{
		"name":           name,
		"auto_init":      true,
		"default_branch": "main",
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", f.baseURL+"/api/v1/user/repos", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "token "+f.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("create repo request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("create repo: status %d: %s", resp.StatusCode, respBody)
	}

	// Return URL with embedded credentials so clone and push both work.
	hostPort := f.baseURL[len("http://"):]
	cloneURL := fmt.Sprintf("http://%s:%s@%s/%s/%s.git",
		giteaUser, giteaPassword, hostPort, giteaUser, name)
	f.repos[name] = cloneURL
	return cloneURL, nil
}

func (f *giteaForge) RepoURL(name string) string {
	hostPort := f.baseURL[len("http://"):]
	return fmt.Sprintf("http://%s:%s@%s/%s/%s.git",
		giteaUser, giteaPassword, hostPort, giteaUser, name)
}

func (f *giteaForge) Close() error {
	return nil
}
