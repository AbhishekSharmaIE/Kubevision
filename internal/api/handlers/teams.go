package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/AbhishekSharmaIE/Kubevision/internal/api/httputil"
	"github.com/AbhishekSharmaIE/Kubevision/internal/api/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TeamHandler serves /api/v1/teams.
type TeamHandler struct {
	pool *pgxpool.Pool
}

// NewTeamHandler constructs TeamHandler.
func NewTeamHandler(pool *pgxpool.Pool) *TeamHandler {
	return &TeamHandler{pool: pool}
}

type createTeamBody struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type updateTeamBody struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

type addMemberBody struct {
	UserID string `json:"userId"`
	Role   string `json:"role,omitempty"`
}

func (h *TeamHandler) userMemberOfTeam(ctx context.Context, userID, teamID uuid.UUID) (bool, error) {
	var ok bool
	err := h.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM team_members WHERE user_id = $1 AND team_id = $2)`, userID, teamID).Scan(&ok)
	return ok, err
}

// List handles GET /teams.
func (h *TeamHandler) List(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	uid := middleware.UserID(c)
	var rows pgx.Rows
	var err error
	if middleware.IsAdmin(c) {
		rows, err = h.pool.Query(ctx, `
			SELECT id, name, COALESCE(description,''), created_at FROM teams ORDER BY name`)
	} else {
		rows, err = h.pool.Query(ctx, `
			SELECT t.id, t.name, COALESCE(t.description,''), t.created_at
			FROM teams t
			INNER JOIN team_members tm ON tm.team_id = t.id
			WHERE tm.user_id = $1
			ORDER BY t.name`, uid)
	}
	if err != nil {
		httputil.ErrorJSON(c, http.StatusInternalServerError, "failed to list teams")
		return
	}
	defer rows.Close()

	list := make([]gin.H, 0)
	for rows.Next() {
		var id uuid.UUID
		var name, desc string
		var created time.Time
		if err := rows.Scan(&id, &name, &desc, &created); err != nil {
			httputil.ErrorJSON(c, http.StatusInternalServerError, "failed to read teams")
			return
		}
		list = append(list, gin.H{
			"id":          id.String(),
			"name":        name,
			"description": desc,
			"createdAt":   created.UTC().Format(time.RFC3339),
		})
	}
	httputil.DataJSON(c, http.StatusOK, list)
}

// Create handles POST /teams (admin).
func (h *TeamHandler) Create(c *gin.Context) {
	var body createTeamBody
	if err := c.ShouldBindJSON(&body); err != nil {
		httputil.ErrorJSON(c, httputil.StatusUnprocessable, "invalid JSON body")
		return
	}
	name := strings.TrimSpace(body.Name)
	if name == "" || len(name) > 255 {
		httputil.ErrorJSON(c, httputil.StatusUnprocessable, "invalid team name")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var id uuid.UUID
	err := h.pool.QueryRow(ctx, `
		INSERT INTO teams (name, description) VALUES ($1, NULLIF(TRIM($2),'')) RETURNING id`,
		name, body.Description).Scan(&id)
	if err != nil {
		httputil.ErrorJSON(c, http.StatusConflict, "could not create team (name may exist)")
		return
	}

	httputil.DataJSON(c, http.StatusCreated, gin.H{
		"id":          id.String(),
		"name":        name,
		"description": strings.TrimSpace(body.Description),
	})
}

// Get handles GET /teams/:id.
func (h *TeamHandler) Get(c *gin.Context) {
	teamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.ErrorJSON(c, http.StatusBadRequest, "invalid team id")
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if !middleware.IsAdmin(c) {
		ok, err := h.userMemberOfTeam(ctx, middleware.UserID(c), teamID)
		if err != nil || !ok {
			httputil.ErrorJSON(c, http.StatusForbidden, "not a member of this team")
			return
		}
	}

	var name, desc string
	var created time.Time
	err = h.pool.QueryRow(ctx, `
		SELECT name, COALESCE(description,''), created_at FROM teams WHERE id = $1`, teamID).
		Scan(&name, &desc, &created)
	if errors.Is(err, pgx.ErrNoRows) {
		httputil.ErrorJSON(c, http.StatusNotFound, "team not found")
		return
	}
	if err != nil {
		httputil.ErrorJSON(c, http.StatusInternalServerError, "failed to load team")
		return
	}

	httputil.DataJSON(c, http.StatusOK, gin.H{
		"id":          teamID.String(),
		"name":        name,
		"description": desc,
		"createdAt":   created.UTC().Format(time.RFC3339),
	})
}

// Update handles PUT /teams/:id (admin).
func (h *TeamHandler) Update(c *gin.Context) {
	teamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.ErrorJSON(c, http.StatusBadRequest, "invalid team id")
		return
	}
	var body updateTeamBody
	if err := c.ShouldBindJSON(&body); err != nil {
		httputil.ErrorJSON(c, httputil.StatusUnprocessable, "invalid JSON body")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	if body.Name != nil {
		n := strings.TrimSpace(*body.Name)
		if n == "" || len(n) > 255 {
			httputil.ErrorJSON(c, httputil.StatusUnprocessable, "invalid team name")
			return
		}
		_, err = h.pool.Exec(ctx, `UPDATE teams SET name = $1 WHERE id = $2`, n, teamID)
		if err != nil {
			httputil.ErrorJSON(c, http.StatusConflict, "could not update team name")
			return
		}
	}
	if body.Description != nil {
		_, err = h.pool.Exec(ctx, `UPDATE teams SET description = NULLIF(TRIM($1),'') WHERE id = $2`, *body.Description, teamID)
		if err != nil {
			httputil.ErrorJSON(c, http.StatusInternalServerError, "could not update description")
			return
		}
	}

	h.Get(c)
}

// Delete handles DELETE /teams/:id (admin).
func (h *TeamHandler) Delete(c *gin.Context) {
	teamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.ErrorJSON(c, http.StatusBadRequest, "invalid team id")
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	cmd, err := h.pool.Exec(ctx, `DELETE FROM teams WHERE id = $1`, teamID)
	if err != nil {
		httputil.ErrorJSON(c, http.StatusInternalServerError, "failed to delete team")
		return
	}
	if cmd.RowsAffected() == 0 {
		httputil.ErrorJSON(c, http.StatusNotFound, "team not found")
		return
	}
	c.Status(http.StatusNoContent)
}

// ListMembers handles GET /teams/:id/members.
func (h *TeamHandler) ListMembers(c *gin.Context) {
	teamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.ErrorJSON(c, http.StatusBadRequest, "invalid team id")
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	if !middleware.IsAdmin(c) {
		ok, err := h.userMemberOfTeam(ctx, middleware.UserID(c), teamID)
		if err != nil || !ok {
			httputil.ErrorJSON(c, http.StatusForbidden, "not a member of this team")
			return
		}
	}

	rows, err := h.pool.Query(ctx, `
		SELECT u.id, u.email, u.name, tm.role
		FROM team_members tm
		INNER JOIN users u ON u.id = tm.user_id
		WHERE tm.team_id = $1
		ORDER BY u.email`, teamID)
	if err != nil {
		httputil.ErrorJSON(c, http.StatusInternalServerError, "failed to list members")
		return
	}
	defer rows.Close()

	list := make([]gin.H, 0)
	for rows.Next() {
		var uid uuid.UUID
		var email, name, role string
		if err := rows.Scan(&uid, &email, &name, &role); err != nil {
			httputil.ErrorJSON(c, http.StatusInternalServerError, "failed to read members")
			return
		}
		list = append(list, gin.H{
			"userId": uid.String(), "email": email, "name": name, "role": role,
		})
	}
	httputil.DataJSON(c, http.StatusOK, list)
}

// AddMember handles POST /teams/:id/members (admin).
func (h *TeamHandler) AddMember(c *gin.Context) {
	teamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.ErrorJSON(c, http.StatusBadRequest, "invalid team id")
		return
	}
	var body addMemberBody
	if err := c.ShouldBindJSON(&body); err != nil {
		httputil.ErrorJSON(c, httputil.StatusUnprocessable, "invalid JSON body")
		return
	}
	userID, err := uuid.Parse(strings.TrimSpace(body.UserID))
	if err != nil {
		httputil.ErrorJSON(c, http.StatusBadRequest, "invalid userId")
		return
	}
	role := strings.ToLower(strings.TrimSpace(body.Role))
	if role == "" {
		role = "member"
	}
	if !validateTeamRole(role) {
		httputil.ErrorJSON(c, httputil.StatusUnprocessable, "role must be admin, member, or viewer")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	_, err = h.pool.Exec(ctx, `
		INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, $3)
		ON CONFLICT (team_id, user_id) DO UPDATE SET role = EXCLUDED.role`,
		teamID, userID, role)
	if err != nil {
		httputil.ErrorJSON(c, http.StatusBadRequest, "could not add member (team or user missing?)")
		return
	}
	httputil.DataJSON(c, http.StatusOK, gin.H{"teamId": teamID.String(), "userId": userID.String(), "role": role})
}

// RemoveMember handles DELETE /teams/:id/members/:userId (admin).
func (h *TeamHandler) RemoveMember(c *gin.Context) {
	teamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.ErrorJSON(c, http.StatusBadRequest, "invalid team id")
		return
	}
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		httputil.ErrorJSON(c, http.StatusBadRequest, "invalid user id")
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	cmd, err := h.pool.Exec(ctx, `DELETE FROM team_members WHERE team_id = $1 AND user_id = $2`, teamID, userID)
	if err != nil {
		httputil.ErrorJSON(c, http.StatusInternalServerError, "failed to remove member")
		return
	}
	if cmd.RowsAffected() == 0 {
		httputil.ErrorJSON(c, http.StatusNotFound, "membership not found")
		return
	}
	c.Status(http.StatusNoContent)
}
