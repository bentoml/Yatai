with (import <nixpkgs> {});
let
  postgresql = pkgs.postgresql_14;
  nodejs = pkgs.nodejs_14;
  go = pkgs.callPackage ./scripts/nix/go.nix { };

  basePackages = [
    postgresql
    nodejs
    go
    pkgs.yarn
    pkgs.jq
    pkgs.make
  ];

  inputs = basePackages
    ++ lib.optional stdenv.isLinux inotify-tools
    ++ lib.optionals stdenv.isDarwin (with darwin.apple_sdk.frameworks; [
        CoreFoundation
        CoreServices
      ]);

in mkShell {
    name = "dev";

    buildInputs = [
        postgresql
    ];

    shellHook = ''
        # setup postgres
        export PGDATA="$PWD/.yatai-db"
        export PGHOST="$PWD"
        export SOCKET_DIRECTORIES="$PWD/.sockets"
        mkdir -p $PGDATA

        if [[ ! $(grep listen_address $PGDATA/postgresql.conf) ]]; then
            initdb -D $PGDATA
            createuser postgres -h localhost
            createdb yatai
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
    '';
}
