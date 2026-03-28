package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/AbhishekSharmaIE/Kubevision/internal/api/httputil"
	"github.com/AbhishekSharmaIE/Kubevision/internal/api/middleware"
	"github.com/AbhishekSharmaIE/Kubevision/internal/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler serves /api/v1/auth/*.
type AuthHandler struct {
	pool *pgxpool.Pool
	jwt  *auth.JWT
	rdb  *redis.Client
}

// NewAuthHandler constructs AuthHandler.
func NewAuthHandler(pool *pgxpool.Pool, j *auth.JWT, rdb *redis.Client) *AuthHandler {
	return &AuthHandler{pool: pool, jwt: j, rdb: rdb}
}

type loginBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshBody struct {
	Token string `json:"token"`
}

// Login handles POST /auth/login.
func (h *AuthHandler) Login(c *gin.Context) {
	var body loginBody
	if err := c.ShouldBindJSON(&body); err != nil {
		httputil.ErrorJSON(c, httputil.StatusUnprocessable, "invalid JSON body")
		return
	}
	email := strings.TrimSpace(strings.ToLower(body.Email))
	if !validateEmail(email) || body.Password == "" {
		httputil.ErrorJSON(c, http.StatusUnauthorized, "invalid credentials")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var uid uuid.UUID
	var hash []byte
	var isAdmin bool
	err := h.pool.QueryRow(ctx, `
		SELECT id, password_hash, COALESCE(is_admin,false) FROM users WHERE email = $1`, email).
		Scan(&uid, &hash, &isAdmin)
	if errors.Is(err, pgx.ErrNoRows) {
		httputil.ErrorJSON(c, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err != nil {
		httputil.ErrorJSON(c, http.StatusInternalServerError, "login failed")
		return
	}
	if len(hash) == 0 || bcrypt.CompareHashAndPassword(hash, []byte(body.Password)) != nil {
		httputil.ErrorJSON(c, http.StatusUnauthorized, "invalid credentials")
		return
	}

	teamIDs, err := h.loadTeamIDs(ctx, uid)
	if err != nil {
		httputil.ErrorJSON(c, http.StatusInternalServerError, "login failed")
		return
	}

	var name string
	_ = h.pool.QueryRow(ctx, `SELECT name FROM users WHERE id = $1`, uid).Scan(&name)

	token, jti, err := h.jwt.IssueToken(uid, email, isAdmin, teamIDs)
	if err != nil {
		httputil.ErrorJSON(c, http.StatusInternalServerError, "could not issue token")
		return
	}
	if err := auth.StoreSession(ctx, h.rdb, uid, jti, h.jwt.AccessTTL()); err != nil {
		httputil.ErrorJSON(c, http.StatusInternalServerError, "could not start session")
		return
	}

	_, _ = h.pool.Exec(ctx, `UPDATE users SET last_login = NOW() WHERE id = $1`, uid)

	httputil.DataJSON(c, http.StatusOK, gin.H{
		"token":     token,
		"tokenType": "Bearer",
		"expiresIn": int(h.jwt.AccessTTL().Seconds()),
		"user": gin.H{
			"id": uid.String(), "email": email, "name": name, "isAdmin": isAdmin,
		},
	})
}

func (h *AuthHandler) loadTeamIDs(ctx context.Context, userID uuid.UUID) ([]string, error) {
	rows, err := h.pool.Query(ctx, `SELECT team_id::text FROM team_members WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		ids = append(ids, s)
	}
	return ids, rows.Err()
}

// Refresh handles POST /auth/refresh (accepts expired access token within refresh window).
func (h *AuthHandler) Refresh(c *gin.Context) {
	var body refreshBody
	_ = c.ShouldBindJSON(&body)
	token := strings.TrimSpace(body.Token)
	if token == "" {
		var ok bool
		token, ok = extractBearer(c)
		if !ok {
			httputil.ErrorJSON(c, httputil.StatusUnprocessable, "token required in body or Authorization header")
			return
		}
	}

	claims, err := h.jwt.ParseForRefresh(token)
	if err != nil {
		httputil.ErrorJSON(c, http.StatusUnauthorized, "invalid or stale token")
		return
	}
	uid, err := uuid.Parse(claims.UserID)
	if err != nil || claims.RegisteredClaims.ID == "" {
		httputil.ErrorJSON(c, http.StatusUnauthorized, "invalid token")
		return
	}
	if !auth.ValidateSession(c.Request.Context(), h.rdb, uid, claims.RegisteredClaims.ID) {
		httputil.ErrorJSON(c, http.StatusUnauthorized, "session revoked")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var email string
	var isAdmin bool
	var name string
	err = h.pool.QueryRow(ctx, `
		SELECT email, COALESCE(is_admin,false), name FROM users WHERE id = $1`, uid).
		Scan(&email, &isAdmin, &name)
	if err != nil {
		httputil.ErrorJSON(c, http.StatusUnauthorized, "user not found")
		return
	}

	teamIDs, err := h.loadTeamIDs(ctx, uid)
	if err != nil {
		httputil.ErrorJSON(c, http.StatusInternalServerError, "refresh failed")
		return
	}

	oldJTI := claims.RegisteredClaims.ID
	if err := auth.RevokeSession(ctx, h.rdb, uid, oldJTI); err != nil {
		httputil.ErrorJSON(c, http.StatusInternalServerError, "refresh failed")
		return
	}

	newTok, newJTI, err := h.jwt.IssueToken(uid, email, isAdmin, teamIDs)
	if err != nil {
		httputil.ErrorJSON(c, http.StatusInternalServerError, "could not issue token")
		return
	}
	if err := auth.StoreSession(ctx, h.rdb, uid, newJTI, h.jwt.AccessTTL()); err != nil {
		httputil.ErrorJSON(c, http.StatusInternalServerError, "could not start session")
		return
	}

	httputil.DataJSON(c, http.StatusOK, gin.H{
		"token":     newTok,
		"tokenType": "Bearer",
		"expiresIn": int(h.jwt.AccessTTL().Seconds()),
		"user": gin.H{
			"id": uid.String(), "email": email, "name": name, "isAdmin": isAdmin,
		},
	})
}

func extractBearer(c *gin.Context) (string, bool) {
	raw := c.GetHeader("Authorization")
	const p = "Bearer "
	if !strings.HasPrefix(raw, p) {
		return "", false
	}
	t := strings.TrimSpace(strings.TrimPrefix(raw, p))
	return t, t != ""
}

// Logout handles POST /auth/logout (requires valid access token).
func (h *AuthHandler) Logout(c *gin.Context) {
	cl := middleware.GetClaims(c)
	if cl == nil {
		httputil.ErrorJSON(c, http.StatusUnauthorized, "not authenticated")
		return
	}
	uid, err := uuid.Parse(cl.UserID)
	if err != nil {
		httputil.ErrorJSON(c, http.StatusUnauthorized, "invalid session")
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	_ = auth.RevokeSession(ctx, h.rdb, uid, cl.RegisteredClaims.ID)
	c.Status(http.StatusNoContent)
}
