use std::fs::File;
use std::io::{self, Read};
use std::path::Path;

fn main() {
    // snake case(変数,関数)
    let file_name = "./asm/Add.asm";
    // match式 & Ok,Errがrustにおけるエラーハンドリングの基本形
    match open_file(file_name) {
        Ok(inputs) => {
            for input in inputs {
                println!("{}", input);
            }
        }
        Err(e) => {
            println!("file open error {}", e);
        }
    }
}

// pointer引数を取ることで関数実行による所有権の移転が起こさず関数にstruct引数を渡す
// エラー発生しうる関数の戻り値はResult<T, E>型(golangでいう res, err みたいな感じ)
fn open_file(name: &str) -> io::Result<Vec<String>> {
    // fn open_file(name: &str) -> io::Result<Vec<String>> {
    let path = Path::new(name);
    // エラー発生しうる関数の後ろに?をつけることで、Ok時は後続処理、error時はearly return処理を簡潔に書ける
    // https://doc.rust-jp.rs/book-ja/ch09-02-recoverable-errors-with-result.html#%E3%82%A8%E3%83%A9%E3%83%BC%E5%A7%94%E8%AD%B2%E3%81%AE%E3%82%B7%E3%83%A7%E3%83%BC%E3%83%88%E3%82%AB%E3%83%83%E3%83%88-%E6%BC%94%E7%AE%97%E5%AD%90
    let mut file = File::open(path)?;
    let mut data = String::new();
    // 関数にmut付き引数で渡すと値の書き換えも実行関数内でできる
    file.read_to_string(&mut data)?;

    if data.is_empty() {
        return Err(io::Error::new(io::ErrorKind::Other, "error: file is empty"));
    }
    // 戻り値は 式(文の区切りを意味する ; をつけるとエラーになる)
    // closure引数の書き方|line|
    // closure関数が1行の場合は {} をJS likeに省略可能
    Ok(data.lines().map(|line| line.to_string()).collect())
}
