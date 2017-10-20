# skybin

This is a prototype of SkyBin's most basic functionality: storing files on
peers' machine.

It's pretty basic but includes what I believe could evolve into a working design
for the storage system.

It supports:

 - Creating a directory to store SkyBin metadata and peer content (a "repo").
 - Running a server to store peers' content.
 - Storing, listing, and retrieving files from the network.

To try it, first set up Go and the protoc compiler for Google protocol buffers.
This can be a little tricky the first time through, but google should give good
help.  Then do the following:

```
# Clone the repo into a skybin directory inside your go workspace.
cd $GOPATH/src && git clone https://github.com/AlexSteele/skybin.git

# Enter the repo directory.
cd skybin

# Grab dependencies.
go get

# Build the skybin binary.
make skybin

# See the available commands.
./skybin

# Create a repo. Use the -home option to set the directory (default=~/.skybin)
./skybin init -home repo1

# Tell skybin where the repo is and start a storage server in the background.
export SKYBIN_HOME=$PWD/repo1
./skybin server &

# Create a second repo.
./skybin init -home repo2
export SKYBIN_HOME=$PWD/repo2

# Tell the second repo about the storage server.
echo "[{'ID': 'provider1', 'Addr': '127.0.0.1:8002'}]" >> repo2/providers.json

# Store a file!
echo "Hello world" >> hello.txt
./skybin put hello.txt

# List stored files.
./skybin list

# Download hello.txt to stdout.
./skybin get hello.txt
```

