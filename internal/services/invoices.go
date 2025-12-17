package services

import (
	"context"
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
		INSERT INTO invoices (id, person_id, type, total, discount, net_total, notes, created_at, updated_at)
		    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	invoiceId := uuid.New()

	_, err = tx.Exec(context.Background(), query,
		invoiceId,
		inv.PersonID,
		inv.Type,
		inv.Total,
		inv.Discount,
		inv.NetTotal,
		inv.Notes,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		tx.Rollback(context.Background())
		return err
	}
	for _, item := range inv.Items {
		itemQuery := `
			INSERT INTO invoice_items (id, invoice_id, description, price, product_id, count, total, discount, net_total)
			    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
		id := uuid.New()
		_, err = tx.Exec(context.Background(), itemQuery,
			id,
			invoiceId,
			item.Description,
			item.Price,
			item.ProductID,
			item.Count,
			item.Total,
			item.Discount,
			item.NetTotal,
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
		    id,
		    person_id,
		    type,
		    total,
		    net_total,
		    discount,
		    notes,
		    created_at,
		    updated_at
		FROM
		    invoices
		WHERE
		    id = $1`
	var inv Invoice
	err := s.db.QueryRow(context.Background(), query, id).Scan(
		&inv.ID,
		&inv.PersonID,
		&inv.Type,
		&inv.Total,
		&inv.NetTotal,
		&inv.Discount,
		&inv.Notes,
		&inv.CreatedAt,
		&inv.UpdatedAt,
	)
	if err != nil {
		return Invoice{}, err
	}

	return inv, nil
}

func (s *Service) ListInvoices() ([]Invoice, error) {
	query := `
		SELECT
		    i.id,
		    i.person_id,
		    CONCAT(p.last_name, ' ', p.first_name) AS person_name,
		    i.type,
		    i.total,
		    i.discount,
		    i.notes,
		    i.number,
		    json_agg(json_build_object('id', ii.id, 'invoiceID', ii.invoice_id, 'description', ii.description, 'price', CAST(ii.price AS INTEGER), 'productID', ii.product_id, 'count', ii.count)) AS items,
		    i.created_at,
		    i.updated_at
		FROM
		    invoices AS i
		    LEFT JOIN persons p ON i.person_id = p.id
		    LEFT JOIN invoice_items ii ON i.id = ii.invoice_id
		GROUP BY
		    i.id,
		    p.last_name,
		    p.first_name`
	rows, err := s.db.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invoices []Invoice

	for rows.Next() {
		var inv Invoice
		if err := rows.Scan(
			&inv.ID,
			&inv.PersonID,
			&inv.PersonName,
			&inv.Type,
			&inv.Total,
			&inv.Discount,
			&inv.Notes,
			&inv.Number,
			&inv.Items,
			&inv.CreatedAt,
			&inv.UpdatedAt,
		); err != nil {
			return nil, err
		}

		invoices = append(invoices, inv)
	}

	return invoices, nil
}

func (s *Service) EditInvoice(id string, invoice Invoice) error {
	ctx := context.Background()

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	// ✅ Proper UPDATE query
	updateInvoiceQuery := `
		UPDATE
		    invoices
		SET
		    person_id = $1,
		    type = $2,
		    total = $3,
		    notes = $4
		WHERE
		    id = $5`
	_, err = tx.Exec(ctx, updateInvoiceQuery,
		invoice.PersonID,
		invoice.Type,
		invoice.Total,
		invoice.Notes,
		id,
	)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	// ✅ Track item IDs for deletion
	itemIDs := make([]uuid.UUID, 0, len(invoice.Items))

	for _, item := range invoice.Items {
		if item.ID == uuid.Nil {
			item.ID = uuid.New() // assign new ID if not provided
		}
		itemIDs = append(itemIDs, item.ID)

		// ✅ UPSERT (insert or update on conflict)
		itemQuery := `
			INSERT INTO invoice_items (id, invoice_id, description, price, product_id, count, total)
			    VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (id)
			    DO UPDATE SET
			        description = EXCLUDED.description,
			        price = EXCLUDED.price,
			        product_id = EXCLUDED.product_id,
			        count = EXCLUDED.count,
			        total = EXCLUDED.total`
		_, err = tx.Exec(ctx, itemQuery,
			item.ID,
			invoice.ID,
			item.Description,
			item.Price,
			item.ProductID,
			item.Count,
			item.Total,
		)
		if err != nil {
			tx.Rollback(ctx)
			return err
		}
	}

	// ✅ Delete items not in the new invoice
	deleteQuery := `
		DELETE FROM invoice_items
		WHERE invoice_id = $1
		    AND id NOT IN (
		        SELECT
		            UNNEST($2::UUID[]))`
	_, err = tx.Exec(ctx, deleteQuery, invoice.ID, itemIDs)
	if err != nil {
		tx.Rollback(ctx)
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

func (s *Service) AddInvoiceItem(ctx context.Context, item InvoiceItem) error {
	// compute item.Total = Price * Count; update invoice total

	return nil
}

func (s *Service) RemoveInvoiceItem(ctx context.Context, id int64) error {
	return nil
}

func (s *Service) ListInvoiceItems(ctx context.Context, invoiceID int64) ([]InvoiceItem, error) {
	return nil, nil
}
