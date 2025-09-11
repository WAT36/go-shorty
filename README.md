# go-shorty

標準ライブラリだけで動く、超シンプルな URL 短縮サービス（学習用）。

- REST API + リダイレクト + JSON 永続化
- テンプレート一枚の簡易 UI
- Go 1.21+ / 依存ゼロ
- Docker 対応 / CI 付き

## 使い方

```bash
go run ./cmd/shorty
# http://localhost:8080
```

環境変数:
PORT（デフォルト 8080）
SHORTY_DB（デフォルト data/urls.json）

API

```
POST /api/shorten → {"url":"<URL>", "custom":"<任意コード>"}
成功: {"code":"abc123","url":"..."}
GET /api/list → 全件
DELETE /api/{code} → 削除
GET /{code} → 302 でリダイレクト
```

開発

```bash
make run
make build
make test
```

Docker

```bash
docker build -t shorty:latest .
docker run --rm -p 8080:8080 -v $PWD/data:/app/data shorty:latest
```
