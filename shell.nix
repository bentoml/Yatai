{ sources ? import nix/sources.nix
 , pkgs ? import sources.nixpkgs { overlays = [] ; config = {}; }
}:

with pkgs;
let
  lib = import <nixpkgs/lib>;
  inherit (lib) optional optionals;

  # custom defined packages in nixpkgs
  go = callPackage ./nix/go.nix { pkgs=pkgs; };
  nodejs = nodejs-14_x;
  postgresql = postgresql_14;

  # postgres definition
  pg_root = builtins.toString ./. + "/.yatai_db";
  pg_user = "postgres";
  pg_host = "localhost";
  pg_db = "yatai";

  # base requirements
  basePackages = with pkgs; [
    # custom defined go version
    go
    # TODO: lock version with niv
    postgresql

    # Without this, we see a whole bunch of warnings about LANG, LC_ALL and locales in general.
    # The solution is from: https://github.com/NixOS/nix/issues/318#issuecomment-52986702
    glibcLocales

    nodejs
    yarn
    jq
    gnumake
    git
    coreutils
  ];

  requiredPackages = basePackages
    ++ lib.optional stdenv.isLinux inotify-tools
    ++ lib.optionals stdenv.isDarwin (with darwin.apple_sdk.frameworks; [
        CoreFoundation
        CoreServices
      ]);

  env = buildEnv {
    name = "build-env";
    paths = requiredPackages;
  };

in
  stdenv.mkDerivation rec {
    name = "yatai-dev";

    phases = ["nobuild"];
    buildInputs = [env];

    shellHook = ''
        # "nix-shell --pure" resets LANG to POSIX, this breaks "make TAGS".
        export LANG="en_US.UTF-8"
        # /bin/ps cannot be found in the build environment.
        export PATH="/bin:/usr/bin:/usr/local/bin:/usr/sbin:$PATH"

        # setup for dashboard
        alias scripts='jq ".scripts" dashboard/package.json'

        make fe-deps be-deps

        export PGDATA=${pg_root}
        export SOCKET_DIRECTORIES="$PWD/sockets"
        mkdir $SOCKET_DIRECTORIES

        if [ ! -d "$PGDATA" ]; then
          initdb --auth=trust --auth-host=trust >/dev/null
          echo "unix_socket_directories = '$SOCKET_DIRECTORIES'" >> $PGDATA/postgresql.conf
          createuser ${pg_user} --createdb -h ${pg_host}
          createdb ${pg_db} -h ${pg_host} k-O ${pg_user}
        fi

        pg_ctl -l $PGDATA/logfile start

        function end {
          echo "Shutting down the database..."
          pg_ctl stop
          echo "Removing directories..."
          rm -rf $SOCKET_DIRECTORIES
        }
        trap end EXIT
    '';
    enableParallelBuilding = true;

    LOCALE_ARCHIVE = if stdenv.isLinux then "${glibcLocales}/lib/locale/locale-archive" else "";

    nobuild = ''
      echo Do not run this derivation with nix-build, it can only be used with nix-shell
    '';
}
