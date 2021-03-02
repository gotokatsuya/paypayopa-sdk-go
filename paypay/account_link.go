package paypay

import (
	"context"
	"encoding/base64"
	"net/http"

	"github.com/dgrijalva/jwt-go"
)

type CreateAccountLinkQRCodeRequest struct {
	Scopes       []string `json:"scopes"`
	Nonce        string   `json:"nonce"`
	RedirectType string   `json:"redirectType,omitempty"`
	RedirectURL  string   `json:"redirectUrl"`
	ReferenceID  string   `json:"referenceId"`
	PhoneNumber  string   `json:"phoneNumber,omitempty"`
	DeviceID     string   `json:"deviceId,omitempty"`
	UserAgent    string   `json:"userAgent,omitempty"`
}

type CreateAccountLinkQRCodeResponse struct {
	ResultInfo struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		CodeID  string `json:"codeId"`
	} `json:"resultInfo"`
	Data struct {
		LinkQRCodeURL string `json:"linkQRCodeURL"`
	} `json:"data"`
}

func (c *Client) CreateAccountLinkQRCode(ctx context.Context, req *CreateAccountLinkQRCodeRequest) (*CreateAccountLinkQRCodeResponse, *http.Response, error) {
	path := "/v1/qr/sessions"
	httpReq, err := c.NewRequest(http.MethodPost, path, req)
	if err != nil {
		return nil, nil, err
	}
	resp := new(CreateAccountLinkQRCodeResponse)
	httpResp, err := c.Do(ctx, httpReq, resp)
	if err != nil {
		return nil, httpResp, err
	}
	return resp, httpResp, nil
}

type ResponseToken struct {
	jwt.StandardClaims
	Result              string `json:"result"`
	ProfileIdentifier   string `json:"profileIdentifier"`
	Nonce               string `json:"nonce"`
	UserAuthorizationID string `json:"userAuthorizationId"`
	ReferenceID         string `json:"referenceId"`
}

func (c *Client) ParseResponseToken(tokenString string) (*ResponseToken, error) {
	token, err := jwt.ParseWithClaims(tokenString, &ResponseToken{}, func(token *jwt.Token) (interface{}, error) {
		return base64.StdEncoding.DecodeString(c.apiSecret)
	})
	if err != nil {
		return nil, err
	}
	return token.Claims.(*ResponseToken), nil
}
