package utility

import (
	"strings"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/text/gstr"
)

type Session struct {
	Email        string   `json:"email"`
	AccessToken  string   `json:"accessToken"`
	RefreshToken string   `json:"refreshToken"`
	PlanType     string   `json:"planType"`
	TeamIds      []string `json:"teamIds"`
	FreeWithGpts bool     `json:"freeWithGpts"`
	// 标记是否被禁用
	Disabled bool `json:"disabled"`
}

// ParseSession parses the session from the given JSON string.
func ParseSession(jsonStr string) (*Session, error) {
	session := &Session{
		Disabled:     false,
		FreeWithGpts: false,
	}
	sessionJson := gjson.New(jsonStr)
	detail := sessionJson.Get("detail").String()
	if detail != "" {
		if detail == "密码不正确!" || gstr.Contains(detail, "account_deactivated") || gstr.Contains(detail, "mfa_bypass") || gstr.Contains(detail, "两步验证") {
			session.Disabled = true
		}
		return session, gerror.Newf("invalid session: %s", detail)
	}
	session.RefreshToken = sessionJson.Get("refresh_token").String()
	session.AccessToken = sessionJson.Get("access_token").String()
	email, err := ParserAccessToken(session.AccessToken)
	if err != nil {
		return session, err
	}
	session.Email = email
	session.PlanType = sessionJson.Get("accountCheckInfo.plan_type").String()
	session.TeamIds = sessionJson.Get("accountCheckInfo.team_ids").Strings()
	if session.PlanType == "free" {
		features := strings.Join(sessionJson.Get("accounts_info.accounts.default.features").Strings(), ",")
		if gstr.Contains(features, "gizmo_interact_unpaid") {
			session.FreeWithGpts = true
		} else {
			session.FreeWithGpts = false
		}
	}

	return session, nil
}
