package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	chi "github.com/go-chi/chi/v5"
	"github.com/kelseyhightower/envconfig"
	qrcode "github.com/skip2/go-qrcode"
	"go.uber.org/zap"
	oauth2 "golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

// shipperizer in ~/shipperizer/hydra-rock/hack/flow-test on IAM-596 ● λ http http://127.0.0.1:4444/.well-known/openid-configuration
// HTTP/1.1 200 OK
// Cache-Control: private, no-cache, no-store, must-revalidate
// Content-Length: 1976
// Content-Type: application/json; charset=utf-8
// Date: Thu, 30 Nov 2023 12:56:08 GMT

// {
//     "authorization_endpoint": "https://localhost:4444/oauth2/auth",
//     "backchannel_logout_session_supported": true,
//     "backchannel_logout_supported": true,
//     "claims_parameter_supported": false,
//     "claims_supported": [
//         "sub"
//     ],
//     "code_challenge_methods_supported": [
//         "plain",
//         "S256"
//     ],
//     "credentials_endpoint_draft_00": "https://localhost:4444/credentials",
//     "credentials_supported_draft_00": [
//         {
//             "cryptographic_binding_methods_supported": [
//                 "jwk"
//             ],
//             "cryptographic_suites_supported": [
//                 "PS256",
//                 "RS256",
//                 "ES256",
//                 "PS384",
//                 "RS384",
//                 "ES384",
//                 "PS512",
//                 "RS512",
//                 "ES512",
//                 "EdDSA"
//             ],
//             "format": "jwt_vc_json",
//             "types": [
//                 "VerifiableCredential",
//                 "UserInfoCredential"
//             ]
//         }
//     ],
//     "device_authorization_endpoint": "https://localhost:4444/oauth2/device/auth",
//     "end_session_endpoint": "https://localhost:4444/oauth2/sessions/logout",
//     "frontchannel_logout_session_supported": true,
//     "frontchannel_logout_supported": true,
//     "grant_types_supported": [
//         "authorization_code",
//         "implicit",
//         "client_credentials",
//         "refresh_token",
//         "urn:ietf:params:oauth:grant-type:device_code"
//     ],
//     "id_token_signed_response_alg": [
//         "RS256"
//     ],
//     "id_token_signing_alg_values_supported": [
//         "RS256"
//     ],
//     "issuer": "https://localhost:4444/",
//     "jwks_uri": "https://localhost:4444/.well-known/jwks.json",
//     "request_object_signing_alg_values_supported": [
//         "none",
//         "RS256",
//         "ES256"
//     ],
//     "request_parameter_supported": true,
//     "request_uri_parameter_supported": true,
//     "require_request_uri_registration": true,
//     "response_modes_supported": [
//         "query",
//         "fragment"
//     ],
//     "response_types_supported": [
//         "code",
//         "code id_token",
//         "id_token",
//         "token id_token",
//         "token",
//         "token id_token code"
//     ],
//     "revocation_endpoint": "https://localhost:4444/oauth2/revoke",
//     "scopes_supported": [
//         "offline_access",
//         "offline",
//         "openid"
//     ],
//     "subject_types_supported": [
//         "public"
//     ],
//     "token_endpoint": "https://localhost:4444/oauth2/token",
//     "token_endpoint_auth_methods_supported": [
//         "client_secret_post",
//         "client_secret_basic",
//         "private_key_jwt",
//         "none"
//     ],
//     "userinfo_endpoint": "https://localhost:4444/userinfo",
//     "userinfo_signed_response_alg": [
//         "RS256"
//     ],
//     "userinfo_signing_alg_values_supported": [
//         "none",
//         "RS256"
//     ]
// }

// shipperizer in ~/shipperizer/hydra on feat_dev_grants_2x λ ./hydra create client --endpoint http://127.0.0.1:4445 --name noauth --grant-type authorization_code,refresh_token,urn:ietf:params:oauth:grant-type:device_code --response-type code,id_token --scope openid,offline --redirect-uri http://localhost:1337/hello --token-endpoint-auth-method none
// CLIENT ID     5a9d0eaf-2f78-4b66-b63d-49838ed33f19
// CLIENT SECRET
// GRANT TYPES     authorization_code, refresh_token, urn:ietf:params:oauth:grant-type:device_code
// RESPONSE TYPES  code, id_token
// SCOPE           openid offline
// AUDIENCE
// REDIRECT URIS   http://localhost:1337/hello
// shipperizer in ~/shipperizer/hydra on feat_dev_grants_2x λ ./hydra create client --endpoint http://127.0.0.1:4445 --name auth --grant-type authorization_code,refresh_token,urn:ietf:params:oauth:grant-type:device_code --response-type code,id_token --scope openid,offline --redirect-uri http://localhost:8000/api/ready
// CLIENT ID     efc05555-e960-4aa2-bc0d-291144e15963
// CLIENT SECRET   cjMbxFlwS1BHXEyVpR-3JLoTPN
// GRANT TYPES     authorization_code, refresh_token, urn:ietf:params:oauth:grant-type:device_code
// RESPONSE TYPES  code, id_token
// SCOPE           openid offline
// AUDIENCE
// REDIRECT URIS   http://localhost:8000/api/ready

type ProviderType int

const (
	Hydra ProviderType = iota
	Github
)

// EnvSpec is the basic environment configuration setup needed for the app to start
type EnvSpec struct {
	OAuthClientID     string       `envconfig:"oauth_client_id"`
	OAuthClientSecret string       `envconfig:"oauth_client_secret"`
	CallbackURI       string       `envconfig:"callback_uri" default:"http://localhost:8000/api/ready"`
	Scopes            []string     `envconfig:"scopes" default:"openid,offline"`
	Provider          ProviderType `envconfig:"provider" default:"0"`
	AuthURL           string       `envconfig:"auth_url" default:"http://localhost:4444/oauth2/auth"`
	TokenURL          string       `envconfig:"token_url" default:"http://localhost:4444/oauth2/token"`
	DeviceAuthURL     string       `envconfig:"device_auth_url" default:"http://localhost:4444/oauth2/device/auth"`
}

func router(logger *zap.SugaredLogger) *chi.Mux {
	router := chi.NewMux()

	router.Get(
		"/api/ready",
		func(w http.ResponseWriter, r *http.Request) {
			logger.Infof("query params: %v", r.URL.Query())
			w.WriteHeader(http.StatusOK)
		},
	)

	return router
}

func deviceFlow(specs *EnvSpec, logger *zap.SugaredLogger) {
	config := new(oauth2.Config)
	config.ClientID = specs.OAuthClientID
	config.ClientSecret = specs.OAuthClientSecret
	config.Scopes = specs.Scopes
	// config.RedirectURL = specs.CallbackURI

	switch specs.Provider {
	case Github:
		config.Endpoint = github.Endpoint
	case Hydra:
		config.Endpoint = oauth2.Endpoint{
			AuthURL:       specs.AuthURL,
			TokenURL:      specs.TokenURL,
			DeviceAuthURL: specs.DeviceAuthURL,
		}
	}

	for {
		verifier := oauth2.GenerateVerifier()
		challenge := oauth2.S256ChallengeFromVerifier(verifier)

		logger.Debugf("Verifier: %s - Challenge: %s", verifier, challenge)
		ctx := context.Background()

		response, err := config.DeviceAuth(
			ctx,
			// oauth2.SetAuthURLParam("response_type", "code"),
			// oauth2.SetAuthURLParam("code_verifier", verifier),
			// oauth2.SetAuthURLParam("code_challenge", challenge),
			// oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		)

		if err != nil {
			logger.Errorf(err.Error())
			logger.Warn("sleeping for 60s")
			time.Sleep(60 * time.Second)
			continue
		}

		logger.Debugf("response: %v", response)
		logger.Infof(
			`
############################################################
add the following to your /etc/hosts
"127.0.0.1 iam.internal"
############################################################
run the following command: $(KUBECTL) get secret -o yaml iam-tls | yq '.data'
copy the ca.crt and tls.crt into /usr/local/share/ca-certificates/ and run update-ca-certificates
to get those certs added to the system pool (and trust them), you might need to do
the same (trust) in your chrome/firefox/safari browser
after that you should be able to point openssl or certigo to the forwarded ingress on your localhost (port 8443) and 
verify that the cert is valid
############################################################
use the hydra cli to create a client:
hydra create client --endpoint http://iam.internal:4445 --name auth --grant-type "urn:ietf:params:oauth:grant-type:device_code,authorization_code" --response-type "code" --scope openid,offline
and then swap the client-id in the flow-test-hydra configmap
bounce the flow-test pod to pick up the changes
############################################################
please enter code %s at %s
or use following command: http --verify=/usr/local/share/ca-certificates/ca.crt %s user_code==%s --follow
############################################################
	`,
			response.UserCode,
			response.VerificationURI,
			response.VerificationURI,
			response.UserCode,
		)
		if qr, err := qrcode.New(response.VerificationURIComplete, qrcode.Low); err == nil {
			logger.Infof("############################################################")
			logger.Infof("or scan this %s", qr.ToString(true))
			logger.Infof("############################################################")
		}
		token, err := config.DeviceAccessToken(ctx, response)

		if err != nil {
			logger.Warn(err, token)
		} else {
			logger.Infof("Succeeded, Token: %v", token)
		}

		logger.Info("device flow done...one way or the other")
		logger.Infof("############################################################")
	}
}

// example taken from https://github.com/supercairos/oauth-device-flow-client-sample/blob/master/src/index.ts
func main() {
	var logger *zap.SugaredLogger

	if _log, err := zap.NewDevelopment(); err != nil {
		logger = zap.NewNop().Sugar()
	} else {
		logger = _log.Sugar()
	}

	specs := new(EnvSpec)

	if err := envconfig.Process("", specs); err != nil {
		panic(fmt.Errorf("issues with environment sourcing: %s", err))
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%v", 9000),
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      router(logger),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			logger.Fatal(err)
		}
	}()

	go deviceFlow(specs, logger)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)

	logger.Desugar().Sync()

	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	logger.Info("Shutting down")
	os.Exit(0)
}
