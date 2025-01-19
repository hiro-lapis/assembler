

### debug

When debug rust file, vscode has to open rust directory, not assembler dir.

### cmd

・現在いるディレクトリのmain.rs を実行

```
cargo run
```

・新しいプロジェクトを作成(現在いるディレクトリに新しいディレクトリ作成)

```
cargo new --bin newProject
```

・Cargo初期化(プロジェクト作成済だけどCargo.toml,Cargo.lockがない場合)

```
cargo init
```

### recommended extention

debug:[`CodeLLDB`](https://marketplace.visualstudio.com/items?itemName=vadimcn.vscode-lldb)