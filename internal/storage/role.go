package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	ssoerrors "sso-service/internal/lib/errors"
	"sso-service/internal/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type RoleDataBase struct {
	db *Database
	// TO DO: Redis Cache
}

func NewRoleDataBase(db *Database) *RoleDataBase {
	return &RoleDataBase{
		db: db,
	}
}

// CheckPermission проверяет, есть ли у субъекта доступ к объекту с данным разрешением
func (roleDB *RoleDataBase) CheckPermission(
	ctx context.Context,
	subjectID, objectID uuid.UUID,
	permissionName string,
) (bool, error) {
	const op = "storage.role.CheckPermission"

	var permID uuid.UUID
	err := roleDB.db.GetContext(ctx, &permID,
		"SELECT id FROM permissions WHERE name = $1", permissionName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, fmt.Errorf("%s: %w", op, ssoerrors.ErrPermissionNotFound)
		}
		return false, fmt.Errorf("%s: %w", op, err)
	}

	found, err := roleDB.bfsCheck(ctx, subjectID, objectID, permID)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return found, nil
}

func (roleDB *RoleDataBase) bfsCheck(
	ctx context.Context,
	startID, targetID, permissionID uuid.UUID,
) (bool, error) {
	const op = "storage.role.bfsCheck"

	visited := make(map[uuid.UUID]bool)
	queue := []uuid.UUID{startID}

	for len(queue) > 0 {
		currentID := queue[0]
		queue = queue[1:]

		if visited[currentID] {
			continue
		}
		visited[currentID] = true
		// Проверяем прямые отношения от currentID к targetID
		direct, err := roleDB.checkDirectRelations(ctx, currentID, targetID, permissionID)
		if err != nil {
			return false, fmt.Errorf("%s: %w", op, err)
		}
		if direct {
			return true, nil
		}

		// Получаем все отношения, где currentID является source
		relations, err := roleDB.getOutboundRelations(ctx, currentID)
		if err != nil {
			return false, fmt.Errorf("%s: %w", op, err)
		}

		// Добавляем всех "соседей" в очередь
		for _, rel := range relations {
			if !visited[rel.TargetID] {
				queue = append(queue, rel.TargetID)
			}
		}
	}

	return false, nil
}

func (roleDB *RoleDataBase) checkDirectRelations(
	ctx context.Context,
	sourceID, targetID, permissionID uuid.UUID,
) (bool, error) {
	const op = "storage.role.checkDirectRelations"

	query := `
        SELECT EXISTS(
            SELECT 1 
            FROM relations r
            JOIN permission_assignments pa ON pa.relation_type = r.relation_type
            WHERE r.source_id = $1 
            AND r.target_id = $2
            AND pa.permission_id = $3
        )
    `
	var exists bool
	err := roleDB.db.GetContext(ctx, &exists, query, sourceID, targetID, permissionID)
	if err != nil {
		return false, fmt.Errorf("failed to check direct relations: %s: %w", op, err)
	}
	return exists, nil
}

func (roleDB *RoleDataBase) getOutboundRelations(
	ctx context.Context,
	sourceID uuid.UUID,
) ([]models.Relation, error) {
	const op = "storage.role.getOutboundRelations"

	query := `
        SELECT id, source_id, target_id, relation_type, created_at
        FROM relations
        WHERE source_id = $1
    `
	var relations []models.Relation
	err := roleDB.db.SelectContext(ctx, &relations, query, sourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get outbound relations: %s: %w", op, err)
	}
	return relations, nil
}

// CreateEntity создает новую сущность
func (roleDB *RoleDataBase) CreateEntity(ctx context.Context, id uuid.UUID, entityType string) error {
	const op = "storage.role.CreateEntity"

	return roleDB.db.WithTransaction(ctx, func(tx *sqlx.Tx) error {
		query := `INSERT INTO entities (id, type) VALUES ($1, $2)`
		_, err := tx.ExecContext(ctx, query, id, entityType)
		if err != nil {
			return fmt.Errorf("failed to create entity: %s: %w", op, err)
		}
		return nil
	})
}

// DeleteEntity удаляет сущность
func (roleDB *RoleDataBase) DeleteEntity(ctx context.Context, id uuid.UUID) error {
	const op = "storage.role.DeleteEntity"

	return roleDB.db.WithTransaction(ctx, func(tx *sqlx.Tx) error {
		query := `DELETE FROM entities WHERE id = $1`
		result, err := tx.ExecContext(ctx, query, id)
		if err != nil {
			return fmt.Errorf("failed to delete entity: %s: %w", op, err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %s: %w", op, err)
		}
		if rowsAffected == 0 {
			return fmt.Errorf("%s: %w", op, ssoerrors.ErrNotFound)
		}

		return nil
	})
}

// CreateRelation создает отношение между сущностями
func (roleDB *RoleDataBase) CreateRelation(
	ctx context.Context,
	sourceID, targetID uuid.UUID,
	relationType string,
) error {
	const op = "storage.role.CreateRelation"

	return roleDB.db.WithTransaction(ctx, func(tx *sqlx.Tx) error {
		// Проверяем существование сущностей
		var exists bool
		err := tx.GetContext(ctx, &exists,
			`SELECT EXISTS(SELECT 1 FROM entities WHERE id = $1)`, sourceID)
		if err != nil || !exists {
			return fmt.Errorf("source entity not found: %s: %w", op, err)
		}

		err = tx.GetContext(ctx, &exists,
			`SELECT EXISTS(SELECT 1 FROM entities WHERE id = $1)`, targetID)
		if err != nil || !exists {
			return fmt.Errorf("target entity not found: %s: %w", op, err)
		}

		// Проверяем, не существует ли уже такое отношение
		err = tx.GetContext(ctx, &exists, `
			SELECT EXISTS(
				SELECT 1 FROM relations 
				WHERE source_id = $1 AND target_id = $2 AND relation_type = $3
			)`, sourceID, targetID, relationType)
		if err != nil {
			return fmt.Errorf("failed to check relation existence: %s: %w", op, err)
		}
		if exists {
			return fmt.Errorf("%s: %w", op, ssoerrors.ErrRelationExists)
		}

		// Создаем отношение
		query := `
			INSERT INTO relations (id, source_id, target_id, relation_type)
			VALUES ($1, $2, $3, $4)
		`
		_, err = tx.ExecContext(ctx, query, uuid.New().String(), sourceID, targetID, relationType)
		if err != nil {
			return fmt.Errorf("failed to create relation: %s: %w", op, err)
		}

		return nil
	})
}

// DeleteRelation удаляет отношение
func (roleDB *RoleDataBase) DeleteRelation(
	ctx context.Context,
	sourceID, targetID uuid.UUID,
	relationType string,
) error {
	const op = "storage.role.DeleteRelation"

	return roleDB.db.WithTransaction(ctx, func(tx *sqlx.Tx) error {
		query := `
			DELETE FROM relations 
			WHERE source_id = $1 AND target_id = $2 AND relation_type = $3
		`
		result, err := tx.ExecContext(ctx, query, sourceID, targetID, relationType)
		if err != nil {
			return fmt.Errorf("failed to delete relation: %s: %w", op, err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %s: %w", op, err)
		}
		if rowsAffected == 0 {
			return fmt.Errorf("%s: %w", op, ssoerrors.ErrNotFound)
		}

		return nil
	})
}

// AddPermission добавляет новое разрешение
func (roleDB *RoleDataBase) AddPermission(ctx context.Context, name, description string) (uuid.UUID, error) {
	const op = "storage.role.AddPermission"

	var id uuid.UUID
	err := roleDB.db.WithTransaction(ctx, func(tx *sqlx.Tx) error {
		// Проверяем, не существует ли уже разрешение с таким именем
		var exists bool
		err := tx.GetContext(ctx, &exists,
			`SELECT EXISTS(SELECT 1 FROM permissions WHERE name = $1)`, name)
		if err != nil {
			return fmt.Errorf("failed to check permission existence: %s: %w", op, err)
		}
		if exists {
			return fmt.Errorf("failed to check permission existence: %s: %w", op, ssoerrors.ErrPermissionExists)
		}

		// Создаем разрешение
		id = uuid.New()
		query := `
			INSERT INTO permissions (id, name, description)
			VALUES ($1, $2, $3)
			RETURNING id
		`
		err = tx.GetContext(ctx, &id, query, id, name, description)
		if err != nil {
			return fmt.Errorf("failed to add permission: %s: %w", op, err)
		}

		return nil
	})

	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}
	return id, nil
}

// AssignPermission назначает разрешение для типа отношения
func (roleDB *RoleDataBase) AssignPermission(ctx context.Context, permissionID uuid.UUID, relationType string) error {
	const op = "storage.role.AssignPermission"

	return roleDB.db.WithTransaction(ctx, func(tx *sqlx.Tx) error {
		// Проверяем существование разрешения
		var exists bool
		err := tx.GetContext(ctx, &exists,
			`SELECT EXISTS(SELECT 1 FROM permissions WHERE id = $1)`, permissionID)
		if err != nil || !exists {
			return fmt.Errorf("permission not found: %s: %w", op, err)
		}

		// Проверяем, не назначено ли уже это разрешение
		err = tx.GetContext(ctx, &exists, `
			SELECT EXISTS(
				SELECT 1 FROM permission_assignments 
				WHERE relation_type = $1 AND permission_id = $2
			)`, relationType, permissionID)
		if err != nil {
			return fmt.Errorf("failed to check assignment existence: %s: %w", op, err)
		}
		if exists {
			return fmt.Errorf("%s: %w", op, ssoerrors.ErrPermissionExists)
		}

		// Назначаем разрешение
		query := `
			INSERT INTO permission_assignments (relation_type, permission_id)
			VALUES ($1, $2)
		`
		_, err = tx.ExecContext(ctx, query, relationType, permissionID)
		if err != nil {
			return fmt.Errorf("failed to assign permission: %s: %w", op, err)
		}

		return nil
	})
}

// RevokePermission отзывает разрешение у типа отношения
func (roleDB *RoleDataBase) RevokePermission(ctx context.Context, permissionID uuid.UUID, relationType string) error {
	const op = "storage.role.RevokePermission"

	return roleDB.db.WithTransaction(ctx, func(tx *sqlx.Tx) error {
		// Проверяем существование разрешения
		var exists bool
		err := tx.GetContext(ctx, &exists,
			`SELECT EXISTS(SELECT 1 FROM permissions WHERE id = $1)`, permissionID)
		if err != nil || !exists {
			return fmt.Errorf("%s: %w", op, ssoerrors.ErrPermissionNotFound)
		}

		// Проверяем, существует ли такое назначение
		err = tx.GetContext(ctx, &exists, `
            SELECT EXISTS(
                SELECT 1 FROM permission_assignments 
                WHERE relation_type = $1 AND permission_id = $2
            )`, relationType, permissionID)
		if err != nil {
			return fmt.Errorf("failed to check assignment existence: %s: %w", op, err)
		}
		if !exists {
			return fmt.Errorf("%s: %w", op, ssoerrors.ErrPermissionNotAssigned)
		}

		// Отзываем разрешение
		query := `
            DELETE FROM permission_assignments 
            WHERE relation_type = $1 AND permission_id = $2
        `
		result, err := tx.ExecContext(ctx, query, relationType, permissionID)
		if err != nil {
			return fmt.Errorf("failed to revoke permission: %s: %w", op, err)
		}

		// Проверяем, что была удалена хотя бы одна запись
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %s: %w", op, err)
		}
		if rowsAffected == 0 {
			return ssoerrors.ErrPermissionNotAssigned
		}

		return nil
	})
}

// GetAllPermissions возвращает все разрешения в системе
func (roleDB *RoleDataBase) GetAllPermissions(ctx context.Context) (*[]models.Permission, error) {
	const op = "storage.role.GetAllPermissions"

	var permissions []models.Permission
	query := `SELECT id, name, description, created_at FROM permissions`
	err := roleDB.db.SelectContext(ctx, &permissions, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all permissions: %s: %w", op, err)
	}
	return &permissions, nil
}

// GetUserRelations возвращает все отношения пользователя
func (roleDB *RoleDataBase) GetUserRelations(ctx context.Context, userID uuid.UUID) (*[]models.Relation, error) {
	const op = "storage.role.GetUserRelations"

	var relations []models.Relation
	query := `
		SELECT id, source_id, target_id, relation_type, created_at 
		FROM relations 
		WHERE source_id = $1
	`
	err := roleDB.db.SelectContext(ctx, &relations, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user relations: %s: %w", op, err)
	}
	return &relations, nil
}

// GetPermissionByCode возвращает разрешение по name

func (roleDB *RoleDataBase) GetPermissionByName(
	ctx context.Context,
	name string,
) (*models.Permission, error) {
	const op = "storage.role.GetPermissionByName"

	var perm models.Permission
	err := roleDB.db.GetContext(ctx, &perm,
		"SELECT id, name, description FROM permissions WHERE name = $1", name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrPermissionNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &perm, nil
}

// GetPermissionsForRelationType возвращает разрешения для типа отношения
func (roleDB *RoleDataBase) GetPermissionsForRelationType(
	ctx context.Context,
	relationType string,
) (*[]models.Permission, error) {
	const op = "storage.role.GetPermissionsForRelationType"

	var permissions []models.Permission
	query := `
		SELECT p.id, p.name, p.description, p.created_at
		FROM permissions p
		JOIN permission_assignments pa ON pa.permission_id = p.id
		WHERE pa.relation_type = $1
	`
	err := roleDB.db.SelectContext(ctx, &permissions, query, relationType)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions for relation type: %s: %w", op, err)
	}
	return &permissions, nil
}

// GetUserPermissions возвращает все разрешения пользователя
func (roleDB *RoleDataBase) GetUserPermissions(ctx context.Context, userID uuid.UUID) (*[]models.Permission, error) {
	const op = "storage.role.GetUserPermissions"

	query := `
		SELECT DISTINCT p.id, p.name, p.description, p.created_at
		FROM relations r
		JOIN permission_assignments pa ON pa.relation_type = r.relation_type
		JOIN permissions p ON p.id = pa.permission_id
		WHERE r.source_id = $1
	`

	var permissions []models.Permission
	err := roleDB.db.SelectContext(ctx, &permissions, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %s: %w", op, err)
	}

	return &permissions, nil
}

// GetEntityRelations возвращает все отношения для сущности
func (roleDB *RoleDataBase) GetEntityRelations(ctx context.Context, entityID uuid.UUID) (*[]models.Relation, error) {
	const op = "storage.role.GetEntityRelations"

	query := `
		SELECT id, source_id, target_id, relation_type, created_at
		FROM relations
		WHERE source_id = $1 OR target_id = $1
	`

	var relations []models.Relation
	err := roleDB.db.SelectContext(ctx, &relations, query, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get entity relations: %s: %w", op, err)
	}

	return &relations, nil
}
