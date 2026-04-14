{
  pkgs ? import <nixpkgs> { },
}:

let
  ruby = pkgs.ruby_2_7.withPackages (ps: [
    ps.webrick
  ]);
in
pkgs.mkShell {
  buildInputs = [
    pkgs.libeb
    pkgs.ruby_2_7
    pkgs.git
    pkgs.gcc
    pkgs.gnumake
  ];

  packages = [ ruby ];

  shellHook = ''
    export PATH=${ruby}/bin:$PATH
    export LD_LIBRARY_PATH=${pkgs.libeb}/lib:$LD_LIBRARY_PATH
    export CPATH=${pkgs.libeb}/include:$CPATH

    # rubyeb19のコンパイルと インストール（初回のみ）
    if ! ${ruby}/bin/ruby -e "require 'eb'" 2>/dev/null; then
      echo "Building and installing rubyeb19..."
      TMPDIR=$(mktemp -d)
      cd "$TMPDIR"
      ${pkgs.git}/bin/git clone https://github.com/kubo/rubyeb19.git
      cd rubyeb19
      ${ruby}/bin/ruby extconf.rb --with-eb-dir=${pkgs.libeb}
      ${pkgs.gnumake}/bin/make
      ${pkgs.gnumake}/bin/make install
      cd /
      rm -rf "$TMPDIR"
      echo "rubyeb19 installation complete"
    fi

    cd /home/asano/test/letmesee/ruby-letmesee

    # shebang fix - 完全パスを使用
    if [ -f index.rb ]; then
      sed -i "1s|^#!.*|#!${ruby}/bin/ruby|" index.rb
    fi

    # index.cgi 用意
    if [ -f index.rb ] && [ ! -f index.cgi ]; then
      cp index.rb index.cgi
      chmod +x index.cgi
    elif [ -f index.cgi ]; then
      sed -i "1s|^#!.*|#!${ruby}/bin/ruby|" index.cgi
      chmod +x index.cgi
    fi

    echo "Starting WEBrick on http://localhost:8080/index.cgi"

    exec ${ruby}/bin/ruby -rwebrick -e '
      s = WEBrick::HTTPServer.new(
        Port: 8080,
        DocumentRoot: Dir.pwd
      )

      s.mount "/index.cgi",
        WEBrick::HTTPServlet::CGIHandler,
        File.join(Dir.pwd, "index.cgi"),
        "${ruby}/bin/ruby"

      trap("INT"){ s.shutdown }
      s.start
    '
  '';
}
