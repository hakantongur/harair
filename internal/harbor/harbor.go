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
	Base     string
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
	if insecure {
		return "http://" + b
	}
	return "https://" + b
}

func New(base, user, pass string, insecure bool) *Client {
	nb := normalizeBase(base, insecure)

	tr := http.DefaultTransport.(*http.Transport).Clone()
	// allow self-signed when explicitly marked insecure over https
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

// --- small HTTP helper ---

func (c *Client) getJSON(u string, out any) error {
	req, _ := http.NewRequest(http.MethodGet, u, nil)
	if c.User != "" || c.Pass != "" {
		req.SetBasicAuth(c.User, c.Pass)
	}
	resp, err := c.httpc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("harbor GET %s: status %s", u, resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// --- API shapes we use ---

type Repository struct {
	Name string `json:"name"` // "project/repo" or just "repo"
}

type Tag struct {
	Name string `json:"name"`
}

type Artifact struct {
	Digest string `json:"digest"`
	Tags   []Tag  `json:"tags"`
}

// --- API methods ---

func (c *Client) ListRepos(project string) ([]Repository, error) {
	const pageSize = 100
	page := 1
	var all []Repository

	for {
		var chunk []Repository
		u := fmt.Sprintf("%s/api/v2.0/projects/%s/repositories?page=%d&page_size=%d",
			c.Base, url.PathEscape(project), page, pageSize)

		if err := c.getJSON(u, &chunk); err != nil {
			return nil, err
		}
		all = append(all, chunk...)
		if len(chunk) < pageSize {
			break
		}
		page++
	}
	return all, nil
}

func (c *Client) ListArtifacts(project, repo string) ([]Artifact, error) {
	const pageSize = 100
	page := 1
	var all []Artifact

	for {
		var chunk []Artifact
		u := fmt.Sprintf("%s/api/v2.0/projects/%s/repositories/%s/artifacts?page=%d&page_size=%d&with_tag=true",
			c.Base, url.PathEscape(project), url.PathEscape(repo), page, pageSize)

		if err := c.getJSON(u, &chunk); err != nil {
			return nil, err
		}
		all = append(all, chunk...)
		if len(chunk) < pageSize {
			break
		}
		page++
	}
	return all, nil
}
