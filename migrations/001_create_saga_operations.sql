CREATE TABLE IF NOT EXISTS saga_operations (
  id UUID PRIMARY KEY,
  saga_id UUID NOT NULL,
  order_id UUID NOT NULL,
  operation VARCHAR(40) NOT NULL,
  status VARCHAR(40) NOT NULL,
  result_payload JSONB,
  processed_at TIMESTAMP,
  UNIQUE (saga_id, operation)
);
