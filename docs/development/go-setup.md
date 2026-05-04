# Go Setup

## Objetivo

O projeto usa Go como runtime principal. Para evitar falhas de permissao em ambientes sandbox, comandos Go devem usar caches locais ao repositorio.

## Uso recomendado

Use o wrapper do repositorio:

```sh
./scripts/go version
./scripts/go list ./...
./scripts/go test ./...
```

O wrapper configura:

- `GOPATH=.go`
- `GOMODCACHE=.go/pkg/mod`
- `GOCACHE=.go/cache`
- `GOTOOLCHAIN=local`

Os diretorios `.go/` e `.tools/` nao devem ser versionados.

Para sobrescrever os caches locais, use `ORCHESTRAOS_GOPATH`, `ORCHESTRAOS_GOMODCACHE` ou `ORCHESTRAOS_GOCACHE`. O wrapper nao herda `GOPATH`, `GOMODCACHE` ou `GOCACHE` globais por padrao.

## Instalacao local opcional

Se o sistema nao tiver um `go` funcional, instale uma toolchain local no repositorio:

```sh
./scripts/install-go-local.sh
```

Por padrao o script instala `go1.26.2` em `.tools/go`. Para usar outra versao:

```sh
GO_VERSION=1.26.2 ./scripts/install-go-local.sh
```

Depois disso, `./scripts/go` usa automaticamente `.tools/go/bin/go`.

## Diagnostico

Para verificar o ambiente usado pelo projeto:

```sh
./scripts/go env GOROOT GOPATH GOMODCACHE GOCACHE GOTOOLCHAIN
```

Se `go list ./...` falhar com erro de codigo, trate o erro no pacote indicado. Se falhar tentando escrever em `/home/.../go` ou `/home/.../.cache`, use `./scripts/go` em vez do binario global diretamente.

## Postgres local

O `docker-compose.yml` publica o Postgres do projeto em `localhost:55432`, mantendo a porta `5432` dentro do container. Essa porta local evita conflito com instalacoes Postgres ja existentes na maquina.

Exemplo:

```sh
docker compose up -d postgres
./scripts/go run ./cmd/orchestraos --db-port 55432 migrate up
TEST_DB_DSN="host=localhost port=55432 user=orchestraos password=orchestraos dbname=orchestraos sslmode=disable" ./scripts/go test ./...
```
