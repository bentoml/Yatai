# Create the standard environment.
source $stdenv/setup
# Extract the source code.
mkdir -p $out/opt
tar -C $out/opt -xzf $src
# Create place to store the binaries.
mkdir -p $out/bin
# Make symlinks to the binaries.
ln -s $out/opt/go/bin/go $out/bin/go
ln -s $out/opt/go/bin/gofmt $out/bin/gofmt
