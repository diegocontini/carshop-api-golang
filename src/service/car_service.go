package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"utfpr.edu.br/carshop-api/src/domain"
	"utfpr.edu.br/carshop-api/src/infra/sqlc"
)

type CarService struct {
	pool *pgxpool.Pool
	q    *sqlc.Queries
}

func NewCarService(pool *pgxpool.Pool) *CarService {
	return &CarService{pool: pool, q: sqlc.New(pool)}
}

func (s *CarService) List(ctx context.Context) ([]domain.Car, error) {
	cars, err := s.q.ListCars(ctx)
	if err != nil {
		return nil, fmt.Errorf("list cars: %w", err)
	}
	if len(cars) == 0 {
		return []domain.Car{}, nil
	}

	ids := make([]int64, len(cars))
	for i, c := range cars {
		ids[i] = c.ID
	}
	imgs, err := s.q.ListCarImagesByCarIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("list car images: %w", err)
	}

	byCar := make(map[int64][]domain.CarImage, len(cars))
	for _, im := range imgs {
		if im.CarID == nil {
			continue
		}
		byCar[*im.CarID] = append(byCar[*im.CarID], domain.CarImage{ID: im.ID, URL: im.Url, CarID: im.CarID})
	}

	out := make([]domain.Car, len(cars))
	for i, c := range cars {
		out[i] = carFromSQLC(c, byCar[c.ID])
	}
	return out, nil
}

func (s *CarService) Get(ctx context.Context, id int64) (domain.Car, error) {
	row, err := s.q.GetCar(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Car{}, domain.ErrNotFound
		}
		return domain.Car{}, fmt.Errorf("get car: %w", err)
	}
	imgs, err := s.q.ListCarImagesByCarIDs(ctx, []int64{id})
	if err != nil {
		return domain.Car{}, fmt.Errorf("get car images: %w", err)
	}
	return carFromSQLC(row, imagesFromSQLC(imgs)), nil
}

func (s *CarService) Create(ctx context.Context, in domain.Car) (domain.Car, error) {
	var created domain.Car
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		row, err := q.CreateCar(ctx, sqlc.CreateCarParams{
			New:         in.New,
			Brand:       in.Brand,
			Model:       in.Model,
			Year:        in.Year,
			Price:       in.Price,
			Color:       in.Color,
			Km:          in.Km,
			Description: in.Description,
		})
		if err != nil {
			return fmt.Errorf("insert car: %w", err)
		}
		newImages := make([]sqlc.BulkInsertCarImagesParams, len(in.Images))
		for i, im := range in.Images {
			carID := row.ID
			newImages[i] = sqlc.BulkInsertCarImagesParams{Url: im.URL, CarID: &carID}
		}
		if len(newImages) > 0 {
			if _, err := q.BulkInsertCarImages(ctx, newImages); err != nil {
				return fmt.Errorf("bulk insert car images: %w", err)
			}
		}
		imgs, err := q.ListCarImagesByCarIDs(ctx, []int64{row.ID})
		if err != nil {
			return fmt.Errorf("reload car images: %w", err)
		}
		created = carFromSQLC(row, imagesFromSQLC(imgs))
		return nil
	})
	if err != nil {
		return domain.Car{}, err
	}
	return created, nil
}

func (s *CarService) Update(ctx context.Context, id int64, in domain.Car) (domain.Car, error) {
	var updated domain.Car
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		row, err := q.UpdateCar(ctx, sqlc.UpdateCarParams{
			ID:          id,
			New:         in.New,
			Brand:       in.Brand,
			Model:       in.Model,
			Year:        in.Year,
			Price:       in.Price,
			Color:       in.Color,
			Km:          in.Km,
			Description: in.Description,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return domain.ErrNotFound
			}
			return fmt.Errorf("update car: %w", err)
		}

		existingIDs, err := q.ListCarImageIDsByCarID(ctx, &id)
		if err != nil {
			return fmt.Errorf("list existing image ids: %w", err)
		}
		existing := make(map[int64]struct{}, len(existingIDs))
		for _, eid := range existingIDs {
			existing[eid] = struct{}{}
		}

		var updIDs []int64
		var updURLs []string
		var toInsert []sqlc.BulkInsertCarImagesParams

		for _, im := range in.Images {
			if im.ID != 0 {
				if _, ok := existing[im.ID]; ok {
					updIDs = append(updIDs, im.ID)
					updURLs = append(updURLs, im.URL)
					continue
				}
			}
			carID := row.ID
			toInsert = append(toInsert, sqlc.BulkInsertCarImagesParams{Url: im.URL, CarID: &carID})
		}

		if len(updIDs) > 0 {
			if err := q.BulkUpdateCarImages(ctx, sqlc.BulkUpdateCarImagesParams{
				Ids:   updIDs,
				Urls:  updURLs,
				CarID: &row.ID,
			}); err != nil {
				return fmt.Errorf("bulk update car images: %w", err)
			}
		}
		if len(toInsert) > 0 {
			if _, err := q.BulkInsertCarImages(ctx, toInsert); err != nil {
				return fmt.Errorf("bulk insert car images: %w", err)
			}
		}

		imgs, err := q.ListCarImagesByCarIDs(ctx, []int64{row.ID})
		if err != nil {
			return fmt.Errorf("reload car images: %w", err)
		}
		updated = carFromSQLC(row, imagesFromSQLC(imgs))
		return nil
	})
	if err != nil {
		return domain.Car{}, err
	}
	return updated, nil
}

func (s *CarService) Delete(ctx context.Context, id int64) error {
	n, err := s.q.DeleteCar(ctx, id)
	if err != nil {
		return fmt.Errorf("delete car: %w", err)
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func carFromSQLC(c sqlc.Car, imgs []domain.CarImage) domain.Car {
	return domain.Car{
		ID:          c.ID,
		New:         c.New,
		Brand:       c.Brand,
		Model:       c.Model,
		Year:        c.Year,
		Price:       c.Price,
		Color:       c.Color,
		Km:          c.Km,
		Description: c.Description,
		Images:      imgs,
	}
}

func imagesFromSQLC(rows []sqlc.CarImage) []domain.CarImage {
	out := make([]domain.CarImage, len(rows))
	for i, r := range rows {
		out[i] = domain.CarImage{ID: r.ID, URL: r.Url, CarID: r.CarID}
	}
	return out
}
