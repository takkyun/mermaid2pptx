# mermaid2pptx への貢献

*English version: [CONTRIBUTING.md](CONTRIBUTING.md).*

mermaid2pptx の改善にご関心をお寄せいただきありがとうございます。Issue・
Pull Request を歓迎します。

## 基本ルール

- **標準ライブラリのみ。** 本ツールは本番依存を持たず、その方針を維持します。
  モジュールの `require` を追加する場合は、まず Issue で相談してください
  (`python-pptx` や `mermaid-cli` は外部の開発/検証ツールであり、Go の依存では
  ありません)。
- **コードコメントは英語。** 周囲のコードと同程度の密度で、*なぜ* を説明します
  (*何を* ではなく)。
- **コミットメッセージは英語。**
- アーキテクチャや詳細な規約は [AGENTS.ja.md](AGENTS.ja.md) を参照。

## 開発環境

Go が必要です(バージョンは [go.mod](go.mod) を参照)。`mermaid-cli`(`mmdc`)は
`.mmd` 入力の変換やサンプル SVG の再生成をするときだけ必要です。

```sh
go build -o mermaid2pptx ./cmd/mermaid2pptx
```

## Pull Request を出す前に

以下をすべて実行し、クリーンであることを確認してください。

```sh
gofmt -l .                    # 何も出力されないこと
go vet ./...
go test ./...
GOOS=windows GOARCH=amd64 go build -o /dev/null ./cmd/mermaid2pptx  # クロスコンパイル確認
```

同じチェックが、すべての push / PR で CI でも実行されます。

### 生成される .pptx に影響する変更の場合

XML が well-formed であるテストが通っても、スライドの見た目が正しい保証には
なりません。ジェネレータを変更したときは:

1. サンプルを再生成し、同期を保つ:
   ```sh
   ./mermaid2pptx -f sample/graph*.svg
   ```
2. 出力を確認し(例: `python-pptx`)、少なくとも 1 つの `.pptx` を PowerPoint で
   開いて、図形・コネクタが配置され編集可能であることを確認する。
3. `TestConnectorEndpointsMatchSVG` をグリーンに保つ — コネクタ幾何の要。
   `slide.go` の `fitTransform` が、ジェネレータとテストで共有する px→EMU 変換の
   唯一の基準です。

新しいダイアグラム種別を追加する場合は、[AGENTS.ja.md](AGENTS.ja.md) の
「新しいダイアグラム種別を追加する」節を参照してください。

## Pull Request の指針

- 1 つの PR は 1 つの変更に集中させ、コミットは小さく自己完結させる。
- 何を・なぜ変えたかを説明し、生成物が変わったかどうかを明記する。
- ビルド済みバイナリや `_sandbox/` 以下はコミットしない(いずれも git-ignore 済み)。
- 顧客名・口座番号その他の個人を特定できる情報を、コード・サンプル・生成物に
  含めない。

## 貢献のライセンス

貢献いただいた時点で、その貢献はプロジェクトと同じ
[Apache License, Version 2.0](LICENSE) の下でライセンスされることに同意した
ものとみなします。
