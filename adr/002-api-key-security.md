# ADR-002: API Key Security

## Status
Accepted

## Context
APIキーを安全に管理する必要がある。平文保存は情報漏洩のリスクがある。

## Decision
- APIキーはbcryptでハッシュ化して保存
- 新規作成時のみ平文を表示（再表示不可）
- プレフィックスと最後の4文字は平文で保存（識別用）
- 例: `claude_************abcd`

## Consequences
- データベースが漏洩してもAPIキーは安全
- ユーザーは作成時にキーを保存する必要がある
- プレフィックスによる用途識別が可能