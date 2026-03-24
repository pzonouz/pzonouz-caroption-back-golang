package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Service) CreateInvoice(inv Invoice) error {
	tx, err := s.db.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return err
	}

	query := `
		INSERT INTO invoices (id, person_id, type, discount, notes,date, created_at, updated_at)
		    VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	invoiceId := uuid.New()

	_, err = tx.Exec(context.Background(), query,
		invoiceId,
		inv.PersonID,
		inv.Type,
		inv.Discount,
		inv.Notes,
		inv.Date,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		tx.Rollback(context.Background())

		return err
	}

	for _, item := range inv.Items {
		itemQuery := `
			INSERT INTO invoice_items (id, invoice_id, description, price, product_id, count, discount)
			    VALUES ($1, $2, $3, $4, $5, $6, $7)`
		id := uuid.New()

		_, err = tx.Exec(context.Background(), itemQuery,
			id,
			invoiceId,
			item.Description,
			item.Price,
			item.ProductID,
			item.Count,
			item.Discount,
		)
		if err != nil {
			tx.Rollback(context.Background())

			return err
		}
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) GetInvoice(id string) (Invoice, error) {
	query := `
	SELECT
    invoices.id,
    invoices.person_id,
    CONCAT(persons.name, ' ', persons.first_name) AS person_name,
    invoices.type,
    invoices.discount,
    invoices.notes,
    invoices.number,
    COALESCE(json_agg(
        json_build_object(
            'id', invoice_items.id,
            'invoiceID', invoice_items.invoice_id,
            'description', invoice_items.description,
            'price', invoice_items.price,
            'productID', invoice_items.product_id,
            'productName', products.name,
            'count', invoice_items.count
        )
    )FILTER (WHERE invoice_items.id IS NOT NULL),'[]') AS items,
    invoices.date,
    invoices.created_at,
    invoices.updated_at
FROM
    invoices
    LEFT JOIN persons ON invoices.person_id = persons.id
    LEFT JOIN invoice_items ON invoices.id = invoice_items.invoice_id
    LEFT JOIN products ON invoice_items.product_id = products.id
WHERE
    invoices.id = $1
GROUP BY
    invoices.id,
    persons.name,
    persons.first_name;
	`

	var inv Invoice

	err := s.db.QueryRow(context.Background(), query, id).Scan(
		&inv.ID,
		&inv.PersonID,
		&inv.PersonName,
		&inv.Type,
		&inv.Discount,
		&inv.Notes,
		&inv.Number,
		&inv.Items,
		&inv.Date,
		&inv.CreatedAt,
		&inv.UpdatedAt,
	)
	if err != nil {
		return Invoice{}, err
	}

	return inv, nil
}

func (s *Service) ListInvoicesWithSortFilterPagination(
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
				`ORDER BY invoices.%s COLLATE "fa-IR-x-icu" %s`,
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
					"invoices."+filter,
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
    invoices.id,
    invoices.person_id,
    CONCAT(persons.name, ' ', persons.first_name) AS person_name,
    invoices.type,
    invoices.discount,
    invoices.notes,
    invoices.number,
    COALESCE(json_agg(
        json_build_object(
            'id', invoice_items.id,
            'invoiceID', invoice_items.invoice_id,
            'description', invoice_items.description,
            'price', invoice_items.price,
            'productID', invoice_items.product_id,
            'count', invoice_items.count,
		        'productName', products.name
        )
    )FILTER (WHERE invoice_items.id IS NOT NULL),'[]') AS items,
    invoices.date,
    invoices.created_at,
    invoices.updated_at
FROM
    invoices
    LEFT JOIN persons ON invoices.person_id = persons.id
    LEFT JOIN invoice_items ON invoices.id = invoice_items.invoice_id
    LEFT JOIN products ON invoice_items.product_id = products.id
    %s
GROUP BY
    invoices.id,
    persons.name,
    persons.first_name
    %s %s;
		`,
		filterBy.String(), orderBy, pagedBy)

	rows, err := s.db.Query(context.Background(), query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
	defer rows.Close()

	var invoices []Invoice

	for rows.Next() {
		var invoice Invoice
		if err := rows.Scan(&invoice.ID, &invoice.PersonID, &invoice.PersonName, &invoice.Type, &invoice.Discount, &invoice.Notes, &invoice.Number, &invoice.Items, &invoice.Date, &invoice.CreatedAt, &invoice.UpdatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		invoices = append(invoices, invoice)
	}

	w.Header().Add("Content-Type", "application/json")

	newQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM
		    invoices
		    LEFT JOIN persons ON invoices.person_id = persons.id
		%s
		`, filterBy.String())
	row := s.db.QueryRow(context.Background(), newQuery)

	var Count int32

	err = row.Scan(&Count)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	var invoicesWithTotalCount struct {
		Rows       []Invoice `json:"rows"`
		TotalCount int32     `json:"totalCount"`
	}

	invoicesWithTotalCount.Rows = invoices
	invoicesWithTotalCount.TotalCount = Count

	err = json.NewEncoder(w).Encode(invoicesWithTotalCount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
}

func (s *Service) EditInvoice(id string, invoice Invoice) error {
	ctx := context.Background()

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	// ✅ Defer rollback ensures the tx is closed on panic or early return.
	// (If tx is already committed, rollback safely does nothing in pgx)
	defer tx.Rollback(ctx)

	updateInvoiceQuery := `
		UPDATE
		    invoices
		SET
		    person_id = $1,
		    type = $2,
		    notes = $3,
	      date  = $4
		WHERE
		    id = $5`

	_, err = tx.Exec(ctx, updateInvoiceQuery,
		invoice.PersonID,
		invoice.Type,
		invoice.Notes,
		invoice.Date,
		id,
	)
	if err != nil {
		return err
	}

	itemIDs := make([]uuid.UUID, 0, len(invoice.Items))

	for _, item := range invoice.Items {
		if item.ID == uuid.Nil {
			item.ID = uuid.New()
		}

		itemIDs = append(itemIDs, item.ID)
		fmt.Println(item.Discount.Value())

		itemQuery := `
			INSERT INTO invoice_items (id, invoice_id, description, price, product_id, count, discount)
			    VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (id)
			    DO UPDATE SET
			        description = EXCLUDED.description,
			        price = EXCLUDED.price,
			        product_id = EXCLUDED.product_id,
			        count = EXCLUDED.count,
			        discount = EXCLUDED.discount`

		_, err = tx.Exec(ctx, itemQuery,
			item.ID,
			id,
			item.Description,
			item.Price,
			item.ProductID,
			item.Count,
			item.Discount,
		)
		if err != nil {
			return err
		}
	}

	// ✅ Safer array comparison for empty slices
	deleteQuery := `
		DELETE FROM invoice_items
		WHERE invoice_id = $1
		    AND id != ALL($2::UUID[])`

	_, err = tx.Exec(ctx, deleteQuery, id, itemIDs) // ✅ Changed from invoice.ID to id
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *Service) DeleteInvoice(id string) error {
	query := `
		DELETE FROM invoices
		WHERE id = $1`

	_, err := s.db.Exec(context.Background(), query, id)
	if err != nil {
		return err
	}

	return nil
}
