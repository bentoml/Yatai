{ pkgs ? import <nixpkgs> {} }:

with pkgs;
let
  lib = import <nixpkgs/lib>;
  inherit (lib) optional optionals;
  postgresql = pkgs.postgresql_14;
  go = pkgs.callPackage ./nix/go.nix { pkgs=pkgs; };

  basePackages = [
    go
    postgresql
    pkgs.jq
    pkgs.yarn
    pkgs.gnumake
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
        if [[ ! $(grep listen_address $PGDATA/postgresql.conf) ]]; then
            initdb -D $PGDATA
            cat >> "$PGDATA/postgresql.conf" <<-EOF
listen_addresses = 'localhost'
port = 5432
unix_socket_directories = '$PGHOST'
EOF
        fi

        pg_ctl -D $PGDATA -l $PGDATA/logfile start

        if [[ ! $(psql -l | grep yatai) ]]; then
            createdb yatai
            createuser -s postgres -h localhost
        fi

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
