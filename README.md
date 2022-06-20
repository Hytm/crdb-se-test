# crdb-se-test

## Schema:

> CREATE TABLE public.feeds (  
  id UUID NOT NULL DEFAULT gen_random_uuid(),  
  url STRING NOT NULL,  
  frequency INT8 NULL,  
  last_update TIMESTAMPTZ NULL,  
  CONSTRAINT "primary" PRIMARY KEY (id ASC)  
)

> CREATE TABLE public.recipes (  
  id UUID NOT NULL DEFAULT gen_random_uuid(),  
  title STRING NULL,  
  thumbnail STRING NULL,  
  url STRING NULL,  
  CONSTRAINT "primary" PRIMARY KEY (id ASC)  
)

## Apps:

### Feed Manager
The Feed Manager app is used to add or delete feeds to read from.
It includes an asynchronous function to refresh the list of recipes based on feed. The refresh can be forced by API calls, or is done automatically on a schedule.

### Read Recipe API
Single endpoint to get the list of recipes.

## Load generation

### Postman
Use the API to add, delete and force the update on a first runner. The second runner is constantly looking at the list.
Duration: 10 to 60 mins
Time between call: 0 to 1000 ms

### WRK
Every endpoint was hit by a dedicated wrk process.
Duration: 30mins
Concurrency: 16
Threads: 2