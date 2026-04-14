{
  pkgs ? import <nixpkgs> { },
}:

let
  ruby = pkgs.ruby.withPackages (ps: [
    ps.webrick
  ]);
in
pkgs.mkShell {
  packages = [ ruby ];

  shellHook = ''
    export PATH=${ruby}/bin:$PATH

    # shebang fix
    if [ -f index.rb ]; then
      sed -i "1s|^#!.*|#!${ruby}/bin/ruby|" index.rb
    fi

    # index.cgi 用意
    if [ -f index.rb ] && [ ! -f index.cgi ]; then
      cp index.rb index.cgi
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
