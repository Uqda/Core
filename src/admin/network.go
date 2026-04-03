package admin

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/hjson/hjson-go/v4"

	"github.com/Uqda/Core/src/config"
	"github.com/Uqda/Core/src/core"
)

// CreateNetworkRequest creates a private invite group and returns an encoded token.
type CreateNetworkRequest struct {
	Name         string   `json:"name"`
	Peers        []string `json:"peers"`
	ExpiresHours int      `json:"expiresHours"`
	Admin        string   `json:"admin,omitempty"`
}

// CreateNetworkResponse is returned by createNetwork.
type CreateNetworkResponse struct {
	Token   string `json:"token"`
	Network string `json:"network"`
	Warning string `json:"warning,omitempty"`
}

// JoinNetworkRequest applies an invite token on this node.
type JoinNetworkRequest struct {
	Token            string `json:"token"`
	OwnerAdminAuth   string `json:"ownerAdminAuth,omitempty"`
}

// JoinNetworkResponse is returned by joinNetwork.
type JoinNetworkResponse struct {
	Network    string `json:"network"`
	Registered bool   `json:"registered"`
	Message    string `json:"message,omitempty"`
}

// ListNetworksResponse lists configured private networks.
type ListNetworksResponse struct {
	Networks []config.PrivateNetwork `json:"networks"`
}

// LeaveNetworkRequest removes a private network by name.
type LeaveNetworkRequest struct {
	Name string `json:"name"`
}

// LeaveNetworkResponse confirms leaveNetwork.
type LeaveNetworkResponse struct {
	OK   bool   `json:"ok"`
	Name string `json:"name"`
}

// InviteRegisterRequest is sent to the network owner to add a member key.
type InviteRegisterRequest struct {
	Token    string `json:"token"`
	GuestKey string `json:"guestKey"`
}

// InviteRegisterResponse confirms inviteRegister.
type InviteRegisterResponse struct {
	OK       bool   `json:"ok"`
	Network  string `json:"network"`
	GuestKey string `json:"guestKey"`
}

// SetupNetworkHandlers registers private-network admin commands.
func (a *AdminSocket) SetupNetworkHandlers() {
	_ = a.AddHandler(
		"createNetwork", "Create a private invite network and return a shareable token", []string{"name", "peers", "expiresHours", "admin"},
		func(in json.RawMessage) (interface{}, error) {
			req := &CreateNetworkRequest{}
			if err := json.Unmarshal(in, &req); err != nil {
				return nil, err
			}
			return a.createNetworkHandler(req)
		},
	)
	_ = a.AddHandler(
		"joinNetwork", "Join a private network using an invite token", []string{"token", "ownerAdminAuth"},
		func(in json.RawMessage) (interface{}, error) {
			req := &JoinNetworkRequest{}
			if err := json.Unmarshal(in, &req); err != nil {
				return nil, err
			}
			return a.joinNetworkHandler(req)
		},
	)
	_ = a.AddHandler(
		"listNetworks", "List configured private networks", []string{},
		func(in json.RawMessage) (interface{}, error) {
			return a.listNetworksHandler()
		},
	)
	_ = a.AddHandler(
		"leaveNetwork", "Remove a private network by name", []string{"name"},
		func(in json.RawMessage) (interface{}, error) {
			req := &LeaveNetworkRequest{}
			if err := json.Unmarshal(in, &req); err != nil {
				return nil, err
			}
			return a.leaveNetworkHandler(req)
		},
	)
	_ = a.AddHandler(
		"inviteRegister", "As network owner, register a guest public key from an invite token", []string{"token", "guestKey"},
		func(in json.RawMessage) (interface{}, error) {
			req := &InviteRegisterRequest{}
			if err := json.Unmarshal(in, &req); err != nil {
				return nil, err
			}
			return a.inviteRegisterHandler(req)
		},
	)
}

func (a *AdminSocket) loadConfigFile() (*config.NodeConfig, error) {
	if a.configFilePath == "" {
		return nil, fmt.Errorf("configuration file path is not set (start uqda with -useconffile)")
	}
	f, err := os.Open(a.configFilePath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	cfg := config.GenerateConfig()
	if _, err := cfg.ReadFrom(f); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (a *AdminSocket) saveConfigFile(cfg *config.NodeConfig) error {
	if a.configFilePath == "" {
		return fmt.Errorf("configuration file path is not set")
	}
	bs, err := hjson.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(a.configFilePath, bs, 0644)
}

func (a *AdminSocket) createNetworkHandler(req *CreateNetworkRequest) (*CreateNetworkResponse, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if len(req.Peers) == 0 {
		return nil, fmt.Errorf("peers is required (at least one bootstrap URI)")
	}
	hours := req.ExpiresHours
	if hours <= 0 {
		hours = 24
	}
	cfg, err := a.loadConfigFile()
	if err != nil {
		return nil, err
	}
	for _, pn := range cfg.PrivateNetworks {
		if pn.Name == name {
			return nil, fmt.Errorf("private network %q already exists", name)
		}
	}
	pass, err := randomHex(32)
	if err != nil {
		return nil, err
	}
	ownerHex := hex.EncodeToString(a.core.PublicKey())
	expires := time.Now().Add(time.Duration(hours) * time.Hour).Unix()
	tok := &core.InviteToken{
		V:        1,
		Net:      name,
		Peers:    append([]string(nil), req.Peers...),
		Password: pass,
		OwnerKey: ownerHex,
		Expires:  expires,
		Admin:    strings.TrimSpace(req.Admin),
	}
	tokenStr, err := core.EncodeInviteToken(tok)
	if err != nil {
		return nil, err
	}
	pn := config.PrivateNetwork{
		Name:        name,
		Password:    pass,
		Peers:       append([]string(nil), req.Peers...),
		AllowedKeys: []string{ownerHex},
		CreatedAt:   time.Now().Unix(),
		IsOwner:     true,
	}
	cfg.PrivateNetworks = append(cfg.PrivateNetworks, pn)
	for i := range cfg.MulticastInterfaces {
		cfg.MulticastInterfaces[i].Password = pass
	}
	if err := a.saveConfigFile(cfg); err != nil {
		return nil, err
	}
	// Owner trusts self via member list; add runtime keys from saved config
	a.reapplyPrivateNetworkKeys(cfg)
	for _, p := range req.Peers {
		u, err := url.Parse(p)
		if err != nil {
			return nil, fmt.Errorf("peer URI: %w", err)
		}
		if err := a.core.AddPeer(u, ""); err != nil {
			a.log.Warnf("createNetwork: addPeer %s: %v", p, err)
		}
	}
	return &CreateNetworkResponse{
		Token:   tokenStr,
		Network: name,
		Warning: "Restart the uqda process for multicast password changes to take effect.",
	}, nil
}

func (a *AdminSocket) joinNetworkHandler(req *JoinNetworkRequest) (*JoinNetworkResponse, error) {
	t, err := core.DecodeInviteToken(req.Token)
	if err != nil {
		return nil, err
	}
	if err := core.ValidateInviteToken(t); err != nil {
		return nil, err
	}
	guestHex := hex.EncodeToString(a.core.PublicKey())
	ownerPub, err := parseHexPub(t.OwnerKey)
	if err != nil {
		return nil, err
	}
	guestPub := a.core.PublicKey()

	cfg, err := a.loadConfigFile()
	if err != nil {
		return nil, err
	}
	for _, pn := range cfg.PrivateNetworks {
		if pn.Name == t.Net {
			return nil, fmt.Errorf("already joined or defined private network %q", t.Net)
		}
	}
	pn := config.PrivateNetwork{
		Name:        t.Net,
		Password:    t.Password,
		Peers:       append([]string(nil), t.Peers...),
		AllowedKeys: []string{t.OwnerKey, guestHex},
		CreatedAt:   time.Now().Unix(),
		IsOwner:     false,
	}
	cfg.PrivateNetworks = append(cfg.PrivateNetworks, pn)
	for i := range cfg.MulticastInterfaces {
		cfg.MulticastInterfaces[i].Password = t.Password
	}
	if err := a.saveConfigFile(cfg); err != nil {
		return nil, err
	}
	a.core.AddPrivateNetworkAllowedKey(ownerPub)
	a.core.AddPrivateNetworkAllowedKey(guestPub)
	a.reapplyPrivateNetworkKeys(cfg)

	for _, p := range t.Peers {
		u, err := url.Parse(p)
		if err != nil {
			return nil, fmt.Errorf("peer URI: %w", err)
		}
		if err := a.core.AddPeer(u, ""); err != nil {
			a.log.Warnf("joinNetwork: addPeer %s: %v", p, err)
		}
	}

	registered := false
	msg := ""
	if ep := strings.TrimSpace(t.Admin); ep != "" {
		if err := dialInviteRegister(ep, req.OwnerAdminAuth, req.Token, guestHex); err != nil {
			msg = "Joined locally; could not notify owner: " + err.Error()
		} else {
			registered = true
		}
	} else {
		msg = "Joined locally; token had no admin endpoint — add your public key on the owner if needed: " + guestHex
	}

	return &JoinNetworkResponse{
		Network:    t.Net,
		Registered: registered,
		Message:    msg,
	}, nil
}

func (a *AdminSocket) listNetworksHandler() (*ListNetworksResponse, error) {
	cfg, err := a.loadConfigFile()
	if err != nil {
		return nil, err
	}
	return &ListNetworksResponse{Networks: append([]config.PrivateNetwork(nil), cfg.PrivateNetworks...)}, nil
}

func (a *AdminSocket) leaveNetworkHandler(req *LeaveNetworkRequest) (*LeaveNetworkResponse, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	cfg, err := a.loadConfigFile()
	if err != nil {
		return nil, err
	}
	var out []config.PrivateNetwork
	found := false
	for _, pn := range cfg.PrivateNetworks {
		if pn.Name == name {
			found = true
			continue
		}
		out = append(out, pn)
	}
	if !found {
		return nil, fmt.Errorf("private network %q not found", name)
	}
	cfg.PrivateNetworks = out
	if err := a.saveConfigFile(cfg); err != nil {
		return nil, err
	}
	a.reapplyPrivateNetworkKeys(cfg)
	return &LeaveNetworkResponse{OK: true, Name: name}, nil
}

func (a *AdminSocket) inviteRegisterHandler(req *InviteRegisterRequest) (*InviteRegisterResponse, error) {
	t, err := core.DecodeInviteToken(req.Token)
	if err != nil {
		return nil, err
	}
	if err := core.ValidateInviteOwner(t, a.core.PublicKey()); err != nil {
		return nil, err
	}
	guest, err := parseHexPub(req.GuestKey)
	if err != nil {
		return nil, fmt.Errorf("guestKey: %w", err)
	}
	cfg, err := a.loadConfigFile()
	if err != nil {
		return nil, err
	}
	guestHex := hex.EncodeToString(guest)
	foundNet := false
	for i := range cfg.PrivateNetworks {
		if cfg.PrivateNetworks[i].Name != t.Net {
			continue
		}
		foundNet = true
		dup := false
		for _, ex := range cfg.PrivateNetworks[i].AllowedKeys {
			if strings.EqualFold(ex, guestHex) {
				dup = true
				break
			}
		}
		if !dup {
			cfg.PrivateNetworks[i].AllowedKeys = append(cfg.PrivateNetworks[i].AllowedKeys, guestHex)
		}
		break
	}
	if !foundNet {
		return nil, fmt.Errorf("private network %q not found on this node", t.Net)
	}
	if err := a.saveConfigFile(cfg); err != nil {
		return nil, err
	}
	a.core.AddPrivateNetworkAllowedKey(guest)
	a.reapplyPrivateNetworkKeys(cfg)
	return &InviteRegisterResponse{OK: true, Network: t.Net, GuestKey: guestHex}, nil
}

func (a *AdminSocket) reapplyPrivateNetworkKeys(cfg *config.NodeConfig) {
	keys, err := config.PrivateNetworkAllowedKeyHex(cfg)
	if err != nil {
		a.log.Warnf("reapplyPrivateNetworkKeys: %v", err)
		return
	}
	pks := make([]ed25519.PublicKey, 0, len(keys))
	for _, k := range keys {
		pks = append(pks, ed25519.PublicKey(k))
	}
	a.core.ReplacePrivateNetworkAllowedKeys(pks)
}

func parseHexPub(s string) (ed25519.PublicKey, error) {
	b, err := hex.DecodeString(strings.TrimSpace(s))
	if err != nil {
		return nil, err
	}
	if len(b) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid public key length")
	}
	return ed25519.PublicKey(b), nil
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func dialInviteRegister(endpoint, auth, tokenStr, guestKeyHex string) error {
	u, err := url.Parse(endpoint)
	if err != nil {
		return err
	}
	if u.Scheme != "tcp" && u.Scheme != "" {
		return fmt.Errorf("admin endpoint must be tcp://host:port")
	}
	host := u.Host
	if host == "" {
		host = endpoint
	}
	conn, err := net.DialTimeout("tcp", host, 15*time.Second)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	args, _ := json.Marshal(map[string]string{
		"token":    tokenStr,
		"guestKey": guestKeyHex,
	})
	req := AdminSocketRequest{
		Name:      "inviteRegister",
		Arguments: args,
		Auth:      auth,
	}
	enc := json.NewEncoder(conn)
	if err := enc.Encode(&req); err != nil {
		return err
	}
	var resp AdminSocketResponse
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return err
	}
	if resp.Status != "success" {
		if resp.Error != "" {
			return fmt.Errorf("%s", resp.Error)
		}
		return fmt.Errorf("inviteRegister failed")
	}
	return nil
}
