package rbac

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Valid permission strings for cluster_permissions.permission.
const (
	PermRead  = "read"
	PermWrite = "write"
	PermAdmin = "admin"
)

// Rank maps permission to numeric strength (higher is more privileged).
func Rank(p string) int {
	switch p {
	case PermRead:
		return 1
	case PermWrite:
		return 2
	case PermAdmin:
		return 3
	default:
		return 0
	}
}

// ParsePermission returns an error if p is not a known permission.
func ParsePermission(p string) error {
	if Rank(p) == 0 {
		return fmt.Errorf("permission must be read, write, or admin")
	}
	return nil
}

// UserIsAdmin returns true if the user row has is_admin.
func UserIsAdmin(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) (bool, error) {
	var admin bool
	err := pool.QueryRow(ctx, `SELECT COALESCE(is_admin, false) FROM users WHERE id = $1`, userID).Scan(&admin)
	if err != nil {
		return false, err
	}
	return admin, nil
}

// UserClusterAccess returns the highest permission rank the user has for cluster/namespace
// via team membership and cluster_permissions. Admins get PermAdmin rank without rows.
// Namespace matches exact key, or permission row namespace '*' (all namespaces).
func UserClusterAccess(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID, clusterID, namespace string) (maxRank int, err error) {
	admin, err := UserIsAdmin(ctx, pool, userID)
	if err != nil {
		return 0, err
	}
	if admin {
		return Rank(PermAdmin), nil
	}
	const q = `
SELECT cp.permission
FROM cluster_permissions cp
INNER JOIN team_members tm ON tm.team_id = cp.team_id
WHERE tm.user_id = $1
  AND cp.cluster_id = $2
  AND (cp.namespace = $3 OR cp.namespace = '*')
`
	rows, err := pool.Query(ctx, q, userID, clusterID, namespace)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	for rows.Next() {
		var perm string
		if err := rows.Scan(&perm); err != nil {
			return 0, err
		}
		if r := Rank(perm); r > maxRank {
			maxRank = r
		}
	}
	return maxRank, rows.Err()
}

// UserSatisfiesPermission returns true if effective rank >= required permission.
func UserSatisfiesPermission(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID, clusterID, namespace, required string) (bool, error) {
	if err := ParsePermission(required); err != nil {
		return false, err
	}
	need := Rank(required)
	got, err := UserClusterAccess(ctx, pool, userID, clusterID, namespace)
	if err != nil {
		return false, err
	}
	return got >= need, nil
}
