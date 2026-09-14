// Package api wires HTTP routes to LDAPService operations.
package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/jasonmiller-cc/parameters-core/pkg/response"
	"github.com/jasonmiller-cc/parameters-ldap/internal/service"
)

// Handler owns all HTTP route registrations for the LDAP API.
type Handler struct {
	svc *service.LDAPService
}

// New creates a Handler backed by svc.
func New(svc *service.LDAPService) *Handler {
	return &Handler{svc: svc}
}

// Register mounts all API routes onto mux.
func (h *Handler) Register(mux *http.ServeMux) {
	// Users
	mux.HandleFunc("GET /api/v1/users", h.listUsers)
	mux.HandleFunc("GET /api/v1/users/{uid}", h.getUser)
	mux.HandleFunc("POST /api/v1/users", h.createUser)
	mux.HandleFunc("PUT /api/v1/users/{uid}", h.updateUser)
	mux.HandleFunc("DELETE /api/v1/users/{uid}", h.deleteUser)

	// Groups
	mux.HandleFunc("GET /api/v1/groups", h.listGroups)
	mux.HandleFunc("GET /api/v1/groups/{cn}", h.getGroup)
	mux.HandleFunc("POST /api/v1/groups", h.createGroup)
	mux.HandleFunc("POST /api/v1/groups/{cn}/members", h.addGroupMember)
	mux.HandleFunc("DELETE /api/v1/groups/{cn}/members/{uid}", h.removeGroupMember)

	// Auth
	mux.HandleFunc("POST /api/v1/auth/bind", h.testBind)
}

// ---- users ----------------------------------------------------------------

func (h *Handler) listUsers(w http.ResponseWriter, r *http.Request) {
	opts := service.ListOptions{
		Search: r.URL.Query().Get("search"),
	}
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			opts.Limit = n
		}
	}
	if opts.Limit == 0 {
		opts.Limit = 100
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			opts.Offset = n
		}
	}

	users, total, err := h.svc.ListUsers(r.Context(), opts)
	if err != nil {
		response.Err(w, err)
		return
	}
	response.List(w, users, total, opts.Limit, opts.Offset)
}

func (h *Handler) getUser(w http.ResponseWriter, r *http.Request) {
	uid := pathSegment(r, "uid")
	user, err := h.svc.GetUser(r.Context(), uid)
	if err != nil {
		response.Err(w, err)
		return
	}
	response.OK(w, user)
}

func (h *Handler) createUser(w http.ResponseWriter, r *http.Request) {
	var req service.CreateUserRequest
	if err := decodeJSON(r, &req); err != nil {
		response.ErrStatus(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	user, err := h.svc.CreateUser(r.Context(), req)
	if err != nil {
		response.Err(w, err)
		return
	}
	response.Created(w, user)
}

func (h *Handler) updateUser(w http.ResponseWriter, r *http.Request) {
	uid := pathSegment(r, "uid")
	var req service.UpdateUserRequest
	if err := decodeJSON(r, &req); err != nil {
		response.ErrStatus(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	user, err := h.svc.UpdateUser(r.Context(), uid, req)
	if err != nil {
		response.Err(w, err)
		return
	}
	response.OK(w, user)
}

func (h *Handler) deleteUser(w http.ResponseWriter, r *http.Request) {
	uid := pathSegment(r, "uid")
	if err := h.svc.DeleteUser(r.Context(), uid); err != nil {
		response.Err(w, err)
		return
	}
	response.NoContent(w)
}

// ---- groups ---------------------------------------------------------------

func (h *Handler) listGroups(w http.ResponseWriter, r *http.Request) {
	groups, err := h.svc.ListGroups(r.Context())
	if err != nil {
		response.Err(w, err)
		return
	}
	response.List(w, groups, len(groups), len(groups), 0)
}

func (h *Handler) getGroup(w http.ResponseWriter, r *http.Request) {
	cn := pathSegment(r, "cn")
	group, err := h.svc.GetGroup(r.Context(), cn)
	if err != nil {
		response.Err(w, err)
		return
	}
	response.OK(w, group)
}

func (h *Handler) createGroup(w http.ResponseWriter, r *http.Request) {
	var req service.CreateGroupRequest
	if err := decodeJSON(r, &req); err != nil {
		response.ErrStatus(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	group, err := h.svc.CreateGroup(r.Context(), req)
	if err != nil {
		response.Err(w, err)
		return
	}
	response.Created(w, group)
}

func (h *Handler) addGroupMember(w http.ResponseWriter, r *http.Request) {
	cn := pathSegment(r, "cn")
	var body struct {
		UID string `json:"uid"`
	}
	if err := decodeJSON(r, &body); err != nil {
		response.ErrStatus(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	if err := h.svc.AddGroupMember(r.Context(), cn, body.UID); err != nil {
		response.Err(w, err)
		return
	}
	response.NoContent(w)
}

func (h *Handler) removeGroupMember(w http.ResponseWriter, r *http.Request) {
	cn := pathSegment(r, "cn")
	uid := pathSegment(r, "uid")
	if err := h.svc.RemoveGroupMember(r.Context(), cn, uid); err != nil {
		response.Err(w, err)
		return
	}
	response.NoContent(w)
}

// ---- auth -----------------------------------------------------------------

func (h *Handler) testBind(w http.ResponseWriter, r *http.Request) {
	var body struct {
		DN       string `json:"dn"`
		Password string `json:"password"`
	}
	if err := decodeJSON(r, &body); err != nil {
		response.ErrStatus(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	if err := h.svc.TestBind(r.Context(), body.DN, body.Password); err != nil {
		response.ErrStatus(w, http.StatusUnauthorized, "bind_failed", err.Error())
		return
	}
	response.OK(w, map[string]string{"status": "ok"})
}

// ---- helpers --------------------------------------------------------------

// pathSegment extracts a named wildcard from an http.ServeMux Go 1.22 pattern.
func pathSegment(r *http.Request, name string) string {
	// Go 1.22+ net/http supports r.PathValue for named wildcards.
	v := r.PathValue(name)
	if v != "" {
		return v
	}
	// Fallback: extract last segment of the path for simple cases.
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

// decodeJSON decodes a JSON request body (up to 1 MiB) into dst.
func decodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}
