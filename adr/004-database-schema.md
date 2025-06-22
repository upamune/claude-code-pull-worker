# ADR-004: Database Schema Design

## Status
Accepted

## Context
SQLiteを使用して、Webhookエンドポイント、APIキー、実行履歴を管理する。

## Decision
以下のテーブル構造を採用：

### webhooks
- UUID、名前、説明、Claude Code実行オプション、通知設定

### api_keys
- ハッシュ化されたキー、プレフィックス、説明、webhook_id（外部キー）

### execution_histories
- 実行時刻、ステータス、プロンプト、レスポンス、エラー、実行時間

### global_settings
- 全体的な設定（デフォルト通知設定など）

## Consequences
- 正規化されたスキーマで整合性を保証
- 実行履歴により監査とデバッグが可能
- 将来の拡張に対応しやすい構造