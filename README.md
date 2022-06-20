# crdb-se-test

Schema:

CREATE TABLE public.feeds (
  id UUID NOT NULL DEFAULT gen_random_uuid(),
  url STRING NOT NULL,
  frequency INT8 NULL,
  last_update TIMESTAMPTZ NULL,
  CONSTRAINT "primary" PRIMARY KEY (id ASC)
)

CREATE TABLE public.recipes (
  id UUID NOT NULL DEFAULT gen_random_uuid(),
  title STRING NULL,
  thumbnail STRING NULL,
  url STRING NULL,
  CONSTRAINT "primary" PRIMARY KEY (id ASC)
)