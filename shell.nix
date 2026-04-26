{
  pkgs ? import <nixpkgs> { },
}:

pkgs.mkShell {
  buildInputs = [
    pkgs.libeb
    pkgs.autoconf
    pkgs.automake
    pkgs.pkg-config
    pkgs.ffmpeg.dev
  ];

  shellHook = ''
    export CGO_CFLAGS="-I${pkgs.libeb}/include"
    export CGO_LDFLAGS="-L${pkgs.libeb}/lib"
  '';
}
