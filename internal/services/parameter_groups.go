package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/pzonouz/pzonouz-caroption-back-golang/internal/utils"
)

func (s *Service) ListParameterGroupsWithSortFilterPagination(
	sort string,
	sortDirection string,
	filters []string,
	filterOperands []string,
	filterConditions []string,
	countInPage string,
	offset string,
	w http.ResponseWriter,
) {
	orderBy := ""

	if sort != "" {
		if sort == "category_name" {
			orderBy = fmt.Sprintf(
				`ORDER BY categories.name COLLATE "fa-IR-x-icu" %s`,
				sortDirection,
			)
		} else {
			orderBy = fmt.Sprintf(
				`ORDER BY parameter_groups.%s COLLATE "fa-IR-x-icu" %s`,
				sort,
				sortDirection,
			)
		}
	}

	var filterBy strings.Builder
	if len(filters) > 0 {
		filterBy.WriteString("WHERE ")
	}
	for index, filter := range filters {
		filterOperand := filterOperands[index]
		filterCondition := filterConditions[index]

		if filterOperand == "contains" {
			filterOperand = "LIKE"

			filterCondition = "%" + filterCondition + "%"
		}

		if len(filters) != 0 {
			if filter == "category_name" {
				filterBy.WriteString(fmt.Sprintf(
					`categories.name %s '%s'`,
					// filter,
					filterOperand,
					filterCondition,
				))
			} else {
				filterBy.WriteString(fmt.Sprintf(
					`%s %s '%s'`,
					"parameter_groups."+filter,
					filterOperand,
					filterCondition,
				))
			}
		}

		if len(filters)-1 > index {
			filterBy.WriteString(" AND ")
		}
	}

	pagedBy := ""
	offsetNum := 0

	if countInPage != "" {
		limit, err := strconv.Atoi(countInPage)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		if offset != "" {
			offsetNum, err = strconv.Atoi(offset)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)

				return
			}
		}

		pagedBy = fmt.Sprintf(`LIMIT %d OFFSET %d`, limit, offsetNum)
	}

	query := fmt.Sprintf(`
		SELECT
		    parameter_groups.id,
		    parameter_groups.name,
		    parameter_groups.category_id,
		    categories.name,
		    parameter_groups.created_at
		FROM
		    parameter_groups
		    LEFT JOIN categories ON parameter_groups.category_id = categories.id
		%s %s %s
		`, filterBy.String(), orderBy, pagedBy)

	rows, err := s.db.Query(context.Background(), query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
	defer rows.Close()

	var parameterGroups []ParameterGroup

	for rows.Next() {
		var parameterGroup ParameterGroup
		if err := rows.Scan(&parameterGroup.ID, &parameterGroup.Name, &parameterGroup.CategoryId, &parameterGroup.CategoryName, &parameterGroup.CreatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		parameterGroups = append(parameterGroups, parameterGroup)
	}

	w.Header().Add("Content-Type", "application/json")

	newQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM
		    parameter_groups
		%s
		`, filterBy.String())
	row := s.db.QueryRow(context.Background(), newQuery)

	var Count int32

	err = row.Scan(&Count)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	var parameterGroupsWithTotalCount struct {
		Rows       []ParameterGroup `json:"rows"`
		TotalCount int32            `json:"totalCount"`
	}

	parameterGroupsWithTotalCount.Rows = parameterGroups
	parameterGroupsWithTotalCount.TotalCount = Count

	err = json.NewEncoder(w).Encode(parameterGroupsWithTotalCount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
}

func (s *Service) GetParameterGroup(id string) (ParameterGroup, error) {
	var parameterGroup ParameterGroup

	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return parameterGroup, err
	}

	query := "SELECT id,name,category_id,created_at FROM parameter_groups WHERE id=$1"
	row := s.db.QueryRow(context.Background(), query, parsedUUID)

	err = row.Scan(
		&parameterGroup.ID,
		&parameterGroup.Name,
		&parameterGroup.CategoryId,
		&parameterGroup.CreatedAt,
	)
	if err != nil {
		return parameterGroup, err
	}

	return parameterGroup, nil
}

func (s *Service) CreateParameterGroup(parameterGroup ParameterGroup) error {
	query := "INSERT INTO parameter_groups (id,name,category_id) VALUES ($1,$2,$3);"
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
		parameterGroup.CategoryId,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) EditParameterGroup(id string, parameterGroup ParameterGroup) error {
	query := "UPDATE parameter_groups SET name=$1,category_id=$2 WHERE id=$3;"
	validate := utils.NewValidate()

	err := validate.Struct(parameterGroup)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(
		context.Background(),
		query,
		parameterGroup.Name,
		parameterGroup.CategoryId,
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
