# let me see...

<img src="static/favicon.svg" width="100" align="right" />


 電子辞書を検索するウェブアプリケーション 'let me see...' のGoによる再実装です。
'let me see...(Go)' は aehlke/ebライブラリを介して、辞書を検索することができます。

モバイル、ダークモード対応あり ^_^

![](./.github/assets/sample-01.png)


## 目的

* ユーザが、epwing辞書を引けるWebサーバを作る
* エージェントがepwing辞書を引けるAPIを作る

## 計画

- ダークモードへの対応、横幅の修正
- フロントコードの整理
- EPWINGのライブラリをGoで書き直せそうか調べる => eblibで要点をしぼっても半月はかかる作業になりそう。見送り。


## 開発

- 通常の環境でビルドしようとして以下のようなエラーが出る。nix-shell環境で行う必要がある。
```bash
go build -o letmesee .
go: downloading golang.org/x/text v0.36.0
# letmesee
In file included from ./eb.go:7:
./hooks.h:4:10: fatal error: eb/eb.h: No such file or directory
    4 | #include <eb/eb.h>
      |          ^~~~~~~~~
compilation terminated.
make: *** [Makefile:3: build] エラー 1
```

フック関数の制約

hooks.cのフック関数をGoで書き直すには以下の制約がある
- 関数ポインタの要件: libebのフックシステムはCの関数ポインタを期待し、eb_set_hook()を通じて登録されます hooks.c:334-375 。Goの関数をCの関数ポインタとして渡すには、//exportディレクティブを使用してGo関数をCにエクスポートする必要がありますが、これは複雑でエラーが発生しやすいです。
- libebデータ構造への直接アクセス: フック関数はEB_Book *book、EB_Appendix *appなどのlibeb固有のデータ構造を直接操作し、eb_write_text_string()などのlibeb関数を呼び出します hooks.c:9-13 。これらの操作をGoから行うには、すべてのlibeb構造体と関数をCGOでラップする必要があります。
- コールバックの複雑さ: フック関数はlibebから呼び出されるコールバックであり、特定のシグネチャを持つ必要があります hooks.c:8-14 。GoのコールバックをCの関数ポインタとして渡すことは可能ですが、パフォーマンスの低下や複雑なメモリ管理が発生します。


## 先行技術
* http://openlab.ring.gr.jp/edict/letmesee/index.html.ja
* https://github.com/kurema/forkedLetMeSee



## 参考
- “aehlke/eb: fork of eb-3.1 (aka eblib, aka libeb) EPWING library written in C, the version required by the Python ebmodule-2.0”. GitHub, [https://github.com/aehlke/eb](https://github.com/aehlke/eb), (Accessed 2026-07-13)
