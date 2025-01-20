

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

### Rust style guide
[reference](https://doc.rust-lang.org/book/title-page.html)
[reference ja](https://doc.rust-jp.rs/book-ja/title-page.html)
[coding style guide](https://doc.rust-lang.org/nightly/style-guide/)


### recommended extention

debug:[`CodeLLDB`](https://marketplace.visualstudio.com/items?itemName=vadimcn.vscode-lldb)
format:[`rust-analyzer`](https://marketplace.visualstudio.com/items?itemName=rust-lang.rust-analyzer)
[format rule](https://rust-lang.github.io/rustfmt/?version=master&search=)