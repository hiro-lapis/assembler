use std::fs::File;
use std::io::{self, Read};
use std::path::Path;

fn main() {
    // snake case(変数,関数)
    let file_name = "./asm/Add.asm";
    // match式 & Ok,Errがrustにおけるエラーハンドリングの基本形
    // https://doc.rust-jp.rs/book-ja/ch02-00-guessing-game-tutorial.html#%E4%B8%8D%E6%AD%A3%E3%81%AA%E5%85%A5%E5%8A%9B%E3%82%92%E5%87%A6%E7%90%86%E3%81%99%E3%82%8B
    match open_file(file_name) {
        Ok(inputs) => {
            // for でも所有権に注意する必要あり
            for input in &inputs {
                println!("{}", input);
            }
            let mut parser = Parser::new(inputs);
            parser.next();
            parser.next();
            parser.next();
            parser.next();
            parser.reset();
            println!("parser {:?}", parser.current_line());
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
    // &str, Stringは標準ライブラリ
    // https://doc.rust-jp.rs/book-ja/ch08-02-strings.html#%E6%96%87%E5%AD%97%E5%88%97%E3%81%A8%E3%81%AF
    let mut data = String::new();
    // 関数にmut付き引数で渡すと値の書き換えも実行関数内でできる
    file.read_to_string(&mut data)?;

    if data.is_empty() {
        return Err(io::Error::new(io::ErrorKind::Other, "error: file is empty"));
    }
    // 最後に書く戻り値は 式(文の区切りを意味する ; をつけるとエラーになる)
    // https://doc.rust-jp.rs/book-ja/ch03-03-how-functions-work.html#%E6%88%BB%E3%82%8A%E5%80%A4%E3%81%AE%E3%81%82%E3%82%8B%E9%96%A2%E6%95%B0
    // closure引数の書き方|line|
    // closure関数が1行の場合は {} をJS likeに省略可能
    Ok(data.lines().map(|line| line.to_string()).collect())
}

// ==比較できるようにする場合derive設定必要
// https://zenn.dev/labbase/articles/d0f080f6cbe8f0
#[derive(PartialEq)]
enum InstructionType {
    AInstruction,
    CInstruction,
    LInstruction,
}

// std::fmt::Display
// https://doc.rust-jp.rs/book-ja/ch05-02-example-structs.html#%E3%83%88%E3%83%AC%E3%82%A4%E3%83%88%E3%81%AE%E5%B0%8E%E5%87%BA%E3%81%A7%E6%9C%89%E7%94%A8%E3%81%AA%E6%A9%9F%E8%83%BD%E3%82%92%E8%BF%BD%E5%8A%A0%E3%81%99%E3%82%8B
// https://qiita.com/TakedaTakumi/items/7936a19979e46fc1b780

#[derive(Debug)]
struct Parser {
    pub instructions: Vec<String>,
    current_line: i32, // 標準的なint型
}
impl Parser {
    // 関連関数はコンストラクタ実装で使われることが多い
    // https://doc.rust-jp.rs/book-ja/ch05-03-method-syntax.html#%E9%96%A2%E9%80%A3%E9%96%A2%E6%95%B0
    pub fn new(lines: Vec<String>) -> Self {
        let mut list: Vec<String> = Vec::new();
        // let list: Vec<String> = [].to_vec();
        for line in lines {
            let mut v = String::from("");
            // Stringの各種メソッド
            // https://doc.rust-lang.org/std/string/struct.String.html#method.is_empty
            if line.is_empty() {
                // if line.len() == 0 {
                continue;
            }
            // if let について
            // https://qiita.com/plotter/items/0d8dc2782f21178d64f1
            // https://doc.rust-jp.rs/book-ja/ch18-01-all-the-places-for-patterns.html#%E6%9D%A1%E4%BB%B6%E5%88%86%E5%B2%90if-let%E5%BC%8F
            if let Some(index) = line.find("//") {
                if index == 0 {
                    continue;
                }
                v = line[..index].to_string();
            } else {
                v = line.to_string();
            }
            if v.is_empty() {
                continue;
            }
            v = v.trim().to_string();
            list.push(v);
        }

        Parser {
            instructions: list,
            current_line: 0,
        }
    }
    fn substring(s: &str, start: usize, length: usize) -> &str {
        if length == 0 {
            return "";
        }

        let mut ci = s.char_indices();
        let byte_start = match ci.nth(start) {
            Some(x) => x.0,
            None => return "",
        };

        match ci.nth(length - 1) {
            Some(x) => &s[byte_start..x.0],
            None => &s[byte_start..],
        }
    }
    // メソッドは常に&self を第一引数にとる
    // setterの場合は &mut self
    // https://doc.rust-jp.rs/book-ja/ch05-03-method-syntax.html#%E3%83%A1%E3%82%BD%E3%83%83%E3%83%89%E8%A8%98%E6%B3%95
    fn current_line(&self) -> i32 {
        self.current_line
    }
    fn next(&mut self) {
        self.current_line += 1
    }
    fn reset(&mut self) {
        self.current_line = 0
    }
    fn instruction_type(&self) -> InstructionType {
        // 借用
        let line = &self.instructions[self.current_line as usize];
        // charsはmulti byte対応の文字列変換.slice index参照より対応範囲が広い
        // unwrapはSome(v)をアンラップ
        if line.chars().nth(0).unwrap().to_string() == "@".to_string() {
            return InstructionType::AInstruction;
        }
        let last_index = line.len() - 1;
        if line.chars().nth(0).unwrap().to_string() == "(".to_string()
            && line.chars().nth(last_index).unwrap().to_string() == ")".to_string()
        {
            return InstructionType::LInstruction;
        }
        InstructionType::CInstruction
    }
    fn label(&self) -> String {
        if self.instruction_type() == InstructionType::LInstruction {
            let char_length = self.instructions[self.current_line as usize].len() - 1;
            let line = &self.instructions[self.current_line as usize].to_string();
            return Parser::substring(line, 1, char_length).to_string();
        }
        String::new()
    }
    fn symbol(&self) -> String {
        if self.instruction_type() == InstructionType::AInstruction {
            // let char_length = self.instructions[self.current_line as usize].len() - 1;
            let line = &self.instructions[self.current_line as usize].to_string();
            //
            return line.chars().skip(1).collect();
            // return line.chars().skip(1).collect::<String>();
        }
        String::new()
    }
}
