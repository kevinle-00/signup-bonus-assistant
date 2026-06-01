-- +goose Up
CREATE TABLE recommendation_runs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  input_snapshot JSONB NOT NULL,
  result_snapshot JSONB NOT NULL,
  best_card_offer_id UUID REFERENCES card_offers(id),
  estimated_year_one_value_cents INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX recommendation_runs_created_at_idx ON recommendation_runs (created_at DESC);
CREATE INDEX recommendation_runs_best_card_offer_id_idx ON recommendation_runs (best_card_offer_id);

-- +goose Down
DROP TABLE IF EXISTS recommendation_runs;
