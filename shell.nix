{ pkgs ? import <nixpkgs> {} }:

with pkgs;
let
  lib = import <nixpkgs/lib>;
  inherit (lib) optional optionals;
  go = pkgs.callPackage ./nix/go.nix { pkgs=pkgs; };

  basePackages = with pkgs; [
    # custom defined go version
    go
    # retrieve from nixpkgs
    postgresql_14
    nodejs-14_x
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

in pkgs.mkShell {
    name = "dev";

    buildInputs = requiredPackages;

    shellHook = ''
        # setup postgres
        export PGDATA="$PWD/.yatai-db"
        export PGHOST="$PWD"
        export SOCKET_DIRECTORIES="$PWD/.sockets"
        mkdir -p $PGDATA

        if [[ ! -d $PGDATA ]]; then
          initdb -D $PGDATA
          createuser postgres -h localhost
          createdb yatai
        fi

        if [[ ! $(grep listen_address $PGDATA/postgresql.conf) ]]; then
            cat >> "$PGDATA/postgresql.conf" <<-EOF
listen_addresses = 'localhost'
port = 5432
unix_socket_directories = '$PGHOST'
EOF
        fi

        pg_ctl -l $PGDATA/logfile start

        function end {
          echo "Shutting down the database..."
          pg_ctl stop
          echo "Removing directories..."
          rm -rf $SOCKET_DIRECTORIES
        }
        trap end EXIT

        # setup for dashboard
        alias scripts='jq ".scripts" dashboard/package.json'

        make fe-deps be-deps
    '';
}
