{
  buildGoModule,
  lib,
}:
buildGoModule {
  pname = "untitled-dm";
  version = "0.0.1";
  src = lib.fileset.toSource {
    root = ../.;
    fileset = lib.fileset.difference ../. ../nix;
  };
  vendorHash = "sha256-uwBJAqN4sIepiiJf9lCDumLqfKJEowQO2tOiSWD3Fig=";
  meta = {
    description = "An untitled display manager";
    homepage = "https://github.com/jtompkin/untitled-dm";
    license = lib.licenses.mit;
    mainProgram = "untitled-dm";
  };
}
