{ pkgs, stdenv, fetchurl }:

let
  toGoKernel = platform:
    if platform.isDarwin then "darwin"
    else platform.parsed.kernel.name;
    hashes = {
      # Use `print-hashes.sh ${version}` to generate the list below
      # https://raw.githubusercontent.com/NixOS/nixpkgs/master/pkgs/development/compilers/go/print-hashes.sh
      darwin-amd64 = "765c021e372a87ce0bc58d3670ab143008dae9305a79e9fa83440425529bb636";
      darwin-arm64 = "ffe45ef267271b9681ca96ca9b0eb9b8598dd82f7bb95b27af3eef2461dc3d2c";
      linux-386 = "982487a0264626950c635c5e185df68ecaadcca1361956207578d661a7b03bee";
      linux-amd64 = "550f9845451c0c94be679faf116291e7807a8d78b43149f9506c1b15eb89008c";
      linux-arm64 = "06f505c8d27203f78706ad04e47050b49092f1b06dc9ac4fbee4f0e4d015c8d4";
      linux-armv6l = "aa0d5516c8bd61654990916274d27491cfa229d322475502b247a8dc885adec5";
      linux-ppc64le = "b821ff58d088c61adc5d7376179a342f325d8715a06abdeb6974f6450663ee60";
      linux-s390x = "7d1727e08fef295f48aed2b8124a07e3752e77aea747fcc7aeb8892b8e2f2ad2";

    };

  toGoCPU = platform: {
    "i686" = "386";
    "x86_64" = "amd64";
    "aarch64" = "arm64";
    "armv6l" = "armv6l";
    "armv7l" = "armv6l";
    "powerpc64le" = "ppc64le";
  }.${platform.parsed.cpu.name} or (throw "Unsupported CPU ${platform.parsed.cpu.name}");

  version = "1.17.3";

  toGoPlatform = platform: "${toGoKernel platform}-${toGoCPU platform}";

  platform = toGoPlatform stdenv.hostPlatform;
in
stdenv.mkDerivation {
  pname = "go";
  version = version;
  src = fetchurl {
    url = "https://golang.org/dl/go${version}.${platform}.tar.gz";
    sha256 = hashes.${platform} or (throw "Missing Go bootstrap hash for platform ${platform}");
  };
  builder = ./go-install.sh;
  system = builtins.currentSystem;
}
