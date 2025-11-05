package harbor

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	Base     string // full base, e.g. http://localhost:8081 or https://harbor1.local
	User     string
	Pass     string
	Insecure bool
	httpc    *http.Client
}

func normalizeBase(base string, insecure bool) string {
	b := strings.TrimRight(base, "/")
	if strings.HasPrefix(b, "http://") || strings.HasPrefix(b, "https://") {
		return b
	}
	// If no scheme provided:
	if insecure {
		return "http://" + b // mock/local typically http
	}
	return "https://" + b // default secure for real Harbor
}

func New(base, user, pass string, insecure bool) *Client {
	nb := normalizeBase(base, insecure)

	tr := http.DefaultTransport.(*http.Transport).Clone()
	// If user explicitly chooses https with self-signed certs, allow skipping verification
	if strings.HasPrefix(nb, "https://") && insecure {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec
	}

	return &Client{
		Base:     nb,
		User:     user,
		Pass:     pass,
		Insecure: insecure,
		httpc: &http.Client{
			Timeout:   60 * time.Second,
			Transport: tr,
		},
	}
}

type Repo struct {
	Name string `json:"name"`
}

func (c *Client) ListRepos(project string) ([]Repo, error) {
	u := fmt.Sprintf("%s/api/v2.0/projects/%s/repositories?page=1&page_size=5000",
		c.Base, url.PathEscape(project))
	req, _ := http.NewRequest(http.MethodGet, u, nil)
	if c.User != "" || c.Pass != "" {
		req.SetBasicAuth(c.User, c.Pass)
	}
	resp, err := c.httpc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("harbor repos: status %s", resp.Status)
	}
	var repos []Repo
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, err
	}
	return repos, nil
}

type Tag struct {
	Name string `json:"name"`
}

type Artifact struct {
	Digest string `json:"digest"`
	Tags   []Tag  `json:"tags"`
}

func (c *Client) ListArtifacts(project, repo string) ([]Artifact, error) {
	u := fmt.Sprintf("%s/api/v2.0/projects/%s/repositories/%s/artifacts?page=1&page_size=200",
		c.Base, url.PathEscape(project), url.PathEscape(repo))
	req, _ := http.NewRequest(http.MethodGet, u, nil)
	if c.User != "" || c.Pass != "" {
		req.SetBasicAuth(c.User, c.Pass)
	}
	resp, err := c.httpc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("harbor artifacts: status %s", resp.Status)
	}
	var arts []Artifact
	if err := json.NewDecoder(resp.Body).Decode(&arts); err != nil {
		return nil, err
	}
	return arts, nil
}
