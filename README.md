# let me see...

<img src="static/favicon.svg" width="100" align="right" />


 電子辞書を検索するウェブアプリケーション 'let me see...' のGoによる再実装です。
'let me see...(Go)' は aehlke/ebライブラリを介して、辞書を検索することができます。

![](./.github/assets/sample-01.png)

>[!NOTE]
> このリポジトリは、Giteaのミラーであり、ソフトウェアの提供を目的としていません。  
> このコードは、以下のプロジェクトのソースコードを参考にしたものであり開発途中のもです。
>* http://openlab.ring.gr.jp/edict/letmesee/index.html.ja
>* https://github.com/kurema/forkedLetMeSee
> 
>ライセンスにはご注意ください。  

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



## 参考
- “aehlke/eb: fork of eb-3.1 (aka eblib, aka libeb) EPWING library written in C, the version required by the Python ebmodule-2.0”. GitHub, [https://github.com/aehlke/eb](https://github.com/aehlke/eb), (Accessed 2026-07-13)
