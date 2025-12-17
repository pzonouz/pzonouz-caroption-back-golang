package services

import (
	"context"

	"github.com/google/uuid"
)

func (s *Service) ListPersons() ([]Person, error) {
	query := `
	SELECT
	    * 
	FROM
	    persons`

	rows, err := s.db.Query(context.Background(), query)
	if err != nil {
		return []Person{}, err
	}
	defer rows.Close()

	var persons []Person

	for rows.Next() {
		var person Person
		if err := rows.Scan(&person.ID, &person.FirstName, &person.LastName, &person.Address, &person.PhoneNumber, &person.CreatedAt, &person.UpdatedAt); err != nil {
			return []Person{}, err
		}

		persons = append(persons, person)
	}

	return persons, nil
}

func (s *Service) GetPerson(id string) (Person, error) {
	var person Person

	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return person, err
	}

	query := "SELECT * FROM persons WHERE id=$1"
	row := s.db.QueryRow(context.Background(), query, parsedUUID)

	err = row.Scan(&person.ID, &person.FirstName, &person.LastName, &person.Address, &person.PhoneNumber, &person.CreatedAt, &person.UpdatedAt)
	if err != nil {
		return person, err
	}

	return person, nil
}

func (s *Service) CreatePerson(person Person) error {
	query := `
		INSERT INTO persons (id, first_name, last_name, address, phone_number)
		    VALUES ($1, $2, $3, $4, $5)`

	id := uuid.New()

	_, err := s.db.Exec(context.Background(), query, id, person.FirstName, person.LastName, person.Address, person.PhoneNumber)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) EditPerson(id string, person Person) error {
	query := "UPDATE persons SET first_name=$1,last_name=$2,address=$3,phone_number=$4 WHERE id=$5;"

	_, err := s.db.Exec(context.Background(), query, person.FirstName, person.LastName, person.Address, person.PhoneNumber, id)
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
