# Claude Code Pull Worker

スマートフォンからClaude Codeを遠隔実行し、結果をDiscordに通知するシステム

## 機能

- **複数のWebhookエンドポイント管理**: UUID形式のエンドポイント（`/webhooks/:uuid`）を動的に作成・管理
- **APIキー認証**: 各エンドポイントに複数のAPIキーを設定可能（bcryptでハッシュ化）
- **Web管理画面**: HTMX + Tailwind CSSによる直感的なUI
- **実行履歴**: 全ての実行を記録し、統計情報を表示
- **Claude Code設定**: エンドポイントごとに異なるClaude Code実行オプションを設定
- **通知設定**: Discord通知をエンドポイント個別/グローバルで設定
- **systemdサービス生成**: `systemd-install`サブコマンドでサービスファイルを自動生成

## セットアップ

### 1. 依存関係のインストール

```bash
go mod download
```

### 2. 環境設定

`.env`ファイルを作成:

```env
DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/YOUR_WEBHOOK_URL
PORT=8080
API_KEY=your-secret-key-here
CLAUDE_TIMEOUT=5m
```

### 3. アプリケーションの起動

```bash
# 開発環境
go run cmd/server/main.go

# ビルドして実行
go build -o claude-code-pull-worker cmd/server/main.go
./claude-code-pull-worker
```

管理画面: http://localhost:8081/

### 4. Tailscaleのセットアップ

```bash
# Tailscaleのインストール
curl -fsSL https://tailscale.com/install.sh | sh

# ログイン
tailscale up

# サービスを外部に公開する場合
tailscale serve https / http://localhost:8080
```

### 5. systemdサービスの設定

```bash
# サービスファイルを生成
./claude-code-pull-worker systemd-install --user=your-username

# サービスをインストール
sudo cp claude-code-pull-worker.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable claude-code-pull-worker
sudo systemctl start claude-code-pull-worker
```


## 使い方

### 1. 管理画面でWebhookエンドポイントを作成

1. http://localhost:8081/ にアクセス
2. "New Webhook"をクリック
3. 名前、説明、Claude Codeオプションを設定
4. 作成されたエンドポイントのUUIDをメモ

### 2. APIキーを生成

1. 作成したWebhookの"API Keys"をクリック
2. "New Key"をクリックして説明を入力
3. 生成されたAPIキーを安全に保存（再表示不可）

### API使用方法

#### エンドポイント

- `POST /webhooks/{uuid}` - Claude Codeを実行
- `GET /` - 管理画面
- `GET /health` - ヘルスチェック

#### リクエスト例

```bash
curl -X POST https://your-tailscale-name.ts.net/webhooks/YOUR-UUID-HERE \
  -H "Content-Type: application/json" \
  -H "X-API-Key: claude_your-generated-api-key" \
  -d '{
    "prompt": "List all files in the current directory",
    "context": "Working in a Go project"
  }'
```

#### レスポンス例

```json
{
  "success": true,
  "timestamp": "2024-06-22T10:00:00Z",
  "prompt": "List all files in the current directory",
  "response": "main.go\ngo.mod\ngo.sum\nREADME.md",
  "execution_time": "2.34s"
}
```

## スマートフォンからの利用

### iOS ショートカットの設定

1. ショートカットアプリを開く
2. 新規ショートカットを作成
3. 「Webリクエスト」アクションを追加
4. URLを設定: `https://your-tailscale-name.ts.net/webhook`
5. メソッド: POST
6. ヘッダー:
   - `Content-Type`: `application/json`
   - `X-API-Key`: `your-secret-key`
7. 本文: 
   ```json
   {
     "prompt": "テキストを入力"
   }
   ```

### Android Taskerの設定

1. 新しいタスクを作成
2. HTTP Requestアクションを追加
3. 設定:
   - Method: POST
   - URL: `https://your-tailscale-name.ts.net/webhook`
   - Headers: `Content-Type:application/json|X-API-Key:your-secret-key`
   - Body: `{"prompt":"%input"}`

## セキュリティ

- Tailscaleによるネットワークレベルの保護
- API Keyによる認証（オプション）
- 実行タイムアウトの設定
- HTTPSによる通信の暗号化（Tailscale serve使用時）

## トラブルシューティング

### Claude Codeが見つからない

```bash
# PATHにclaude実行ファイルが含まれているか確認
which claude

# systemdサービスの場合、Environmentに適切なPATHを設定
```

### Discord通知が届かない

- Webhook URLが正しいか確認
- Discord側でWebhookが有効になっているか確認
- ログでエラーメッセージを確認

### Tailscale接続エラー

```bash
# Tailscaleの状態確認
tailscale status

# サービスの再起動
sudo systemctl restart tailscaled
```