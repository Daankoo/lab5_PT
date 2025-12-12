package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	db "github.com/Daankoo/lab5_PT/db/sqlc"
)

type mockStore struct {
	users  map[int32]db.User
	nextID int32
}

func newMockStore() *mockStore {
	return &mockStore{
		users:  make(map[int32]db.User),
		nextID: 1,
	}
}

func (m *mockStore) CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error) {
	u := db.User{
		ID:    m.nextID,
		Name:  arg.Name,
		Email: arg.Email,
		Age:   arg.Age,
	}
	m.users[m.nextID] = u
	m.nextID++
	return u, nil
}

func (m *mockStore) ListUsers(ctx context.Context) ([]db.User, error) {
	result := make([]db.User, 0, len(m.users))
	for _, u := range m.users {
		result = append(result, u)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result, nil
}

func (m *mockStore) GetUser(ctx context.Context, id int32) (db.User, error) {
	u, ok := m.users[id]
	if !ok {
		return db.User{}, sql.ErrNoRows
	}
	return u, nil
}

func (m *mockStore) UpdateUser(ctx context.Context, arg db.UpdateUserParams) (db.User, error) {
	_, ok := m.users[arg.ID]
	if !ok {
		return db.User{}, sql.ErrNoRows
	}
	u := db.User{
		ID:    arg.ID,
		Name:  arg.Name,
		Email: arg.Email,
		Age:   arg.Age,
	}
	m.users[arg.ID] = u
	return u, nil
}

func (m *mockStore) DeleteUser(ctx context.Context, id int32) error {
	delete(m.users, id)
	return nil
}

func newTestServer() (*Server, *mockStore) {
	ms := newMockStore()
	s := NewServer(ms)
	return s, ms
}

func TestHandleCreateUser_Valid(t *testing.T) {
	srv, store := newTestServer()

	body := `{"name":"Alice","email":"alice@example.com","age":22}`
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
	w := httptest.NewRecorder()

	srv.handleCreateUser(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	var u db.User
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if u.Name != "Alice" || u.Email != "alice@example.com" || u.Age != 22 {
		t.Fatalf("unexpected user in response: %+v", u)
	}

	if len(store.users) != 1 {
		t.Fatalf("expected 1 user in store, got %d", len(store.users))
	}
}

func TestHandleCreateUser_InvalidData(t *testing.T) {
	srv, store := newTestServer()

	body := `{"name":"","email":"bad@example.com","age":10}`
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
	w := httptest.NewRecorder()

	srv.handleCreateUser(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}

	if len(store.users) != 0 {
		t.Fatalf("expected 0 users in store, got %d", len(store.users))
	}
}

func TestHandleListUsers(t *testing.T) {
	srv, store := newTestServer()

	_, _ = store.CreateUser(context.Background(), db.CreateUserParams{
		Name:  "Alice",
		Email: "alice@example.com",
		Age:   22,
	})
	_, _ = store.CreateUser(context.Background(), db.CreateUserParams{
		Name:  "Bob",
		Email: "bob@example.com",
		Age:   30,
	})

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	w := httptest.NewRecorder()

	srv.handleListUsers(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var users []db.User
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}

	if users[0].Name != "Alice" || users[1].Name != "Bob" {
		t.Fatalf("unexpected users order/content: %+v", users)
	}
}

func TestHandleGetUser_Found(t *testing.T) {
	srv, store := newTestServer()

	u, _ := store.CreateUser(context.Background(), db.CreateUserParams{
		Name:  "Alice",
		Email: "alice@example.com",
		Age:   22,
	})

	req := httptest.NewRequest(http.MethodGet, "/users/"+strconv.Itoa(int(u.ID)), nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", strconv.Itoa(int(u.ID)))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	srv.handleGetUser(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var got db.User
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if got.ID != u.ID || got.Name != u.Name {
		t.Fatalf("unexpected user: %+v", got)
	}
}

func TestHandleGetUser_NotFound(t *testing.T) {
	srv, _ := newTestServer()

	req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	srv.handleGetUser(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
}

func TestHandleUpdateUser_Valid(t *testing.T) {
	srv, store := newTestServer()

	u, _ := store.CreateUser(context.Background(), db.CreateUserParams{
		Name:  "Alice",
		Email: "alice@example.com",
		Age:   22,
	})

	body := `{"name":"Alice Updated","email":"alice2@example.com","age":25}`
	req := httptest.NewRequest(http.MethodPut, "/users/"+strconv.Itoa(int(u.ID)), strings.NewReader(body))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", strconv.Itoa(int(u.ID)))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	srv.handleUpdateUser(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var got db.User
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if got.Name != "Alice Updated" || got.Email != "alice2@example.com" || got.Age != 25 {
		t.Fatalf("unexpected updated user: %+v", got)
	}

	stored := store.users[u.ID]
	if stored.Name != "Alice Updated" {
		t.Fatalf("store not updated, got: %+v", stored)
	}
}

func TestHandleUpdateUser_NotFound(t *testing.T) {
	srv, _ := newTestServer()

	body := `{"name":"X","email":"x@example.com","age":30}`
	req := httptest.NewRequest(http.MethodPut, "/users/999", strings.NewReader(body))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "999")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	srv.handleUpdateUser(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
}

func TestHandleDeleteUser(t *testing.T) {
	srv, store := newTestServer()

	u, _ := store.CreateUser(context.Background(), db.CreateUserParams{
		Name:  "Alice",
		Email: "alice@example.com",
		Age:   22,
	})

	req := httptest.NewRequest(http.MethodDelete, "/users/"+strconv.Itoa(int(u.ID)), nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", strconv.Itoa(int(u.ID)))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	srv.handleDeleteUser(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, resp.StatusCode)
	}

	if len(store.users) != 0 {
		t.Fatalf("expected 0 users in store after delete, got %d", len(store.users))
	}
}

