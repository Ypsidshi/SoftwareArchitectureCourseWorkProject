package repository

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"coursework/deal-service/internal/domain"
)

type AdminContractsFilter struct {
	PaymentStatus string
	Status        string
}

func (r *Repository) ListContracts(ctx context.Context, page, pageSize int, filter AdminContractsFilter) ([]domain.Contract, int, error) {
	where, args := []string{"1=1"}, []any{}
	if s := strings.TrimSpace(filter.PaymentStatus); s != "" {
		args = append(args, s)
		where[0] += fmt.Sprintf(" AND payment_status = $%d", len(args))
	}
	if s := strings.TrimSpace(filter.Status); s != "" {
		args = append(args, s)
		where[0] += fmt.Sprintf(" AND status = $%d", len(args))
	}
	whereSQL := where[0]

	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM deal.contracts WHERE `+whereSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limitParam := len(args) + 1
	offsetParam := len(args) + 2
	args = append(args, pageSize, (page-1)*pageSize)

	const selectCols = `id, resident_id, room_id, manager_id, start_date, end_date, amount, currency,
       status, payment_status, payment_error, invoice_id, created_at, updated_at`
	q := `SELECT ` + selectCols + ` FROM deal.contracts WHERE ` + whereSQL + ` ORDER BY created_at DESC LIMIT $` +
		strconv.Itoa(limitParam) + ` OFFSET $` + strconv.Itoa(offsetParam)

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]domain.Contract, 0, pageSize)
	for rows.Next() {
		var item domain.Contract
		if err := scanContract(rows, &item); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func scanContract(sc interface {
	Scan(dest ...any) error
}, item *domain.Contract) error {
	return sc.Scan(
		&item.ID, &item.ResidentID, &item.RoomID, &item.ManagerID,
		&item.StartDate, &item.EndDate, &item.Amount, &item.Currency,
		&item.Status, &item.PaymentStatus, &item.PaymentError, &item.InvoiceID,
		&item.CreatedAt, &item.UpdatedAt,
	)
}
