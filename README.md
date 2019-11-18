# ES Hit
[![Build Status](https://travis-ci.org/leyantech/es-hit.svg?branch=master)](https://travis-ci.org/leyantech/es-hit)

Query ES then send hits number to Graphite

Support 2 types of search rules:

- Static rules which defined inside configuration file.
- Use saved search from kibana, which have specific prefix.

## How to Run it

- Define query rules inside config.toml
- glide up
- Run for testing  `go run main.go -config conf/config.toml -verbose`

## Systemd config

- copy `conf/es-hit.service` to `/etc/sytemd/system/es-hit.service`
- Modify `es-hit.service` for binary and configuration file location
- `systemctl daemon-reload`
- `systemctl enable es-hit; systemctl start es-hit`
- Check log: `journalctl -u es-hit -f`

## TODO

- Support other field as term search condition.
- kibana rule need support filter
