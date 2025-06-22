# ADR-003: Notification Configuration

## Status
Accepted

## Context
通知設定をグローバルとエンドポイント個別の両方で管理する必要がある。

## Decision
- 通知設定は階層構造とする
  - グローバル設定（デフォルト）
  - エンドポイント個別設定（オーバーライド）
- JSON形式で設定を保存
- 通知タイプ（discord, slack等）を拡張可能な構造

## Consequences
- 柔軟な通知設定が可能
- エンドポイントごとに異なる通知先を設定可能
- 新しい通知タイプの追加が容易