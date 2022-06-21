package main

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/golang-jwt/jwt"
	"github.com/lestrrat-go/jwx/jwk"
)

var denyAllPolicy = events.APIGatewayCustomAuthorizerResponse{
	PrincipalID: "user",
	PolicyDocument: events.APIGatewayCustomAuthorizerPolicy{
		Version: "2012-10-17",
		Statement: []events.IAMPolicyStatement{
			{
				Action:   []string{"execute-api:Invoke"},
				Effect:   "Deny",
				Resource: []string{"*"},
			},
		},
	},
}

func main() {
	lambda.Start(handleAuthorization)
}

func handleAuthorization(ctx context.Context, event events.APIGatewayCustomAuthorizerRequest) (events.APIGatewayCustomAuthorizerResponse, error) {
	tokenStr := strings.Replace(event.AuthorizationToken, "Bearer ", "", 1)

	jwksURL := os.Getenv("OIDC_JWKS_URL")
	issuer := os.Getenv("OIDC_ISSUER")

	audienceStr := os.Getenv("OIDC_AUDIENCE")
	audience := strings.Split(audienceStr, ",")

	scopesStr := os.Getenv("OIDC_SCOPES")
	validScopes := strings.Split(scopesStr, ",")

	claims := JWTAuthorizerClaims{}
	_, err := jwt.ParseWithClaims(tokenStr, &claims, func(t *jwt.Token) (interface{}, error) {
		keyID := t.Header["kid"].(string)

		set, err := jwk.Fetch(ctx, jwksURL)
		if err != nil {
			return nil, err
		}

		if key, found := set.LookupKeyID(keyID); found {
			var k interface{}
			if err := key.Raw(&k); err == nil {
				return k, nil
			}
		}
		return nil, fmt.Errorf("unable to find key %s in keySet", keyID)
	})
	if err != nil {
		log.Printf("error validating token: %+v", err)
		return events.APIGatewayCustomAuthorizerResponse{}, errors.New("Unauthorized")
	}

	if !claims.VerifyIssuer(issuer, true) {
		log.Printf("issuer mismatch %s %s", issuer, claims.Issuer)
		return events.APIGatewayCustomAuthorizerResponse{}, errors.New("Unauthorized")
	}

	hasAudience := false
	for _, a := range audience {
		hasAudience = hasAudience || claims.VerifyAudience(a, true)
	}

	if !hasAudience {
		log.Printf("missing audience %+v in [%s]", audience, claims.Audience)
		return events.APIGatewayCustomAuthorizerResponse{}, errors.New("Unauthorized")
	}

	ok, scopes := claims.VerifyScope(validScopes)
	if !ok {
		log.Printf("missing audience %+v in [%s]", validScopes, claims.Scope)
		return events.APIGatewayCustomAuthorizerResponse{}, errors.New("Unauthorized")
	}

	authorization := NewAuthorization(event.MethodArn, claims.Email, scopes)

	return authorization, nil
}

type JWTAuthorizerClaims struct {
	FirstName string `json:"given_name"`
	LastName  string `json:"family_name"`
	Email     string `json:"email"`
	Scope     string `json:"scp,omitempty"`
	jwt.StandardClaims
}

func (c *JWTAuthorizerClaims) VerifyScope(scopes []string) (bool, []string) {
	result := false
	foundScopes := []string{}
	jwtScope := strings.Split(c.Scope, " ")
	for _, js := range jwtScope {
		for _, s := range scopes {
			if subtle.ConstantTimeCompare([]byte(js), []byte(s)) != 0 {
				result = true
				foundScopes = append(foundScopes, s)
			}
		}
	}

	return result, foundScopes
}

type APIPermission struct {
	Verb     string
	Resource string
}

type APIPermissions []APIPermission

var permissions = map[string]APIPermissions{
	"network.read": {
		{Verb: "GET", Resource: ""},
		{Verb: "GET", Resource: "api/v1/networks"},
		{Verb: "GET", Resource: "api/v1/networks/*"},
		{Verb: "GET", Resource: "api/v1/networks/*/subnets"},
		{Verb: "GET", Resource: "api/v1/pools"},
		{Verb: "GET", Resource: "api/v1/pools/*"},
		{Verb: "GET", Resource: "api/v1/providers"},
		{Verb: "GET", Resource: "api/v1/providers/*"},
	},
	"network.admin": {
		{Verb: "POST", Resource: "api/v1/networks"},
		{Verb: "PUT", Resource: "api/v1/networks/*"},
		{Verb: "DELETE", Resource: "api/v1/networks/*"},
		{Verb: "POST", Resource: "api/v1/pools"},
		{Verb: "DELETE", Resource: "api/v1/pools/*"},
		{Verb: "POST", Resource: "api/v1/providers"},
		{Verb: "DELETE", Resource: "api/v1/providers/*"},
	},
}

func NewAuthorization(methodARN, user string, scopes []string) events.APIGatewayCustomAuthorizerResponse {

	// "arn:aws:execute-api:{regionId}:{accountId}:{apiId}/{stage}/{httpVerb}/[{resource}/[{child-resources}]]"
	parts := strings.Split(methodARN, ":")
	apiGatewayParts := strings.Split(parts[5], "/")
	region := parts[3]
	accountID := parts[4]
	apiID := apiGatewayParts[0]
	stage := apiGatewayParts[1]

	policyStms := []events.IAMPolicyStatement{}
	for _, s := range scopes {
		if perms, ok := permissions[s]; ok {
			for _, p := range perms {
				r := fmt.Sprintf("arn:aws:execute-api:%s:%s:%s/%s/%s/%s",
					region,
					accountID,
					apiID,
					stage,
					p.Verb,
					p.Resource)
				stm := events.IAMPolicyStatement{
					Action:   []string{"execute-api:Invoke"},
					Effect:   "Allow",
					Resource: []string{r},
				}
				policyStms = append(policyStms, stm)
			}
		}
	}

	if len(policyStms) == 0 {
		return denyAllPolicy
	}

	return events.APIGatewayCustomAuthorizerResponse{
		PrincipalID: user,
		PolicyDocument: events.APIGatewayCustomAuthorizerPolicy{
			Version:   "2012-10-17",
			Statement: policyStms,
		},
	}
}
