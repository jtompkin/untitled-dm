{
  buildGoModule,
  lib,
}:
buildGoModule {
  pname = "untitled-dm";
  version = "1.0.0";
  src = lib.fileset.toSource {
    root = ../.;
    fileset = lib.fileset.difference ../. ../nix;
  };
  vendorHash = "sha256-qrX55UC7IMOZS8yDB+JIf5fAatfsRaMl38T1rDKHSAg=";
  meta = {
    description = "An untitled display manager";
    homepage = "https://github.com/jtompkin/untitled-dm";
    license = lib.licenses.mit;
    mainProgram = "untitled-dm";
  };
}
