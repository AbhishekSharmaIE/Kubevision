package handlers

import (
	"net/mail"
	"strings"

	"github.com/AbhishekSharmaIE/Kubevision/internal/rbac"
)

func validateEmail(s string) bool {
	s = strings.TrimSpace(s)
	if len(s) == 0 || len(s) > 255 {
		return false
	}
	if _, err := mail.ParseAddress(s); err != nil {
		return false
	}
	return true
}

func validateName(s string) bool {
	s = strings.TrimSpace(s)
	return len(s) > 0 && len(s) <= 255
}

func validateTeamRole(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "admin", "member", "viewer":
		return true
	default:
		return false
	}
}

func validatePermissionString(s string) error {
	return rbac.ParsePermission(s)
}

func clampLimitOffset(limit, offset, maxLimit int) (int, int) {
	if limit <= 0 {
		limit = 50
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}
