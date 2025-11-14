package services

import (
	"context"

	"github.com/google/uuid"

	"github.com/pzonouz/pzonouz-caroption-back-golang/internal/utils"
)

func (s *Service) ListParentEntities() ([]Entity, error) {
	query := `
		SELECT
		    p.id,
		    p.name,
		    p.parent_id,
		    p.description,
		    p.priority,
		    i.image_url,
		    p.image_id,
		    p.entity_slug,
		    p.show,
		    p.created_at,
		    p.updated_at
		FROM
		    entities p
		    LEFT JOIN entities c ON c.parent_id = p.id
		    LEFT JOIN images i ON p.image_id = i.id
		WHERE
		    p.parent_id IS NULL
		GROUP BY
		    p.id,
		    p.name,
		    p.parent_id,
		    p.created_at,
		    i.image_url
		ORDER BY
		    p.priority`

	rows, err := s.db.Query(context.Background(), query)
	if err != nil {
		return []Entity{}, err
	}
	defer rows.Close()

	var entities []Entity

	for rows.Next() {
		var entity Entity
		if err := rows.Scan(&entity.ID, &entity.Name, &entity.ParentID, &entity.Description, &entity.Priority, &entity.ImageUrl, &entity.ImageID, &entity.EntitySlug, &entity.Show, &entity.CreatedAt, &entity.UpdatedAt); err != nil {
			return []Entity{}, err
		}

		entities = append(entities, entity)
	}

	if err := rows.Err(); err != nil {
		return []Entity{}, err
	}

	return entities, nil
}

func (s *Service) ListEntities() ([]Entity, error) {
	query := `
		SELECT
		    c.id,
		    c.name,
		    c.parent_id,
		    p.name AS parent_name,
		    c.description,
		    c.priority,
		    c.image_id,
		    i.image_url,
		    c.entity_slug,
		    c.show,
		    c.created_at,
		    c.updated_at
		FROM
		    entities AS c
		    LEFT JOIN entities p ON c.parent_id = p.id
		    LEFT JOIN images i ON c.image_id = i.id
		GROUP BY
		    c.id,
		    i.image_url,
		    p.name;
		
		`

	rows, err := s.db.Query(context.Background(), query)
	if err != nil {
		return []Entity{}, err
	}
	defer rows.Close()

	var entities []Entity

	for rows.Next() {
		var entity Entity
		if err := rows.Scan(&entity.ID, &entity.Name, &entity.ParentID, &entity.ParentName, &entity.Description, &entity.Priority, &entity.ImageID, &entity.ImageUrl, &entity.EntitySlug, &entity.Show, &entity.CreatedAt, &entity.UpdatedAt); err != nil {
			return []Entity{}, err
		}

		// entity.Children = []ChildEntity{}

		entities = append(entities, entity)
	}

	return entities, nil
}

func (s *Service) GetEntity(id string) (Entity, error) {
	var Entity Entity

	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return Entity, err
	}

	query := `
		SELECT
		    id,
		    name,
		    parent_id,
		    description,
		    priority,
		    image_id,
		    entity_slug,
		    show,
		    created_at,
		    updated_at
		FROM
		    entities
		WHERE
		    id = $1`
	row := s.db.QueryRow(context.Background(), query, parsedUUID)

	err = row.Scan(
		&Entity.ID,
		&Entity.Name,
		&Entity.ParentID,
		&Entity.Description,
		&Entity.Priority,
		&Entity.ImageID,
		&Entity.EntitySlug,
		&Entity.Show,
		&Entity.CreatedAt,
		&Entity.UpdatedAt,
	)
	if err != nil {
		return Entity, err
	}

	return Entity, nil
}

func (s *Service) GetEntityBySlug(slug string) (Entity, error) {
	var Entity Entity

	query := `
		SELECT
		    id,
		    name,
		    parent_id,
		    description,
		    priority,
		    image_id,
		    entity_slug,
		    show,
		    created_at,
		    updated_at
		FROM
		    entities
		WHERE
		    entity_slug = $1`
	row := s.db.QueryRow(context.Background(), query, slug)

	err := row.Scan(
		&Entity.ID,
		&Entity.Name,
		&Entity.ParentID,
		&Entity.Description,
		&Entity.Priority,
		&Entity.ImageID,
		&Entity.EntitySlug,
		&Entity.Show,
		&Entity.CreatedAt,
		&Entity.UpdatedAt,
	)
	if err != nil {
		return Entity, err
	}

	return Entity, nil
}

func (s *Service) CreateEntity(Entity Entity) error {
	query := `
		INSERT INTO entities (id, name, parent_id, description, priority, image_id, entity_slug, show)
		    VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	validate := utils.NewValidate()

	err := validate.Struct(Entity)
	if err != nil {
		return err
	}

	id := uuid.New()

	_, err = s.db.Exec(
		context.Background(),
		query,
		id,
		Entity.Name,
		Entity.ParentID,
		Entity.Description,
		Entity.Priority,
		Entity.ImageID,
		Entity.EntitySlug,
		Entity.Show,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) EditEntity(id string, Entity Entity) error {
	query := `
		UPDATE
		    entities
		SET
		    name = $1,
		    parent_id = $2,
		    image_id = $3,
		    priority = $4,
		    entity_slug = $5,
		    show = $6
		WHERE
		    id = $7`
	validate := utils.NewValidate()

	err := validate.Struct(Entity)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(
		context.Background(),
		query,
		Entity.Name,
		Entity.ParentID,
		Entity.ImageID,
		Entity.Priority,
		Entity.EntitySlug,
		Entity.Show,
		id,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) DeleteEntity(id string) error {
	query := `
		DELETE FROM entities
		WHERE id = $1`

	_, err := s.db.Exec(context.Background(), query, id)
	if err != nil {
		return err
	}

	return nil
}
