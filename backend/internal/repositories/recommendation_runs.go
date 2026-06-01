package repositories

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RecommendationRunRepository interface {
	CreateRecommendationRun(ctx context.Context, run RecommendationRun) error
}

type RecommendationRun struct {
	InputSnapshot              any
	ResultSnapshot             any
	BestCardOfferID            string
	EstimatedYearOneValueCents int
}

type PostgresRecommendationRunRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRecommendationRunRepository(pool *pgxpool.Pool) *PostgresRecommendationRunRepository {
	return &PostgresRecommendationRunRepository{pool: pool}
}

func (r *PostgresRecommendationRunRepository) CreateRecommendationRun(ctx context.Context, run RecommendationRun) error {
	inputSnapshot, err := json.Marshal(run.InputSnapshot)
	if err != nil {
		return fmt.Errorf("encode input snapshot: %w", err)
	}
	resultSnapshot, err := json.Marshal(run.ResultSnapshot)
	if err != nil {
		return fmt.Errorf("encode result snapshot: %w", err)
	}

	_, err = r.pool.Exec(ctx, `
		INSERT INTO recommendation_runs (
			input_snapshot,
			result_snapshot,
			best_card_offer_id,
			estimated_year_one_value_cents
		) VALUES ($1::jsonb, $2::jsonb, NULLIF($3, '')::uuid, $4)
	`, inputSnapshot, resultSnapshot, run.BestCardOfferID, run.EstimatedYearOneValueCents)
	if err != nil {
		return fmt.Errorf("insert recommendation run: %w", err)
	}

	return nil
}
