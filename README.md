# ES Hit

Query ES then send hits number to Graphite

## How to Run it

- Define query rules inside config.toml
- `es-hit -config config/config.toml -verbose`

## TODO

- Support other field as term search condition
- Rules store inside DB, user can define it by themself