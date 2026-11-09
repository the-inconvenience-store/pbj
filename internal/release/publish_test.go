package release_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/the-inconvenience-store/pbj/internal/release"
)

func TestPublicationConvergesWithoutDuplicateTapCommits(t *testing.T) {
	source := filepath.Join(t.TempDir(), "pbj_1.2.3_source.tar.gz")
	if err := os.WriteFile(source, []byte("source archive fixture"), 0o600); err != nil {
		t.Fatal(err)
	}
	prepared, err := release.Prepare("v1.2.3", source)
	if err != nil {
		t.Fatal(err)
	}
	for name, initialSHA := range map[string]string{"create": "", "update": "original"} {
		t.Run(name, func(t *testing.T) { assertPublicationConverges(t, prepared, initialSHA) })
	}
}

func assertPublicationConverges(t *testing.T, prepared *release.Release, initialSHA string) {
	t.Helper()
	store := &tapFixture{t: t, sha: initialSHA, content: []byte("class Pbj < Formula\n  version \"1.0.0\"\nend\n")}
	server := httptest.NewServer(store)
	defer server.Close()
	publisher := release.Publisher{Client: server.Client(), Endpoint: server.URL, Token: "fixture-token"}
	changed, err := publisher.Publish(context.Background(), prepared)
	if err != nil || !changed {
		t.Fatalf("first publication changed=%t err=%v", changed, err)
	}
	changed, err = publisher.Publish(context.Background(), prepared)
	if err != nil || changed {
		t.Fatalf("repeated publication changed=%t err=%v", changed, err)
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	if store.commits != 1 {
		t.Fatalf("same release produced %d tap commits, want one", store.commits)
	}
}

func TestPublicationDoesNotRollBackNewerTapRelease(t *testing.T) {
	source := filepath.Join(t.TempDir(), "pbj_1.2.3_source.tar.gz")
	if err := os.WriteFile(source, []byte("source archive fixture"), 0o600); err != nil {
		t.Fatal(err)
	}
	prepared, err := release.Prepare("v1.2.3", source)
	if err != nil {
		t.Fatal(err)
	}
	store := &tapFixture{t: t, sha: "newer", content: []byte("class Pbj < Formula\n  version \"2.0.0\"\nend\n")}
	server := httptest.NewServer(store)
	defer server.Close()
	publisher := release.Publisher{Client: server.Client(), Endpoint: server.URL, Token: "fixture-token"}
	changed, err := publisher.Publish(context.Background(), prepared)
	if err == nil || changed {
		t.Fatalf("older release replaced the current tap: changed=%t err=%v", changed, err)
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	if store.commits != 0 {
		t.Fatal("rejected downgrade modified the tap")
	}
}

type tapFixture struct {
	t       *testing.T
	mu      sync.Mutex
	content []byte
	sha     string
	commits int
}

func (s *tapFixture) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	switch r.Method {
	case http.MethodGet:
		if s.sha == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err := json.NewEncoder(w).Encode(map[string]string{"sha": s.sha, "content": base64.StdEncoding.EncodeToString(s.content)}); err != nil {
			s.t.Error(err)
		}
	case http.MethodPut:
		s.update(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *tapFixture) update(w http.ResponseWriter, r *http.Request) {
	var update struct {
		SHA     string `json:"sha"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		s.t.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if update.SHA != s.sha {
		w.WriteHeader(http.StatusConflict)
		return
	}
	decoded, err := base64.StdEncoding.DecodeString(update.Content)
	if err != nil {
		s.t.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	created := s.sha == ""
	s.content = decoded
	s.sha = "updated"
	s.commits++
	if created {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}
