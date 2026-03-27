package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/AbhishekSharmaIE/Kubevision/internal/api/httputil"
	"github.com/AbhishekSharmaIE/Kubevision/internal/api/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// UserHandler serves /api/v1/users.
type UserHandler struct {
	pool *pgxpool.Pool
}

// NewUserHandler constructs UserHandler.
func NewUserHandler(pool *pgxpool.Pool) *UserHandler {
	return &UserHandler{pool: pool}
}

type createUserBody struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password,omitempty"`
	IsAdmin  *bool  `json:"isAdmin,omitempty"`
}

type updateUserBody struct {
	Name     *string `json:"name,omitempty"`
	Email    *string `json:"email,omitempty"`
	Password *string `json:"password,omitempty"`
	IsAdmin  *bool   `json:"isAdmin,omitempty"`
}

func (h *UserHandler) countUsers(ctx context.Context) (int64, error) {
	var n int64
	err := h.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&n)
	return n, err
}

// Create handles POST /users — first user becomes admin without auth; later requires admin.
func (h *UserHandler) Create(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var body createUserBody
	if err := c.ShouldBindJSON(&body); err != nil {
		httputil.ErrorJSON(c, httputil.StatusUnprocessable, "invalid JSON body")
		return
	}
	if !validateEmail(body.Email) {
		httputil.ErrorJSON(c, httputil.StatusUnprocessable, "invalid email")
		return
	}
	if !validateName(body.Name) {
		httputil.ErrorJSON(c, httputil.StatusUnprocessable, "invalid name")
		return
	}

	n, err := h.countUsers(ctx)
	if err != nil {
		httputil.ErrorJSON(c, http.StatusInternalServerError, "failed to check users")
		return
	}

	isAdmin := false
	if body.IsAdmin != nil {
		isAdmin = *body.IsAdmin
	}
	if n == 0 {
		isAdmin = true
	} else {
		if middleware.UserID(c) == uuid.Nil {
			httputil.ErrorJSON(c, http.StatusUnauthorized, "authenticate as admin to create users")
			return
		}
		if !middleware.IsAdmin(c) {
			httputil.ErrorJSON(c, http.StatusForbidden, "admin required to create users")
			return
		}
	}

	var hash []byte
	if body.Password != "" {
		if len(body.Password) < 8 {
			httputil.ErrorJSON(c, httputil.StatusUnprocessable, "password must be at least 8 characters")
			return
		}
		b, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
		if err != nil {
			httputil.ErrorJSON(c, http.StatusInternalServerError, "could not hash password")
			return
		}
		hash = b
	}

	var id uuid.UUID
	err = h.pool.QueryRow(ctx, `
		INSERT INTO users (email, name, password_hash, is_admin)
		VALUES (LOWER(TRIM($1)), TRIM($2), $3, $4)
		RETURNING id`,
		body.Email, body.Name, nullableBytes(hash), isAdmin).Scan(&id)
	if err != nil {
		httputil.ErrorJSON(c, http.StatusConflict, "could not create user (email may already exist)")
		return
	}

	httputil.DataJSON(c, http.StatusCreated, gin.H{
		"id":             id.String(),
		"email":          body.Email,
		"name":           body.Name,
		"isAdmin":        isAdmin,
		"passwordSet":    len(hash) > 0,
	})
}

func nullableBytes(b []byte) any {
	if len(b) == 0 {
		return nil
	}
	return b
}

// List handles GET /users (admin only).
func (h *UserHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, offset = clampLimitOffset(limit, offset, 200)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	rows, err := h.pool.Query(ctx, `
		SELECT id, email, name, COALESCE(is_admin,false), created_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		httputil.ErrorJSON(c, http.StatusInternalServerError, "failed to list users")
		return
	}
	defer rows.Close()

	list := make([]gin.H, 0)
	for rows.Next() {
		var id uuid.UUID
		var email, name string
		var admin bool
		var created time.Time
		if err := rows.Scan(&id, &email, &name, &admin, &created); err != nil {
			httputil.ErrorJSON(c, http.StatusInternalServerError, "failed to read users")
			return
		}
		list = append(list, gin.H{
			"id":        id.String(),
			"email":     email,
			"name":      name,
			"isAdmin":   admin,
			"createdAt": created.UTC().Format(time.RFC3339),
		})
	}
	var total int64
	_ = h.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&total)
	httputil.DataJSONWithMeta(c, http.StatusOK, list, gin.H{"total": total, "limit": limit, "offset": offset})
}

// Get handles GET /users/:id (self or admin).
func (h *UserHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.ErrorJSON(c, http.StatusBadRequest, "invalid user id")
		return
	}
	self := middleware.UserID(c)
	if id != self && !middleware.IsAdmin(c) {
		httputil.ErrorJSON(c, http.StatusForbidden, "cannot view other users")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var email, name string
	var admin bool
	var created time.Time
	err = h.pool.QueryRow(ctx, `
		SELECT email, name, COALESCE(is_admin,false), created_at FROM users WHERE id = $1`, id).
		Scan(&email, &name, &admin, &created)
	if errors.Is(err, pgx.ErrNoRows) {
		httputil.ErrorJSON(c, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		httputil.ErrorJSON(c, http.StatusInternalServerError, "failed to load user")
		return
	}

	httputil.DataJSON(c, http.StatusOK, gin.H{
		"id":        id.String(),
		"email":     email,
		"name":      name,
		"isAdmin":   admin,
		"createdAt": created.UTC().Format(time.RFC3339),
	})
}

// Update handles PUT /users/:id.
func (h *UserHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.ErrorJSON(c, http.StatusBadRequest, "invalid user id")
		return
	}
	self := middleware.UserID(c)
	if id != self && !middleware.IsAdmin(c) {
		httputil.ErrorJSON(c, http.StatusForbidden, "cannot update other users")
		return
	}

	var body updateUserBody
	if err := c.ShouldBindJSON(&body); err != nil {
		httputil.ErrorJSON(c, httputil.StatusUnprocessable, "invalid JSON body")
		return
	}
	if body.IsAdmin != nil && !middleware.IsAdmin(c) {
		httputil.ErrorJSON(c, http.StatusForbidden, "only admins can change isAdmin")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	if body.Email != nil {
		if !validateEmail(*body.Email) {
			httputil.ErrorJSON(c, httputil.StatusUnprocessable, "invalid email")
			return
		}
		_, err = h.pool.Exec(ctx, `UPDATE users SET email = LOWER(TRIM($1)) WHERE id = $2`, *body.Email, id)
		if err != nil {
			httputil.ErrorJSON(c, http.StatusConflict, "could not update email")
			return
		}
	}
	if body.Name != nil {
		if !validateName(*body.Name) {
			httputil.ErrorJSON(c, httputil.StatusUnprocessable, "invalid name")
			return
		}
		_, err = h.pool.Exec(ctx, `UPDATE users SET name = TRIM($1) WHERE id = $2`, *body.Name, id)
		if err != nil {
			httputil.ErrorJSON(c, http.StatusInternalServerError, "could not update name")
			return
		}
	}
	if body.Password != nil && *body.Password != "" {
		if len(*body.Password) < 8 {
			httputil.ErrorJSON(c, httputil.StatusUnprocessable, "password must be at least 8 characters")
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(*body.Password), bcrypt.DefaultCost)
		if err != nil {
			httputil.ErrorJSON(c, http.StatusInternalServerError, "could not hash password")
			return
		}
		_, err = h.pool.Exec(ctx, `UPDATE users SET password_hash = $1 WHERE id = $2`, hash, id)
		if err != nil {
			httputil.ErrorJSON(c, http.StatusInternalServerError, "could not update password")
			return
		}
	}
	if body.IsAdmin != nil {
		_, err = h.pool.Exec(ctx, `UPDATE users SET is_admin = $1 WHERE id = $2`, *body.IsAdmin, id)
		if err != nil {
			httputil.ErrorJSON(c, http.StatusInternalServerError, "could not update admin flag")
			return
		}
	}

	h.Get(c)
}

// Delete handles DELETE /users/:id (admin only).
func (h *UserHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.ErrorJSON(c, http.StatusBadRequest, "invalid user id")
		return
	}
	if id == middleware.UserID(c) {
		httputil.ErrorJSON(c, http.StatusBadRequest, "cannot delete your own account via this endpoint")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	cmd, err := h.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		httputil.ErrorJSON(c, http.StatusInternalServerError, "failed to delete user")
		return
	}
	if cmd.RowsAffected() == 0 {
		httputil.ErrorJSON(c, http.StatusNotFound, "user not found")
		return
	}
	c.Status(http.StatusNoContent)
}
