package release

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// TapEndpoint is the sole production destination for generated formulas.
const TapEndpoint = "https://api.github.com/repos/the-inconvenience-store/homebrew-tap/contents/Formula/pbj.rb"

// Publisher writes one formula through GitHub's contents interface. Client and
// Endpoint are supplied so the external HTTP seam can be exercised independently.
// The caller supplies a bounded Client and credentials; token values are never logged.
type Publisher struct {
	Client   *http.Client
	Endpoint string
	Token    string
}

// Publish converges on this release's generated formula. Identical content is a
// no-op; changed content is updated with the current blob SHA (no forced overwrite).
func (p Publisher) Publish(ctx context.Context, release *Release) (bool, error) {
	if p.Token == "" {
		return false, errors.New("HOMEBREW_TAP_TOKEN must grant Contents read/write on the-inconvenience-store/homebrew-tap")
	}
	current, sha, err := p.contents(ctx)
	if err != nil {
		return false, err
	}
	formula := release.Formula()
	if bytes.Equal(current, formula) {
		return false, nil
	}
	if sha != "" {
		if err := release.rejectDowngrade(current); err != nil {
			return false, err
		}
	}
	identity := commitIdentity{Name: "github-actions[bot]", Email: "41898282+github-actions[bot]@users.noreply.github.com"}
	update := contentsUpdate{Message: "chore: update pbj to " + release.version, Content: base64.StdEncoding.EncodeToString(formula), SHA: sha, Author: identity, Committer: identity}
	body, err := json.Marshal(update)
	if err != nil {
		return false, fmt.Errorf("encode formula update: %w", err)
	}
	if err := p.update(ctx, body); err != nil {
		return false, err
	}
	return true, nil
}

type commitIdentity struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type contentsUpdate struct {
	Message   string         `json:"message"`
	Content   string         `json:"content"`
	SHA       string         `json:"sha,omitempty"`
	Author    commitIdentity `json:"author"`
	Committer commitIdentity `json:"committer"`
}

func (p Publisher) contents(ctx context.Context) (data []byte, sha string, err error) {
	response, err := p.request(ctx, http.MethodGet, nil)
	if err != nil {
		return nil, "", err
	}
	defer func() { err = errors.Join(err, response.Body.Close()) }()
	if response.StatusCode == http.StatusNotFound {
		return nil, "", nil
	}
	if response.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("read tap formula: GitHub returned %s", response.Status)
	}
	var file struct {
		Content string `json:"content"`
		SHA     string `json:"sha"`
	}
	const maxContentsBytes = 1 << 20
	if err := json.NewDecoder(io.LimitReader(response.Body, maxContentsBytes)).Decode(&file); err != nil {
		return nil, "", fmt.Errorf("decode tap formula: %w", err)
	}
	if file.SHA == "" {
		return nil, "", errors.New("GitHub formula response has no blob SHA")
	}
	decoded, err := base64.StdEncoding.DecodeString(file.Content)
	if err != nil {
		return nil, "", fmt.Errorf("decode tap formula bytes: %w", err)
	}
	return decoded, file.SHA, nil
}

func (p Publisher) update(ctx context.Context, body []byte) (err error) {
	response, err := p.request(ctx, http.MethodPut, body)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, response.Body.Close()) }()
	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusCreated {
		return fmt.Errorf("publish tap formula: GitHub returned %s", response.Status)
	}
	return nil
}

func (p Publisher) request(ctx context.Context, method string, body []byte) (*http.Response, error) {
	request, err := http.NewRequestWithContext(ctx, method, p.Endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("prepare GitHub request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+p.Token)
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	request.Header.Set("User-Agent", "pbj-release")
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := p.Client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("GitHub request failed: %w", err)
	}
	return response, nil
}
