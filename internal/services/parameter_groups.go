package services

import (
	"context"

	"github.com/google/uuid"

	"github.com/pzonouz/pzonouz-caroption-back-golang/internal/utils"
)

func (s *Service) ListParameterGroups() ([]ParameterGroup, error) {
	query := `SELECT p.id,p.name,p.Entity_id,c.name,p.created_at 
	FROM parameter_groups AS p
	LEFT JOIN categories c ON p.Entity_id=c.id;
	`

	rows, err := s.db.Query(context.Background(), query)
	if err != nil {
		return []ParameterGroup{}, err
	}
	defer rows.Close()

	var parameterGroups []ParameterGroup

	for rows.Next() {
		var parameterGroup ParameterGroup
		if err := rows.Scan(&parameterGroup.ID, &parameterGroup.Name, &parameterGroup.EntityId, &parameterGroup.EntityName, &parameterGroup.CreatedAt); err != nil {
			return []ParameterGroup{}, err
		}

		parameterGroups = append(parameterGroups, parameterGroup)
	}

	return parameterGroups, nil
}

func (s *Service) GetParameterGroup(id string) (ParameterGroup, error) {
	var parameterGroup ParameterGroup

	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return parameterGroup, err
	}

	query := "SELECT id,name,Entity_id,created_at FROM parameter_groups WHERE id=$1"
	row := s.db.QueryRow(context.Background(), query, parsedUUID)

	err = row.Scan(
		&parameterGroup.ID,
		&parameterGroup.Name,
		&parameterGroup.EntityId,
		&parameterGroup.CreatedAt,
	)
	if err != nil {
		return parameterGroup, err
	}

	return parameterGroup, nil
}

func (s *Service) CreateParameterGroup(parameterGroup ParameterGroup) error {
	query := "INSERT INTO parameter_groups (id,name,Entity_id) VALUES ($1,$2,$3);"
	validate := utils.NewValidate()

	err := validate.Struct(parameterGroup)
	if err != nil {
		return err
	}

	id := uuid.New()

	_, err = s.db.Exec(
		context.Background(),
		query,
		id,
		parameterGroup.Name,
		parameterGroup.EntityId,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) EditParameterGroup(id string, parameterGroup ParameterGroup) error {
	query := "UPDATE parameter_groups SET name=$1,Entity_id=$2 WHERE id=$3;"
	validate := utils.NewValidate()

	err := validate.Struct(parameterGroup)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(
		context.Background(),
		query,
		parameterGroup.Name,
		parameterGroup.EntityId,
		id,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) DeleteParameterGroup(id string) error {
	query := "DELETE FROM parameter_groups WHERE id=$1"

	_, err := s.db.Exec(context.Background(), query, id)
	if err != nil {
		return err
	}

	return nil
}
