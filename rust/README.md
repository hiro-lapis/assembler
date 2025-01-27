

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

### Rust guide
[reference](https://doc.rust-lang.org/book/title-page.html)
[reference ja](https://doc.rust-jp.rs/book-ja/title-page.html)
[rust by example](https://doc.rust-jp.rs/rust-by-example-ja/index.html)
[coding style guide](https://doc.rust-lang.org/nightly/style-guide/)
[Rust 入門](https://zenn.dev/mebiusbox/books/22d4c1ed9b0003/viewer/661cf1)

### recommended extention

debug:[`CodeLLDB`](https://marketplace.visualstudio.com/items?itemName=vadimcn.vscode-lldb)
format:[`rust-analyzer`](https://marketplace.visualstudio.com/items?itemName=rust-lang.rust-analyzer)
[format rule](https://rust-lang.github.io/rustfmt/?version=master&search=)

variables are immutable by default. even Vect, Result
in fn, last return value can be a expression
when u declare const, the const must receive type annotation
Rustacean = rust arean, the jargon refers to rust coders

let が基本immutable でちょいちょい間違える
引数に渡すときも書き換えたい時はmutをつける必要がある。
golangがよほどシンプルな言語だというのがわかった
公式リファレンスがとてもよい。疑問に思ったことを調べると、だいたいどこを見ればいいかが書いてある
use で使ってないのあっても怒られないのは意外

今回の場合、String::fromとto_stringは全く同じことをします。 従って、どちらを選ぶかは、スタイル次第
どちらを使用するとか、コード規約どうする？
https://doc.rust-jp.rs/book-ja/ch08-02-strings.html#%E6%96%B0%E8%A6%8F%E6%96%87%E5%AD%97%E5%88%97%E3%82%92%E7%94%9F%E6%88%90%E3%81%99%E3%82%8B