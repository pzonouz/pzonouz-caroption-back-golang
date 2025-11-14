package services

import (
	"context"

	"github.com/google/uuid"

	"github.com/pzonouz/pzonouz-caroption-back-golang/internal/utils"
)

func (s *Service) ListParameters() ([]Parameter, error) {
	query := `SELECT p.id,p.name,p.description,p.type,p.parameter_group_id,p.selectables,p.priority,p.created_at 
	FROM parameters AS p ORDER BY p.priority`

	rows, err := s.db.Query(context.Background(), query)
	if err != nil {
		return []Parameter{}, err
	}
	defer rows.Close()

	var parameters []Parameter

	for rows.Next() {
		var parameter Parameter
		if err := rows.Scan(&parameter.ID, &parameter.Name, &parameter.Description, &parameter.Type, &parameter.ParameterGroupId, &parameter.Selectables, &parameter.Priority, &parameter.CreatedAt); err != nil {
			return []Parameter{}, err
		}

		parameters = append(parameters, parameter)
	}

	return parameters, nil
}

func (s *Service) ListParametersByEntity(Entity_id string) ([]Parameter, error) {
	query := `
	SELECT
	  p.id,
	  p.name,
	  p.description,
	  p.type,
	  p.parameter_group_id,
	  p.selectables,
	  p.priority,
	  p.created_at
	FROM parameters AS p
	JOIN parameter_groups AS pg ON pg.id = p.parameter_group_id
	JOIN categories AS c ON c.id = pg.Entity_id
	WHERE c.id IN (SELECT cc.parent_id FROM categories as cc WHERE cc.id = $1 )
	ORDER BY p.priority
	`

	rows, err := s.db.Query(context.Background(), query, Entity_id)
	if err != nil {
		return []Parameter{}, err
	}
	defer rows.Close()

	var parameters []Parameter

	for rows.Next() {
		var parameter Parameter
		if err := rows.Scan(&parameter.ID, &parameter.Name, &parameter.Description, &parameter.Type, &parameter.ParameterGroupId, &parameter.Selectables, &parameter.Priority, &parameter.CreatedAt); err != nil {
			return []Parameter{}, err
		}

		parameters = append(parameters, parameter)
	}

	return parameters, nil
}

func (s *Service) GetParameter(id string) (Parameter, error) {
	var parameter Parameter

	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return parameter, err
	}

	query := `SELECT id,name,description,type,parameter_group_id,selectables,priority,created_at FROM parameters WHERE id=$1`
	row := s.db.QueryRow(context.Background(), query, parsedUUID)

	err = row.Scan(
		&parameter.ID,
		&parameter.Name,
		&parameter.Description,
		&parameter.Type,
		&parameter.ParameterGroupId,
		&parameter.Selectables,
		&parameter.Priority,
		&parameter.CreatedAt,
	)
	if err != nil {
		return parameter, err
	}

	return parameter, nil
}

func (s *Service) CreateParameter(parameter Parameter) error {
	query := `INSERT INTO parameters (id,name,description,type,parameter_group_id,selectables,priority) VALUES ($1,$2,$3,$4,$5,$6,$7)`
	validate := utils.NewValidate()

	err := validate.Struct(parameter)
	if err != nil {
		return err
	}

	id := uuid.New()

	_, err = s.db.Exec(
		context.Background(),
		query,
		id,
		parameter.Name,
		parameter.Description,
		parameter.Type,
		parameter.ParameterGroupId,
		parameter.Selectables,
		parameter.Priority,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) EditParameter(id string, parameter Parameter) error {
	query := `UPDATE parameters SET name=$1,description=$2,type=$3,parameter_group_id=$4,selectables=$5,priority=$6 WHERE id=$7`
	validate := utils.NewValidate()

	err := validate.Struct(parameter)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(
		context.Background(),
		query,
		parameter.Name,
		parameter.Description,
		parameter.Type,
		parameter.ParameterGroupId,
		parameter.Selectables,
		parameter.Priority,
		id,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) DeleteParameter(id string) error {
	query := "DELETE FROM parameters WHERE id=$1"

	_, err := s.db.Exec(context.Background(), query, id)
	if err != nil {
		return err
	}

	return nil
}
