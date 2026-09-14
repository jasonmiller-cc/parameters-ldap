// Package service provides LDAP directory operations for users and groups.
// The stub implementations return example data; replace the dial/search calls
// with real go-ldap operations to connect to FreeIPA / OpenLDAP in production.
package service

import (
	"context"
	"crypto/tls"
	"fmt"

	ldap "github.com/go-ldap/ldap/v3"

	"github.com/jasonmiller-cc/parameters-ldap/internal/config"
)

// User represents a directory user entry.
type User struct {
	DN       string   `json:"dn"`
	UID      string   `json:"uid"`
	CN       string   `json:"cn"`
	SN       string   `json:"sn"`
	Mail     string   `json:"mail"`
	MemberOf []string `json:"member_of"`
}

// Group represents a directory group entry.
type Group struct {
	DN          string   `json:"dn"`
	CN          string   `json:"cn"`
	Description string   `json:"description"`
	Members     []string `json:"members"`
}

// CreateUserRequest is the body for POST /api/v1/users.
type CreateUserRequest struct {
	UID      string `json:"uid"`
	CN       string `json:"cn"`
	SN       string `json:"sn"`
	Mail     string `json:"mail"`
	Password string `json:"password"`
}

// UpdateUserRequest is the body for PUT /api/v1/users/{uid}.
type UpdateUserRequest struct {
	CN   string `json:"cn,omitempty"`
	SN   string `json:"sn,omitempty"`
	Mail string `json:"mail,omitempty"`
}

// CreateGroupRequest is the body for POST /api/v1/groups.
type CreateGroupRequest struct {
	CN          string `json:"cn"`
	Description string `json:"description"`
}

// ListOptions carries pagination and search parameters.
type ListOptions struct {
	Search string
	Limit  int
	Offset int
}

// LDAPService wraps LDAP directory operations.
type LDAPService struct {
	cfg *config.Config
}

// New creates an LDAPService from the given config.
func New(cfg *config.Config) *LDAPService {
	return &LDAPService{cfg: cfg}
}

// dial opens and binds an LDAP connection using the service config.
// Callers are responsible for calling conn.Close().
func (s *LDAPService) dial() (*ldap.Conn, error) {
	tlsCfg := &tls.Config{
		InsecureSkipVerify: s.cfg.LDAP.TLSSkipVerify, //nolint:gosec // intentional per config
	}
	conn, err := ldap.DialURL(s.cfg.LDAP.URL, ldap.DialWithTLSConfig(tlsCfg))
	if err != nil {
		return nil, fmt.Errorf("ldap dial %s: %w", s.cfg.LDAP.URL, err)
	}
	if s.cfg.LDAP.BindDN != "" {
		if err := conn.Bind(s.cfg.LDAP.BindDN, s.cfg.LDAP.BindPassword); err != nil {
			_ = conn.Close()
			return nil, fmt.Errorf("ldap bind: %w", err)
		}
	}
	return conn, nil
}

// Ping verifies the LDAP connection is alive.
// Stub: skips dial when URL is not configured.
func (s *LDAPService) Ping(_ context.Context) error {
	if s.cfg.LDAP.URL == "" {
		return nil // not configured; treat as healthy for local dev
	}
	conn, err := s.dial()
	if err != nil {
		return err
	}
	_ = conn.Close()
	return nil
}

// ListUsers returns a page of users matching opts.Search.
func (s *LDAPService) ListUsers(_ context.Context, opts ListOptions) ([]User, int, error) {
	users := []User{
		{
			DN:       fmt.Sprintf("uid=jdoe,%s", s.cfg.LDAP.UserBaseDN),
			UID:      "jdoe",
			CN:       "John Doe",
			SN:       "Doe",
			Mail:     "jdoe@example.com",
			MemberOf: []string{"cn=engineers,cn=groups,dc=example,dc=com"},
		},
	}
	return users, len(users), nil
}

// GetUser returns the user with the given uid.
func (s *LDAPService) GetUser(_ context.Context, uid string) (*User, error) {
	if uid == "" {
		return nil, fmt.Errorf("uid is required")
	}
	return &User{
		DN:       fmt.Sprintf("uid=%s,%s", uid, s.cfg.LDAP.UserBaseDN),
		UID:      uid,
		CN:       "Example User",
		SN:       "User",
		Mail:     uid + "@example.com",
		MemberOf: []string{},
	}, nil
}

// CreateUser adds a new user to the directory.
func (s *LDAPService) CreateUser(_ context.Context, req CreateUserRequest) (*User, error) {
	if req.UID == "" {
		return nil, fmt.Errorf("uid is required")
	}
	return &User{
		DN:       fmt.Sprintf("uid=%s,%s", req.UID, s.cfg.LDAP.UserBaseDN),
		UID:      req.UID,
		CN:       req.CN,
		SN:       req.SN,
		Mail:     req.Mail,
		MemberOf: []string{},
	}, nil
}

// UpdateUser modifies attributes on an existing user.
func (s *LDAPService) UpdateUser(_ context.Context, uid string, req UpdateUserRequest) (*User, error) {
	if uid == "" {
		return nil, fmt.Errorf("uid is required")
	}
	return &User{
		DN:       fmt.Sprintf("uid=%s,%s", uid, s.cfg.LDAP.UserBaseDN),
		UID:      uid,
		CN:       req.CN,
		SN:       req.SN,
		Mail:     req.Mail,
		MemberOf: []string{},
	}, nil
}

// DeleteUser removes a user from the directory.
func (s *LDAPService) DeleteUser(_ context.Context, uid string) error {
	if uid == "" {
		return fmt.Errorf("uid is required")
	}
	return nil
}

// ListGroups returns all groups.
func (s *LDAPService) ListGroups(_ context.Context) ([]Group, error) {
	groups := []Group{
		{
			DN:          fmt.Sprintf("cn=engineers,%s", s.cfg.LDAP.GroupBaseDN),
			CN:          "engineers",
			Description: "Engineering team",
			Members:     []string{"uid=jdoe," + s.cfg.LDAP.UserBaseDN},
		},
	}
	return groups, nil
}

// GetGroup returns the group with the given cn.
func (s *LDAPService) GetGroup(_ context.Context, cn string) (*Group, error) {
	if cn == "" {
		return nil, fmt.Errorf("cn is required")
	}
	return &Group{
		DN:          fmt.Sprintf("cn=%s,%s", cn, s.cfg.LDAP.GroupBaseDN),
		CN:          cn,
		Description: "Example group",
		Members:     []string{},
	}, nil
}

// CreateGroup adds a new group to the directory.
func (s *LDAPService) CreateGroup(_ context.Context, req CreateGroupRequest) (*Group, error) {
	if req.CN == "" {
		return nil, fmt.Errorf("cn is required")
	}
	return &Group{
		DN:          fmt.Sprintf("cn=%s,%s", req.CN, s.cfg.LDAP.GroupBaseDN),
		CN:          req.CN,
		Description: req.Description,
		Members:     []string{},
	}, nil
}

// AddGroupMember adds uid to the members list of group cn.
func (s *LDAPService) AddGroupMember(_ context.Context, cn, uid string) error {
	if cn == "" || uid == "" {
		return fmt.Errorf("cn and uid are required")
	}
	return nil
}

// RemoveGroupMember removes uid from the members list of group cn.
func (s *LDAPService) RemoveGroupMember(_ context.Context, cn, uid string) error {
	if cn == "" || uid == "" {
		return fmt.Errorf("cn and uid are required")
	}
	return nil
}

// TestBind attempts an LDAP bind with the provided dn and password.
func (s *LDAPService) TestBind(_ context.Context, dn, password string) error {
	if dn == "" || password == "" {
		return fmt.Errorf("dn and password are required")
	}
	// TODO: dial s.cfg.LDAP.URL and attempt a simple bind with dn/password.
	return nil
}
