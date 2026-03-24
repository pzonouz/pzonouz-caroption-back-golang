package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

func (s *Service) ListPersonsWithSortFilterPagination(
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
		if sort == "parent_name" {
			orderBy = fmt.Sprintf(
				`ORDER BY p.name COLLATE "fa-IR-x-icu" %s`,
				sortDirection,
			)
		} else {
			orderBy = fmt.Sprintf(
				`ORDER BY persons.%s COLLATE "fa-IR-x-icu" %s`,
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
			if filter == "parent_name" {
				filterBy.WriteString(fmt.Sprintf(
					`p.name %s '%s'`,
					filterOperand,
					filterCondition,
				))
			} else {
				filterBy.WriteString(fmt.Sprintf(
					`%s %s '%s'`,
					"persons."+filter,
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
		    persons.id,
		    persons.first_name,
		    persons.name,
		    persons.address,
		    persons.phone_number,
		    persons.created_at,
		    persons.updated_at
		FROM
		    persons
		%s %s %s
		`, filterBy.String(), orderBy, pagedBy)

	rows, err := s.db.Query(context.Background(), query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
	defer rows.Close()

	var persons []Person

	for rows.Next() {
		var person Person
		if err := rows.Scan(&person.ID, &person.FirstName, &person.Name, &person.Address, &person.PhoneNumber, &person.CreatedAt, &person.UpdatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		persons = append(persons, person)
	}

	w.Header().Add("Content-Type", "application/json")

	newQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM
		    persons
		%s
		`, filterBy.String())
	row := s.db.QueryRow(context.Background(), newQuery)

	var Count int32

	err = row.Scan(&Count)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	var personsWithTotalCount struct {
		Rows       []Person `json:"rows"`
		TotalCount int32    `json:"totalCount"`
	}

	personsWithTotalCount.Rows = persons
	personsWithTotalCount.TotalCount = Count

	err = json.NewEncoder(w).Encode(personsWithTotalCount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
}

func (s *Service) GetPerson(id string) (Person, error) {
	var person Person

	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return person, err
	}

	query := "SELECT * FROM persons WHERE id=$1"
	row := s.db.QueryRow(context.Background(), query, parsedUUID)

	err = row.Scan(
		&person.ID,
		&person.FirstName,
		&person.Name,
		&person.Address,
		&person.PhoneNumber,
		&person.CreatedAt,
		&person.UpdatedAt,
	)
	if err != nil {
		return person, err
	}

	return person, nil
}

func (s *Service) CreatePerson(person Person) error {
	query := `
		INSERT INTO persons (id, first_name, name, address, phone_number)
		    VALUES ($1, $2, $3, $4, $5)`

	id := uuid.New()

	_, err := s.db.Exec(
		context.Background(),
		query,
		id,
		person.FirstName,
		person.Name,
		person.Address,
		person.PhoneNumber,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) EditPerson(id string, person Person) error {
	query := "UPDATE persons SET first_name=$1,name=$2,address=$3,phone_number=$4 WHERE id=$5;"

	_, err := s.db.Exec(
		context.Background(),
		query,
		person.FirstName,
		person.Name,
		person.Address,
		person.PhoneNumber,
		id,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) DeletePerson(id string) error {
	query := "DELETE FROM persons WHERE id=$1"

	_, err := s.db.Exec(context.Background(), query, id)
	if err != nil {
		return err
	}

	return nil
}
