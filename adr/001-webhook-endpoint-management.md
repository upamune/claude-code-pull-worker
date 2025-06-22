# ADR-001: Webhook Endpoint Management

## Status
Accepted

## Context
現在のシステムは単一のWebhookエンドポイント（`/webhook`）のみをサポートしているが、複数のエンドポイントを動的に管理する必要がある。

## Decision
- UUID v4を使用して各Webhookエンドポイントを識別する
- パスパターン: `/webhooks/:uuid`
- 各エンドポイントはClaude Codeの実行オプションを持つ
- エンドポイントごとに複数のAPIキーを設定可能

## Consequences
- 複数のプロジェクトや用途で同一サーバーを共有可能
- エンドポイントごとに異なる設定で実行可能
- UUID使用によりエンドポイントの推測が困難