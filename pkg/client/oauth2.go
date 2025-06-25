package client

import (
	"context"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/int128/oauth2cli"
	"github.com/pkg/browser"
	"golang.org/x/oauth2"
	"golang.org/x/sync/errgroup"
)

type OAuth2Authorizer struct {
	cfg   oauth2cli.Config
	ready chan string

	o *OAuth2AuthorizerOptions
}

type OAuth2AuthorizerOptions struct {
	ClientID string
	Issuer   string
	Scopes   []string
	TokenDir string
}

func NewOAuth2Authorizer(o *OAuth2AuthorizerOptions) (*OAuth2Authorizer, error) {
	pkceVerifier := oauth2.GenerateVerifier()
	ready := make(chan string, 1)
	cfg := oauth2cli.Config{
		OAuth2Config: oauth2.Config{
			ClientID: o.ClientID,
			Endpoint: oauth2.Endpoint{
				AuthURL:  fmt.Sprintf("%s/authorize", o.Issuer),
				TokenURL: fmt.Sprintf("%s/token", o.Issuer),
			},
			Scopes: append([]string{"openid"}, o.Scopes...),
		},
		AuthCodeOptions:      []oauth2.AuthCodeOption{oauth2.S256ChallengeOption(pkceVerifier)},
		TokenRequestOptions:  []oauth2.AuthCodeOption{oauth2.VerifierOption(pkceVerifier)},
		LocalServerReadyChan: ready,
		Logf:                 log.Printf,
	}

	return &OAuth2Authorizer{
		cfg:   cfg,
		ready: ready,
		o:     o,
	}, nil
}

func (o *OAuth2Authorizer) Close() {
	close(o.ready)
}

func (o *OAuth2Authorizer) NewClient(ctx context.Context, t *oauth2.Token) *http.Client {
	return o.cfg.OAuth2Config.Client(ctx, t)
}

func (o *OAuth2Authorizer) GetToken(ctx context.Context) (*oauth2.Token, error) {
	h, err := o.tokenHash()
	if err != nil {
		return nil, err
	}
	t, err := o.GetFromCache(ctx)
	if err != nil {
		log.Printf("no token found on cache: %v", err)
		return o.NewToken(ctx)
	}

	valid, err := o.Valid(t.AccessToken)
	if err != nil {
		log.Printf("fail to validate token: %v", err)
	}
	if valid {
		return t, nil
	}

	t, err = o.RefreshToken(ctx, t.RefreshToken)
	if err != nil {
		log.Printf("error refreshing token: %v", err)
		return o.NewToken(ctx)
	}

	err = o.SaveToCache(ctx, t)
	if err != nil {
		log.Printf("failed to save token %s", h)
		return t, nil
	}
	return t, nil
}

func (o *OAuth2Authorizer) NewToken(ctx context.Context) (*oauth2.Token, error) {
	t, err := o.Authorize(ctx)
	if err != nil {
		return nil, err
	}
	err = o.SaveToCache(ctx, t)
	if err != nil {
		log.Printf("failed to save token: %+v", err)
		return t, nil
	}
	return t, nil
}

type cacheKey struct {
	ClientID string
	Scopes   []string
}

type entity struct {
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

func (o *OAuth2Authorizer) tokenHash() (string, error) {
	key := cacheKey{
		ClientID: o.cfg.OAuth2Config.ClientID,
		Scopes:   o.cfg.OAuth2Config.Scopes,
	}

	s := sha256.New()
	e := gob.NewEncoder(s)
	if err := e.Encode(&key); err != nil {
		return "", fmt.Errorf("could not encode the key: %w", err)
	}
	h := hex.EncodeToString(s.Sum(nil))
	return h, nil
}

func (o *OAuth2Authorizer) GetFromCache(ctx context.Context) (*oauth2.Token, error) {
	id, err := o.tokenHash()
	if err != nil {
		return nil, err
	}
	p := filepath.Join(o.o.TokenDir, id)
	f, err := os.Open(p)
	if err != nil {
		return nil, fmt.Errorf("could not open file %s: %w", p, err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			log.Printf("error closing file %s: %v", p, closeErr)
		}
	}()
	d := json.NewDecoder(f)
	var e entity
	if err := d.Decode(&e); err != nil {
		return nil, fmt.Errorf("invalid json file %s: %w", p, err)
	}
	return &oauth2.Token{
		AccessToken:  e.AccessToken,
		RefreshToken: e.RefreshToken,
	}, nil
}

func (o *OAuth2Authorizer) SaveToCache(ctx context.Context, t *oauth2.Token) error {
	id, err := o.tokenHash()
	if err != nil {
		return err
	}
	p := filepath.Join(o.o.TokenDir, id)
	f, err := os.Create(p)
	if err != nil {
		return fmt.Errorf("could not open file %s: %w", p, err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			log.Printf("error closing file %s: %v", p, closeErr)
		}
	}()
	enc := json.NewEncoder(f)
	e := entity{
		AccessToken:  t.AccessToken,
		RefreshToken: t.RefreshToken,
	}
	if err := enc.Encode(&e); err != nil {
		return fmt.Errorf("failed to encode json %s: %w", p, err)
	}
	return nil
}

func (o *OAuth2Authorizer) Valid(t string) (bool, error) {
	p := new(jwt.Parser)
	claims := jwt.RegisteredClaims{}
	_, _, err := p.ParseUnverified(t, &claims)
	if err != nil {
		log.Printf("Something went wrong!")
		return false, err
	}
	exp, err := claims.GetExpirationTime()
	if err != nil {
		log.Printf("Could not get expiration time: %v", err)
		return false, err
	}
	if exp == nil {
		log.Printf("No expiration time found in token claims")
		return false, nil
	}
	if exp.Before(time.Now()) {
		log.Printf("Expired!")
		return false, nil
	}
	return true, nil
}

func (o *OAuth2Authorizer) Authorize(ctx context.Context) (*oauth2.Token, error) {
	var out *oauth2.Token
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		select {
		case url := <-o.ready:
			log.Printf("Open %s", url)
			if err := browser.OpenURL(url); err != nil {
				log.Printf("could not open the browser: %s", err)
			}
			return nil
		case <-ctx.Done():
			return fmt.Errorf("context done while waiting for authorization: %w", ctx.Err())
		}
	})
	eg.Go(func() error {
		token, err := oauth2cli.GetToken(ctx, o.cfg)
		if err != nil {
			return fmt.Errorf("authorization code flow error: %w", err)
		}
		out = token
		return nil
	})
	if err := eg.Wait(); err != nil {
		return nil, fmt.Errorf("authorization error: %w", err)
	}
	return out, nil
}

func (o *OAuth2Authorizer) RefreshToken(ctx context.Context, rt string) (*oauth2.Token, error) {
	opts := []oauth2.AuthCodeOption{
		oauth2.SetAuthURLParam("grant_type", "refresh_token"),
		oauth2.SetAuthURLParam("refresh_token", rt),
		// AzureAD requires scope to be set
		oauth2.SetAuthURLParam("scope", strings.Join(o.cfg.OAuth2Config.Scopes, " ")),
	}

	t, err := o.cfg.OAuth2Config.Exchange(ctx, "", opts...)
	if err != nil {
		return nil, err
	}

	return t, nil
}
