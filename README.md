# ES Hit

Query ES then send hits number to Graphite

## How to Run it

- Define query rules inside config.toml
- Run it
  - `go run main.go -config conf/config.toml -verbose`

## Systemd config

- copy `systemd/es-hit.service` to `/etc/sytemd/system/es-hit.service`
- Modify es-hit.service for binary and configuration file location
- `systemctl daemon-reload`
- `systemctl enable es-hit; systemctl start es-hit`
- Check log: `journalctl -u es-hit -f`

## TODO

- Support other field as term search condition.
- Rules store inside DB, user can define it by themself.
